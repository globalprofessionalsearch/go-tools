package jwtauth

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/globalprofessionalsearch/go-tools/auth"
)

// JWTAuthenticator receives a JWT token, and is expected to return an object
// that will be stored in the request context.  If an error is returned, it's
// encouraged to return one of the errors defined in the auth package.
type JWTAuthenticator func(token *jwt.Token) (interface{}, error)

// NewJWTAuthenticator creates a middleware that will detect an incoming
// JWT Token in the specified location, call a user-define function for validating
// the token, and store a returned object in the request context.
func NewJWTAuthenticator(keyname, contextKey string, failFn auth.ErrorHandler, authFn JWTAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			req, err := checkJWT(keyname, contextKey, r, authFn)
			if err != nil {
				failFn(rw, req, err)
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}

// NewJWTAuthenticatorMiddleware creates a negroni-style middleware that will detect an incoming
// JWT token in the specified location, call a user-define function for validating
// the token, and store a returned object in the request context.
func NewJWTAuthenticatorMiddleware(keyname, contextKey string, failFn auth.ErrorHandler, authFn JWTAuthenticator) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		req, err := checkJWT(keyname, contextKey, r, authFn)
		if err != nil {
			failFn(rw, req, err)
			return
		}
		next(rw, req)
	}
}

func checkJWT(keyname, contextKey string, r *http.Request, authFn JWTAuthenticator) (*http.Request, error) {

}
