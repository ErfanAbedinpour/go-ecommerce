package storage

import (
	"io"
	"strings"

	"app/internal/config"
)

type UploadResult struct {
	URL         string
	Filename    string
	Size        int64
	ContentType string
}

type Uploader interface {
	Save(filename string, contentType string, size int64, data io.Reader) (*UploadResult, error)
}

func NewUploader(cfg config.UploadConfig) Uploader {
	if cfg.Provider == "s3" {
		return NewS3Uploader(cfg)
	}
	return NewLocalUploader(cfg)
}

func isAllowedType(cfg config.UploadConfig, contentType string) bool {
	if contentType == "" {
		return false
	}
	for _, allowed := range cfg.AllowedTypes {
		if strings.EqualFold(contentType, allowed) {
			return true
		}
	}
	return false
}
