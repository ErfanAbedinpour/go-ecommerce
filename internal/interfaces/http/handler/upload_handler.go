package handler

import (
	"log/slog"
	"net/http"

	"app/internal/infrastructure/storage"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
)

var _ = dtoresponse.UploadResponse{}

// UploadHandler handles file upload HTTP endpoints.
type UploadHandler struct {
	uploader *storage.Uploader
	log      *slog.Logger
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uploader *storage.Uploader, log *slog.Logger) *UploadHandler {
	return &UploadHandler{uploader: uploader, log: log}
}

// Upload godoc
// @Summary      Upload file
// @Description  Upload an image file (product image, logo, favicon). Returns a public URL.
// @Tags         uploads
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Image file (jpeg, png, webp, gif)"
// @Success      201   {object}  dtoresponse.UploadResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/uploads [post]
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Error(w, r, h.log, apperror.Validation("invalid multipart form", nil))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, r, h.log, apperror.Validation("file is required", map[string]string{
			"file": "is required",
		}))
		return
	}
	defer file.Close()

	result, err := h.uploader.Save(header.Filename, header.Header.Get("Content-Type"), header.Size, file)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.UploadResponse{
		URL:         result.URL,
		Filename:    result.Filename,
		Size:        result.Size,
		ContentType: result.ContentType,
	})
}
