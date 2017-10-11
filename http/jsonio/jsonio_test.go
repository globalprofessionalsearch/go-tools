package jsonio

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type person struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

func TestUnmarshalRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/", strings.NewReader(`{"name":"Foobert", "age":70}`))
	require.Nil(t, err)
	var in person
	err = UnmarshalRequest(req, &in)
	require.Nil(t, err)
	require.Equal(t, "Foobert", in.Name)
	require.Equal(t, 70, in.Age)
}

func TestRespond(t *testing.T) {
	out := person{"Foobert", 70}
	exp := `{"name":"Foobert","age":70}`
	rw := httptest.NewRecorder()
	Respond(rw, 200, out)
	res := rw.Result()
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	require.Equal(t, exp, string(body))
}

func TestRespondErrors(t *testing.T) {
	exp := `{"errors":["foo","bar"]}`
	rw := httptest.NewRecorder()
	RespondErrors(rw, 400, errors.New("foo"), errors.New("bar"))
	res := rw.Result()
	require.Equal(t, 400, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	require.Equal(t, exp, string(body))
}
