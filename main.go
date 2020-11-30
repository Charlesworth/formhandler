package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	headerKeyContentType     = "Content-Type"
	headerValFormURLEncoded  = "application/x-www-form-urlencoded"
	headerValApplicationJSON = "application/json"
	headerValFormMultipart   = "multipart/form-data"
	megabyte                 = 1_048_576
)

func main() {
	port := ":8080"
	addr := "localhost" + port
	r := mux.NewRouter()

	r.HandleFunc("/form", handleForm).Methods(http.MethodPost)
	formSubmissionEndpoint := addr + "/form"

	r.HandleFunc("/simple", handleTemplate("formTemplates/simple.tmpl", formSubmissionEndpoint)).Methods(http.MethodGet)
	r.HandleFunc("/singleFile", handleTemplate("formTemplates/singleFile.tmpl", formSubmissionEndpoint)).Methods(http.MethodGet)
	r.HandleFunc("/multiFile", handleTemplate("formTemplates/multiFile.tmpl", formSubmissionEndpoint)).Methods(http.MethodGet)
	r.HandleFunc("/complex", handleTemplate("formTemplates/complex.tmpl", formSubmissionEndpoint)).Methods(http.MethodGet)

	http.ListenAndServe(port, r)
}

func handleTemplate(tmplFile string, formEnpoint string) func(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(tmplFile))
	templateData := struct{ Address string }{formEnpoint}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request to template %s", tmplFile)
		tmpl.Execute(w, templateData)
	}
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	results, files, err := getFormContent(w, r)

	if err != nil {
		log.Println("Error: ", err.Error())
		w.WriteHeader(400)
		fmt.Fprintf(w, err.Error())
		return
	}

	fmt.Printf("Form Results (len %v): %+v\n", len(results), results)
	fmt.Printf("Form Files (len %v): %+v\n", len(files), files)
	w.WriteHeader(200)
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

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// TODO: pass back malformed request
func getFormContent(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {

	switch contentType := getContentType(r.Header); contentType {

	case headerValApplicationJSON:
		// limit the body size to 1MB as json encoded forms do not include files
		r.Body = http.MaxBytesReader(w, r.Body, megabyte)
		results, err = parseApplicationJSON(r)

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
func parseApplicationJSON(r *http.Request) (results map[string][]string, err error) {
	dec := json.NewDecoder(r.Body)
	jsonContent := map[string]interface{}{}
	decodeErr := dec.Decode(&jsonContent)
	if decodeErr != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(decodeErr, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return nil, &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(decodeErr, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return nil, &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(decodeErr, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return nil, &malformedRequest{status: http.StatusBadRequest, msg: msg}

		// TODO: for checking struct required tags with "DisallowUnknownFields()"
		// case strings.HasPrefix(err.Error(), "json: unknown field "):
		// 	fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		// 	msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
		// 	return nil, &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(decodeErr, io.EOF):
			msg := "Request body must not be empty"
			return nil, &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case decodeErr.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return nil, &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			// TODO: return a generic error
			return nil, decodeErr
		}
	}

	secondDecodeErr := dec.Decode(&struct{}{})
	if secondDecodeErr != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return nil, &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return parseMapInterface(jsonContent)
}

func parseMapInterface(mapInterface map[string]interface{}) (results map[string][]string, err error) {
	results = make(map[string][]string)

	for key, interfaceValue := range mapInterface {
		switch value := interfaceValue.(type) {
		// string unmarshals JSON strings
		case string:
			// TODO: check the string for escape stuff
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
