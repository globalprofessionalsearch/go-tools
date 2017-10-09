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

func appRouter() http.Handler {
	// create the authenticators and authorizers
	apikeyAuthenticator := apikeyauth.CreateAPIKeyAuthenticator("Key", "ApiClient", auth.DefaultErrorHandler, appAuthenticateAPIKey)
	authClient := auth.CreateClientAuthorizer("ApiClient", auth.DefaultErrorHandler)
	authPerms := auth.CreatePermissionsAuthorizer("ApiClient", auth.DefaultErrorHandler)
	router := mux.NewRouter()
	router.HandleFunc("/public", appHttpHandler).Methods("GET")
	router.HandleFunc("/private", authClient(appHttpHandler)).Methods("GET")
	router.HandleFunc("/private/users", authPerms(appHttpHandler, "users.read")).Methods("GET")
	router.HandleFunc("/private/users", authPerms(appHttpHandler, "users.read", "users.write")).Methods("POST")

	// create the main app handler by wrapping the router
	// in the various authenticator middlewares
	handler := apikeyAuthenticator(router)

	n := negroni.Classic()
	n.UseHandler(handler)
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

func TestAppApikeyAuth(t *testing.T) {
	// test bad key

	// test good key 1

	// test good key 2
}

func TestAppJWTAuth(t *testing.T) {
	// TODO: test w/ variety of good & bad JWT tokens
	t.Skip("Not implemented...")
}
