package toolbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_readJSON(t *testing.T) {
	var testApp Tools

	// create a sample JSON file and add it to body
	sampleJSON := map[string]interface{}{
		"foo": "bar",
	}
	body, _ := json.Marshal(sampleJSON)

	// declare a variable to read the decoded json into
	var decodedJSON struct {
		Foo string `json:"foo"`
	}

	// create a request with the body
	req, err := http.NewRequest("POST", "/", bytes.NewReader(body))
	if err != nil {
		t.Log("Error", err)
	}

	// create a test response recorder, which satisfies the requirements
	// for a ResponseWriter
	rr := httptest.NewRecorder()
	defer req.Body.Close()

	// call readJSON and check for an error
	err = testApp.ReadJSON(rr, req, &decodedJSON)
	if err != nil {
		t.Error("failed to decode json", err)
	}
}

func Test_writeJSON(t *testing.T) {
	var testApp Tools

	rr := httptest.NewRecorder()
	payload := JsonResponse{
		Error:   false,
		Message: "foo",
	}

	headers := make(http.Header)
	headers.Add("FOO", "BAR")
	err := testApp.WriteJSON(rr, http.StatusOK, payload, headers)
	if err != nil {
		t.Errorf("failed to write JSON: %v", err)
	}
}

func Test_errorJSON(t *testing.T) {
	var testApp Tools

	rr := httptest.NewRecorder()
	err := testApp.ErrorJSON(rr, errors.New("some error"))
	if err != nil {
		t.Error(err)
	}

	var requestPayload JsonResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&requestPayload)
	if err != nil {
		t.Error("received error when decoding errorJSON payload:", err)
	}

	if !requestPayload.Error {
		t.Error("error set to false in response from errorJSON, and should be set to true")
	}
}

func Test_randomString(t *testing.T) {
	var testApp Tools

	s := testApp.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}
