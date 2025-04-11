package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yuri-NagaSaki/ImageFlow/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Image format constants
const (
	FormatAVIF     = "avif"
	FormatWebP     = "webp"
	FormatOriginal = "original"
)

// detectBestFormat determines optimal image format based on Accept headers
func detectBestFormat(r *http.Request) string {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "image/avif") {
		return FormatAVIF
	}
	if strings.Contains(accept, "image/webp") {
		return FormatWebP
	}
	return FormatOriginal
}

// determineOrientation selects orientation based on device type and request parameters
func determineOrientation(r *http.Request, deviceType string) string {
	orientation := r.URL.Query().Get("orientation")
	if orientation == "" {
		if deviceType == utils.DeviceMobile {
			return "portrait"
		} else {
			return "landscape"
		}
	}
	return orientation
}

// getContentType returns the appropriate Content-Type based on format and filename
func getContentType(format string, filename string) string {
	if format == FormatAVIF {
		return "image/avif"
	}
	if format == FormatWebP {
		return "image/webp"
	}

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

// setImageResponseHeaders sets standard HTTP headers for image responses
func setImageResponseHeaders(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Vary", "Accept, User-Agent")
}

// getFormattedImagePath constructs the path to an image with the given format
func getFormattedImagePath(format string, orientation string, filename string) string {
	switch format {
	case FormatAVIF:
		return fmt.Sprintf("%s/avif/%s.avif", orientation, filename)
	case FormatWebP:
		return fmt.Sprintf("%s/webp/%s.webp", orientation, filename)
	default:
		// Keep original format and path
		// Determine extension by checking if filename already has one
		extension := filepath.Ext(filename)
		if extension == "" {
			extension = ".jpg" // Default to jpg if no extension
		}
		return fmt.Sprintf("original/%s/%s%s", orientation, filename, extension)
	}
}

// RandomImageHandler serves random images from S3 storage
func RandomImageHandler(s3Client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := os.Getenv("S3_BUCKET")

		// Determine device type and orientation
		deviceType := utils.DetectDeviceType(r)
		orientation := determineOrientation(r, deviceType)

		// Get tag parameter for filtering
		tag := r.URL.Query().Get("tag")

		// Build the prefix for original images directory
		prefix := fmt.Sprintf("original/%s/", orientation)

		// List objects in the directory
		output, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket: &bucket,
			Prefix: aws.String(prefix),
		})

		if err != nil {
			log.Printf("Error listing objects: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Filter image files
		var imageObjects []string
		for _, obj := range output.Contents {
			if utils.IsImageFile(*obj.Key) {
				// If tag filtering is requested, check metadata
				if tag != "" {
					// Extract ID from key
					fileBaseName := filepath.Base(*obj.Key)
					id := strings.TrimSuffix(fileBaseName, filepath.Ext(fileBaseName))

					// Get metadata to check tags
					metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
					if err != nil || metadata == nil {
						// Skip if metadata not found or error
						continue
					}

					// Check if image has the requested tag
					hasTag := false
					for _, imgTag := range metadata.Tags {
						if imgTag == tag {
							hasTag = true
							break
						}
					}

					if !hasTag {
						// Skip if image doesn't have the tag
						continue
					}
				}

				imageObjects = append(imageObjects, *obj.Key)
			}
		}

		if len(imageObjects) == 0 {
			http.Error(w, "No images found", http.StatusNotFound)
			return
		}

		// Select a random image
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomIndex := rng.Intn(len(imageObjects))
		originalKey := imageObjects[randomIndex]
		log.Printf("Selected random image: %s", originalKey)

		// Extract base filename without extension
		fileBaseName := filepath.Base(originalKey)
		filename := strings.TrimSuffix(fileBaseName, filepath.Ext(fileBaseName))

		// Determine best format for client
		bestFormat := detectBestFormat(r)

		// Preserve PNG format for transparency if original is PNG
		isPNG := strings.HasSuffix(strings.ToLower(originalKey), ".png")
		if isPNG && bestFormat == FormatOriginal {
			// Use original PNG
			imageKey := originalKey
			log.Printf("Serving original PNG: %s", imageKey)

			// Get image from S3
			data, err := s3Client.GetObject(r.Context(), &s3.GetObjectInput{
				Bucket: &bucket,
				Key:    aws.String(imageKey),
			})

			if err != nil {
				log.Printf("Error getting image %s: %v", imageKey, err)
				http.Error(w, "Image not found", http.StatusNotFound)
				return
			}
			defer data.Body.Close()

			// Set response headers with PNG content type
			setImageResponseHeaders(w, "image/png")

			// Copy image data to response
			if _, err := io.Copy(w, data.Body); err != nil {
				log.Printf("Error sending image: %v", err)
				return
			}
			return
		}

		// Get image path
		imageKey := getFormattedImagePath(bestFormat, orientation, filename)
		log.Printf("Serving image format: %s, path: %s", bestFormat, imageKey)

		// Get image from S3
		data, err := s3Client.GetObject(r.Context(), &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    aws.String(imageKey),
		})

		if err != nil {
			log.Printf("Error getting image %s: %v", imageKey, err)
			// Fall back to original format if specific format not available
			if bestFormat != FormatOriginal {
				log.Printf("Falling back to original image format")
				data, err = s3Client.GetObject(r.Context(), &s3.GetObjectInput{
					Bucket: &bucket,
					Key:    aws.String(originalKey),
				})

				if err != nil {
					log.Printf("Error getting original image: %v", err)
					http.Error(w, "Image not found", http.StatusNotFound)
					return
				}

				// Use original format
				bestFormat = FormatOriginal
				imageKey = originalKey
			} else {
				http.Error(w, "Image not found", http.StatusNotFound)
				return
			}
		}
		defer data.Body.Close()

		// Set response headers
		contentType := getContentType(bestFormat, imageKey)
		setImageResponseHeaders(w, contentType)

		// Copy image data to response
		if _, err := io.Copy(w, data.Body); err != nil {
			log.Printf("Error sending image: %v", err)
			return
		}
	}
}

// LocalRandomImageHandler serves random images from local storage
func LocalRandomImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get local storage path
		localPath := os.Getenv("LOCAL_STORAGE_PATH")

		// Determine device type
		deviceType := utils.DetectDeviceType(r)

		// Get tag parameter for filtering
		tag := r.URL.Query().Get("tag")
		log.Printf("Random image request with tag: %s, device type: %s", tag, deviceType)

		// Determine orientation based on device type
		// This is the core design principle: desktop gets landscape, mobile gets portrait
		orientation := "landscape" // Default for desktop
		if deviceType == utils.DeviceMobile {
			orientation = "portrait" // Mobile gets portrait
		}

		log.Printf("Using orientation: %s based on device type: %s", orientation, deviceType)

		// Find all images with the specified tag
		var matchingImages []*utils.ImageMetadata
		var err error

		if tag != "" {
			// When filtering by tag, get all images with that tag
			// First get the image IDs with the tag
			imageIDs, err := findImagesWithTagDebug(tag, "local", localPath)
			if err != nil {
				log.Printf("Error finding images with tag %s: %v", tag, err)
				http.Error(w, "Error finding images", http.StatusInternalServerError)
				return
			}

			// Convert image IDs to metadata objects
			for _, id := range imageIDs {
				metadata, err := utils.MetadataManager.GetMetadata(context.Background(), id)
				if err == nil && metadata != nil {
					matchingImages = append(matchingImages, metadata)
				}
			}

			log.Printf("Found %d images with tag: %s", len(matchingImages), tag)

			if len(matchingImages) == 0 {
				log.Printf("No images found with tag: %s", tag)
				http.Error(w, "No images found", http.StatusNotFound)
				return
			}

			// Filter by the determined orientation (based on device type)
			var filteredImages []*utils.ImageMetadata
			for _, img := range matchingImages {
				if img.Orientation == orientation {
					filteredImages = append(filteredImages, img)
				}
			}

			// Only use filtered images if we found any
			if len(filteredImages) > 0 {
				matchingImages = filteredImages
				log.Printf("Filtered to %d images with orientation: %s", len(matchingImages), orientation)
			} else {
				log.Printf("No images found with tag %s and orientation %s", tag, orientation)
				http.Error(w, "No images found matching criteria", http.StatusNotFound)
				return
			}
		} else {
			// Without tag filtering, use the standard approach

			// Read files from the orientation directory
			originalDir := filepath.Join(localPath, "original", orientation)
			log.Printf("Looking for images in directory: %s", originalDir)

			files, err := os.ReadDir(originalDir)
			if err != nil {
				log.Printf("Error reading directory %s: %v", originalDir, err)
				http.Error(w, "No images found", http.StatusNotFound)
				return
			}

			// Convert files to metadata objects
			for _, file := range files {
				if !file.IsDir() && utils.IsImageFile(file.Name()) {
					id := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
					matchingImages = append(matchingImages, &utils.ImageMetadata{
						ID:          id,
						Orientation: orientation,
						Paths: struct {
							Original string `json:"original"`
							Auto	 string `json:"auto"`
							WebP     string `json:"webp"`
							AVIF     string `json:"avif"`
						}{
							Original: filepath.Join("original", orientation, file.Name()),
						},
					})
				}
			}

			if len(matchingImages) == 0 {
				log.Printf("No images found in directory: %s", originalDir)
				http.Error(w, "No images found", http.StatusNotFound)
				return
			}
		}

		// Select a random image from matching images
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomIndex := rng.Intn(len(matchingImages))
		selectedImage := matchingImages[randomIndex]
		log.Printf("Selected random image: %s, orientation: %s", selectedImage.ID, selectedImage.Orientation)

		// Determine best format for client
		bestFormat := detectBestFormat(r)
		log.Printf("Best format for client: %s", bestFormat)

		// Get the image path based on the format
		var imagePath string
		var contentType string

		// Check if the image is a PNG (for transparency)
		isPNG := false
		if selectedImage.Format == "png" {
			isPNG = true
		} else {
			// Try to detect from the original path
			isPNG = strings.HasSuffix(strings.ToLower(selectedImage.Paths.Original), ".png")
		}

		// For PNG images with transparency, use original format
		if isPNG && bestFormat == FormatOriginal {
			imagePath = filepath.Join(localPath, selectedImage.Paths.Original)
			contentType = "image/png"
			log.Printf("Using original PNG for transparency: %s", imagePath)
		} else {
			// Use the appropriate format based on browser support
			switch bestFormat {
			case FormatAVIF:
				imagePath = filepath.Join(localPath, selectedImage.Orientation, "avif", selectedImage.ID+".avif")
				contentType = "image/avif"
			case FormatWebP:
				imagePath = filepath.Join(localPath, selectedImage.Orientation, "webp", selectedImage.ID+".webp")
				contentType = "image/webp"
			default:
				// Use original format
				imagePath = filepath.Join(localPath, selectedImage.Paths.Original)
				contentType = getContentType(FormatOriginal, imagePath)
			}
			log.Printf("Using format %s, path: %s", bestFormat, imagePath)

			// Check if file exists, fall back to original if needed
			if _, err := os.Stat(imagePath); os.IsNotExist(err) && bestFormat != FormatOriginal {
				log.Printf("Format %s not available, falling back to original", bestFormat)
				imagePath = filepath.Join(localPath, selectedImage.Paths.Original)
				contentType = getContentType(FormatOriginal, imagePath)
			}
		}

		// Read and serve the image
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			log.Printf("Error reading image %s: %v", imagePath, err)
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}

		// Set response headers
		setImageResponseHeaders(w, contentType)

		// Send image data
		if _, err := w.Write(imageData); err != nil {
			log.Printf("Error sending image: %v", err)
			return
		}
	}
}
