package auth

import (
	"errors"
	"net/http"
)

var (
	// ErrAuthenticationRequired is returned when authentication credentials are required, but
	// were either missing or invalid
	ErrAuthenticationRequired = errors.New("authentication required")
	// ErrAuthorizationFailed is returned when authentication worked, but
	// authorization is still denied
	ErrAuthorizationFailed = errors.New("authorization failed")
)

// ErrPermissionDenied is returned when a specific permission
// check failed in a permissions authorizer
type ErrPermissionDenied struct {
	perm string
}

func (e ErrPermissionDenied) Error() string {
	return "permission denied: " + e.perm
}

// Permission returns the name of the specific permission
// that failed in the permission authorizer
func (e ErrPermissionDenied) Permission() string {
	return e.perm
}

// Client is a basic interface expected in the request context by
// the client authorizer.  It must be identifiable in some way via
// the `Id` method.
type Client interface {
	Id() string
}

// Authorizer is the interface expected by the permissions authorizer middleware.
// It must provide a way of checking specific permissions.
type Authorizer interface {
	HasPermission(perm string) (bool, error)
}

// ErrorHandler is called when an error occurs in authenticator or authorizer
// middlewares
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// StandardErrorHandler provides a default implementation for use in
// authorizer handlers.
//
// NOTE: it could make sense to follow other conventions here and provide a function
// for creating the StandardErrorHandler.
// doing this would allow an instance of `log.Logger` to be optionally passed in, and
// more configurable behavior generally without requiring the app to implement a custom
// error handler.  Food for thought.
func StandardErrorHandler(w http.ResponseWriter, r *http.Request, e error) {
	if e == ErrAuthenticationRequired {
		w.WriteHeader(401)
		w.Write([]byte("Authentication required"))
		return
	}

	if e == ErrAuthorizationFailed {
		w.WriteHeader(403)
		w.Write([]byte("Access denied"))
		return
	}

	if _, ok := e.(ErrPermissionDenied); ok {
		w.WriteHeader(403)
		w.Write([]byte("Access denied"))
		return
	}

	// NOTE: perhaps it makes sense to handle redirects via custom error type here, but
	// I can't decide if that feels gross or not.

	// this is an error that the auth system doesn't know anything about, which
	// means it's probably bad
	http.Error(w, "Internal error", 500)
}

// NewClientAuthorizer returns an authorization middleware that requires a Client
// be set in the request context at the specified key. The client instance must have an
// identifier of some sort set, meaning it cannot be an empty string.
func NewClientAuthorizer(keyname string, failFn ErrorHandler) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := checkClient(keyname, r)
			if err != nil {
				failFn(rw, r, err)
				return
			}
			handler.ServeHTTP(rw, r)
		})
	}
}

// NewClientAuthorizerMiddleware returns a negroni-style authorization middleware that requires a Client
// be set in the request context at the specified key. The client instance must have an
// identifier of some sort set, meaning it cannot be an empty string.
func NewClientAuthorizerMiddleware(keyname string, failFn ErrorHandler) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		_, err := checkClient(keyname, r)
		if err != nil {
			failFn(rw, r, err)
			return
		}
		next(rw, r)
	}
}

func checkClient(keyname string, req *http.Request) (bool, error) {
	c := req.Context().Value(keyname)
	client, ok := c.(Client)
	if !ok {
		return false, ErrAuthenticationRequired
	}
	if "" == client.Id() {
		return false, ErrAuthorizationFailed
	}
	return true, nil
}

// NewPermissionsAuthorizer return an authorization middleware that requires an Authorizer
// be set in the request context at the specified key.  The middleware facilitates wrapping
// `http.HandlerFunc`s with permission checks, which will only execute if the Authorizer
// grants all specified permissions.
func NewPermissionsAuthorizer(keyname string, failFn ErrorHandler) func(http.Handler, ...string) http.Handler {
	return func(handler http.Handler, perms ...string) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_, err := checkPermissions(keyname, r, perms...)
			if err != nil {
				failFn(rw, r, err)
				return
			}
			handler.ServeHTTP(rw, r)
		})
	}
}

// NewPermissionsAuthorizerMiddleware returns a negroni-style middleware factory for invoking
// permission checks
func NewPermissionsAuthorizerMiddleware(keyname string, failFn ErrorHandler) func(...string) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(perms ...string) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
		return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			_, err := checkPermissions(keyname, r, perms...)
			if err != nil {
				failFn(rw, r, err)
				return
			}
			next(rw, r)
		}
	}
}

func checkPermissions(keyname string, req *http.Request, perms ...string) (bool, error) {
	a := req.Context().Value(keyname)
	// must actually have an authorizer to check - if not, the request must not
	// have been authenticated
	authorizer, ok := a.(Authorizer)
	if !ok {
		return false, ErrAuthenticationRequired
	}

	// check each permission - end early if any one is denied
	for _, perm := range perms {
		if allowed, err := authorizer.HasPermission(perm); err != nil {
			return false, err
		} else if !allowed {
			return false, ErrPermissionDenied{perm}
		}
	}

	return true, nil
}

// NewBasicApiClient return a new BasicApiClient with the specified
// id and permissions list.
func NewBasicApiClient(id string, perms []string) BasicApiClient {
	return BasicApiClient{id, perms}
}

// BasicApiClient implements both the Client and Authorizer interfaces, so
// can easily be used for either type of authentication and authorization.
type BasicApiClient struct {
	id    string
	perms []string
}

func (b BasicApiClient) Id() string {
	return b.id
}

func (b BasicApiClient) HasPermission(perm string) (bool, error) {
	for _, p := range b.perms {
		if p == perm {
			return true, nil
		}
	}
	return false, nil
}
