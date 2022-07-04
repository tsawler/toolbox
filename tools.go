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
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321_+"

// Tools is the type for this package. Create a variable of this type and you have access
// to all the methods with the receiver type *Tools.
type Tools struct {
	MaxJSONSize      int      // maximum siz of JSON file we'll process
	MaxFileSize      int      // maximum size of uploaded files in bytes
	AllowedFileTypes []string // allowed file types for upload (e.g. image/jpeg)
}

// JSONResponse is the type used for sending JSON around
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ReadJSON tries to read the body of a request and converts it into JSON
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1024 * 1024 // one megabyte
	if t.MaxJSONSize != 0 {
		maxBytes = t.MaxJSONSize
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(data)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// WriteJSON takes a response status code and arbitrary data and writes a json response to the client
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
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
func (t *Tools) PushJSONToRemote(client *http.Client, uri string, data interface{}) (int, error) {
	// create json we'll send
	jsonData, err := json.Marshal(data)
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

// UploadedFile is a struct used for the uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

// UploadOneFile uploads one file to a specified directory, and gives it a random name.
// It returns the newly named file, the original file name, and potentially an error.
// If the optional last parameter is set to true, then we will not rename the file, but
// will use the original file name.
func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	// check to see if we are renaming the file with the optional last parameter
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	// parse the form so we have access to the file
	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("the uploaded file is too big")
	}
	var uploadedFile UploadedFile

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			infile, err := hdr.Open()
			if err != nil {
				return nil, err
			}
			defer infile.Close()

			buff := make([]byte, 512)
			_, err = infile.Read(buff)
			if err != nil {
				return nil, err
			}

			allowed := false
			filetype := http.DetectContentType(buff)
			if len(t.AllowedFileTypes) > 0 {
				for _, x := range t.AllowedFileTypes {
					if strings.EqualFold(filetype, x) {
						allowed = true
					}
				}
			} else {
				allowed = true
			}

			if !allowed {
				return nil, errors.New("the uploaded file type is not permitted")
			}

			_, err = infile.Seek(0, 0)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			if renameFile {
				uploadedFile.NewFileName = t.RandomString(25) + filepath.Ext(hdr.Filename)
			} else {
				uploadedFile.NewFileName = hdr.Filename
			}
			uploadedFile.OriginalFileName = hdr.Filename

			var outfile *os.File
			defer outfile.Close()

			if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); nil != err {
				return nil, err
			} else {
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize
			}
		}

	}
	return &uploadedFile, nil
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
