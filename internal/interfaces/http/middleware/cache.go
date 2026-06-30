package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"app/internal/infrastructure/cache"
)

type CacheResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *CacheResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CacheResponseWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

// CacheMiddleware caches the response of GET requests.
func Cache(redisCache *cache.RedisCache, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || redisCache == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key based on URL path and query
			hash := sha256.Sum256([]byte(r.URL.String()))
			cacheKey := "cache:" + hex.EncodeToString(hash[:])

			ctx := r.Context()
			cachedData, err := redisCache.Get(ctx, cacheKey)
			if err == nil && cachedData != nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				w.Write(cachedData)
				return
			}

			cw := &CacheResponseWriter{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
				Body:           &bytes.Buffer{},
			}

			next.ServeHTTP(cw, r)

			if cw.StatusCode >= 200 && cw.StatusCode < 300 {
				_ = redisCache.Set(ctx, cacheKey, cw.Body.Bytes(), ttl)
			}
		})
	}
}
