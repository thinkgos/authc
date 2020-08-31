package authj

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/casbin/casbin/v2"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type ctxAuthKey struct{}

// NewAuthorizer returns the authorizer
// uses a Casbin enforcer and Subject function as input
func NewAuthorizer(e *casbin.Enforcer) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// checks the userName,path,method permission combination from the request.
			allowed, err := e.Enforce(subject(r), r.URL.Path, r.Method)
			if err != nil {
				renderJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"code":    http.StatusInternalServerError,
					"message": "Permission validation errors occur!",
				})
				return
			} else if !allowed {
				// the 403 Forbidden to the client
				renderJSON(w, http.StatusForbidden, map[string]interface{}{
					"code":    http.StatusForbidden,
					"message": "Permission denied!",
				})
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// subject returns the value associated with this context for subjectCtxKey,
func subject(r *http.Request) string {
	val, _ := r.Context().Value(ctxAuthKey{}).(string)
	return val
}

func renderJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	content, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(content)
	if err != nil {
		panic(err)
	}
}

// ContextWithSubject return a copy of parent in which the value associated with
// subjectCtxKey is subject.
func ContextWithSubject(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, ctxAuthKey{}, subject)
}
