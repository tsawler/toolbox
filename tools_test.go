package toolbox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// RoundTripFunc is used to satisfy the interface requirements for http.Client
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip is used to satisfy the interface requirements for http.Client
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestTools_PushJSONToRemote(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		return &http.Response{
			StatusCode: http.StatusOK,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	})

	var testApp Tools
	var foo struct {
		Bar string `json:"bar"`
	}
	foo.Bar = "bar"
	_, err := testApp.PushJSONToRemote(client, "http://example.com/some/path", foo)
	if err != nil {
		t.Error("failed to call remote url", err)
	}
}

func TestTools_ReadJSON(t *testing.T) {
	var testApp Tools
	testApp.MaxFileSize = 1048576 * 2

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

	// create json with two json entries
	badJSON := `
		{
			"foo": "bar"
		}
		{
			"alpha": "beta"
		}`

	// create a request with the body
	req, err = http.NewRequest("POST", "/", bytes.NewReader([]byte(badJSON)))
	if err != nil {
		t.Log("Error", err)
	}

	err = testApp.ReadJSON(rr, req, &decodedJSON)
	if err == nil {
		t.Error("did not get an error with bad json")
	}
}

func TestTools_WriteJSON(t *testing.T) {
	var testApp Tools

	rr := httptest.NewRecorder()
	payload := JSONResponse{
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

func TestTools_ErrorJSON(t *testing.T) {
	var testApp Tools

	rr := httptest.NewRecorder()
	err := testApp.ErrorJSON(rr, errors.New("some error"))
	if err != nil {
		t.Error(err)
	}

	var requestPayload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	err = decoder.Decode(&requestPayload)
	if err != nil {
		t.Error("received error when decoding ErrorJSON payload:", err)
	}

	if !requestPayload.Error {
		t.Error("error set to false in response from ErrorJSON, and should be set to true")
	}

	// test with status
	err = testApp.ErrorJSON(rr, errors.New("another error"), http.StatusServiceUnavailable)
	if err != nil {
		t.Error(err)
	}
}

func TestTools_RandomString(t *testing.T) {
	var testApp Tools

	s := testApp.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}

func TestTools_DownloadStaticFile(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	var testApp Tools

	testApp.DownloadStaticFile(rr, req, "./testdata", "tgg.jpg", "gatsby.jpg")

	res := rr.Result()
	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "240366" {
		t.Error("wrong content length of", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachment; filename=\"gatsby.jpg\"" {
		t.Error("wrong content disposition of", res.Header["Content-Disposition"][0])
	}

	_, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
}

func TestTools_UploadOneFile(t *testing.T) {
	// set up a pipe to avoid buffering
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer writer.Close()
		// create the form data field 'fileupload'
		part, err := writer.CreateFormFile("file", "./testdata/img.png")
		if err != nil {
			t.Error(err)
		}

		f, err := os.Open("./testdata/img.png")
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		if err != nil {
			t.Error("error decoding image", err)
		}

		err = png.Encode(part, img)
		if err != nil {
			t.Error(err)
		}
	}()

	// read from the pipe which receives data
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	var testTools Tools

	newFile, _, err := testTools.UploadOneFile(request, "./testdata/uploads/")
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", newFile)); os.IsNotExist(err) {
		t.Error("Expected file to exist", err)
	}

	// clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", newFile))
}

func TestTools_CreateDirIfNotExist(t *testing.T) {
	var testTool Tools

	err := testTool.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	err = testTool.CreateDirIfNotExist("./testdata/myDir")
	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/myDir")
}
