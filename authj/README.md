# authj

authj is an authorization middleware for net/http, it's based on
[casbin](https://github.com/casbin/casbin).

[![GoDoc](https://godoc.org/github.com/thinkgos/http-middlewares/authj?status.svg)](https://godoc.org/github.com/thinkgos/http-middlewares/authj)

## Installation

```bash
    go get github.com/thinkgos/http-middlewares/authj
```

## Simple Example

```Go
package main

import (
    "net/http"

    "github.com/casbin/casbin/v2"
    "github.com/thinkgos/http-middlewares/authj"
)

func main() {
    // load the casbin model and policy from files, database is also supported.
    e ,err := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")
    if err!= nil{
        panic(err)
    }
    // define your router, and use the Casbin authj middleware.
    // the access that is denied by authz will return HTTP 403 error.
    // before you use middleware2 your should use middleware1 to set subject
    middleware1 := func(next http.HandlerFunc) http.HandlerFunc{
        return func(w http.ResponseWriter, r *http.Request) {
            next.ServeHTTP(w, r.WithContext(authj.ContextWithSubject(r.Context(), "admin")))
        }
    }
    middleware2 := authj.NewAuthorizer(e,authj.ContextSubject)
}
```

## Documentation

The authorization determines a request based on ``{subject, object, action}``, which means what ``subject`` can perform what ``action`` on what ``object``. In this plugin, the meanings are:

1. ``subject``: the logged-on user name
2. ``object``: the URL path for the web resource like "dataset1/item1"
3. ``action``: HTTP method like GET, POST, PUT, DELETE, or the high-level actions you defined like "read-file", "write-blog"

For how to write authorization policy and other details, please refer to [the Casbin's documentation](https://github.com/casbin/casbin).

## Getting Help

- [Gin-authz](https://github.com/gin-contrib/authz)
- [Gin-authj](https://github.com/thinkgos/gin-middlewares/authj)

## License

This project is under MIT License. See the [LICENSE](LICENSE) file for the full license text.
