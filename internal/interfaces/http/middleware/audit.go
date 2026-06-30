package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	domain "app/internal/domain/audit"
	"app/internal/domain/user"
)

// AuditResponseWriter captures the status code and response body for audit logging.
type AuditResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *AuditResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *AuditResponseWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditLog middleware logs state-changing actions by admin users.
func AuditLog(repo domain.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only audit state-changing requests
			if r.Method == http.MethodGet || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Capture request body
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset body for next handlers
			}

			// Wrap response writer
			aw := &AuditResponseWriter{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
				Body:           &bytes.Buffer{},
			}

			next.ServeHTTP(aw, r)

			// Only log successful actions or specific validations? Log all for now if they are state-changing
			if aw.StatusCode >= 200 && aw.StatusCode < 400 {
				ctx := r.Context()
				
				// Get user from context
				userID, err := GetUserID(ctx)
				if err != nil {
					return
				}
				role, err := GetUserRole(ctx)
				if err != nil || role != user.RoleAdmin {
					return
				}

				// Basic resource extraction from URL path (e.g. /api/v1/admin/products/123 -> products, 123)
				parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/admin/"), "/")
				resourceType := "unknown"
				resourceID := ""
				
				if len(parts) > 0 && parts[0] != "" {
					resourceType = parts[0]
				}
				if len(parts) > 1 {
					resourceID = parts[1]
				}

				// Async save audit log
				go func() {
					log := &domain.AuditLog{
						AdminUserID:  userID.String(),
						Action:       r.Method,
						ResourceType: resourceType,
						ResourceID:   resourceID,
						NewValue:     reqBody, // Save request body as new value
						IPAddress:    r.RemoteAddr,
						UserAgent:    r.UserAgent(),
					}

					// Optional: Try to parse response for created ID if it was a POST
					if r.Method == http.MethodPost && resourceID == "" {
						var resData struct {
							ID string `json:"id"`
						}
						if err := json.Unmarshal(aw.Body.Bytes(), &resData); err == nil && resData.ID != "" {
							log.ResourceID = resData.ID
						}
					}

					_ = repo.Create(context.Background(), log)
				}()
			}
		})
	}
}
