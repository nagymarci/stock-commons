package reqid

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func verifyRequestIdExists(t *testing.T) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetRequestId(r) == "" {
			t.Fatalf("requestId not found in request context")
		}
	})
}

const TestRequestIdHeaderKey = "TEST-Request-Id"

func addRequestIdToTestHeader(t *testing.T) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(TestRequestIdHeaderKey, GetRequestId(r))
	})
}

func TestRequestIdMiddleware(t *testing.T) {
	t.Run("puts request id to http request context", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		mux := http.NewServeMux()
		mux.Handle("/test", ReqIdMiddleware(verifyRequestIdExists(t)))

		mux.ServeHTTP(rec, req)
	})

	t.Run("puts request id to http response header", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		mux := http.NewServeMux()
		mux.Handle("/test", ReqIdMiddleware(addRequestIdToTestHeader(t)))

		mux.ServeHTTP(rec, req)

		res := rec.Result()

		if res.Header.Values(RequestIdHttpHeaderKey)[0] != res.Header.Values(TestRequestIdHeaderKey)[0] {
			t.Fatalf("expected [%s], got [%s]", res.Header.Values(TestRequestIdHeaderKey)[0], res.Header.Values(RequestIdHttpHeaderKey)[0])
		}
	})
}
