// Package recording - Storage layer for recordings
package recording

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// StorageBackend defines the storage interface
type StorageBackend interface {
	Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, int64, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

// S3Storage implements StorageBackend using S3-compatible storage
type S3Storage struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	prefix     string
	endpoint   string
	region     string
}

// S3Config contains S3 storage configuration
type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	Prefix    string
	UseSSL    bool
}

// NewS3Storage creates a new S3 storage backend
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		prefix:    cfg.Prefix,
		endpoint:  cfg.Endpoint,
		region:    cfg.Region,
	}, nil
}

// Upload stores a recording file
func (s *S3Storage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	fullKey := filepath.Join(s.prefix, key)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(fullKey),
		Body:          data,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
		Metadata: map[string]string{
			"x-amz-meta-uploaded-at": time.Now().UTC().Format(time.RFC3339),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// Download retrieves a recording file
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	fullKey := filepath.Join(s.prefix, key)

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, *result.ContentLength, nil
}

// Delete removes a recording file
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	fullKey := filepath.Join(s.prefix, key)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := filepath.Join(s.prefix, key)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return true, nil
}

// GetURL generates a presigned URL for temporary access
func (s *S3Storage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	fullKey := filepath.Join(s.prefix, key)

	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	}, func(o *s3.PresignOptions) {
		o.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// LocalStorage implements StorageBackend using local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	return &LocalStorage{basePath: basePath}, nil
}

// Upload stores a recording file locally
func (s *LocalStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	path := filepath.Join(s.basePath, key)

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Download retrieves a recording file locally
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	path := filepath.Join(s.basePath, key)

	file, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, fmt.Errorf("failed to stat file: %w", err)
	}

	return file, info.Size(), nil
}

// Delete removes a recording file locally
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path := filepath.Join(s.basePath, key)
	return os.Remove(path)
}

// Exists checks if a file exists locally
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	path := filepath.Join(s.basePath, key)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// GetURL returns a local file path (not a URL for local storage)
func (s *LocalStorage) GetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return filepath.Join(s.basePath, key), nil
}

// StorageManager manages recording storage with multiple backends
type StorageManager struct {
	primary   StorageBackend
	backup    StorageBackend
	local     StorageBackend
	config    StorageManagerConfig
}

// StorageManagerConfig contains storage manager settings
type StorageManagerConfig struct {
	EnableBackup     bool
	EnableLocalCache bool
	MaxFileSize      int64
	CompressionLevel int
}

// NewStorageManager creates a new storage manager
func NewStorageManager(primary StorageBackend, backup StorageBackend, config StorageManagerConfig) *StorageManager {
	return &StorageManager{
		primary: primary,
		backup:  backup,
		config:  config,
	}
}

// SaveRecording saves a recording to storage
func (sm *StorageManager) SaveRecording(ctx context.Context, recordingID uuid.UUID, data []byte) error {
	key := sm.generateKey(recordingID)
	_ = calculateChecksum(data) // reserved for integrity verification

	// Upload to primary
	if err := sm.primary.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), "application/x-asciicast"); err != nil {
		return fmt.Errorf("failed to upload to primary: %w", err)
	}

	// Upload to backup if enabled
	if sm.config.EnableBackup && sm.backup != nil {
		if err := sm.backup.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), "application/x-asciicast"); err != nil {
			// Log but don't fail - primary succeeded
			fmt.Printf("Warning: backup upload failed: %v\n", err)
		}
	}

	return nil
}

// GetRecording retrieves a recording from storage
func (sm *StorageManager) GetRecording(ctx context.Context, recordingID uuid.UUID) (io.ReadCloser, int64, error) {
	key := sm.generateKey(recordingID)

	// Try primary first
	reader, size, err := sm.primary.Download(ctx, key)
	if err == nil {
		return reader, size, nil
	}

	// Fallback to backup
	if sm.backup != nil {
		return sm.backup.Download(ctx, key)
	}

	return nil, 0, err
}

// DeleteRecording removes a recording from all backends
func (sm *StorageManager) DeleteRecording(ctx context.Context, recordingID uuid.UUID) error {
	key := sm.generateKey(recordingID)

	var errs []error

	if err := sm.primary.Delete(ctx, key); err != nil {
		errs = append(errs, fmt.Errorf("primary: %w", err))
	}

	if sm.backup != nil {
		if err := sm.backup.Delete(ctx, key); err != nil {
			errs = append(errs, fmt.Errorf("backup: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to delete from some backends: %v", errs)
	}

	return nil
}

// GetPresignedURL generates a temporary access URL
func (sm *StorageManager) GetPresignedURL(ctx context.Context, recordingID uuid.UUID, expires time.Duration) (string, error) {
	key := sm.generateKey(recordingID)
	return sm.primary.GetURL(ctx, key, expires)
}

// generateKey generates the storage key for a recording
func (sm *StorageManager) generateKey(recordingID uuid.UUID) string {
	now := time.Now()
	return filepath.Join(
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%s.cast", recordingID.String()),
	)
}

// calculateChecksum calculates SHA256 checksum
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
