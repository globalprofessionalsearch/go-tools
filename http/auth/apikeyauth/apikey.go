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

// NewAPIKeyAuthenticator creates a middleware that will detect an incoming
// Api Key in the specified location, call a user-define function for validating
// the api key, and store a returned object in the request context.
func NewAPIKeyAuthenticator(keyname, contextKey string, failFn auth.ErrorHandler, authFn APIKeyAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			req, err := checkAPIKey(keyname, contextKey, r, authFn)
			if err != nil {
				failFn(rw, req, err)
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}

// NewAPIKeyAuthenticatorMiddleware creates a negroni-style middleware that will detect an incoming
// Api Key in the specified location, call a user-define function for validating
// the api key, and store a returned object in the request context.
func NewAPIKeyAuthenticatorMiddleware(keyname, contextKey string, failFn auth.ErrorHandler, authFn APIKeyAuthenticator) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		req, err := checkAPIKey(keyname, contextKey, r, authFn)
		if err != nil {
			failFn(rw, req, err)
			return
		}
		next(rw, req)
	}
}

func checkAPIKey(keyname string, contextKey string, r *http.Request, authFn APIKeyAuthenticator) (*http.Request, error) {
	authHeader := r.Header.Get("Authorization")

	// no auth header, continue on
	if authHeader == "" {
		return r, nil
	}

	// no api key sent, continue on
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || authHeaderParts[0] != keyname {
		return r, nil
	}

	// validate the api key, and get something back
	obj, err := authFn(authHeaderParts[1])
	if err != nil {
		return r, err
	}
	if obj == nil {
		return r, errors.New("authenticator returned nil, should return error instead")
	}

	// return new req w/ altered context
	return r.WithContext(context.WithValue(r.Context(), contextKey, obj)), nil
}
