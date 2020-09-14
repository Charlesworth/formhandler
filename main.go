package main

import (
	"fmt"
	"net/http"
)

type ContactDetails struct {
	Email   string
	Subject string
	Message string
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(500)
		return
	}

	details := ContactDetails{
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}

	// do something with details
	fmt.Printf("%+v\n", details)

	w.WriteHeader(200)
}

func main() {
	http.HandleFunc("/", handlePost)

	http.ListenAndServe(":8080", nil)
}
