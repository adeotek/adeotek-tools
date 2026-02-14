package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adeotek/adeotek-tools/sql-toolbox/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Uploader defines the interface for uploading files to S3
type S3Uploader interface {
	Upload(localPath, key string) error
}

// S3UploadService implements S3Uploader using AWS SDK v2
type S3UploadService struct {
	client *s3.Client
	bucket string
	prefix string
}

// NewS3UploadService creates a new S3 upload service from config
func NewS3UploadService(s3Config models.S3Config) (*S3UploadService, error) {
	ctx := context.Background()

	var opts []func(*config.LoadOptions) error

	if s3Config.Region != "" {
		opts = append(opts, config.WithRegion(s3Config.Region))
	}

	if s3Config.AccessKeyID != "" && s3Config.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				s3Config.AccessKeyID,
				s3Config.SecretAccessKey,
				"",
			),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	if s3Config.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Config.Endpoint)
			o.UsePathStyle = true // Required for MinIO/S3-compatible
		})
	}

	client := s3.NewFromConfig(cfg, s3Opts...)

	return &S3UploadService{
		client: client,
		bucket: s3Config.Bucket,
		prefix: s3Config.Prefix,
	}, nil
}

// Upload uploads a local file to S3 with the given key
// Note: The key should include any desired prefix (use BuildS3Key to construct it)
func (svc *S3UploadService) Upload(localPath, key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file for upload: %w", err)
	}
	defer file.Close()

	// Key is expected to already include the prefix
	fullKey := key

	contentType := "application/sql"
	if strings.HasSuffix(localPath, ".gz") {
		contentType = "application/gzip"
	}

	_, err = svc.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(svc.bucket),
		Key:         aws.String(fullKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// BuildS3Key constructs the S3 object key from a local file path, database name, and optional prefix
func BuildS3Key(prefix, dbName, filePath string) string {
	key := dbName + "/" + filepath.Base(filePath)
	if prefix != "" {
		return prefix + key
	}
	return key
}
