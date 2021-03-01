package requestid

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRequestID(t *testing.T) {
	tests := map[string]struct {
		requestIDHeader  string
		request          func() *http.Request
		expectedResponse string
	}{
		"Retrieves Request Id from default header": {
			"X-Request-Id",
			func() *http.Request {
				req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/", nil)
				req.Header.Add("X-Request-Id", "req-123456")

				return req
			},
			"RequestID: req-123456",
		},
		"Retrieves Request Id from custom header": {
			"X-Trace-Id",
			func() *http.Request {
				req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/", nil)
				req.Header.Add("X-Trace-Id", "trace:abc123")

				return req
			},
			"RequestID: trace:abc123",
		},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()

		r := chi.NewRouter()

		r.Use(RequestID(WithRequestIDHeader(test.requestIDHeader), WithNextRequestID(NextRequestID)))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			requestID := FromRequestID(r.Context())
			response := fmt.Sprintf("RequestID: %s", requestID)

			w.Write([]byte(response)) // nolint: errcheck
		})
		r.ServeHTTP(w, test.request())

		if w.Body.String() != test.expectedResponse {
			t.Fatalf("RequestID was not the expected value")
		}
	}
}

func BenchmarkNextRequestID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NextRequestID()
	}
}
