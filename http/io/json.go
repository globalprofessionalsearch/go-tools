package json

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func UnmarshalRequestJson(r *http.Request, target interface{}) error {
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

func RespondJson(w http.ResponseWriter, code int, val interface{}) {
	json, err := json.Marshal(val)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(json)
}

func RespondJsonErrors(w http.ResponseWriter, code int, errs ...error) {
	errors := []string{}
	for _, err := range errs {
		errors = append(errors, err.Error())
	}

	RespondJson(w, code, map[string][]string{"errors": errors})
}
