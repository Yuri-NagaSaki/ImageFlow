package handlers

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Yuri-NagaSaki/ImageFlow/config"
	"github.com/Yuri-NagaSaki/ImageFlow/utils"
)

// UploadResult represents the result of an image upload
type UploadResult struct {
	Filename    string            `json:"filename"`
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Orientation string            `json:"orientation,omitempty"`
	Format      string            `json:"format,omitempty"`
	URLs        map[string]string `json:"urls,omitempty"`
	ExpiryTime  string            `json:"expiryTime,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// getPublicURL constructs a public-facing URL for accessing an image
func getPublicURL(key string) string {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "local" {
		return fmt.Sprintf("/images/%s", key)
	}
	// For S3 storage
	customDomain := os.Getenv("CUSTOM_DOMAIN")
	if customDomain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(customDomain, "/"), key)
	}
	// Fallback to S3 endpoint with bucket name
	endpoint := strings.TrimSuffix(os.Getenv("S3_ENDPOINT"), "/")
	bucket := os.Getenv("S3_BUCKET")
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, key)
}

// determineImageOrientation classifies an image as landscape or portrait
// Square images and portrait images are classified as portrait
func determineImageOrientation(img image.Config) string {
	if img.Width > img.Height {
		return "landscape"
	}
	return "portrait"
}

// UploadHandler handles image uploads, converting them to multiple formats
func UploadHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workerSemaphore := make(chan struct{}, cfg.WorkerThreads)
		log.Printf("Using %d parallel worker threads for image processing", cfg.WorkerThreads)

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form with default max upload size (32MB)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Get uploaded files
		files := r.MultipartForm.File["images[]"]
		if len(files) == 0 {
			http.Error(w, "No files uploaded", http.StatusBadRequest)
			return
		}

		// Check if number of uploaded files exceeds the maximum limit
		if len(files) > cfg.MaxUploadCount {
			http.Error(w, fmt.Sprintf("Too many files uploaded. Maximum allowed is %d files", cfg.MaxUploadCount), http.StatusBadRequest)
			return
		}

		// Get expiry time parameter (in minutes)
		expiryMinutes := 0 // Default: never expire
		if expiryParam := r.FormValue("expiryMinutes"); expiryParam != "" {
			if minutes, err := strconv.Atoi(expiryParam); err == nil && minutes >= 0 {
				expiryMinutes = minutes
			} else {
				log.Printf("Invalid expiryMinutes parameter: %s, using default: %d", expiryParam, expiryMinutes)
			}
		}

		// Get tags parameter
		var tags []string
		if tagsParam := r.FormValue("tags"); tagsParam != "" {
			// Split by comma and trim spaces
			for _, tag := range strings.Split(tagsParam, ",") {
				trimmedTag := strings.TrimSpace(tag)
				if trimmedTag != "" {
					tags = append(tags, trimmedTag)
				}
			}
			log.Printf("Image tags: %v", tags)
		}

		// Create result array and wait group
		results := make([]UploadResult, 0, len(files))
		resultsMutex := sync.Mutex{}
		var wgFiles sync.WaitGroup

		for _, fileHeader := range files {
			// Add each file to wait group
			wgFiles.Add(1)

			// Start goroutine to process each file
			go func(fileHeader *multipart.FileHeader) {
				defer wgFiles.Done()

				// Acquire worker slot
				workerSemaphore <- struct{}{}
				defer func() { <-workerSemaphore }()

				// Calculate expiry time
				var expiryTime time.Time
				if expiryMinutes > 0 {
					expiryTime = time.Now().Add(time.Duration(expiryMinutes) * time.Minute)
					log.Printf("Image will expire at: %v", expiryTime)
				}

				file, err := fileHeader.Open()
				if err != nil {
					log.Printf("Error opening file: %v", err)
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename: fileHeader.Filename,
						Status:   "error",
						Message:  fmt.Sprintf("Error opening file: %v", err),
					})
					resultsMutex.Unlock()
					return
				}
				defer file.Close()

				// Read image configuration to determine orientation
				img, _, err := image.DecodeConfig(file)
				if err != nil {
					log.Printf("Error reading image configuration: %v", err)
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename: fileHeader.Filename,
						Status:   "error",
						Message:  fmt.Sprintf("Error reading image configuration: %v", err),
					})
					resultsMutex.Unlock()
					return
				}
				orientation := determineImageOrientation(img)

				// Reset file pointer
				if _, err := file.Seek(0, 0); err != nil {
					log.Printf("Error resetting file pointer: %v", err)
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename: fileHeader.Filename,
						Status:   "error",
						Message:  fmt.Sprintf("Error resetting file pointer: %v", err),
					})
					resultsMutex.Unlock()
					return
				}

				// Read file content
				data := make([]byte, fileHeader.Size)
				if _, err := file.Read(data); err != nil {
					log.Printf("Error reading file: %v", err)
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename: fileHeader.Filename,
						Status:   "error",
						Message:  fmt.Sprintf("Error reading file: %v", err),
					})
					resultsMutex.Unlock()
					return
				}

				// Generate unique filename
				timestamp := time.Now().Format("20060102_150405")
				filename := fmt.Sprintf("%s_%d", timestamp, time.Now().UnixNano()%10000)
				imageID := filename

				// Detect image format
				imgFormat, err := utils.DetectImageFormat(data)
				if err != nil {
					log.Printf("Error detecting image format: %v", err)
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename: fileHeader.Filename,
						Status:   "error",
						Message:  fmt.Sprintf("Error detecting image format: %v", err),
					})
					resultsMutex.Unlock()
					return
				}

				// Check if it's a GIF file
				var originalKey string
				if imgFormat.Format == "gif" {
					// Store GIF in special directory
					originalKey = filepath.Join("gif", filename+imgFormat.Extension)
				} else {
					// Store other formats in original directory
					originalKey = filepath.Join("original", orientation, filename+imgFormat.Extension)
				}

				if err := utils.Storage.Store(r.Context(), originalKey, data); err != nil {
					log.Printf("Error storing original image: %v", err)
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename: fileHeader.Filename,
						Status:   "error",
						Message:  fmt.Sprintf("Error storing original file: %v", err),
					})
					resultsMutex.Unlock()
					return
				}
				log.Printf("Original image stored: %s", originalKey)

				// Check if it's a GIF file
				var wg sync.WaitGroup
				var webpURL, avifURL string

				if imgFormat.Format == "gif" {
					// GIF files are not converted, only store original format
					log.Printf("GIF file stored in special directory: %s", originalKey)

					var expiryTimeStr string
					if !expiryTime.IsZero() {
						expiryTimeStr = expiryTime.Format(time.RFC3339)
					}

					metadata := &utils.ImageMetadata{
						ID:           imageID,
						OriginalName: fileHeader.Filename,
						UploadTime:   time.Now(),
						Format:       imgFormat.Format,
						Tags:         tags,
					}

					if !expiryTime.IsZero() {
						metadata.ExpiryTime = expiryTime
					}

					metadata.Paths.Original = originalKey

					if err := utils.MetadataManager.SaveMetadata(r.Context(), metadata); err != nil {
						log.Printf("Warning: Failed to save metadata for GIF image %s: %v", imageID, err)
					}

					// Add success result to results array
					resultsMutex.Lock()
					results = append(results, UploadResult{
						Filename:   fileHeader.Filename,
						Status:     "success",
						Message:    "GIF file uploaded successfully",
						Format:     imgFormat.Format,
						ExpiryTime: expiryTimeStr,
						Tags:       tags,
						URLs: map[string]string{
							"original": getPublicURL(originalKey),
							"webp":     "",
							"avif":     "",
						},
					})
					resultsMutex.Unlock()
					return
				}

				// For non-GIF files, proceed with normal conversion
				// Create channels to store conversion results
				webpResultCh := make(chan []byte, 1)
				avifResultCh := make(chan []byte, 1)

				// WebP conversion
				wg.Add(1)
				go func() {
					defer wg.Done()
					webpData, err := utils.ConvertToWebP(data)
					if err != nil {
						log.Printf("WebP conversion error: %v", err)
						webpResultCh <- nil
						return
					}
					webpResultCh <- webpData
				}()

				// AVIF conversion
				wg.Add(1)
				go func() {
					defer wg.Done()
					avifData, err := utils.ConvertToAVIF(data)
					if err != nil {
						log.Printf("AVIF conversion error: %v", err)
						avifResultCh <- nil
						return
					}
					avifResultCh <- avifData
				}()

				// Wait for all conversions to complete
				wg.Wait()

				// Process WebP result
				webpData := <-webpResultCh
				if webpData != nil {
					webpKey := filepath.Join(orientation, "webp", filename+".webp")
					if err := utils.Storage.Store(r.Context(), webpKey, webpData); err != nil {
						log.Printf("Error storing WebP image: %v", err)
					} else {
						log.Printf("WebP image stored: %s", webpKey)
						webpURL = getPublicURL(webpKey)
					}
				}

				// Process AVIF result
				avifData := <-avifResultCh
				if avifData != nil {
					avifKey := filepath.Join(orientation, "avif", filename+".avif")
					if err := utils.Storage.Store(r.Context(), avifKey, avifData); err != nil {
						log.Printf("Error storing AVIF image: %v", err)
					} else {
						log.Printf("AVIF image stored: %s", avifKey)
						avifURL = getPublicURL(avifKey)
					}
				}

				// Get URL for original image
				originalURL := getPublicURL(originalKey)

				// Set WebP and AVIF URLs with defaults if conversion failed
				if webpURL == "" {
					webpURL = originalURL // Fallback to original when conversion fails
				}
				if avifURL == "" {
					avifURL = originalURL // Fallback to original when conversion fails
				}

				var expiryTimeStr string
				if !expiryTime.IsZero() {
					expiryTimeStr = expiryTime.Format(time.RFC3339)
				}

				metadata := &utils.ImageMetadata{
					ID:           imageID,
					OriginalName: fileHeader.Filename,
					UploadTime:   time.Now(),
					Format:       imgFormat.Format,
					Orientation:  orientation,
					Tags:         tags,
				}

				if !expiryTime.IsZero() {
					metadata.ExpiryTime = expiryTime
				}

				metadata.Paths.Original = originalKey
				metadata.Paths.Auto = filepath.Join("auto/image/", originalKey)
				if webpURL != originalURL {
					metadata.Paths.WebP = filepath.Join(orientation, "webp", imageID+".webp")
				}
				if avifURL != originalURL {
					metadata.Paths.AVIF = filepath.Join(orientation, "avif", imageID+".avif")
				}

				// Save Metadata
				if err := utils.MetadataManager.SaveMetadata(r.Context(), metadata); err != nil {
					log.Printf("Warning: Failed to save metadata for image %s: %v", imageID, err)
				}

				// Add success result to results array
				resultsMutex.Lock()
				results = append(results, UploadResult{
					Filename:    fileHeader.Filename,
					Status:      "success",
					Message:     "File uploaded and converted successfully",
					Orientation: orientation,
					Format:      imgFormat.Format,
					ExpiryTime:  expiryTimeStr,
					Tags:        tags,
					URLs: map[string]string{
						"original": originalURL,
						"webp":     webpURL,
						"avif":     avifURL,
					},
				})
				resultsMutex.Unlock()
			}(fileHeader) // Pass fileHeader to goroutine
		}

		// Wait for all files to be processed
		wgFiles.Wait()

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": results,
		})
	}
}
