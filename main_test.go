package main

import (
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestHandleSimpleForm(t *testing.T) {
// 	req, err := http.NewRequest(http.MethodPost, "/", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handleSimpleForm)

// 	handler.ServeHTTP(rr, req)

// 	assert.Equal(t, http.StatusOK, rr.Code, "handler expected success for basically anything!")
// }

// func TestHandleSimpleFormWithValidation(t *testing.T) {
// 	req, err := http.NewRequest(http.MethodPost, "/", nil)
// 	assert.NoError(t, err)

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handleFormWithMultiFile)

// 	handler.ServeHTTP(rr, req)

// 	assert.Equal(t, http.StatusBadRequest, rr.Code, "expected an error for no request body")
// }

func TestGetFormContent(t *testing.T) {
	var formContentTests = []struct {
		testName               string
		testRequestConstructor func() (*http.Request, error)
		expectedResultOutput   map[string][]string
		expectedFileOutput     map[string][]*multipart.FileHeader
	}{
		{"url encoded form with no fields", func() (*http.Request, error) {
			form := url.Values{}
			r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r, err
		}, map[string][]string{}, nil},

		{"url encoded form with only empty fields", func() (*http.Request, error) {
			form := url.Values{"field1": {}, "field2": {}}
			r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r, err
		}, map[string][]string{}, nil},

		{"url encoded form with some filled and some empty fields", func() (*http.Request, error) {
			form := url.Values{"field1": {"value1"}, "field2": {}}
			r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r, err
		}, map[string][]string{"field1": {"value1"}}, nil},

		{"url encoded form with no empty fields", func() (*http.Request, error) {
			form := url.Values{"field1": {"value1"}, "field2": {"value2"}}
			r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r, err
		}, map[string][]string{"field1": {"value1"}, "field2": {"value2"}}, nil},

		{"multipart form with no empty fields", func() (*http.Request, error) {
			form := url.Values{"field1": {"value1"}, "field2": {"value2"}}
			r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r, err
		}, map[string][]string{"field1": {"value1"}, "field2": {"value2"}}, nil},
	}

	for _, tt := range formContentTests {
		t.Run(tt.testName, func(t *testing.T) {
			r, err := tt.testRequestConstructor()
			assert.NoError(t, err)

			results, files, err := getFormContent(r)

			assert.Equal(t, tt.expectedResultOutput, results, "unexpected parsed form results")
			assert.Equal(t, tt.expectedFileOutput, files, "unexpected form files")
		})
	}
}
