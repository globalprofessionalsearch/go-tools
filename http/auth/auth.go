package auth

import (
	"errors"
	"net/http"
)

var (
	// ErrAuthenticationRequired is returned when
	ErrAuthenticationRequired = errors.New("authentication required")
	// ErrAuthorizationFailed is returned when
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

// Authorizer is the interface expected by the permissions authorizer.
// It must provide a way of checking specific permissions.
type Authorizer interface {
	HasPermission(perm string) (bool, error)
}

// ErrorHandler is called when
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// DefaultErrorHandler provides a default implementation for use in
// authorizer handlers.
//
// NOTE: it could make sense to follow other conventions here and provide a function
// for creating the DefaultErrorHandler, which may be renamed to StandardErrorHandler.
// doing this would allow an instance of `log.Logger` to be optionally passed in, and
// more configurable behavior generally without requiring the app to implement a custom
// error handler.  Food for thought.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, e error) {
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

// CreateClientAuthorizer returns an authorization middleware that requires a Client
// be set in the request context at the specified key. The client instance must have an
// identifier of some sort set, meaning it cannot be an empty string.
func CreateClientAuthorizer(keyname string, failFn ErrorHandler) func(http.HandlerFunc) http.HandlerFunc {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			c := r.Context().Value(keyname)
			client, ok := c.(Client)
			if !ok {
				failFn(rw, r, ErrAuthenticationRequired)
				return
			}
			if "" == client.Id() {
				failFn(rw, r, ErrAuthorizationFailed)
				return
			}
			handler(rw, r)
		}
	}
}

// CreatePermissionsAuthorizer return an authorization middleware that requires an Authorizer
// be set in the request context at the specified key.  The middleware facilitates wrapping
// `http.HandlerFunc`s with permission checks, which will only execute if the Authorizer
// grants all specified permissions.
func CreatePermissionsAuthorizer(keyname string, failFn ErrorHandler) func(http.HandlerFunc, ...string) http.HandlerFunc {
	return func(handler http.HandlerFunc, perms ...string) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			a := r.Context().Value(keyname)
			// must actually have an authorizer to check - if not, the request must not
			// have been authenticated
			authorizer, ok := a.(Authorizer)
			if !ok {
				failFn(rw, r, ErrAuthenticationRequired)
				return
			}

			// check each permission - end early if any one is denied
			for _, perm := range perms {
				if allowed, err := authorizer.HasPermission(perm); err != nil {
					failFn(rw, r, err)
					return
				} else if !allowed {
					failFn(rw, r, ErrPermissionDenied{perm})
					return
				}
			}

			// made it, must be ok to proceed
			handler(rw, r)
		}
	}
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
