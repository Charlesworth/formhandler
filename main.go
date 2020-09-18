package main

import (
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const (
	headerKeyContentType    = "Content-Type"
	headerValFormURLEncoded = "application/x-www-form-urlencoded"
	headerValFormMultipart  = "multipart/form-data"
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
	results, files, err := getFormContent(r)

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

func getFormContent(r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error) {
	contentType := r.Header.Get(headerKeyContentType)
	if contentType == headerValFormURLEncoded {
		results, err := parseFormURLEncoded(r)
		return results, nil, err
	}
	if strings.HasPrefix(contentType, headerValFormMultipart) {
		return parseFormMultipart(r)
	}
	return nil, nil, fmt.Errorf(`Unsupported content type "%v", please use "%v" or "%v"`, contentType, headerValFormMultipart, headerValFormURLEncoded)
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
