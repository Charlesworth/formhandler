package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handleSimpleForm).Methods(http.MethodPost)

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

func handleSimpleFormWithValidation(w http.ResponseWriter, r *http.Request) {

	type ContactDetails struct {
		Email   string
		Subject string
		Message string
	}

	details := ContactDetails{
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}

	fmt.Printf("%+v\n", details)

	w.WriteHeader(200)
}
