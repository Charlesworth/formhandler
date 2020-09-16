# skunkworksForms: example code for handling HTML forms in Go

Playing around exhaustively with accepting forms via a Go web server for another project.
The code should handle:

- all valid form input types
- `application/x-www-form-urlencoded` and `multipart/form-data` form enctypes
- exclusively non-`GET` forms
- forms uploading a single file
- forms uploading multiple files via a single input using the `multiple` attribute
- forms uploading multiple files via multiple inputs
- form post success responses (i.e. a "thank you" page)
- simple server side form validation
- dynamic server side form validation
- restricting submissions to a given request domain
- protect against: posting massive form body
- protect against: large amounts of files via `multiple` attribute file inputs
- protect against: CSRF
- protect against: slowloris
- protect against: injection via form names and values

## Form HTTP requests

### Content-Type

Form post requests will either contain Content-Type `application/x-www-form-urlencoded` or `multipart/form-data`, both need to be supported.

> if you have binary (non-alphanumeric) data (or a significantly sized payload) to transmit, use multipart/form-data. Otherwise, use application/x-www-form-urlencoded ([source](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type))

When handling forms, it is possile to send files via `application/x-www-form-urlencoded` but it is highly inefficient, resulting in request sizes at least 3x bigger for each non-alphanumeric byte ([source](https://stackoverflow.com/questions/4007969/application-x-www-form-urlencoded-or-multipart-form-data)).
For this reason it is useful to simply refuse requests with files that use `application/x-www-form-urlencoded`.

You could also use JS to encode your forms into whatever you wanted, i.e. encode to JSON or XML and send that.
This repo is not covering those use cases.

### Form HTML inputs

All of these input types should be supported, including the cases with support the `multiple` HTML attribute ([source](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input)).

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

### Method

What happens when you don't define a form "method" attribute? Is there a default?

## Links

### Go form handling code specifics

- <https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad>
- <https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html>
- <https://ayada.dev/posts/multipart-requests-in-go/>
- <https://www.reddit.com/r/golang/comments/apf6l5/multiple_files_upload_using_gos_standard_library/>

## Useful Go libraries

Value sanitizer:

- <https://github.com/leebenson/conform> Don't need this, it's the form validation responsibility

Request to struct:

- <https://github.com/gorilla/schema> X no file support, requires structs
- <https://github.com/go-playground/form> X no file support, requires structs
- <https://github.com/mholt/binding> This could work for `application/x-www-form-urlencoded` although not convinced for multipart

Validation:

- <https://github.com/go-playground/validator> X requires a struct
- <https://github.com/asaskevich/govalidator> Calling the validation functions could be useful
- <https://github.com/go-ozzo/ozzo-validation> Validator can be constructed dynamically, provides a nice error message and uses govalidator under the covers
- <https://github.com/astaxie/beego/tree/develop/validation> Very basic and specific to Beego framework

CSRF protection middleware

- <https://github.com/gorilla/csrf>
- <https://github.com/justinas/nosurf>
