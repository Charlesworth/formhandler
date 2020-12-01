# formhandler

A http handler that given a form request, parses and outputs the form content and files.

```language: go
// GetFormContent accepts a request of content type "application/x-www-form-urlencoded",
// "application/json" or "multipart/form-data", parses the body and returns the form data
// and files contained in the request
func GetFormContent(
 w http.ResponseWriter,
 r *http.Request,
) (
 results map[string][]string,
 files map[string][]*multipart.FileHeader,
 err error,
)

// GetFormContentWithConfig operates the same as GetFormContent but with added config options:
// - maxFormSize: The maximum size in bytes a form request can be (applies to JSON and URL encoded forms, which cannot have files attached)
// - maxFormWithFilesSize: The maximum size in bytes a form request with attached files can be (applies to multipart/form-data encoded forms, which can have files attached)
// - maxMemory: Given a form request body is parsed, maxMemory bytes of its file parts are stored in memory, with the remainder stored on disk in temporary files (applies to multipart/form-data encoded forms, which can have files attached)
func GetFormContentWithConfig(
 maxFormSize int64,
 maxFormWithFilesSize int64,
 maxMemory int64,
) func(w http.ResponseWriter, r *http.Request) (results map[string][]string, files map[string][]*multipart.FileHeader, err error)
```

## Form requests

### Accepted HTTP Content-Type

Requests must have a `Content-Type` of:

- `multipart/form-data`
- `application/json`
- `application/x-www-form-urlencoded`

Only `multipart/form-data` supports file uploads:

> if you have binary (non-alphanumeric) data (or a significantly sized payload) to transmit, use multipart/form-data. Otherwise, use application/x-www-form-urlencoded ([source](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type))

It is possible to send files via `application/x-www-form-urlencoded` and `application/json` but it is highly inefficient, resulting in request sizes at least 3x bigger for each non-alphanumeric byte ([source](https://stackoverflow.com/questions/4007969/application-x-www-form-urlencoded-or-multipart-form-data)). formhandler doesn't support this.

### Accepted HTTP Methods

formhandler does not check the request type, only the `Content-Type`, checking methods is left up to the user.

### Form input

All HTML input types are supported, including the cases with support the `multiple` HTML attribute ([source](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input)).

Input types:

- button
- checkbox
- color
- date
- datetime-local
- email
- file
- image
- month
- number
- password
- radio
- range
- search
- tel
- text
- time
- url
- week

Non inputs types:

- Multi-line text fields with `textarea`
- Drop-down select with `select`

The handler removes fields with no values set when parsing the form request.

### JSON input

formhandler also accepts JSON input via the `application/json` content type. It is very strict with the JSON it accepts, to try and conform the JSON structure to match what a traditional `multipart/form-data` or `application/x-www-form-urlencoded` uploads.

The JSON object can only have value types of `string` and `array<string>`. Any JSON values with `number`, `null`, `object` or `arrays<number | null | object>` types will be rejected.

Valid JSON:

```language: json
{"name": "charlie", "number_choices": ["1", "4"]}
```

Invalid JSON:

```language: json
{"name": 123, "number_choices": [1, 1.2, null], "empty": null}
```

## HTTP Security

The handler protects against users posting massive request bodies by default and also with configurable request body size and usable memory limits.
Precautions should be taken by the user:

- With server HTTP timeouts being set on the server ([useful source](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/))
- Form content is not sanitized at all, if storing or serving form content make sure to use SQL, HTML or any other applicable sanitization technique
- CSRF where the form backend on a separate domain from the web backend

## Stuff that could be added

- Additional tests (described in the test file)
- Sentinel errors
- Add a JSON Schema to describe the object accepted
