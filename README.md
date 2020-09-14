# skunkworksForms: a form learning and experimentation repo

I want to play around with hadling form submissions in Go, including:

- different request body content types
- single and multi file uploads
- dynamic form validation (no struct tags)
- ignoring values stored in the URL query string
- protection against common attack vectors

This should help me form the basis around a good design for safely handling form submissions.

## Goals

Example code for:
- multipart forms
- excluding query string values when parsing forms
- simple server side form validation
- dynamic server side form validation
- forms uploading a single file
- forms uploading a multiple files
- forms input files using "accept" tag for certain file types
- forms input files using "multiple" tag
- form post success responses
- form post success thank you page
- restricting submissions to a given request domain
- Protect against: posting massive form body
- Protect against: CSRF
- Protect against: Slowloris
- Protect against: injection via form names and values

## Form requests

### Content-Type

Form post requests will either contain Content-Type `application/x-www-form-urlencoded` or `multipart/form-data`, both need to be supported.

> if you have binary (non-alphanumeric) data (or a significantly sized payload) to transmit, use multipart/form-data. Otherwise, use application/x-www-form-urlencoded

You can also use JS to encode your forms into whatever you wanted, i.e. encode to JSON or XML and send that.

When handling forms, it is possile to send files via `application/x-www-form-urlencoded` but it is highly inefficient, resulting in request sizes at least 3x bigger for each non-alphanumeric byte. For this reason it may be useful to simply refuse requests with files that use `application/x-www-form-urlencoded`

Sources:

- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
- https://stackoverflow.com/questions/4007969/application-x-www-form-urlencoded-or-multipart-form-data

## Go handling forms

- https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad


## Libraries

good general resources:
- https://awesome-go.com/#forms

Sanitizer (probs not required):
- https://github.com/leebenson/conform

Request to struct:
- https://github.com/gorilla/schema
- https://github.com/go-playground/form
- https://github.com/mholt/binding

validation:
- https://github.com/go-playground/validator
- https://github.com/asaskevich/govalidator
- https://github.com/go-ozzo/ozzo-validation
- https://github.com/astaxie/beego/tree/develop/validation

CSRF protection middleware
- https://github.com/gorilla/csrf
- https://github.com/justinas/nosurf
