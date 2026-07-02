package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"app/pkg/apperror"
	"app/pkg/i18n"
)

var translator = i18n.NewTranslator(i18n.LocaleEN)

// Init configures response localization from application config.
func Init(locale string) {
	translator = i18n.NewTranslator(locale)
}

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

// Error writes a structured error response with statusCode and request path.
func Error(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	appErr := apperror.AsAppError(err)

	if appErr.Status >= http.StatusInternalServerError && log != nil {
		log.Error("internal error",
			slog.String("code", appErr.Code),
			slog.String("path", requestPath(r)),
			slog.String("error", appErr.Error()),
		)
		appErr = apperror.Internal("an unexpected error occurred")
	}

	appErr = translator.Translate(appErr)
	JSON(w, appErr.Status, apperror.NewErrorResponse(appErr, requestPath(r)))
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Path
}
