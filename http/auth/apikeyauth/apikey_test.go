package apikeyauth

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/globalprofessionalsearch/go-tools/http/auth"
	"github.com/stretchr/testify/require"
)

func authenticateApiKey(key string) (interface{}, error) {
	if key == "good-api-key" {
		return auth.NewBasicApiClient("good-api-key", []string{}), nil
	}
	return nil, auth.ErrAuthenticationRequired
}

// closing over a handler so I can export the received
// request and test how it was modified
func createTestHandler(t *testing.T, req *http.Request) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// export the received request so it can be tested
		*req = *r
		rw.WriteHeader(200)
		rw.Write([]byte("Hello world!"))
	})
}

func TestNewAPIKeyAuthenticator(t *testing.T) {
	// create the authenticator middleware
	authenticate := NewAPIKeyAuthenticator("Key", "ApiClient", auth.StandardErrorHandler, authenticateApiKey)

	// test case: no key - 200 ok
	t.Run("no api key", func(t *testing.T) {
		var receivedReq http.Request
		ts := httptest.NewServer(authenticate(createTestHandler(t, &receivedReq)))
		defer ts.Close()
		r, _ := http.NewRequest("GET", "", nil)
		res := runReq(t, ts, r)
		require.Equal(t, 200, res.StatusCode)
		out := readRes(t, res)
		require.Equal(t, "Hello world!", out)
		val := receivedReq.Context().Value("ApiClient")
		require.Nil(t, val)
	})

	t.Run("invalid api key sent", func(t *testing.T) {
		var receivedReq http.Request
		ts := httptest.NewServer(authenticate(createTestHandler(t, &receivedReq)))
		defer ts.Close()
		r, _ := http.NewRequest("GET", "", nil)
		r.Header.Set("Authorization", "Key bad-api-key")
		res := runReq(t, ts, r)
		require.Equal(t, 401, res.StatusCode)
		out := readRes(t, res)
		require.Equal(t, "Authentication required", out)
		val := receivedReq.Context().Value("ApiClient")
		require.Nil(t, val)
	})

	t.Run("good api key sent", func(t *testing.T) {
		var receivedReq http.Request
		ts := httptest.NewServer(authenticate(createTestHandler(t, &receivedReq)))
		defer ts.Close()
		r, _ := http.NewRequest("GET", "", nil)
		r.Header.Set("Authorization", "Key good-api-key")
		res := runReq(t, ts, r)
		require.Equal(t, 200, res.StatusCode)
		out := readRes(t, res)
		require.Equal(t, "Hello world!", out)
		val := receivedReq.Context().Value("ApiClient")
		require.NotNil(t, val)
		require.Equal(t, "good-api-key", val.(auth.BasicApiClient).Id())
	})
}

func runReq(t *testing.T, ts *httptest.Server, req *http.Request) *http.Response {
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	req.URL = u
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func readRes(t *testing.T, res *http.Response) string {
	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
