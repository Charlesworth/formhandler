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

// ParseError is the error returned from parsing the request that can be used
// to produce a http error response with it's status and message
type ParseError struct {
	status int
	msg    string
}

func (pe *ParseError) Error() string {
	return pe.msg
}

func generate(formSize int64, formWithFilesSize int64) func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err *ParseError) {
	// TODO: pass back malformed request
	return func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err *ParseError) {

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
			err = &ParseError{status: http.StatusUnsupportedMediaType, msg: fmt.Sprintf("Content-Type header is required")}

		default:
			err = &ParseError{status: http.StatusUnsupportedMediaType, msg: fmt.Sprintf("Content-Type header %s is unsupported", contentType)}
		}

		return results, files, err
	}
}

// TODO: pass back malformed request
func getFormContent(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err *ParseError) {

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
		err = &ParseError{status: http.StatusUnsupportedMediaType, msg: fmt.Sprintf("Content-Type header is required")}

	default:
		err = &ParseError{status: http.StatusUnsupportedMediaType, msg: fmt.Sprintf("Content-Type header %s is unsupported", contentType)}
	}

	return results, files, err
}

func parseApplicationJSON(reader io.Reader) (results map[string][]string, err *ParseError) {
	dec := json.NewDecoder(reader)
	jsonContent := map[string]interface{}{}
	decodeErr := dec.Decode(&jsonContent)
	if decodeErr != nil {
		var syntaxError *json.SyntaxError

		switch {
		case errors.As(decodeErr, &syntaxError):
			return nil, &ParseError{status: http.StatusBadRequest, msg: fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)}

		case errors.Is(decodeErr, io.ErrUnexpectedEOF):
			return nil, &ParseError{status: http.StatusBadRequest, msg: fmt.Sprintf("Request body contains badly-formed JSON")}

		case errors.Is(decodeErr, io.EOF):
			return nil, &ParseError{status: http.StatusBadRequest, msg: "Request body must not be empty"}

		case decodeErr.Error() == "http: request body too large":
			return nil, &ParseError{status: http.StatusRequestEntityTooLarge, msg: "Request body too large"}

		default:
			// TODO: This is an unkown error here, putting as a server error for now
			return nil, &ParseError{status: http.StatusInternalServerError, msg: "Internal Server Error"}
		}
	}

	secondDecodeErr := dec.Decode(&struct{}{})
	if secondDecodeErr != io.EOF {
		return nil, &ParseError{status: http.StatusBadRequest, msg: "Request body must only contain a single JSON object"}
	}

	return parseMapInterface(jsonContent)
}

func parseMapInterface(mapInterface map[string]interface{}) (results map[string][]string, err *ParseError) {
	results = make(map[string][]string)
	if len(mapInterface) == 0 {
		return nil, &ParseError{status: http.StatusBadRequest, msg: `JSON object contains no fields`}
	}

	for key, interfaceValue := range mapInterface {
		switch value := interfaceValue.(type) {
		// string unmarshals JSON strings
		case string:
			if value == "" {
				return nil, &ParseError{status: http.StatusBadRequest, msg: fmt.Sprintf(`JSON object contains invalid value for field "%s", cannot use an empty string`, key)}
			}
			results[key] = []string{value}

		// []interface{} unmarshals JSON arrays
		case []interface{}:
			if len(value) == 0 {
				return nil, &ParseError{status: http.StatusBadRequest, msg: fmt.Sprintf(`JSON object contains invalid value for field "%s", cannot use an empty array`, key)}
			}

			arrResults := []string{}
			for _, value := range value {
				strValue, ok := value.(string)
				if !ok {
					return nil, &ParseError{status: http.StatusBadRequest, msg: fmt.Sprintf(`JSON object contains invalid array for field "%s", array values must be exclusively strings`, key)}
				}
				arrResults = append(arrResults, strValue)
			}
			results[key] = arrResults

		// reject all other JSON types
		default:
			return nil, &ParseError{status: http.StatusBadRequest, msg: fmt.Sprintf(`JSON object contains invalid value for field "%s", values must be string or []string types`, key)}
		}
	}

	return results, nil
}

func parseFormURLEncoded(r *http.Request) (results map[string][]string, err *ParseError) {
	// Body reader size is capped at 10MB when using ParseForm()
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		return nil, &ParseError{status: http.StatusBadRequest, msg: `Invalid URL encoded form`}
	}

	results = r.Form
	reduceUnansweredFields(results)

	return results, nil
}

func parseFormMultipart(r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err *ParseError) {
	// TODO: use the same as the max bytes reader here
	parseFormErr := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
	if parseFormErr != nil {
		return nil, nil, &ParseError{status: http.StatusBadRequest, msg: `Invalid URL encoded form`}
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
