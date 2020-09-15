package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	port := ":8080"
	addr := "localhost" + port
	r := mux.NewRouter()

	r.HandleFunc("/simple", handleSimpleForm).Methods(http.MethodPost)
	r.HandleFunc("/simple", handleTemplate("formTemplates/simple.tmpl", addr+"/simple")).Methods(http.MethodGet)

	r.HandleFunc("/file", handleFormWithMultiFile).Methods(http.MethodPost)
	r.HandleFunc("/singleFile", handleTemplate("formTemplates/singleFile.tmpl", addr+"/file")).Methods(http.MethodGet)
	r.HandleFunc("/multiFile", handleTemplate("formTemplates/multiFile.tmpl", addr+"/file")).Methods(http.MethodGet)

	r.HandleFunc("/complex", handleTemplate("formTemplates/complex.tmpl", addr+"/complex")).Methods(http.MethodGet)
	r.HandleFunc("/complex", handlePrintForm).Methods(http.MethodPost)

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

func handleSimpleForm(w http.ResponseWriter, r *http.Request) {
	type SimpleForm struct {
		Email   string
		Subject string
		Message string
	}

	details := SimpleForm{
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}

	fmt.Printf("%+v\n", details)

	w.WriteHeader(200)
}

func handleFormWithMultiFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
	if err != nil {
		log.Println("can't parse form: ", err.Error())
		w.WriteHeader(400)
		return
	}

	// rework the stdlib FormFile to cycle through the files here
	fhs := r.MultipartForm.File["file"]
	if len(fhs) == 0 {
		log.Println("no files")
		w.WriteHeader(400)
		return
	}

	for _, fileHeader := range fhs {
		fmt.Printf("name: %s, size: %v\n", fileHeader.Filename, fileHeader.Size)
	}

	w.WriteHeader(200)
}

func handlePrintForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
	if err != nil {
		log.Println("can't parse form: ", err.Error())
		w.WriteHeader(400)
		return
	}

	for field, values := range r.PostForm {
		fmt.Printf("field: %v, values: %v\n", field, values)
	}

	for field, fileHeaders := range r.MultipartForm.File {
		fileNames := []string{}
		for _, fileHeader := range fileHeaders {
			if len(fileNames) == 0 {
				fileNames = []string{fileHeader.Filename}
			} else {
				fileNames = append(fileNames, fileHeader.Filename)
			}
		}
		fmt.Printf("field: %v, files: %v\n", field, fileNames)
	}

	w.WriteHeader(200)
}
