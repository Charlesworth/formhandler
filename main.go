package forms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

const (
	headerKeyContentType     = "Content-Type"
	headerValFormURLEncoded  = "application/x-www-form-urlencoded"
	headerValApplicationJSON = "application/json"
	headerValFormMultipart   = "multipart/form-data"

	megabyte = 1_048_576
)

func handleForm(w http.ResponseWriter, r *http.Request) {
	getFormContent(w, r)
}

// isMultipartFormHeader returns if the content-type header is multipart/form-data.
// because multipart/form-data has the field boundaries suffixed to the header,
// we need to check that the prefix of the header is "multipart/form-data"
// and not match on string equality like the other content type header checks.
func isMultipartFormHeader(contentType string) bool {
	return strings.HasPrefix(contentType, headerValFormMultipart)
}

func getContentType(header http.Header) string {
	contentType := header.Get(headerKeyContentType)
	if isMultipartFormHeader(contentType) {
		return headerValFormMultipart
	}
	return contentType
}

// MalformedRequest is the error returned from parsing the request that can be used
// to produce a http error response
type MalformedRequest struct {
	status int
	msg    string
}

func (mr *MalformedRequest) Error() string {
	return mr.msg
}

func generate(formSize int64, formWithFilesSize int64) func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {
	// TODO: pass back malformed request
	return func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {

		switch contentType := getContentType(r.Header); contentType {

		case headerValApplicationJSON:
			r.Body = http.MaxBytesReader(w, r.Body, formSize)
			results, err = parseApplicationJSON(r.Body)

		case headerValFormURLEncoded:
			r.Body = http.MaxBytesReader(w, r.Body, formSize)
			results, err = parseFormURLEncoded(r)

		case headerValFormMultipart:
			r.Body = http.MaxBytesReader(w, r.Body, formWithFilesSize)
			results, files, err = parseFormMultipart(r)

		case "":
			err = fmt.Errorf("Content-Type header is required")
			// http.Error(resp, errMsg, http.StatusUnsupportedMediaType)

		default:
			err = fmt.Errorf("Content-Type header %s is unsupported", contentType)
			// http.Error(resp, errMsg, http.StatusUnsupportedMediaType)
		}

		return results, files, err
	}
}

// TODO: pass back malformed request
func getFormContent(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {

	switch contentType := getContentType(r.Header); contentType {

	case headerValApplicationJSON:
		// limit the body size to 1MB as json encoded forms do not include files
		r.Body = http.MaxBytesReader(w, r.Body, megabyte)
		results, err = parseApplicationJSON(r.Body)

	case headerValFormURLEncoded:
		// limit the body size to 1MB as URL encoded forms do not include files
		r.Body = http.MaxBytesReader(w, r.Body, megabyte)
		results, err = parseFormURLEncoded(r)

	case headerValFormMultipart:
		// limit the body size to 10MB as multipart encoded forms can include files
		r.Body = http.MaxBytesReader(w, r.Body, 10*megabyte)
		results, files, err = parseFormMultipart(r)

	case "":
		err = fmt.Errorf("Content-Type header is required")
		// http.Error(resp, errMsg, http.StatusUnsupportedMediaType)
		// return nil, nil, errors.New(errMsg)

	default:
		err = fmt.Errorf("Content-Type header %s is unsupported", contentType)
		// http.Error(resp, errMsg, http.StatusUnsupportedMediaType)
	}

	return results, files, err
}

// TODO: use malformed error here
func parseApplicationJSON(reader io.Reader) (results map[string][]string, err error) {
	dec := json.NewDecoder(reader)
	jsonContent := map[string]interface{}{}
	decodeErr := dec.Decode(&jsonContent)
	if decodeErr != nil {
		var syntaxError *json.SyntaxError

		switch {
		case errors.As(decodeErr, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return nil, &MalformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(decodeErr, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return nil, &MalformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(decodeErr, io.EOF):
			msg := "Request body must not be empty"
			return nil, &MalformedRequest{status: http.StatusBadRequest, msg: msg}

		case decodeErr.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return nil, &MalformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			// TODO: return a generic error
			return nil, decodeErr
		}
	}

	secondDecodeErr := dec.Decode(&struct{}{})
	if secondDecodeErr != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return nil, &MalformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return parseMapInterface(jsonContent)
}

func parseMapInterface(mapInterface map[string]interface{}) (results map[string][]string, err error) {
	results = make(map[string][]string)
	if len(mapInterface) == 0 {
		return nil, errors.New(`JSON object contains no fields`)
	}

	for key, interfaceValue := range mapInterface {
		switch value := interfaceValue.(type) {
		// string unmarshals JSON strings
		case string:
			// TODO: check the string for escape stuff
			if value == "" {
				return nil, fmt.Errorf(`JSON object contains invalid value for field "%s", cannot use an empty string`, key)
			}
			results[key] = []string{value}

		// []interface{} unmarshals JSON arrays
		case []interface{}:
			if len(value) == 0 {
				return nil, fmt.Errorf(`JSON object contains invalid value for field "%s", cannot use an empty array`, key)
			}

			arrResults := []string{}
			// TODO: unpack the strings
			for _, value := range value {
				strValue, ok := value.(string)
				if !ok {
					// TODO: send back error in malformed format
					return nil, fmt.Errorf(`JSON object contains invalid array for field "%s", array values must be exclusively strings`, key)
				}
				// TODO: check the strings for escape stuff
				arrResults = append(arrResults, strValue)
			}
			results[key] = arrResults

		// reject everything else, we only accept string or []string
		default:
			// TODO: send back error in malformed format
			return nil, fmt.Errorf(`JSON object contains invalid value for field "%s", values must be string or []string types`, key)
		}
	}

	return results, nil
}

func parseFormURLEncoded(r *http.Request) (results map[string][]string, err error) {
	// Body reader size is capped at 10MB when using ParseForm()
	err = r.ParseForm()
	if err != nil {
		// TODO: server or user error?
		return nil, err
	}

	results = r.Form
	reduceUnansweredFields(results)

	return results, nil
}

func parseFormMultipart(r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {
	err = r.ParseMultipartForm(32 << 20) // maxMemory 32MB
	if err != nil {
		// TODO: server or user error?
		return nil, nil, err
	}

	results = r.PostForm
	reduceUnansweredFields(results)

	return results, r.MultipartForm.File, nil
}

func reduceUnansweredFields(results map[string][]string) {
	// unanswered fields are encoded as an empty []string, these are removed
	for field, values := range results {
		if values == nil || len(values) == 0 || (len(values) == 1 && values[0] == "") {
			delete(results, field)
		}
	}
}
