// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package authj

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
)

func testAuthzRequest(t *testing.T, next http.HandlerFunc, user string, path string,
	method string,
	code int) {
	r, _ := http.NewRequest(method, path, nil)
	r.SetBasicAuth(user, "123")
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)

	if w.Code != code {
		t.Errorf("%s, %s, %s: %d, supposed to be %d", user, path, method, w.Code, code)
	}
}

func TestBasic(t *testing.T) {
	e, _ := casbin.NewEnforcer("authj_model.conf", "authj_policy.csv")

	next := func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(ContextWithSubject(r.Context(), "alice"))
		n := NewAuthorizer(e)(func(writer http.ResponseWriter,
			request *http.Request) {
		})
		n.ServeHTTP(w, r)
	}

	testAuthzRequest(t, next, "alice", "/dataset1/resource1", "GET", 200)
	testAuthzRequest(t, next, "alice", "/dataset1/resource1", "POST", 200)
	testAuthzRequest(t, next, "alice", "/dataset1/resource2", "GET", 200)
	testAuthzRequest(t, next, "alice", "/dataset1/resource2", "POST", 403)
}

func TestPathWildcard(t *testing.T) {
	e, _ := casbin.NewEnforcer("authj_model.conf", "authj_policy.csv")
	next := func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(ContextWithSubject(r.Context(), "bob"))

		n := NewAuthorizer(e)(func(writer http.ResponseWriter, request *http.Request) {})
		n.ServeHTTP(w, r)
	}

	testAuthzRequest(t, next, "bob", "/dataset2/resource1", "GET", 200)
	testAuthzRequest(t, next, "bob", "/dataset2/resource1", "POST", 200)
	testAuthzRequest(t, next, "bob", "/dataset2/resource1", "DELETE", 200)
	testAuthzRequest(t, next, "bob", "/dataset2/resource2", "GET", 200)
	testAuthzRequest(t, next, "bob", "/dataset2/resource2", "POST", 403)
	testAuthzRequest(t, next, "bob", "/dataset2/resource2", "DELETE", 403)

	testAuthzRequest(t, next, "bob", "/dataset2/folder1/item1", "GET", 403)
	testAuthzRequest(t, next, "bob", "/dataset2/folder1/item1", "POST", 200)
	testAuthzRequest(t, next, "bob", "/dataset2/folder1/item1", "DELETE", 403)
	testAuthzRequest(t, next, "bob", "/dataset2/folder1/item2", "GET", 403)
	testAuthzRequest(t, next, "bob", "/dataset2/folder1/item2", "POST", 200)
	testAuthzRequest(t, next, "bob", "/dataset2/folder1/item2", "DELETE", 403)
}

func TestRBAC(t *testing.T) {
	e, _ := casbin.NewEnforcer("authj_model.conf", "authj_policy.csv")
	next := func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(ContextWithSubject(r.Context(), "cathy"))

		n := NewAuthorizer(e)(func(writer http.ResponseWriter, request *http.Request) {})
		n.ServeHTTP(w, r)
	}

	// cathy can access all /dataset1/* resources via all methods because it has the dataset1_admin role.
	testAuthzRequest(t, next, "cathy", "/dataset1/item", "GET", 200)
	testAuthzRequest(t, next, "cathy", "/dataset1/item", "POST", 200)
	testAuthzRequest(t, next, "cathy", "/dataset1/item", "DELETE", 200)
	testAuthzRequest(t, next, "cathy", "/dataset2/item", "GET", 403)
	testAuthzRequest(t, next, "cathy", "/dataset2/item", "POST", 403)
	testAuthzRequest(t, next, "cathy", "/dataset2/item", "DELETE", 403)

	// delete all roles on user cathy, so cathy cannot access any resources now.
	_, err := e.DeleteRolesForUser("cathy")
	if err != nil {
		panic(err)
	}
	testAuthzRequest(t, next, "cathy", "/dataset1/item", "GET", 403)
	testAuthzRequest(t, next, "cathy", "/dataset1/item", "POST", 403)
	testAuthzRequest(t, next, "cathy", "/dataset1/item", "DELETE", 403)
	testAuthzRequest(t, next, "cathy", "/dataset2/item", "GET", 403)
	testAuthzRequest(t, next, "cathy", "/dataset2/item", "POST", 403)
	testAuthzRequest(t, next, "cathy", "/dataset2/item", "DELETE", 403)
}
