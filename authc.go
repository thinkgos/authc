package authc

import (
	"encoding/json"
	"net/http"

	"github.com/casbin/casbin/v2"
)

// authorizer stores the casbin handler
type authorizer struct {
	*casbin.Enforcer
	Subject
}

// Subject function get subject
type Subject func(r *http.Request) string

// NewAuthorizer returns the authorizer
// uses a Casbin enforcer and Subject function as input
func NewAuthorizer(e *casbin.Enforcer, s Subject) func(next http.HandlerFunc) http.HandlerFunc {
	a := &authorizer{e, s}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			//checks the userName,path,method permission combination from the request.
			allowed, err := a.Enforce(a.Subject(r), r.URL.Path, r.Method)
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
