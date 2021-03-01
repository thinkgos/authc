package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/thinkgos/http-middlewares/gzap"
	"github.com/thinkgos/http-middlewares/mids"
)

func main() {
	r := chi.NewMux()

	logger, _ := zap.NewProduction()

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	//   - Custom zap field
	r.Use(gzap.Logger(logger,
		gzap.WithTimeFormat(time.RFC3339),
		gzap.WithUTC(true),
		gzap.WithCustomFields(
			func(r *http.Request) zap.Field { return zap.String("custom field1", mids.ClientIP(r)) },
			func(r *http.Request) zap.Field { return zap.String("custom field2", mids.ClientIP(r)) },
		),
	))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	//   - Custom zap field
	r.Use(gzap.Recovery(logger, true,
		gzap.WithCustomFields(
			func(r *http.Request) zap.Field { return zap.String("custom field1", mids.ClientIP(r)) },
			func(r *http.Request) zap.Field { return zap.String("custom field2", mids.ClientIP(r)) },
		),
	))
	// Example ping request.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong " + fmt.Sprint(time.Now().Unix())))
	})

	// Example when panic happen.
	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("An unexpected error happen!")
	})

	// Listen and Server in 0.0.0.0:8080
	http.ListenAndServe(":8080", r)
}
