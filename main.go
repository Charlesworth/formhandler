package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handleSimpleForm).Methods(http.MethodPost)
	r.HandleFunc("/formWithFile", handleFormWithFile).Methods(http.MethodPost)
	r.HandleFunc("/formWithMultiFile", handleFormWithMultiFile).Methods(http.MethodPost)

	http.ListenAndServe(":8080", r)
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

func handleFormWithFile(w http.ResponseWriter, r *http.Request) {
	type FileForm struct {
		Text string
	}

	details := FileForm{
		Text: r.FormValue("text"),
	}

	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		fmt.Println("file error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("name: %s, size: %v\n", fileHeader.Filename, fileHeader.Size)
	fmt.Printf("%+v\n", details)

	w.WriteHeader(200)
}

func handleFormWithMultiFile(w http.ResponseWriter, r *http.Request) {
	type FileForm struct {
		Text string
	}

	details := FileForm{
		Text: r.FormValue("text"),
	}

	// rework the stdlib FormFile to cycle through the files here
	fhs := r.MultipartForm.File["files"]
	for _, fileHeader := range fhs {
		fmt.Printf("name: %s, size: %v\n", fileHeader.Filename, fileHeader.Size)
	}

	fmt.Printf("%+v\n", details)

	w.WriteHeader(200)
}
