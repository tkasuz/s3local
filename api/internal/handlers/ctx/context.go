package ctx

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tkasuz/s3local/internal/config"
	"github.com/tkasuz/s3local/internal/db"
)

type ctxKey string

const (
	// StoreKey is exported for testing purposes
	StoreKey      ctxKey = "store"
	cfgKey        ctxKey = "cfg"
	bucketNameKey ctxKey = "bucketName"
	objectKeyKey  ctxKey = "objectKey"
)

// WithStore injects store into request context
func WithStore(store *db.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), StoreKey, store)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetStore retrieves store from context
func GetStore(ctx context.Context) *db.Store {
	s, _ := ctx.Value(StoreKey).(*db.Store)
	return s
}

// WithConfig injects config into request context
func WithConfig(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), cfgKey, cfg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetConfig retrieves config from context
func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(cfgKey).(*config.Config); ok {
		return cfg
	}
	return nil
}

func WithBucketName() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bucket := chi.URLParam(r, "bucket")
			ctx := context.WithValue(r.Context(), bucketNameKey, bucket)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetBucketName(ctx context.Context) string {
	if bucketName, ok := ctx.Value(bucketNameKey).(string); ok {
		return bucketName
	}
	return ""
}

func WithObjectKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the full key from the URL path
			// URL pattern is /{bucket}/{key...} where key can contain slashes
			bucket := chi.URLParam(r, "bucket")

			// Remove leading slash and bucket name to get the key
			// Example: "/mybucket/folder1/file.txt" -> "folder1/file.txt"
			path := r.URL.Path
			if len(path) > 0 && path[0] == '/' {
				path = path[1:] // Remove leading slash
			}

			// Remove bucket name and the following slash
			prefix := bucket + "/"
			key := ""
			if len(path) > len(prefix) && path[:len(prefix)] == prefix {
				key = path[len(prefix):]
			}

			ctx := context.WithValue(r.Context(), objectKeyKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetObjectKey(ctx context.Context) string {
	if bucketName, ok := ctx.Value(objectKeyKey).(string); ok {
		return bucketName
	}
	return ""
}
