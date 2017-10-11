package auth

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func handler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Write([]byte("Hello world!"))
}

func TestHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/", nil)
	rw := httptest.NewRecorder()
	handler(rw, r)
	res := rw.Result()

	out, _ := ioutil.ReadAll(res.Body)

	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, "Hello world!", string(out))
}

func TestNewClientAuthorizer(t *testing.T) {
	// wrap handler in a client checker
	h := NewClientAuthorizer("ApiClient", DefaultErrorHandler)(http.HandlerFunc(handler))

	// no api client - fail
	r := httptest.NewRequest("GET", "http://example.com/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	res := rw.Result()
	out, _ := ioutil.ReadAll(res.Body)
	require.Equal(t, 401, res.StatusCode)
	require.Equal(t, "Authentication required", string(out))

	// with api client, but no id - fail
	r = httptest.NewRequest("GET", "http://example.com/", nil)
	req := r.WithContext(context.WithValue(r.Context(), "ApiClient", BasicApiClient{}))
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	res = rw.Result()
	out, _ = ioutil.ReadAll(res.Body)
	require.Equal(t, 403, res.StatusCode)
	require.Equal(t, "Access denied", string(out))

	// with api client and an id
	r = httptest.NewRequest("GET", "http://example.com/", nil)
	req = r.WithContext(context.WithValue(r.Context(), "ApiClient", BasicApiClient{id: "some-id"}))
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	res = rw.Result()
	out, _ = ioutil.ReadAll(res.Body)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, "Hello world!", string(out))
}

func TestNewPermissionsAuthorizer(t *testing.T) {
	h := NewPermissionsAuthorizer("ApiClient", DefaultErrorHandler)(http.HandlerFunc(handler), "foo", "bar")

	// no authorizer - fail
	r := httptest.NewRequest("GET", "http://example.com/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	res := rw.Result()
	out, _ := ioutil.ReadAll(res.Body)
	require.Equal(t, 401, res.StatusCode)
	require.Equal(t, "Authentication required", string(out))

	// bad perms - fail
	r = httptest.NewRequest("GET", "http://example.com/", nil)
	req := r.WithContext(context.WithValue(r.Context(), "ApiClient", BasicApiClient{perms: []string{"foo"}}))
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	res = rw.Result()
	out, _ = ioutil.ReadAll(res.Body)
	require.Equal(t, 403, res.StatusCode)
	require.Equal(t, "Access denied", string(out))

	// good perms - pass
	r = httptest.NewRequest("GET", "http://example.com/", nil)
	req = r.WithContext(context.WithValue(r.Context(), "ApiClient", BasicApiClient{perms: []string{"foo", "bar"}}))
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	res = rw.Result()
	out, _ = ioutil.ReadAll(res.Body)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, "Hello world!", string(out))
}
