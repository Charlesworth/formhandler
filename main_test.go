package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFormContent_URLEncoded(t *testing.T) {
	var formContentTests = []struct {
		testName               string
		testRequestConstructor func() (req *http.Request, err error)
		expectedValuesOutput   map[string][]string
	}{

		{
			"No fields",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{})
			},
			map[string][]string{},
		},
		{
			"single value field with [none] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {}})
			},
			map[string][]string{},
		},
		{
			"single value field with [none] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value1"}})
			},
			map[string][]string{"field1": {"value1"}},
		},
		{
			"single value field with [multiple] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value1", "value2"}})
			},
			map[string][]string{"field1": {"value1", "value2"}},
		},
		{
			"multiple value fields with [none] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {}, "field2": {}})
			},
			map[string][]string{},
		},
		{
			"multiple value fields with [one] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value1"}, "field2": {"value2"}})
			},
			map[string][]string{"field1": {"value1"}, "field2": {"value2"}},
		},
		{
			"multiple value fields with [one, none] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value1"}, "field2": {}})
			},
			map[string][]string{"field1": {"value1"}},
		},
		{
			"multiple value fields with [multiple] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value11", "value12"}, "field2": {"value21", "value22"}})
			},
			map[string][]string{"field1": {"value11", "value12"}, "field2": {"value21", "value22"}},
		},
		{
			"multiple value fields with [multiple, none] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value11", "value12"}, "field2": {}})
			},
			map[string][]string{"field1": {"value11", "value12"}},
		},
		{
			"multiple value fields with [multiple, one] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value11", "value12"}, "field2": {"value21"}})
			},
			map[string][]string{"field1": {"value11", "value12"}, "field2": {"value21"}},
		},
		{
			"multiple value fields with [multiple, none, one] value",
			func() (*http.Request, error) {
				return constructURLEncodedForm(url.Values{"field1": {"value11", "value12"}, "field2": {"value21"}, "field3": {}})
			},
			map[string][]string{"field1": {"value11", "value12"}, "field2": {"value21"}},
		},
	}

	for _, tt := range formContentTests {
		t.Run(tt.testName, func(t *testing.T) {
			r, err := tt.testRequestConstructor()
			assert.NoError(t, err, "Error constructing test request")

			results, files, err := getFormContent(r)

			assert.Equal(t, tt.expectedValuesOutput, results, "unexpected parsed form results")
			assert.Empty(t, files, "unexpected file parsed from url encoded form")
		})
	}
}

// func TestGetFormContent_Values(t *testing.T) {
// 	var formContentTests = []struct {
// 		testName               string
// 		testRequestConstructor func() (req *http.Request, cleanupFunc func(), err error)
// 		expectedValuesOutput   map[string][]string
// 	}{

// 		// URL encoded forms
// 		{
// 			"url encoded form with no fields",
// 			func() (*http.Request, func(), error) {
// 				return constructURLEncodedForm(url.Values{})
// 			},
// 			map[string][]string{},
// 		},
// 		// single value field with [none] value
// 		// single value field with [one] value
// 		// single value field with [multiple] value
// 		// multiple value fields with [none] value
// 		// multiple value fields with [one] value
// 		// multiple value fields with [one, none] value
// 		// multiple value fields with [multiple] value
// 		// multiple value fields with [multiple, none] value
// 		// multiple value fields with [multiple, one] value
// 		// multiple value fields with [multiple, none, one] value
// 		{
// 			"url encoded form with only empty fields",
// 			func() (*http.Request, func(), error) {
// 				return constructURLEncodedForm(url.Values{"field1": {}, "field2": {}})
// 			},
// 			map[string][]string{},
// 		},
// 		{
// 			"url encoded form with some filled and some empty fields",
// 			func() (*http.Request, func(), error) {
// 				return constructURLEncodedForm(url.Values{"field1": {"value1"}, "field2": {}})
// 			},
// 			map[string][]string{"field1": {"value1"}},
// 		},
// 		{
// 			"url encoded form with no empty fields",
// 			func() (*http.Request, func(), error) {
// 				return constructURLEncodedForm(url.Values{"field1": {"value1"}, "field2": {"value2"}})
// 			},
// 			map[string][]string{"field1": {"value1"}, "field2": {"value2"}},
// 		},

// 		// Multipart forms
// 		{
// 			"multipart form with no empty fields and a single non-multiple file",
// 			func() (*http.Request, func(), error) {
// 				values := map[string]io.Reader{
// 					"field1": strings.NewReader("value1"),
// 				}
// 				req, err := constructMultipartForm(values)

// 				return req, nil, err
// 			},
// 			map[string][]string{"field1": {"value1"}},
// 		},
// 		{
// 			"multipart form with no empty fields",
// 			func() (*http.Request, func(), error) {
// 				values := map[string]io.Reader{
// 					"field1": strings.NewReader("value1"),
// 				}
// 				req, err := constructMultipartForm(values)

// 				return req, func() {}, err
// 			},
// 			map[string][]string{"field1": {"value1"}},
// 		},

// !!! no fields
// no fields
// !!! no fields

// !!! file fields
// single file field with [none] files
// single file field with [one] files
// single file field with [multiple] files
// multiple file fields with [none] files
// multiple file fields with [one] files
// multiple file fields with [one, none] files
// multiple file fields with [multiple] files
// multiple file fields with [multiple, none] files
// multiple file fields with [multiple, one] files
// multiple file fields with [multiple, none, one] files
// !!! file fields

// !!! value fields
// single value field with [none] value
// single value field with [one] value
// single value field with [multiple] value
// multiple value fields with [none] value
// multiple value fields with [one] value
// multiple value fields with [one, none] value
// multiple value fields with [multiple] value
// multiple value fields with [multiple, none] value
// multiple value fields with [multiple, one] value
// multiple value fields with [multiple, none, one] value
// !!! value fields

// !!! file and value fields

// !!! file and value fields

// 	}

// 	for _, tt := range formContentTests {
// 		t.Run(tt.testName, func(t *testing.T) {
// 			r, cleanup, err := tt.testRequestConstructor()
// 			assert.NoError(t, err)

// 			results, _, err := getFormContent(r)

// 			assert.Equal(t, tt.expectedValuesOutput, results, "unexpected parsed form results")

// 			// TODO: this doesn't account for "multiple" attribute files
// 			// assert.Equal(t, tt.expectedFileCount, files, "unexpected form files")

// 			cleanup()
// 		})
// 	}
// }

func constructURLEncodedForm(values url.Values) (*http.Request, error) {
	r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r, err
}

// {
// 	"multipart form with no empty fields and a single non-multiple file",
// 	func() (*http.Request, func(), error) {
// 		testFile1, testFile1CleanUp, err := tempTestFile("png")
// 		values := map[string]io.Reader{
// 			"file1":  testFile1,
// 			"field1": strings.NewReader("value1"),
// 		}
// 		req, err := constructMultipartForm(values)

// 		return req, testFile1CleanUp, err
// 	},
// 	map[string][]string{"field1": {"value1"}},
// },

func tempTestFile(fileSuffix string) (file *os.File, cleanupFunc func(), err error) {
	file, err = ioutil.TempFile(os.TempDir(), "testFile-*."+fileSuffix)
	if err != nil {
		return nil, nil, err
	}

	// Write to the file
	text := []byte("Writing some text into the temp test file")
	if _, err = file.Write(text); err != nil {
		return nil, nil, err
	}

	// // Close the file
	// if err := file.Close(); err != nil {
	// 	return nil, nil, err
	// }

	return file, func() { os.Remove(file.Name()) }, nil
}

func constructMultipartForm(values map[string]io.Reader) (r *http.Request, err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return nil, err
		}
	}

	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest(http.MethodPost, "/", &b)
	if err != nil {
		return nil, err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, err
}
