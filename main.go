package main

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yuri-NagaSaki/ImageFlow/config"
	"github.com/Yuri-NagaSaki/ImageFlow/handlers"
	"github.com/Yuri-NagaSaki/ImageFlow/utils"
	"github.com/joho/godotenv"
)

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment variable or use a default value
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "*" // Default to allow all origins if not specified
		}

		// Handle multiple origins if specified
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins != "*" {
			// Check if the request origin is in the allowed origins list
			allowedList := strings.Split(allowedOrigins, ",")
			allowed := false
			for _, allowedOrigin := range allowedList {
				if strings.TrimSpace(allowedOrigin) == origin {
					allowed = true
					break
				}
			}

			if allowed {
				// Set the specific origin instead of the wildcard
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Allow credentials when specific origin is set
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				// If origin is not allowed, still set the wildcard for public resources
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			// Set wildcard for all origins
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set the config in the utils package
	utils.SetConfig(cfg)

	// Initialize S3 client only when using S3 storage
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "s3" {
		if err := utils.InitS3Client(); err != nil {
			log.Fatalf("Failed to initialize S3 client: %v", err)
		}
	}

	// Initialize storage provider
	if err := utils.InitStorage(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize metadata store
	if err := utils.InitMetadataStore(); err != nil {
		log.Fatalf("Failed to initialize metadata store: %v", err)
	}

	// Ensure image directories exist
	ensureDirectories(cfg)

	// Initialize and start image cleaner
	utils.InitCleaner()
	log.Printf("Image cleaner started")

	// Configure MIME types
	configureMIMETypes()

	// Create routes
	http.HandleFunc("/api/validate-api-key", handlers.ValidateAPIKey(cfg))
	http.HandleFunc("/api/upload", handlers.RequireAPIKey(cfg, handlers.UploadHandler(cfg)))
	http.HandleFunc("/api/images", handlers.RequireAPIKey(cfg, handlers.ListImagesHandler(cfg)))
	http.HandleFunc("/api/delete-image", handlers.RequireAPIKey(cfg, handlers.DeleteImageHandler(cfg)))
	http.HandleFunc("/api/config", handlers.RequireAPIKey(cfg, handlers.ConfigHandler(cfg)))
	http.HandleFunc("/api/tags", handlers.RequireAPIKey(cfg, handlers.TagsHandler(cfg)))
	http.HandleFunc("/api/debug/tags", handlers.RequireAPIKey(cfg, handlers.DebugTagsHandler(cfg)))

	// Add cleanup trigger endpoint
	http.HandleFunc("/api/trigger-cleanup", handlers.RequireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		utils.TriggerCleanup()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Cleanup process triggered",
		})
	}))

	// Use appropriate random image handler based on storage type
	if storageType == "s3" {
		http.HandleFunc("/api/random", handlers.RandomImageHandler(utils.S3Client))
		http.Handle("/auto/image/", http.StripPrefix("/auto/image/", http.HandlerFunc(handlers.S3AutoImageHandler(utils.S3Client))))
	} else {
		http.HandleFunc("/api/random", handlers.LocalRandomImageHandler())
		http.Handle("/auto/image/", http.StripPrefix("/auto/image/", http.HandlerFunc(handlers.LocalAutoImageHandler())))
		// Serve local images
		localPath := os.Getenv("LOCAL_STORAGE_PATH")
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(".", localPath)
		}
		http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(localPath))))
	}

	// Serve static files
	fs := http.FileServer(http.Dir("static"))

	// Next.js static assets
	http.Handle("/_next/", http.StripPrefix("/_next/", http.FileServer(http.Dir("static/_next"))))

	// Static assets
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Favicon files
	faviconServer := http.FileServer(http.Dir("favicon"))
	http.Handle("/favicon-16.png", faviconServer)
	http.Handle("/favicon-32.png", faviconServer)
	http.Handle("/favicon.ico", faviconServer)

	// Text files
	http.HandleFunc("/index.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.txt")
	})
	http.HandleFunc("/manage.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/manage.txt")
	})

	// Serve upload and management pages
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.ServeFile(w, r, "static/index.html")
		case "/manage":
			http.ServeFile(w, r, "static/manage.html")
		default:
			filePath := filepath.Join("static", r.URL.Path)
			if !filepath.IsAbs(filePath) {
				http.NotFound(w, r)
				return
			}
			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, filePath)
			} else {
				http.NotFound(w, r)
			}
		}
	})

	// Apply CORS middleware to all handlers
	handler := corsMiddleware(http.DefaultServeMux)

	// Start server
	log.Printf("Starting server on %s with %s storage (with CORS enabled)", cfg.ServerAddr, storageType)
	if err := http.ListenAndServe(cfg.ServerAddr, handler); err != nil {
		log.Fatal(err)
	}
}

// configureMIMETypes registers common MIME types
func configureMIMETypes() {
	// Register common MIME types
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".ico", "image/x-icon")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".jpeg", "image/jpeg")
	mime.AddExtensionType(".webp", "image/webp")
	mime.AddExtensionType(".avif", "image/avif")
}

// ensureDirectories creates necessary directory structure for images
func ensureDirectories(cfg *config.Config) {
	dirs := []string{
		filepath.Join(cfg.ImageBasePath, "original", "landscape"),
		filepath.Join(cfg.ImageBasePath, "original", "portrait"),
		filepath.Join(cfg.ImageBasePath, "landscape", "webp"),
		filepath.Join(cfg.ImageBasePath, "landscape", "avif"),
		filepath.Join(cfg.ImageBasePath, "portrait", "webp"),
		filepath.Join(cfg.ImageBasePath, "portrait", "avif"),
		filepath.Join(cfg.ImageBasePath, "gif"),      // Directory for GIF files
		filepath.Join(cfg.ImageBasePath, "metadata"), // Directory for image metadata
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: Failed to create directory %s: %v", dir, err)
		} else {
			log.Printf("Created directory: %s", dir)
		}
	}
}
