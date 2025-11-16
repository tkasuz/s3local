package middlware

import (
	"context"
	"net/http"

	"github.com/tkasuz/s3local/internal/sqlc"
)

type ctxKey string

const sqlcKey ctxKey = "sqlc"

func WithSQLC(q *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), sqlcKey, q)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) *sqlc.Queries {
	q, _ := ctx.Value(sqlcKey).(*sqlc.Queries)
	return q
}
