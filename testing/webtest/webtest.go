// Package webtest provides convenince methods for making requests to an instance
// of `httptest.Server`.
package webtest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Client is a wrapper around a basic `http.Client` with convenience methods for testing
// calls to a `httptest.Server`.  Any internal errors encountered will automatically
// call `t.Fatal`.
//
// Example:
//
//	```go
//	func TestMyServer(t *testing.T) {
// 	}
//	```
//
type Client struct {
	http.Client

	t              *testing.T
	targetServer   *httptest.Server
	defaultHeaders map[string]string
}

// NewClient return a new client for use with a specific `*testing.T`.  Internal errors
// will authomatically end the test by calling `t.Fatal`.
func NewClient(t *testing.T) *Client {
	return &Client{t: t, Client: http.Client{}}
}

// SetDefaultHeaders sets any headers that should automaticaly be set on
// any requests to the target server.  For most cases this would be for
// setting any required `Authorization` headers for authenticating
// multiple requests.
func (c *Client) SetDefaultHeaders(headers map[string]string) *Client {
	c.defaultHeaders = headers
	return c
}

// SetTargetServer sets the target instance of `httptest.Server` that any
// requests should be made to.
func (c *Client) SetTargetServer(target *httptest.Server) *Client {
	c.targetServer = target
	return c
}

// Call runs a request against the target server with the given method, path
// and body.  Failures fail the test, and the response is returned.
func (c Client) Call(method, path string, body io.Reader) *http.Response {
	req := c.NewRequest(method, path, body)
	return c.Do(req)
}

// CallJson runs a request against the target server with the given method, path
// and body.  The body can be anything that `json.Marshal` can handle.
// Internal failures fail the test, and the response is returned.
func (c Client) CallJson(method, path string, body interface{}) *http.Response {
	req := c.NewJsonRequest(method, path, body)
	return c.Do(req)
}

// NewRequest returns a new request configured for the test server, with any default
// headers already set.
func (c Client) NewRequest(method, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		c.t.Fatal(err)
	}
	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	// the url must be absolute, so depends on the test server
	absPath := strings.Join([]string{c.targetServer.URL, path}, "")

	// check for a query string, and ensure it is parsed and encoded properly
	queryParts := strings.Split(absPath, "?")
	if len(queryParts) == 2 {
		parsedQ, err := url.ParseQuery(queryParts[1])
		if err != nil {
			c.t.Fatal(err)
		}
		absPath += "?" + parsedQ.Encode()
	}

	uri, err := url.ParseRequestURI(absPath)
	if err != nil {
		c.t.Fatal(err)
	}
	req.URL = uri
	return req
}

// NewJsonRequest returns a new request, with the given body converted into
// json and json headers already set.  The request has default headers set
// and its path is configured for the available test server.
func (c Client) NewJsonRequest(method, path string, body interface{}) *http.Request {
	jsn, err := json.Marshal(body)
	if err != nil {
		c.t.Fatal(err)
	}
	req := c.NewRequest(method, path, bytes.NewReader(jsn))
	req.Header.Set("Content Type", "application/json")
	return req
}

// Do runs the given request, failing the test in case of interenal error.
// It is assumed that the request has already been configured properly
// for the test server.
func (c Client) Do(req *http.Request) *http.Response {
	res, err := c.Client.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	return res
}

// UnmarshalJsonResponse decodes json in a response into the target. Any
// internal errors automatically fail the test.
func UnmarshalJsonResponse(t *testing.T, res *http.Response, target interface{}) {
	jsn, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(jsn, target); err != nil {
		t.Fatal(err)
	}
}
