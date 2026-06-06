package handler

import (
	"net/http"

	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
)

var _ = dtoresponse.MessageResponse{}

// AdminIndex godoc
// @Summary      Admin API root
// @Description  Returns a welcome message. Requires admin role.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.MessageResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/ [get]
func AdminIndex(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{"message": "ecommerce admin API v1"})
}
