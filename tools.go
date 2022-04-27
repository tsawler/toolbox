package toolbox

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/gabriel-vasile/mimetype"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321_+"

// Tools is the type for this package. Create a variable of this type and you have access
// to all the methods with the receiver type *Tools.
type Tools struct {
	MaxFileSize int
}

// JSONResponse is the type used for sending JSON around
type JSONResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ReadJSON tries to read the body of a request and converts it into JSON
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte
	if t.MaxFileSize > 0 {
		maxBytes = t.MaxFileSize
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

// WriteJSON takes a response status code and arbitrary data and writes a json response to the client
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// ErrorJSON takes an error, and optionally a response status code, and generates and sends
// a json error response
func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return t.WriteJSON(w, statusCode, payload)
}

// RandomString returns a random string of letters of length n
func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

// PushJSONToRemote posts arbitrary json to some url, and returns error,
// if any, as well as the response status code
func (t *Tools) PushJSONToRemote(client *http.Client, uri string, data any) (int, error) {
	// create json we'll send
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return 0, err
	}

	// build the request and set header
	request, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")

	// call the uri
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	return response.StatusCode, nil
}

// DownloadStaticFile downloads a file, and tries to force the browser to avoid displaying it in
// the browser window by setting content-disposition. It also allows specification of the display name.
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, p, file, displayName string) {
	fp := path.Join(p, file)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))

	http.ServeFile(w, r, fp)
}

// UploadOneFile uploads one file to a specified directory, and gives it a random name.
// It returns the newly named file, the original file name, and potentially an error.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string) (string, string, error) {
	// parse the form so we have access to the file
	err := r.ParseMultipartForm(1024 * 1024 * 1024)
	if err != nil {
		return "", "", err
	}

	var filename, fileNameDisplay string
	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			infile, err := hdr.Open()
			if err != nil {
				return "", "", err
			}
			defer infile.Close()

			ext, err := mimetype.DetectReader(infile)
			if err != nil {
				fmt.Println(err)
				return "", "", err
			}

			_, err = infile.Seek(0, 0)
			if err != nil {
				fmt.Println(err)
				return "", "", err
			}

			filename = t.RandomString(25) + ext.Extension()
			fileNameDisplay = hdr.Filename

			var outfile *os.File
			defer outfile.Close()

			if outfile, err = os.Create(uploadDir + filename); nil != err {
				fmt.Println(err)
			} else {
				_, err := io.Copy(outfile, infile)
				if err != nil {
					fmt.Println(err)
					return "", "", err
				}
			}
		}

	}
	return filename, fileNameDisplay, nil
}

// CreateDirIfNotExist creates a directory, and all necessary parent directories, if it does not exist.
func (t *Tools) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}
