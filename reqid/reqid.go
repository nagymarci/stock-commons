package reqid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIdKey           ContextKey = "requestId"
	RequestIdHttpHeaderKey            = "X-Request-Id"
)

func GetRequestId(r *http.Request) string {
	reqID := r.Context().Value(RequestIdKey)

	if ret, ok := reqID.(string); ok {
		return ret
	}

	return ""
}

func ReqIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New()

		ctx := context.WithValue(r.Context(), RequestIdKey, reqID.String())

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

		w.Header().Add(RequestIdHttpHeaderKey, reqID.String())
	})
}
