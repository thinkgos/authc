package ratelimiter

import (
	"net/http"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
)

// RateLimit rate limit
func RateLimit(lmt *limiter.Limiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpError := tollbooth.LimitByRequest(lmt, w, r)
			if httpError != nil {
				w.Header().Add("", lmt.GetMessageContentType())
				w.WriteHeader(httpError.StatusCode)
				_, err := w.Write([]byte(httpError.Message))
				if err != nil {
					panic(err)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
