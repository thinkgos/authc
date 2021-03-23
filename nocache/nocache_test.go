package nocache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_NoCacheHeaders(t *testing.T) {
	responseHeaders := map[string]string{
		"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
		"Pragma":        "no-cache",
		"Expires":       epoch,
	}
	recorder := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)

	m := http.NewServeMux()
	NoCache(m).ServeHTTP(recorder, r)
	for key, value := range responseHeaders {
		if recorder.Header()[key][0] != value {
			t.Errorf("Missing header: %s", key)
		}
	}
}
