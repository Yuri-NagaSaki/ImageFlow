package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Yuri-NagaSaki/ImageFlow/config"
	"github.com/Yuri-NagaSaki/ImageFlow/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ImageInfo represents information about an image
type ImageInfo struct {
	ID          string            `json:"id"`          // Filename without extension
	FileName    string            `json:"filename"`    // Full filename with extension
	URL         string            `json:"url"`         // URL to access the image
	URLs        map[string]string `json:"urls"`        // URLs for all available formats
	Orientation string            `json:"orientation"` // landscape or portrait
	Format      string            `json:"format"`      // original, webp, avif
	Size        int64             `json:"size"`        // File size in bytes
	Path        string            `json:"path"`        // Path relative to storage root
	StorageType string            `json:"storageType"` // "local" or "s3"
	Tags        []string          `json:"tags"`        // Image tags for categorization
}

// PaginatedResponse represents a paginated response with images
type PaginatedResponse struct {
	Success    bool        `json:"success"`    // Whether the request was successful
	Images     []ImageInfo `json:"images"`     // Images for current page
	Page       int         `json:"page"`       // Current page number
	Limit      int         `json:"limit"`      // Number of items per page
	TotalPages int         `json:"totalPages"` // Total number of pages
	Total      int         `json:"total"`      // Total number of images
}

// ListImagesHandler returns a handler for listing images
func ListImagesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get storage type from environment
		storageType := os.Getenv("STORAGE_TYPE")

		// Validate API key
		if !validateAPIKey(w, r, cfg.APIKey) {
			return
		}

		// Parse query parameters
		params := parseQueryParams(r)

		var allImages []ImageInfo
		var err error

		// Handle image listing based on format and storage type
		if params.format == "gif" {
			// For GIF files, orientation doesn't matter
			if storageType == "s3" {
				allImages, err = listS3GIFImages()
			} else {
				allImages, err = listGIFImages(cfg.ImageBasePath)
			}
		} else {
			// Handle regular image formats
			if storageType == "s3" {
				allImages, err = listS3Images(params.orientation, params.format)
			} else {
				allImages, err = listLocalImages(cfg.ImageBasePath, params.orientation, params.format)
			}
		}

		if err != nil {
			log.Printf("Error listing images: %v", err)
			http.Error(w, "获取图片列表失败", http.StatusInternalServerError)
			return
		}

		// Sort images by filename in descending order (newest first)
		sort.Slice(allImages, func(i, j int) bool {
			return allImages[i].FileName > allImages[j].FileName
		})

		// Filter by tag if specified
		if params.tag != "" {
			filteredImages := []ImageInfo{}
			for _, img := range allImages {
				// Check if the image has the requested tag
				for _, tag := range img.Tags {
					if tag == params.tag {
						filteredImages = append(filteredImages, img)
						break
					}
				}
			}
			allImages = filteredImages
		}

		// Generate paginated response
		sendPaginatedResponse(w, allImages, params.page, params.limit)
	}
}

// Query parameters structure
type queryParams struct {
	orientation string
	format      string
	tag         string // Tag to filter by
	page        int
	limit       int
}

// validateAPIKey checks if the provided API key is valid
func validateAPIKey(w http.ResponseWriter, r *http.Request, configAPIKey string) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "未提供授权信息", http.StatusUnauthorized)
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "授权格式无效", http.StatusUnauthorized)
		return false
	}

	apiKey := parts[1]
	if apiKey != configAPIKey {
		http.Error(w, "API密钥无效", http.StatusUnauthorized)
		return false
	}

	return true
}

// parseQueryParams extracts and validates query parameters
func parseQueryParams(r *http.Request) queryParams {
	orientation := r.URL.Query().Get("orientation")
	format := r.URL.Query().Get("format")
	tag := r.URL.Query().Get("tag")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Default values
	if orientation == "" {
		orientation = "all" // all, landscape, portrait
	}
	if format == "" {
		format = "original" // original, webp, avif
	}
	// Tag can be empty, which means no tag filtering

	// Set default pagination values
	page := 1
	limit := 12 // Default items per page

	// Parse page number
	if pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	// Parse limit
	if limitStr != "" {
		limitVal, err := strconv.Atoi(limitStr)
		if err == nil && limitVal > 0 && limitVal <= 50 { // Cap at 50 items per page
			limit = limitVal
		}
	}

	return queryParams{
		orientation: orientation,
		format:      format,
		tag:         tag,
		page:        page,
		limit:       limit,
	}
}

// sendPaginatedResponse creates and sends a paginated JSON response
func sendPaginatedResponse(w http.ResponseWriter, allImages []ImageInfo, page, limit int) {
	// Calculate pagination values
	total := len(allImages)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Ensure page is within valid range
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	// Calculate start and end indices for the current page
	start := (page - 1) * limit
	end := start + limit

	// Ensure end index is within bounds
	if end > total {
		end = total
	}

	// Extract the subset of images for the current page
	var pagedImages []ImageInfo
	if start < total {
		pagedImages = allImages[start:end]
	} else {
		pagedImages = []ImageInfo{} // Empty slice for out of range pages
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send the paginated response
	response := PaginatedResponse{
		Success:    true,
		Images:     pagedImages,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		Total:      total,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// getOrientationAndFormatSlices returns slices of orientations and formats based on input parameters
func getOrientationAndFormatSlices(orientation, format string) ([]string, []string) {
	var orientations []string
	var formats []string

	// Determine which orientations to list
	if orientation == "all" {
		orientations = []string{"landscape", "portrait"}
	} else {
		orientations = []string{orientation}
	}

	// Determine which formats to list
	if format == "all" {
		formats = []string{"original", "webp", "avif"}
	} else {
		formats = []string{format}
	}

	return orientations, formats
}

// getDirPathAndExtensions returns directory path and valid extensions based on format
func getDirPathAndExtensions(basePath, orientVal, formatVal string) (string, []string) {
	var dirPath string
	var extensions []string

	// Construct the path based on format
	if formatVal == "original" {
		dirPath = filepath.Join(basePath, "original", orientVal)
		extensions = []string{".jpg", ".jpeg", ".png", ".webp", ".avif"}
	} else {
		dirPath = filepath.Join(basePath, orientVal, formatVal)
		extension := "." + formatVal
		extensions = []string{extension}
	}

	return dirPath, extensions
}

// constructImageURL creates the appropriate URL for image access in list.go context
func constructImageURL(relPath string) string {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "local" {
		return fmt.Sprintf("/images/%s", strings.ReplaceAll(relPath, "\\", "/"))
	}

	// For S3 storage
	customDomain := os.Getenv("CUSTOM_DOMAIN")
	if customDomain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(customDomain, "/"), strings.ReplaceAll(relPath, "\\", "/"))
	}

	// Fallback to S3 endpoint with bucket name
	endpoint := strings.TrimSuffix(os.Getenv("S3_ENDPOINT"), "/")
	bucket := os.Getenv("S3_BUCKET")
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, strings.ReplaceAll(relPath, "\\", "/"))
}

// hasValidExtension checks if a filename has one of the valid extensions
func hasValidExtension(fileName string, extensions []string) bool {
	lowerFileName := strings.ToLower(fileName)
	for _, ext := range extensions {
		if strings.HasSuffix(lowerFileName, ext) {
			return true
		}
	}
	return false
}

// getImageURLs returns URLs for all available formats of an image
func getImageURLs(id string, orientation string, isGIF bool) map[string]string {
	urls := make(map[string]string)

	// For GIF files, only original format is available
	if isGIF {
		originalKey := filepath.Join("gif", id+".gif")
		urls["original"] = constructImageURL(originalKey)
		urls["auto"] = urls["original"] // Fallback to original
		urls["webp"] = urls["original"] // Fallback to original
		urls["avif"] = urls["original"] // Fallback to original
		return urls
	}

	// For regular images, determine paths for all formats
	// Get the extension of the original file from metadata if available
	originalExt := ".jpg" // Default extension
	metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
	if err == nil && metadata != nil && metadata.Paths.Original != "" {
		originalExt = filepath.Ext(metadata.Paths.Original)
	}

	// Construct original URL
	originalKey := filepath.Join("original", orientation, id+originalExt)
	urls["original"] = constructImageURL(originalKey)
	urls["auto"] = fmt.Sprintf("/auto/image/%s", strings.ReplaceAll(originalKey, "\\", "/"))

	// Construct URLs for converted formats
	webpKey := filepath.Join(orientation, "webp", id+".webp")
	avifKey := filepath.Join(orientation, "avif", id+".avif")

	urls["webp"] = constructImageURL(webpKey)
	urls["avif"] = constructImageURL(avifKey)

	return urls
}

// listLocalImages returns a list of images from the local filesystem
func listLocalImages(basePath, orientation, format string) ([]ImageInfo, error) {
	var images []ImageInfo
	orientations, formats := getOrientationAndFormatSlices(orientation, format)

	// First, collect all original format images
	var originalImages = make(map[string]ImageInfo)
	originalOrientations := []string{"landscape", "portrait"}

	// Get all original format images first
	for _, orientVal := range originalOrientations {
		dirPath := filepath.Join(basePath, "original", orientVal)
		extensions := []string{".jpg", ".jpeg", ".png", ".webp", ".avif"}

		// List files in the directory
		files, err := os.ReadDir(dirPath)
		if err != nil {
			// Skip directory if it doesn't exist
			log.Printf("Warning: Could not read directory %s: %v", dirPath, err)
			continue
		}

		// Process each file in original format
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fileName := file.Name()
			if !hasValidExtension(fileName, extensions) {
				continue
			}

			// Get file info for size
			fileInfo, err := file.Info()
			if err != nil {
				log.Printf("Warning: Could not get file info for %s: %v", fileName, err)
				continue
			}

			// Construct relative path
			relPath := filepath.Join("original", orientVal, fileName)

			// Extract ID (filename without extension)
			id := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			// Get tags from metadata if available
			var tags []string
			metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
			if err == nil && metadata != nil {
				tags = metadata.Tags
			}

			// Get URLs for all formats
			urls := getImageURLs(id, orientVal, false)

			// Store image info in map
			originalImages[id] = ImageInfo{
				ID:          id,
				FileName:    fileName,
				URL:         constructImageURL(relPath),
				URLs:        urls,
				Orientation: orientVal,
				Format:      "original",
				Size:        fileInfo.Size(),
				Path:        relPath,
				StorageType: "local",
				Tags:        tags,
			}
		}
	}

	// Now filter based on user's requested orientation and format
	for _, orientVal := range orientations {
		if format == "original" || formats[0] == "original" {
			// For original format, use the stored original images
			for _, img := range originalImages {
				if orientVal == "all" || img.Orientation == orientVal {
					images = append(images, img)
				}
			}
		} else {
			// For webp/avif formats, use the original images but update format and URL
			for _, img := range originalImages {
				if orientVal == "all" || img.Orientation == orientVal {
					// Create a new image info with the requested format
					newImg := img
					newImg.Format = format
					newImg.URL = img.URLs[format]
					// Update filename to match the requested format
					baseName := strings.TrimSuffix(img.FileName, filepath.Ext(img.FileName))
					newImg.FileName = baseName + "." + format
					// Update path to match the requested format
					if format == "webp" {
						newImg.Path = filepath.Join(orientVal, "webp", baseName+".webp")
					} else if format == "avif" {
						newImg.Path = filepath.Join(orientVal, "avif", baseName+".avif")
					}
					images = append(images, newImg)
				}
			}
		}
	}

	return images, nil
}

// listGIFImages returns a list of GIF images from the special gif directory
func listGIFImages(basePath string) ([]ImageInfo, error) {
	var images []ImageInfo
	dirPath := filepath.Join(basePath, "gif")

	// List files in the directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		// Return empty list if directory doesn't exist
		log.Printf("Warning: Could not read GIF directory %s: %v", dirPath, err)
		return images, nil
	}

	// Filter and collect GIF files
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if !strings.HasSuffix(strings.ToLower(fileName), ".gif") {
			continue
		}

		// Get file info for size
		fileInfo, err := file.Info()
		if err != nil {
			log.Printf("Warning: Could not get file info for %s: %v", fileName, err)
			continue
		}

		// Construct image URL and path
		relPath := filepath.Join("gif", fileName)

		// Extract ID (filename without extension)
		id := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// Get tags from metadata if available
		var tags []string
		metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
		if err == nil && metadata != nil {
			tags = metadata.Tags
		}

		// Get URLs for all formats (for GIF there's only original)
		urls := getImageURLs(id, "", true)

		// Add image info to the result
		images = append(images, ImageInfo{
			ID:          id,
			FileName:    fileName,
			URL:         constructImageURL(relPath),
			URLs:        urls,
			Orientation: "gif", // Use "gif" as orientation to mark it as special
			Format:      "gif",
			Size:        fileInfo.Size(),
			Path:        relPath,
			StorageType: "local",
			Tags:        tags,
		})
	}

	return images, nil
}

// listS3Images returns a list of images from S3 storage
func listS3Images(orientation, format string) ([]ImageInfo, error) {
	// For S3, we don't use basePath, but we need to handle GIF format separately
	if format == "gif" {
		// Fetch GIF images from S3 with gif/ prefix
		return listS3GIFImages()
	}

	var images []ImageInfo
	bucket := os.Getenv("S3_BUCKET")

	// First, collect all original format images
	var originalImages = make(map[string]ImageInfo)
	originalOrientations := []string{"landscape", "portrait"}

	// Get all original format images first
	for _, orientVal := range originalOrientations {
		prefix := "original/" + orientVal + "/"
		extensions := []string{".jpg", ".jpeg", ".png", ".webp", ".avif"}

		// List objects with the prefix
		paginator := s3.NewListObjectsV2Paginator(utils.S3Client, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(prefix),
		})

		// Process each page of results
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(context.Background())
			if err != nil {
				return nil, err
			}

			// Process each object
			for _, obj := range output.Contents {
				key := *obj.Key
				fileName := filepath.Base(key)

				if !hasValidExtension(fileName, extensions) {
					continue
				}

				// Extract ID (filename without extension)
				id := strings.TrimSuffix(fileName, filepath.Ext(fileName))

				// Get tags from metadata if available
				var tags []string
				metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
				if err == nil && metadata != nil {
					tags = metadata.Tags
				}

				// Get URLs for all formats
				urls := getImageURLs(id, orientVal, false)

				// Store image info in map
				originalImages[id] = ImageInfo{
					ID:          id,
					FileName:    fileName,
					URL:         constructImageURL(key),
					URLs:        urls,
					Orientation: orientVal,
					Format:      "original",
					Size:        *obj.Size,
					Path:        key,
					StorageType: "s3",
					Tags:        tags,
				}
			}
		}
	}

	// Now filter based on user's requested orientation and format
	orientations, formats := getOrientationAndFormatSlices(orientation, format)

	for _, orientVal := range orientations {
		if format == "original" || formats[0] == "original" {
			// For original format, use the stored original images
			for _, img := range originalImages {
				if orientVal == "all" || img.Orientation == orientVal {
					images = append(images, img)
				}
			}
		} else {
			// For webp/avif formats, use the original images but update format and URL
			for _, img := range originalImages {
				if orientVal == "all" || img.Orientation == orientVal {
					// Create a new image info with the requested format
					newImg := img
					newImg.Format = format
					newImg.URL = img.URLs[format]
					// Update filename to match the requested format
					baseName := strings.TrimSuffix(img.FileName, filepath.Ext(img.FileName))
					newImg.FileName = baseName + "." + format
					// Update path to match the requested format
					if format == "webp" {
						newImg.Path = filepath.Join(orientVal, "webp", baseName+".webp")
					} else if format == "avif" {
						newImg.Path = filepath.Join(orientVal, "avif", baseName+".avif")
					}
					images = append(images, newImg)
				}
			}
		}
	}

	return images, nil
}

// listS3GIFImages returns a list of GIF images from the special gif directory in S3
func listS3GIFImages() ([]ImageInfo, error) {
	var images []ImageInfo
	bucket := os.Getenv("S3_BUCKET")
	prefix := "gif/"

	// List objects with the gif/ prefix
	paginator := s3.NewListObjectsV2Paginator(utils.S3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	// Process each page of results
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("error listing S3 objects: %v", err)
		}

		for _, obj := range page.Contents {
			key := *obj.Key
			fileName := filepath.Base(key)

			// Skip objects that are not .gif files
			if !strings.HasSuffix(strings.ToLower(fileName), ".gif") {
				continue
			}

			// Extract ID (filename without extension)
			id := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			// Get tags from metadata if available
			var tags []string
			metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
			if err == nil && metadata != nil {
				tags = metadata.Tags
			}

			// Get URLs for all formats (for GIF there's only original)
			urls := getImageURLs(id, "", true)

			// Add to results
			images = append(images, ImageInfo{
				ID:          id,
				FileName:    fileName,
				URL:         constructImageURL(key),
				URLs:        urls,
				Orientation: "gif",
				Format:      "gif",
				Size:        *obj.Size,
				Path:        key,
				StorageType: "s3",
				Tags:        tags,
			})
		}
	}

	return images, nil
}
