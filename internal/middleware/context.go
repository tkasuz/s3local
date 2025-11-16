package middleware

import (
	"context"
	"database/sql"
	"net/http"

	database "github.com/tkasuz/s3local/internal/sqlc"
)

type contextKey string

const queriesKey contextKey = "queries"

// QueriesMiddleware injects sqlc queries into request context
func QueriesMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queries := database.New(db)
			ctx := context.WithValue(r.Context(), queriesKey, queries)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetQueries retrieves queries from context
func GetQueries(ctx context.Context) *database.Queries {
	if queries, ok := ctx.Value(queriesKey).(*database.Queries); ok {
		return queries
	}
	return nil
}
