package handler

import (
	"net/http"

	"app/internal/interfaces/http/response"
)

// AdminIndex godoc
// @Summary      Admin API root
// @Description  Returns a welcome message. Requires admin role.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.MessageResponse
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Router       /api/v1/admin/ [get]
func AdminIndex(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{"message": "ecommerce admin API v1"})
}
