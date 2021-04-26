package requestid

// Ported from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

// Ported from chi's middleware, source:
// https://github.com/go-chi/chi/v5/blob/master/middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

// Key to use when setting the request ID.
type ctxRequestIDKey struct{}

// Config defines the config for RequestID middleware
type Config struct {
	requestIDHeader string
	nextRequestID   func() string
}

// Option RequestID option
type Option func(*Config)

// WithRequestIDHeader optional request id header (default "X-Request-Id")
func WithRequestIDHeader(s string) Option {
	return func(c *Config) {
		c.requestIDHeader = s
	}
}

// WithNextRequestID optional next request id function (default NextRequestID function)
func WithNextRequestID(nextRequestID func() string) Option {
	return func(c *Config) {
		c.nextRequestID = nextRequestID
	}
}

// RequestID is a middleware that injects a request ID into the context of each
// request. if it is empty, set the write head
// - requestIDHeader is the name of the HTTP Header which contains the request id.
// Exported so that it can be changed by developers. (default "X-Request-Id")
// - nextRequestID generates the next request ID.(default NextRequestID)
func RequestID(opts ...Option) func(next http.Handler) http.Handler {
	c := &Config{
		requestIDHeader: "X-Request-ID",
		nextRequestID:   NextRequestID,
	}
	for _, opt := range opts {
		opt(c)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			requestID := r.Header.Get(c.requestIDHeader)
			if requestID == "" {
				requestID = c.nextRequestID()
				w.Header().Add(c.requestIDHeader, requestID)
			}
			ctx = context.WithValue(ctx, ctxRequestIDKey{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromRequestID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func FromRequestID(ctx context.Context) string {
	reqID, ok := ctx.Value(ctxRequestIDKey{}).(string)
	if !ok {
		return ""
	}
	return reqID
}

var prefix string
var sequenceID uint64

// see chi middleware request_id
// A quick note on the statistics here: we're trying to calculate the chance that
// two randomly generated base62 prefixes will collide. We use the formula from
// http://en.wikipedia.org/wiki/Birthday_problem
//
// P[m, n] \approx 1 - e^{-m^2/2n}
//
// We ballpark an upper bound for $m$ by imagining (for whatever reason) a server
// that restarts every second over 10 years, for $m = 86400 * 365 * 10 = 315360000$
//
// For a $k$ character base-62 identifier, we have $n(k) = 62^k$
//
// Plugging this in, we find $P[m, n(10)] \approx 5.75%$, which is good enough for
// our purposes, and is surely more than anyone would ever need in practice -- a
// process that is rebooted a handful of times a day for a hundred years has less
// than a millionth of a percent chance of generating two colliding IDs.

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [20]byte
	var b64 string
	for len(b64) < 16 {
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s-%d-%s-", hostname, os.Getpid(), b64[:10])
}

// NextRequestID generates the next request ID.
// A request ID is a string of the form like {hostname}-{pid}-{init-rand-value}-{sequence},
// where "random" is a base62 random string that uniquely identifies this go
// process, and where the last number is an atomically incremented request
// counter.
func NextRequestID() string {
	return fmt.Sprintf("%s%012d", prefix, atomic.AddUint64(&sequenceID, 1))
}
