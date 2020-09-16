package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleSimpleForm(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSimpleForm)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler expected success for basically anything!")
}

func TestHandleSimpleFormWithValidation(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleFormWithMultiFile)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "expected an error for no request body")
}

func TestWhatever(t *testing.T) {
	// To test a "application/x-www-form-urlencoded" form, encode the form results as url.Values
	form := url.Values{"key": {"Value"}, "id": {"123"}}
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSimpleForm)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code, "expected an error for no request body")
}

func TestWhatever1(t *testing.T) {
	// To test a "application/x-www-form-urlencoded" form, encode the form results as url.Values
	form := url.Values{"email": {"Value"}, "subject": {"asdf"}, "message": {"123"}}
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSimpleForm)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "expected an error for no request body")
	bodyBytes, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		log.Fatal(err)
	}
	assert.Empty(t, string(bodyBytes))
}

func TestWhatever2(t *testing.T) {
	// To test a "application/x-www-form-urlencoded" form, encode the form results as url.Values
	form := url.Values{"subject": {"asdf"}, "message": {"123"}}
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleSimpleForm)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code, "expected an error for no request body")
	bodyBytes, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		log.Fatal(err)
	}
	assert.Contains(t, string(bodyBytes), "email")
}
