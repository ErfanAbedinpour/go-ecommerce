package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"app/pkg/apperror"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// OK writes a 200 JSON response.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 JSON response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes a structured error response.
func Error(w http.ResponseWriter, log *slog.Logger, err error) {
	appErr := apperror.AsAppError(err)

	if appErr.Status >= http.StatusInternalServerError {
		log.Error("internal error",
			slog.String("code", appErr.Code),
			slog.String("error", appErr.Error()),
		)
		appErr = apperror.Internal("an unexpected error occurred")
	}

	JSON(w, appErr.Status, apperror.ErrorResponse{Error: *appErr})
}
