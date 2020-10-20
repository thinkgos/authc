package traceid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Key to use when setting the trace ID.
type ctxTraceIDKey struct{}

// TraceIDHeader is the name of the HTTP Header which contains the trace id.
// Exported so that it can be changed by developers
var TraceIDHeader = "X-Trace-Id"

// TraceID is a middleware that injects a trace ID into the context of each
// request. A trace ID is a string of uuid.
func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(TraceIDHeader)
		if requestID == "" {
			requestID = NextTraceID()
		}
		ctx = context.WithValue(ctx, ctxTraceIDKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromTraceID returns a trace ID from the given context if one is present.
// Returns the empty string if a trace ID cannot be found.
func FromTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(ctxTraceIDKey{}).(string)
	if !ok {
		return ""
	}
	return traceID
}

// NextTraceID generates the next request ID.
func NextTraceID() string {
	return uuid.New().String()
}
