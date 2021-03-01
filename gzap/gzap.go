package gzap

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/thinkgos/http-middlewares/mids"
)

// Option logger/recover option
type Option func(c *Config)

// WithTimeFormat optional a time package format string (e.g. time.RFC3339).
func WithTimeFormat(layout string) Option {
	return func(c *Config) {
		c.timeFormat = layout
	}
}

// WithUTC a boolean stating whether to use UTC time zone or local.(default local).
func WithUTC(b bool) Option {
	return func(c *Config) {
		c.utc = b
	}
}

// WithCustomFields optional custom field
func WithCustomFields(fields ...func(r *http.Request) zap.Field) Option {
	return func(c *Config) {
		c.customFields = fields
	}
}

// WithDisable optional disable this feature.
func WithDisable(b *atomic.Bool) Option {
	return func(c *Config) {
		c.disable = b
	}
}

// Config logger/recover config
type Config struct {
	timeFormat   string
	utc          bool
	disable      *atomic.Bool
	customFields []func(r *http.Request) zap.Field
}

// Logger returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
//
// Requests with errors are logged using zap.Error().
// Requests without errors are logged using zap.Info().
func Logger(logger *zap.Logger, opts ...Option) func(next http.Handler) http.Handler {
	cfg := Config{
		time.RFC3339Nano,
		false,
		atomic.NewBool(false),
		nil,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return func(next http.Handler) http.Handler {
		if cfg.disable.Load() {
			return next
		}

		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// some evil middlewares modify this values
			path := r.URL.Path
			query := r.URL.RawQuery
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			end := time.Now()
			latency := end.Sub(start)
			if cfg.utc {
				end = end.UTC()
			}

			fields := []zap.Field{
				zap.Int("status", ww.Status()),
				zap.String("method", r.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", mids.ClientIP(r)),
				zap.String("user-agent", r.UserAgent()),
				zap.String("time", end.Format(cfg.timeFormat)),
				zap.Duration("latency", latency),
			}
			for _, field := range cfg.customFields {
				fields = append(fields, field(r))
			}
			logger.Info(path, fields...)
		}
		return http.HandlerFunc(fn)
	}
}

// Recovery returns a gin.HandlerFunc (middleware)
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func Recovery(logger *zap.Logger, stack bool, opts ...Option) func(next http.Handler) http.Handler {
	cfg := Config{
		time.RFC3339Nano,
		false,
		atomic.NewBool(false),
		nil,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return func(next http.Handler) http.Handler {
		if cfg.disable.Load() {
			return next
		}
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Check for a broken connection, as it is not really a
					// condition that warrants a panic stack trace.
					var brokenPipe bool
					if ne, ok := err.(*net.OpError); ok {
						if se, ok := ne.Err.(*os.SyscallError); ok {
							if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
								strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
								brokenPipe = true
							}
						}
					}

					httpRequest, _ := httputil.DumpRequest(r, false)
					if brokenPipe {
						logger.Error(r.URL.Path,
							zap.Any("error", err),
							zap.ByteString("request", httpRequest),
						)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					now := time.Now()
					if cfg.utc {
						now = now.UTC()
					}
					fields := []zap.Field{
						zap.String("time", now.Format(cfg.timeFormat)),
						zap.Any("error", err),
						zap.ByteString("request", httpRequest),
					}
					for _, field := range cfg.customFields {
						fields = append(fields, field(r))
					}
					if stack {
						fields = append(fields, zap.ByteString("stack", debug.Stack()))
					}
					logger.Error("[Recovery from panic]", fields...)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
