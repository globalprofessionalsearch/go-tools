package webtest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipleClients(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		t.Fatal("shouldn't be called")
	}))
	defer ts.Close()
	c1 := NewClient(t).SetTargetServer(ts)
	c1.SetDefaultHeaders(map[string]string{"X-FOO": "FOO"})
	c2 := NewClient(t).SetTargetServer(ts)
	c2.SetDefaultHeaders(map[string]string{"X-FOO": "BAR"})

	req1 := c1.NewRequest("GET", "/", nil)
	req2 := c2.NewRequest("GET", "/", nil)

	require.Equal(t, "FOO", req1.Header.Get("X-FOO"))
	require.Equal(t, "BAR", req2.Header.Get("X-FOO"))
}

func TestCall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(`Hello World`))
	}))
	defer ts.Close()
	c := NewClient(t).SetTargetServer(ts)

	res := c.Call("GET", "/", nil)
	require.Equal(t, 200, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, []byte(`Hello World`), body)
}

func TestQuery(t *testing.T) {
	var u *url.URL
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		u = r.URL
	}))

	c := NewClient(t).SetTargetServer(ts)
	c.Call("GET", "/foo?foo=1&bar=2&baz=cheese&baz=fruit&wat", nil)

	q := u.Query()
	require.Equal(t, []string{"1"}, q["foo"])
	require.Equal(t, []string{"2"}, q["bar"])
	require.Equal(t, []string{"cheese","fruit"}, q["baz"])
	require.Equal(t, []string{""}, q["wat"])
}