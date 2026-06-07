package storage

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"app/internal/config"
	"app/pkg/apperror"
)

// Uploader stores uploaded files on the local filesystem.
type Uploader struct {
	cfg config.UploadConfig
}

// NewUploader creates a new Uploader.
func NewUploader(cfg config.UploadConfig) *Uploader {
	return &Uploader{cfg: cfg}
}

// UploadResult holds metadata for a stored file.
type UploadResult struct {
	URL         string
	Filename    string
	Size        int64
	ContentType string
}

// Save validates and stores a file, returning its public URL.
func (u *Uploader) Save(filename string, contentType string, size int64, data io.Reader) (*UploadResult, error) {
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
	if !u.isAllowedType(contentType) {
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
	if err := os.MkdirAll(u.cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	destPath := filepath.Join(u.cfg.Dir, storedName)
	out, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("create upload file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, data)
	if err != nil {
		_ = os.Remove(destPath)
		return nil, fmt.Errorf("write upload file: %w", err)
	}

	baseURL := strings.TrimRight(u.cfg.BaseURL, "/")
	return &UploadResult{
		URL:         baseURL + "/" + storedName,
		Filename:    storedName,
		Size:        written,
		ContentType: contentType,
	}, nil
}

func (u *Uploader) isAllowedType(contentType string) bool {
	if contentType == "" {
		return false
	}
	for _, allowed := range u.cfg.AllowedTypes {
		if strings.EqualFold(contentType, allowed) {
			return true
		}
	}
	return false
}
