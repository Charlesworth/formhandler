# formhandler

## TODO

- make a default handler
- make a handler generator with optional config
- use x/errors
- test url encoded and multipart with no body

## Form requests

### Content-Type

Form post requests will either contain Content-Type `application/x-www-form-urlencoded` or `multipart/form-data`, both should be supported when working with forms.

> if you have binary (non-alphanumeric) data (or a significantly sized payload) to transmit, use multipart/form-data. Otherwise, use application/x-www-form-urlencoded ([source](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type))

It is possile to send files via `application/x-www-form-urlencoded` but it is highly inefficient, resulting in request sizes at least 3x bigger for each non-alphanumeric byte ([source](https://stackoverflow.com/questions/4007969/application-x-www-form-urlencoded-or-multipart-form-data)).
For this reason it is useful to refuse requests with files that use `application/x-www-form-urlencoded`.

You could also use JS to encode your forms into whatever you wanted, i.e. encode to JSON or XML and send that.
This repo is not covering that case.

### Form inputs

All of these HTML input types should be supported, including the cases with support the `multiple` HTML attribute ([source](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input)).

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

To simplify what can be stored on the server, non-empty inputs consist of 4 different types:

- a file: where the file input is used without the `multiple` attribute
- multiple files: where the file input is used with the `multiple` attribute
- a string: any non-file input not using the `multiple` attribute
- an array of strings: any non-file input that supports and uses the `multiple` attribute

For this usecase we can ignore empty fields, they will not be stored and should be removed when parsing the form request.

### HTTP Method

While `GET` and `POST` can be used, I will only accept `method="POST"`, this is easy to handle by limiting the handler to only accept `POST` requests. Delivering form data in the URL could be insecure in some situations so best for my case to avoid entirely.

### Submission responses

For "thank you" pages, just respond with the content in the `POST` response, simple!Redirects are valid too, for when you want to link to another domain:

    w.Header().Set("Location", "https://example.com/")
    w.WriteHeader(302)

## Links

### Useful Go libraries

Value sanitizer:

- <https://github.com/leebenson/conform> Good if you need to sanitise the form data.

Request to struct:

- <https://github.com/gorilla/schema> No file support and requires struct tags.
- <https://github.com/go-playground/form> No file support and requires struct tags.
- <https://github.com/mholt/binding> No struct tags required and supports files via multipart.

Validation:

- <https://github.com/go-playground/validator> Requires struct tags.
- <https://github.com/asaskevich/govalidator> Loads of useful validation functions.
- <https://github.com/go-ozzo/ozzo-validation> Validator can be constructed dynamically, provides nice error messages and uses govalidator under the covers for masses for validation functions.
- <https://github.com/astaxie/beego/tree/develop/validation> Very basic and specific to Beego framework.

CSRF protection middleware

- <https://github.com/gorilla/csrf>
- <https://github.com/justinas/nosurf>

### Go form handling code specifics

- <https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad>
- <https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html>
- <https://ayada.dev/posts/multipart-requests-in-go/>
- <https://www.reddit.com/r/golang/comments/apf6l5/multiple_files_upload_using_gos_standard_library/>

## Vulnerabilities

A couple of vulrabilities to be aware of, the code doesn't contain fixes but I've suggested some.

- posting massive form body
  - limit request size via the http server client
- large amounts of tiny files via `multiple` attribute file inputs
  - No great answer here, setting a sensible limit of files per submission seems to work best
- CSRF
  - use CSRF cookies where you can
- slowloris with the nature large body requests
  - set reasonable request timeouts
- SQL / code injection via form names and values
  - make sure you are checking the content of the form before processing or storing
