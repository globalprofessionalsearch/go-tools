package apikeyauth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/globalprofessionalsearch/go-tools/http/auth"
)

// APIKeyAuthenticator receives a string, and is expected to return an object
// that will be stored in the request context.  If an error is returned, it's
// encouraged to return one of the errors defined in the auth package.
type APIKeyAuthenticator func(key string) (interface{}, error)

// CreateAPIKeyAuthenticator creates a middleware that will detect an incoming
// Api Key in the specified location, call a user-define function for validating
// the api key, and store a returned object in the request context.
func CreateAPIKeyAuthenticator(keyname, contextKey string, failFn auth.ErrorHandler, authFn APIKeyAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// no auth header, continue on
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(rw, r)
				return
			}

			// no api key sent, continue on
			authHeaderParts := strings.Split(authHeader, " ")
			if len(authHeaderParts) != 2 || authHeaderParts[0] != keyname {
				next.ServeHTTP(rw, r)
				return
			}

			// validate the api key, and get something back
			obj, err := authFn(authHeaderParts[1])
			if err != nil {
				failFn(rw, r, err)
				return
			}
			if obj == nil {
				failFn(rw, r, errors.New("authenticator returned nil, should return error instead"))
				return
			}

			// set the received object in the request context, and call the next handler
			next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), contextKey, obj)))
		})
	}
}
