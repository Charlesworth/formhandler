# skunkworksForms: example code for handling HTML forms in Go

Playing around exaustively with accepting forms via a Go web server for another project.
The code should handle:

- all valid form input types
- `application/x-www-form-urlencoded` and `multipart/form-data` form enctypes
- exclusively non-`GET` forms
- simple server side form validation
- dynamic server side form validation
- forms uploading a single file
- forms uploading multiple files via a single input using the `multiple` tag
- forms uploading multiple files via multiple inputs
- form post success responses (i.e. a "thank you" page)
- restricting submissions to a given request domain
- protect against: posting massive form body
- protect against: large amounts of files via `multiple` tag file inputs
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

## Form HTML inputs

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

## Open Questions

- what happens when you don't define a form "method" tag? Is there a default?

## Links

### Go handling forms

- <https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad>
- <https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html>

## Useful Go libraries

Value sanitizer:

- <https://github.com/leebenson/conform>

Request to struct:

- <https://github.com/gorilla/schema>
- <https://github.com/go-playground/form>
- <https://github.com/mholt/binding>

Validation:

- <https://github.com/go-playground/validator>
- <https://github.com/asaskevich/govalidator>
- <https://github.com/go-ozzo/ozzo-validation>
- <https://github.com/astaxie/beego/tree/develop/validation>

CSRF protection middleware

- <https://github.com/gorilla/csrf>
- <https://github.com/justinas/nosurf>
