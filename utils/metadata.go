package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ImageMetadata stores metadata information for images
type ImageMetadata struct {
	ID           string    `json:"id"`           // Image ID (without extension)
	OriginalName string    `json:"originalName"` // Original filename
	UploadTime   time.Time `json:"uploadTime"`   // Upload timestamp
	ExpiryTime   time.Time `json:"expiryTime"`   // Expiry timestamp (if set)
	Format       string    `json:"format"`       // Original format
	Orientation  string    `json:"orientation"`  // Image orientation
	Tags         []string  `json:"tags"`         // Image tags for categorization
	Paths        struct {
		Original string `json:"original"` // Path to original image
		Auto     string `json:"auto"`     // Path to original image
		WebP     string `json:"webp"`     // Path to WebP format
		AVIF     string `json:"avif"`     // Path to AVIF format
	} `json:"paths"`
}

// MetadataStore defines the interface for metadata storage operations
type MetadataStore interface {
	SaveMetadata(ctx context.Context, metadata *ImageMetadata) error
	GetMetadata(ctx context.Context, id string) (*ImageMetadata, error)
	ListExpiredImages(ctx context.Context) ([]*ImageMetadata, error)
	DeleteMetadata(ctx context.Context, id string) error
}

// LocalMetadataStore implements metadata storage for local filesystem
type LocalMetadataStore struct {
	BasePath string
}

// NewLocalMetadataStore creates a new local metadata store
func NewLocalMetadataStore(basePath string) (*LocalMetadataStore, error) {
	// Create metadata directory
	metadataDir := filepath.Join(basePath, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metadata directory: %v", err)
	}
	return &LocalMetadataStore{BasePath: basePath}, nil
}

// SaveMetadata saves image metadata to a local file
func (lms *LocalMetadataStore) SaveMetadata(ctx context.Context, metadata *ImageMetadata) error {
	metadataDir := filepath.Join(lms.BasePath, "metadata")
	metadataPath := filepath.Join(metadataDir, metadata.ID+".json")

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	log.Printf("Metadata saved for image %s", metadata.ID)
	return nil
}

// GetMetadata retrieves image metadata from a local file
func (lms *LocalMetadataStore) GetMetadata(ctx context.Context, id string) (*ImageMetadata, error) {
	metadataPath := filepath.Join(lms.BasePath, "metadata", id+".json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var metadata ImageMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &metadata, nil
}

// ListExpiredImages lists all expired images
func (lms *LocalMetadataStore) ListExpiredImages(ctx context.Context) ([]*ImageMetadata, error) {
	metadataDir := filepath.Join(lms.BasePath, "metadata")
	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata directory: %v", err)
	}

	var expiredImages []*ImageMetadata
	now := time.Now()

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		metadataPath := filepath.Join(metadataDir, file.Name())
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			log.Printf("Error reading metadata file %s: %v", metadataPath, err)
			continue
		}

		var metadata ImageMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			log.Printf("Error unmarshaling metadata from %s: %v", metadataPath, err)
			continue
		}

		// Check if the image has expired
		if !metadata.ExpiryTime.IsZero() && metadata.ExpiryTime.Before(now) {
			expiredImages = append(expiredImages, &metadata)
		}
	}

	return expiredImages, nil
}

// DeleteMetadata deletes image metadata
func (lms *LocalMetadataStore) DeleteMetadata(ctx context.Context, id string) error {
	metadataPath := filepath.Join(lms.BasePath, "metadata", id+".json")
	return os.Remove(metadataPath)
}

// S3MetadataStore implements metadata storage for S3
type S3MetadataStore struct {
	client *S3Storage
	prefix string
}

// NewS3MetadataStore creates a new S3 metadata store
func NewS3MetadataStore(s3Storage *S3Storage) *S3MetadataStore {
	return &S3MetadataStore{
		client: s3Storage,
		prefix: "metadata/",
	}
}

// SaveMetadata saves image metadata to S3
func (sms *S3MetadataStore) SaveMetadata(ctx context.Context, metadata *ImageMetadata) error {
	key := sms.prefix + metadata.ID + ".json"

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := sms.client.Store(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store metadata in S3: %v", err)
	}

	log.Printf("Metadata saved to S3 for image %s", metadata.ID)
	return nil
}

// GetMetadata retrieves image metadata from S3
func (sms *S3MetadataStore) GetMetadata(ctx context.Context, id string) (*ImageMetadata, error) {
	key := sms.prefix + id + ".json"

	data, err := sms.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata from S3: %v", err)
	}

	var metadata ImageMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &metadata, nil
}

// ListExpiredImages lists all expired images in S3
func (sms *S3MetadataStore) ListExpiredImages(ctx context.Context) ([]*ImageMetadata, error) {
	// Get the S3 bucket name
	bucket := os.Getenv("S3_BUCKET")

	// List all metadata objects with the metadata/ prefix
	paginator := s3.NewListObjectsV2Paginator(S3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(sms.prefix),
	})

	var expiredImages []*ImageMetadata
	now := time.Now()

	// Iterate through all pages of results
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("Error listing metadata objects from S3: %v", err)
			return nil, fmt.Errorf("failed to list metadata objects from S3: %v", err)
		}

		// Process each object in the current page
		for _, obj := range page.Contents {
			// Skip if not a JSON file
			if !strings.HasSuffix(*obj.Key, ".json") {
				continue
			}

			// Extract the image ID from the key
			id := strings.TrimSuffix(strings.TrimPrefix(*obj.Key, sms.prefix), ".json")

			// Get the metadata for this image
			metadata, err := sms.GetMetadata(ctx, id)
			if err != nil {
				log.Printf("Error getting metadata for image %s: %v", id, err)
				continue
			}

			// Check if the image has expired
			if !metadata.ExpiryTime.IsZero() && metadata.ExpiryTime.Before(now) {
				// Only log individual expired images when we actually find them
				expiredImages = append(expiredImages, metadata)
			}
		}
	}

	// Only log the count if we found expired images
	if len(expiredImages) > 0 {
		log.Printf("Found %d expired images in S3", len(expiredImages))
	}
	return expiredImages, nil
}

// DeleteMetadata deletes image metadata from S3
func (sms *S3MetadataStore) DeleteMetadata(ctx context.Context, id string) error {
	key := sms.prefix + id + ".json"
	return sms.client.Delete(ctx, key)
}

// Global metadata storage instance
var MetadataManager MetadataStore

// InitMetadataStore initializes the metadata storage
func InitMetadataStore() error {
	storageType := os.Getenv("STORAGE_TYPE")

	if storageType == "s3" {
		// Ensure S3 client is initialized
		if S3Client == nil {
			return fmt.Errorf("S3 client not initialized")
		}

		// Get S3 storage instance
		s3Storage, ok := Storage.(*S3Storage)
		if !ok {
			return fmt.Errorf("failed to get S3 storage instance")
		}

		MetadataManager = NewS3MetadataStore(s3Storage)
		log.Printf("S3 metadata store initialized")
	} else {
		// Local storage
		localPath := os.Getenv("LOCAL_STORAGE_PATH")
		if localPath == "" {
			localPath = "static/images"
		}

		// Ensure path is absolute
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(".", localPath)
		}

		localStore, err := NewLocalMetadataStore(localPath)
		if err != nil {
			return fmt.Errorf("failed to create local metadata store: %v", err)
		}

		MetadataManager = localStore
		log.Printf("Local metadata store initialized at %s", localPath)
	}

	return nil
}
