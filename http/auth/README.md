# Auth #

This package provides tools you can use in your app for handling some authentication and authorization functionality, via basic Api Key and JWT mechanisms.  There is no magic provided - it is up to your application to configure, register and use the functions properly.

## Design ##

When a request comes in, authentication middlewares check for an api key, or a JWT token.  If found, an *authenticator* is called to create an object, which is stored in the request context.  The object is defined by your application, and there are some small interfaces it can implement in order to work with the *authorizers* provided by this package.  The authenticators should be provided by your application - no assumptions are made about how your app should store or validate the credentials.  That said, in most cases your application will probably want to make database queries and return a custom struct that describes the clients in your application.

Next, route *authorization* middlewares can be used when setting up your routing to protect your http handlers, either by requiring that a `Client` be set in the request context, or by requiring that an `Authorizer` has the specified permissions.

As long as your app can create functions to validate incoming api keys, and/or JWT tokens, and return a custom struct that implements the `Client` and `Authorizer` interfaces, you can make use of the *authorization* middlewares provided.

See the tests for basic usage examples.

## Usage Example with Negroni & Gorilla ##

This package provides only basic `net/http` integration, but usage with other libraries is fairly straight forward. For example, to use the auth middlewares in a project with `negroni` using the `gorilla/mux` router, you might do something like this:

```go
// app-specific functions for validating incoming Api Keys and JWT tokens
func authenticateApikey(key string) (interface{}, error) {
	return nil, nil
}

func authenticateJwt(token *jwt.Token) (interface{}, error) {
	return nil, nil
}

// create the authenticators for detecting and verifying credentials
apikeyAuthenticator := auth.CreateApikeyAuthenticator("Key", "ApiClient", auth.DefaultAuthFailedResponder, authenticateApikey)
jwtAuthenticator := auth.CreateJwtAuthenticator("Bearer", "ApiClient", auth.DefaultAuthFailedResponder, authenticateJwt)

// create the authorizers for protecting specific routes
authClient := auth.CreateClientAuthorizer("ApiClient", auth.DefaultAuthFailedResponder)
autPerms := auth.CreatePermissionsAuthorizer("ApiClient", auth.DefaultAuthFailedResponder)

// set up your app routing w/ gorilla/mux router
router := mux.NewRouter()
// completely public route, no authentication required
router.HandleFunc("/", HandleIndex).Methods("GET")
// requires an api client, but no special permissions
router.HandleFunc("/status", authClient(HandleStatus)).Methods("GET")
// requires special permissions
router.HandleFunc("/api/users", authPerms(HandleUsers, 'users.read', 'users.write')).Methods("GET", "POST")

// setup negroni w/ authentication middlewares and router
n := negroni.Classic()
n.Use(negroni.WrapFunc(apikeyAuthenticator))
n.Use(negroni.WrapFunc(jwtAuthenticator))
n.UseHandler(router)
```
