package testing

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func UnmarshalJsonFile(t *testing.T, filepath string, target interface{}) {
	jsonFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(jsonFile, target); err != nil {
		t.Fatal(err)
	}
}
