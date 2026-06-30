package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	appconfig "app/internal/config"
	"app/pkg/apperror"
)

type S3Uploader struct {
	cfg      appconfig.UploadConfig
	uploader *manager.Uploader
}

func NewS3Uploader(cfg appconfig.UploadConfig) *S3Uploader {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.S3Endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           cfg.S3Endpoint,
				SigningRegion: cfg.S3Region,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.S3Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
	)
	
	if err != nil {
		// Log or handle error, but since we're in constructor we'll panic or let it fail on upload
		panic(fmt.Errorf("failed to load S3 config: %w", err))
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // often needed for MinIO or custom endpoints
	})

	uploader := manager.NewUploader(client)

	return &S3Uploader{
		cfg:      cfg,
		uploader: uploader,
	}
}

func (u *S3Uploader) Save(filename string, contentType string, size int64, data io.Reader) (*UploadResult, error) {
	if size <= 0 {
		return nil, apperror.Validation("file is required", map[string]string{
			"file": "is required",
		})
	}

	maxBytes := int64(u.cfg.MaxSizeMB) * 1024 * 1024
	if size > maxBytes {
		return nil, apperror.Validation("file too large", map[string]string{
			"file": fmt.Sprintf("must be at most %d MB", u.cfg.MaxSizeMB),
		})
	}

	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
	}
	if !isAllowedType(u.cfg, contentType) {
		return nil, apperror.Validation("unsupported file type", map[string]string{
			"file": "must be an allowed image type",
		})
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		exts, _ := mime.ExtensionsByType(contentType)
		if len(exts) > 0 {
			ext = exts[0]
		}
	}

	storedName := uuid.New().String() + ext

	result, err := u.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(u.cfg.S3Bucket),
		Key:         aws.String(storedName),
		Body:        data,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &UploadResult{
		URL:         result.Location,
		Filename:    storedName,
		Size:        size,
		ContentType: contentType,
	}, nil
}
