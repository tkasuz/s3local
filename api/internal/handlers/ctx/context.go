package ctx

import (
	"context"
	"net/http"

	"github.com/tkasuz/s3local/internal/config"
	"github.com/tkasuz/s3local/internal/db"
)

type ctxKey string

const (
	// StoreKey is exported for testing purposes
	StoreKey ctxKey = "store"
	cfgKey   ctxKey = "cfg"
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
