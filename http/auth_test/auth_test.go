package authtest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"

	"github.com/globalprofessionalsearch/go-tools/http/auth"
	"github.com/globalprofessionalsearch/go-tools/http/auth/apikeyauth"
	"github.com/globalprofessionalsearch/go-tools/testing/webtest"
)

var (
	// map of api keys to permissions for that key
	goodAPIKeys = map[string][]string{
		"good-key-1": []string{"users.read", "users.write"},
		"good-key-2": []string{"users.read"},
	}
)

// appRouter creates an example router including public and private routes.  Some private routes
// only require that an authenticated user exist, but some require that the
// authenticated user actually have certain permissions.
//
// As new forms of auth are supported by the `auth` package, this test should be
// updated to test a realistic usage example
func appRouter() http.Handler {
	// create the authenticators and authorizers
	apikeyAuthenticator := apikeyauth.NewAPIKeyAuthenticator("Key", "ApiClient", auth.DefaultErrorHandler, appAuthenticateAPIKey)
	authClient := auth.NewClientAuthorizer("ApiClient", auth.DefaultErrorHandler)
	authPerms := auth.NewPermissionsAuthorizer("ApiClient", auth.DefaultErrorHandler)
	appHandler := http.HandlerFunc(appHttpHandler)

	// create app router w/ routes, some protected w/ the authorizers
	router := mux.NewRouter()
	router.Handle("/public", appHandler).Methods("GET")
	router.Handle("/private", authClient(appHandler)).Methods("GET")
	router.Handle("/private/users", authPerms(appHandler, "users.read")).Methods("GET")
	router.Handle("/private/users", authPerms(appHandler, "users.read", "users.write")).Methods("POST")

	// create the main app handler by wrapping the router
	// in the various authenticator middlewares
	handler := apikeyAuthenticator(router)

	n := negroni.New()
	n.UseHandler(handler)
	return n
}

func appRouterWithMiddleware() http.Handler {
	apikeyAuthenticator := negroni.HandlerFunc(apikeyauth.NewAPIKeyAuthenticatorMiddleware("Key", "ApiClient", auth.DefaultErrorHandler, appAuthenticateAPIKey))
	authClient := negroni.HandlerFunc(auth.NewClientAuthorizerMiddleware("ApiClient", auth.DefaultErrorHandler))
	permsChecker := auth.NewPermissionsAuthorizerMiddleware("ApiClient", auth.DefaultErrorHandler)
	authPerms := func(perms ...string) negroni.Handler {
		return negroni.HandlerFunc(permsChecker(perms...))
	}
	appHandler := negroni.Wrap(http.HandlerFunc(appHttpHandler))

	// create app router w/ routes, some protected w/ the authorizer middlewares
	auth := negroni.New()
	auth.Use(authClient)
	router := mux.NewRouter()
	router.HandleFunc("/public", appHttpHandler).Methods("GET")
	router.Handle("/private", auth.With(appHandler)).Methods("GET")
	router.Handle("/private/users", auth.With(authPerms("users.read"), appHandler)).Methods("GET")
	router.Handle("/private/users", auth.With(authPerms("users.read", "users.write"), appHandler)).Methods("POST")

	n := negroni.New()
	n.Use(apikeyAuthenticator)
	n.UseHandler(router)
	return n
}

func appAuthenticateAPIKey(key string) (interface{}, error) {
	for k, perms := range goodAPIKeys {
		if k == key {
			return auth.NewBasicApiClient(k, perms), nil
		}
	}
	return nil, auth.ErrAuthenticationRequired
}

func appHttpHandler(rw http.ResponseWriter, r *http.Request) {
	msg := "Hello world!"
	user, ok := r.Context().Value("ApiClient").(auth.BasicApiClient)
	if ok {
		msg = "Hello " + user.Id()
	}
	rw.WriteHeader(200)
	rw.Write([]byte(msg))
}

func TestAppPublicAuth(t *testing.T) {
	tests := []struct {
		method, path string
		expectedCode int
		expectedText string
	}{
		{"GET", "/public", 200, "Hello world!"},
		{"GET", "/private", 401, "Authentication required"},
		{"GET", "/private/users", 401, "Authentication required"},
		{"POST", "/private/users", 401, "Authentication required"},
	}

	ts := httptest.NewServer(appRouter())
	defer ts.Close()
	tsm := httptest.NewServer(appRouterWithMiddleware())
	defer tsm.Close()
	for _, server := range []*httptest.Server{ts, tsm} {
		ts := server
		for _, test := range tests {
			test := test
			t.Run(fmt.Sprint(test), func(t *testing.T) {
				client := webtest.NewClient(t).SetTargetServer(ts)
				res := client.Call(test.method, test.path, nil)
				require.Equal(t, test.expectedCode, res.StatusCode)
				out, _ := ioutil.ReadAll(res.Body)
				require.Equal(t, test.expectedText, string(out))
			})
		}
	}

}

func TestAppApikeyAuth(t *testing.T) {
	tests := []struct {
		method, path, key string
		code              int
		text              string
	}{
		// bad api key, should all fail
		{"GET", "/public", "bad-api-key", 401, "Authentication required"},
		{"GET", "/private", "bad-api-key", 401, "Authentication required"},
		{"GET", "/private/users", "bad-api-key", 401, "Authentication required"},
		{"POST", "/private/users", "bad-api-key", 401, "Authentication required"},

		// good api key, full permissions
		{"GET", "/public", "good-key-1", 200, "Hello good-key-1"},
		{"GET", "/private", "good-key-1", 200, "Hello good-key-1"},
		{"GET", "/private/users", "good-key-1", 200, "Hello good-key-1"},
		{"POST", "/private/users", "good-key-1", 200, "Hello good-key-1"},

		// good api key, partial permissions
		{"GET", "/public", "good-key-2", 200, "Hello good-key-2"},
		{"GET", "/private", "good-key-2", 200, "Hello good-key-2"},
		{"GET", "/private/users", "good-key-2", 200, "Hello good-key-2"},
		{"POST", "/private/users", "good-key-2", 403, "Access denied"},
	}

	ts := httptest.NewServer(appRouter())
	defer ts.Close()
	tsm := httptest.NewServer(appRouterWithMiddleware())
	defer tsm.Close()
	for _, server := range []*httptest.Server{ts, tsm} {
		ts := server
		for _, test := range tests {
			t.Run(fmt.Sprint(test), func(t *testing.T) {
				client := webtest.NewClient(t).SetTargetServer(ts)
				req := client.NewRequest(test.method, test.path, nil)
				req.Header.Set("Authorization", "Key "+test.key)
				res := client.Do(req)
				require.Equal(t, test.code, res.StatusCode)
				out, _ := ioutil.ReadAll(res.Body)
				require.Equal(t, test.text, string(out))
			})
		}
	}
}

func TestAppJWTAuth(t *testing.T) {
	// TODO: test w/ variety of good & bad JWT tokens
	t.Skip("Not implemented...")
}
