package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// auto format images from S3 storage
func S3AutoImageHandler(s3Client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := os.Getenv("S3_BUCKET")

		// Get path from URL with orientation and filename
		// This assumes the URL path is in the format /auto/image/original/portrait/20250410_100015_9747.png}
		imgPath := strings.TrimPrefix(r.URL.Path, "/auto/image/")

		// Get file name and extension
		fileBaseName := filepath.Base(imgPath)
		// Extract file extension
		fileExt := filepath.Ext(fileBaseName)
		// Extract file name without extension
		filename := strings.TrimSuffix(fileBaseName, fileExt)
		// Extract orientation from path
		orientation := filepath.Base(filepath.Dir(imgPath))
		log.Printf("Extracted orientation: %s filename: %s ext: %s from Path: %s", orientation, filename, fileExt, imgPath)

		// Determine best format for client
		bestFormat := detectBestFormat(r)

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
					Key:    aws.String(imageKey),
				})

				if err != nil {
					log.Printf("Error getting original image: %v", err)
					http.Error(w, "Image not found", http.StatusNotFound)
					return
				}

				// Use original format
				bestFormat = FormatOriginal
				imageKey = imgPath
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

// auto format images from local storage
func LocalAutoImageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get local storage path
		localPath := os.Getenv("LOCAL_STORAGE_PATH")

		// Get path from URL with orientation and filename
		// This assumes the URL path is in the format /auto/image/original/portrait/20250410_100015_9747.png}
		imgPath := strings.TrimPrefix(r.URL.Path, "/auto/image/")

		// Get file name and extension
		fileBaseName := filepath.Base(imgPath)
		// Extract file extension
		fileExt := filepath.Ext(fileBaseName)
		// Extract file name without extension
		filename := strings.TrimSuffix(fileBaseName, fileExt)
		// Extract orientation from path
		orientation := filepath.Base(filepath.Dir(imgPath))
		log.Printf("Extracted orientation: %s filename: %s ext: %s from Path: %s", orientation, filename, fileExt, imgPath)

		// Determine best format for client
		bestFormat := detectBestFormat(r)
		log.Printf("Best format for client: %s", bestFormat)

		// Get the image path based on the format
		var imagePath string

		// Get image path
		imageKey := getFormattedImagePath(bestFormat, orientation, filename)
		// Set response headers
		contentType := getContentType(bestFormat, imageKey)
		imagePath = filepath.Join(localPath, imageKey)

		log.Printf("Using format %s, path: %s", bestFormat, imagePath)

		// Check if file exists, fall back to original if needed
		if _, err := os.Stat(imagePath); os.IsNotExist(err) && bestFormat != FormatOriginal {
			log.Printf("Format %s not available, falling back to original", bestFormat)
			imagePath = filepath.Join(localPath, imgPath)
			contentType = getContentType(FormatOriginal, imagePath)
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
