package formhandler

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
	headerKeyContentType = "Content-Type"

	headerValFormURLEncoded  = "application/x-www-form-urlencoded"
	headerValApplicationJSON = "application/json"
	headerValFormMultipart   = "multipart/form-data"

	megabyte = 1_048_576
)

// GetFormContent accepts a request of content type "application/x-www-form-urlencoded",
// "application/json" or "multipart/form-data", parses the body and returns the form data
// and files contained in the request
func GetFormContent(
	w http.ResponseWriter,
	r *http.Request,
) (
	results map[string][]string,
	files map[string][]*multipart.FileHeader,
	err error,
) {
	return GetFormContentWithConfig(megabyte, megabyte*10, megabyte*10)(w, r)
}

// GetFormContentWithConfig operates the same as GetFormContent but with added config options:
// - maxFormSize: The maximum size in bytes a form request can be (applies to JSON and URL encoded forms, which cannot have files attached)
// - maxFormWithFilesSize: The maximum size in bytes a form request with attached files can be (applies to multipart/form-data encoded forms, which can have files attached)
// - maxMemory: Given a form request body is parsed, maxMemory bytes of its file parts are stored in memory, with the remainder stored on disk in temporary files (applies to multipart/form-data encoded forms, which can have files attached)
func GetFormContentWithConfig(
	maxFormSize int64,
	maxFormWithFilesSize int64,
	maxMemory int64,
) func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {
	return func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {

		switch contentType := getContentType(r.Header); contentType {

		case headerValApplicationJSON:
			r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
			results, err = parseApplicationJSON(r.Body)

		case headerValFormURLEncoded:
			r.Body = http.MaxBytesReader(w, r.Body, maxFormSize)
			results, err = parseFormURLEncoded(r)

		case headerValFormMultipart:
			r.Body = http.MaxBytesReader(w, r.Body, maxFormWithFilesSize)
			results, files, err = parseFormMultipart(r, maxMemory)

		case "":
			err = &ParseError{Status: http.StatusUnsupportedMediaType, Msg: fmt.Sprintf("Content-Type header is required")}

		default:
			err = &ParseError{Status: http.StatusUnsupportedMediaType, Msg: fmt.Sprintf("Content-Type header %s is unsupported", contentType)}
		}

		return results, files, err
	}
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
// to produce a http error response with a status and message
type ParseError struct {
	Status int
	Msg    string
}

func (pe *ParseError) Error() string {
	return pe.Msg
}

func parseApplicationJSON(reader io.Reader) (results map[string][]string, err *ParseError) {
	dec := json.NewDecoder(reader)
	jsonContent := map[string]interface{}{}
	decodeErr := dec.Decode(&jsonContent)
	if decodeErr != nil {
		var syntaxError *json.SyntaxError

		switch {
		case errors.As(decodeErr, &syntaxError):
			return nil, &ParseError{Status: http.StatusBadRequest, Msg: fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)}

		case errors.Is(decodeErr, io.ErrUnexpectedEOF):
			return nil, &ParseError{Status: http.StatusBadRequest, Msg: "Request body contains badly-formed JSON"}

		case errors.Is(decodeErr, io.EOF):
			return nil, &ParseError{Status: http.StatusBadRequest, Msg: "Request body must not be empty"}

		case decodeErr.Error() == "http: request body too large":
			return nil, &ParseError{Status: http.StatusRequestEntityTooLarge, Msg: "Request body too large"}

		default:
			return nil, &ParseError{Status: http.StatusInternalServerError, Msg: "JSON parsing error"}
		}
	}

	secondDecodeErr := dec.Decode(&struct{}{})
	if secondDecodeErr != io.EOF {
		return nil, &ParseError{Status: http.StatusBadRequest, Msg: "Request body must only contain a single JSON object"}
	}

	return parseMapInterface(jsonContent)
}

func parseMapInterface(mapInterface map[string]interface{}) (results map[string][]string, err *ParseError) {
	results = make(map[string][]string)
	if len(mapInterface) == 0 {
		return nil, &ParseError{Status: http.StatusBadRequest, Msg: `JSON object contains no fields`}
	}

	for key, interfaceValue := range mapInterface {
		switch value := interfaceValue.(type) {
		// string unmarshals JSON strings
		case string:
			if value == "" {
				return nil, &ParseError{Status: http.StatusBadRequest, Msg: fmt.Sprintf(`JSON object contains invalid value for field "%s", cannot use an empty string`, key)}
			}
			results[key] = []string{value}

		// []interface{} unmarshals JSON arrays
		case []interface{}:
			if len(value) == 0 {
				return nil, &ParseError{Status: http.StatusBadRequest, Msg: fmt.Sprintf(`JSON object contains invalid value for field "%s", cannot use an empty array`, key)}
			}

			arrResults := []string{}
			for _, value := range value {
				strValue, ok := value.(string)
				if !ok {
					return nil, &ParseError{Status: http.StatusBadRequest, Msg: fmt.Sprintf(`JSON object contains invalid array for field "%s", array values must be exclusively strings`, key)}
				}
				arrResults = append(arrResults, strValue)
			}
			results[key] = arrResults

		// reject all other JSON types
		default:
			return nil, &ParseError{Status: http.StatusBadRequest, Msg: fmt.Sprintf(`JSON object contains invalid value for field "%s", values must be string or []string types`, key)}
		}
	}

	return results, nil
}

func parseFormURLEncoded(r *http.Request) (results map[string][]string, err *ParseError) {
	// Body reader size is capped at 10MB when using ParseForm()
	parseFormErr := r.ParseForm()
	if parseFormErr != nil {
		return nil, &ParseError{Status: http.StatusBadRequest, Msg: `Invalid URL encoded form`}
	}

	results = r.Form
	reduceUnansweredFields(results)

	return results, nil
}

func parseFormMultipart(r *http.Request, maxMemory int64) (results map[string][]string, files map[string][]*multipart.FileHeader, err *ParseError) {
	parseFormErr := r.ParseMultipartForm(maxMemory)
	if parseFormErr != nil {
		return nil, nil, &ParseError{Status: http.StatusBadRequest, Msg: `Invalid URL encoded form`}
	}

	results = r.PostForm
	reduceUnansweredFields(results)

	return results, r.MultipartForm.File, nil
}

// Unanswered fields in URL encoded and multipart forms are encoded as an empty []string,
// this function removes the empty []string from the results
func reduceUnansweredFields(results map[string][]string) {
	for field, values := range results {
		if values == nil || len(values) == 0 || (len(values) == 1 && values[0] == "") {
			delete(results, field)
		}
	}
}
