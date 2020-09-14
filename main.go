package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/simpleForm", handleSimpleForm)

	http.ListenAndServe(":8080", nil)
}

func handleSimpleForm(w http.ResponseWriter, r *http.Request) {
	type ContactDetails struct {
		Email   string
		Subject string
		Message string
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	details := ContactDetails{
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}

	fmt.Printf("%+v\n", details)

	w.WriteHeader(200)
}
