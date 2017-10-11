// Package jsonio contains utilities for handing incoming/outgoing
// json http requests
package jsonio

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// UnmarshalRequest decodes json from a request body into a target
func UnmarshalRequest(r *http.Request, target interface{}) error {
	in, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(in, target)
	if err != nil {
		return err
	}
	return nil
}

// Respond sends a json response, marshaling any sent data into json
func Respond(w http.ResponseWriter, code int, val interface{}) {
	jsn, err := json.Marshal(val)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsn)
}

// RespondErrors sends a json response, marshaling one or more errors
// into an `errors` array
func RespondErrors(w http.ResponseWriter, code int, errs ...error) {
	errors := []string{}
	for _, err := range errs {
		errors = append(errors, err.Error())
	}

	Respond(w, code, map[string][]string{"errors": errors})
}
