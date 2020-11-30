package formhandler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFormContent_JSONEncoded(t *testing.T) {
	var formContentTests = []struct {
		testName               string
		testRequestConstructor func() (req *http.Request, err error)
		expectedValuesOutput   map[string][]string
		expectedError          bool
	}{
		{
			"1 valid string field",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": "value1"}`)
			},
			map[string][]string{"field1": {"value1"}},
			false,
		},
		{
			"2 valid string fields",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": "value1", "field2": "value2"}`)
			},
			map[string][]string{
				"field1": {"value1"},
				"field2": {"value2"},
			},
			false,
		},
		{
			"1 valid string array field",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": ["value1"]}`)
			},
			map[string][]string{"field1": {"value1"}},
			false,
		},
		{
			"2 valid string array fields",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": ["value1"], "field2": ["value2"]}`)
			},
			map[string][]string{
				"field1": {"value1"},
				"field2": {"value2"},
			},
			false,
		},
		{
			"empty string array field",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": []}`)
			},
			nil,
			true,
		},
		{
			"empty string field",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": ""}`)
			},
			nil,
			true,
		},
		{
			"empty string array field with valid string",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": "value1", "field2": []}`)
			},
			nil,
			true,
		},
		{
			"empty string field with valid string array",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": "", "field2": ["value2"]}`)
			},
			nil,
			true,
		},
		{
			"empty object body",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{}`)
			},
			nil,
			true,
		},
		{
			"non json body",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`hello world`)
			},
			nil,
			true,
		},
		{
			"malformed json body",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": value1}`)
			},
			nil,
			true,
		},
		{
			"empty body",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(``)
			},
			nil,
			true,
		},
		{
			"bigger than 1 MB body",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(generateBigJSON())
			},
			nil,
			true,
		},
		{
			"multiple JSON objects",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"1":"1"}{"2":"2"}`)
			},
			nil,
			true,
		},
		{
			"invalid number",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": 1.2}`)
			},
			nil,
			true,
		},
		{
			"invalid null",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": null}`)
			},
			nil,
			true,
		},
		{
			"invalid array contents",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": [1, 1.345, null]}`)
			},
			nil,
			true,
		},
		{
			"invalid child object",
			func() (*http.Request, error) {
				return constructJSONEncodedForm(`{"field1": {"hello": "hi"}}`)
			},
			nil,
			true,
		},
	}

	for _, tt := range formContentTests {
		t.Run(tt.testName, func(t *testing.T) {
			r, err := tt.testRequestConstructor()
			assert.NoError(t, err, "Error constructing test request")

			w := httptest.NewRecorder()
			results, files, err := getFormContent(w, r)

			assert.Equal(t, len(tt.expectedValuesOutput), len(results), "unexpected parsed form results")
			assert.Equal(t, tt.expectedValuesOutput, results, "unexpected parsed form results")
			assert.True(t, (err != nil) == tt.expectedError)

			assert.Empty(t, files, "unexpected file parsed from url encoded form")
		})
	}
}

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

			w := httptest.NewRecorder()
			results, files, err := getFormContent(w, r)

			assert.Equal(t, len(tt.expectedValuesOutput), len(results), "unexpected parsed form results")
			assert.Equal(t, tt.expectedValuesOutput, results, "unexpected parsed form results")

			assert.Empty(t, files, "unexpected file parsed from url encoded form")
		})
	}
}

func TestGetFormContent_Multipart(t *testing.T) {
	var formContentTests = []struct {
		testName               string
		testRequestConstructor func() (req *http.Request, cleanupFunc func(), err error)
		expectedValuesOutput   map[string][]string
		expectedFileCounts     map[string]int
	}{
		{
			"empty form",
			func() (*http.Request, func(), error) {
				values := map[string]io.Reader{}
				req, err := constructMultipartForm(values)

				return req, func() {}, err
			},
			map[string][]string{},
			nil,
		},
		{
			"single field",
			func() (*http.Request, func(), error) {
				values := map[string]io.Reader{
					"field1": strings.NewReader("value1"),
				}
				req, err := constructMultipartForm(values)

				return req, func() {}, err
			},
			map[string][]string{"field1": {"value1"}},
			nil,
		},
		{
			"single field single file",
			func() (*http.Request, func(), error) {
				testFile1, testFile1CleanUp, err := tempTestFile("png")
				values := map[string]io.Reader{
					"file1":  testFile1,
					"field1": strings.NewReader("value1"),
				}
				req, err := constructMultipartForm(values)

				return req, testFile1CleanUp, err
			},
			map[string][]string{"field1": {"value1"}},
			map[string]int{"file1": 1},
		},
		{
			"multiple fields single file",
			func() (*http.Request, func(), error) {
				testFile1, testFile1CleanUp, err := tempTestFile("png")
				values := map[string]io.Reader{
					"file1":  testFile1,
					"field1": strings.NewReader("value1"),
					"field2": strings.NewReader("value2"),
				}
				req, err := constructMultipartForm(values)

				return req, testFile1CleanUp, err
			},
			map[string][]string{"field1": {"value1"}, "field2": {"value2"}},
			map[string]int{"file1": 1},
		},
		{
			"single file",
			func() (*http.Request, func(), error) {
				testFile1, testFile1CleanUp, err := tempTestFile("png")
				values := map[string]io.Reader{
					"file1": testFile1,
				}
				req, err := constructMultipartForm(values)

				return req, testFile1CleanUp, err
			},
			map[string][]string{},
			map[string]int{"file1": 1},
		},
		{
			"multiple file",
			func() (*http.Request, func(), error) {
				testFile1, testFile1CleanUp, err := tempTestFile("png")
				testFile2, testFile2CleanUp, err := tempTestFile("png")
				values := map[string]io.Reader{
					"file1": testFile1,
					"file2": testFile2,
				}
				req, err := constructMultipartForm(values)

				return req, func() {
					testFile1CleanUp()
					testFile2CleanUp()
				}, err
			},
			map[string][]string{},
			map[string]int{"file1": 1, "file2": 1},
		},
		{
			"multiple fields multiple file",
			func() (*http.Request, func(), error) {
				testFile1, testFile1CleanUp, err := tempTestFile("png")
				testFile2, testFile2CleanUp, err := tempTestFile("png")
				values := map[string]io.Reader{
					"file1":  testFile1,
					"file2":  testFile2,
					"field1": strings.NewReader("value1"),
					"field2": strings.NewReader("value2"),
				}
				req, err := constructMultipartForm(values)

				return req, func() {
					testFile1CleanUp()
					testFile2CleanUp()
				}, err
			},
			map[string][]string{"field1": {"value1"}, "field2": {"value2"}},
			map[string]int{"file1": 1, "file2": 1},
		},
	}

	for _, tt := range formContentTests {
		t.Run(tt.testName, func(t *testing.T) {
			r, cleanup, err := tt.testRequestConstructor()
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			results, files, err := getFormContent(w, r)

			assert.Equal(t, tt.expectedValuesOutput, results, "unexpected parsed form results")

			assert.Equal(t, len(files), len(tt.expectedFileCounts), "unexpected files fields present")
			if len(files) > 0 {
				// check the maps match
				for fileName, count := range tt.expectedFileCounts {
					assert.Equal(t, count, len(files[fileName]),
						fmt.Sprintf("Mismatched amount of files parsed for %v", fileName))
				}
			}

			cleanup()
		})
	}
}

func TestInvalidContentType(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)

	r.Header.Set("Content-Type", "application/fake-test-content-type")

	w := httptest.NewRecorder()
	results, files, err := getFormContent(w, r)

	assert.Nil(t, results)
	assert.Nil(t, files)
	assert.NotNil(t, err)
}

func TestMissingContentType(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	results, files, err := getFormContent(w, r)

	assert.Nil(t, results)
	assert.Nil(t, files)
	assert.NotNil(t, err)
}

/*
To test the all combinations multipart form, all of these cases need to be covered:

- no fields
  - no fields

- file fields
  - single file field with [none] files
  - single file field with [one] files
  - single file field with [multiple] files
  - multiple file fields with [none] files
  - multiple file fields with [one] files
  - multiple file fields with [one, none] files
  - multiple file fields with [multiple] files
  - multiple file fields with [multiple, none] files
  - multiple file fields with [multiple, one] files
  - multiple file fields with [multiple, none, one] files

- value fields
  - single value field with [none] value
  - single value field with [one] value
  - single value field with [multiple] value
  - multiple value fields with [none] value
  - multiple value fields with [one] value
  - multiple value fields with [one, none] value
  - multiple value fields with [multiple] value
  - multiple value fields with [multiple, none] value
  - multiple value fields with [multiple, one] value
  - multiple value fields with [multiple, none, one] value

- every combination of file and value fields

*/

func constructJSONEncodedForm(jsonStr string) (*http.Request, error) {
	r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(jsonStr))
	r.Header.Set("Content-Type", "application/json")
	return r, err
}

func constructURLEncodedForm(values url.Values) (*http.Request, error) {
	r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r, err
}

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

	return file, func() { os.Remove(file.Name()) }, nil
}

func constructMultipartForm(values map[string]io.Reader) (r *http.Request, err error) {
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

	// closing the multipart writer adds the terminating boundary
	w.Close()

	req, err := http.NewRequest(http.MethodPost, "/", &b)
	if err != nil {
		return nil, err
	}

	// set the content type, this will contain the boundary
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, err
}

// generates a 1_091_902 byte valid json string
func generateBigJSON() string {
	var sb strings.Builder

	sb.WriteString(`{`)
	for i := 0; i < (megabyte / 8 / 2); i++ {
		sb.WriteString(fmt.Sprintf(`"%v": "%v",`, i, i))
	}
	sb.WriteString(`"a": "a"}`)

	bigJSON := sb.String()
	return bigJSON
}
