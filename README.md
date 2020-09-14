# skunkworksForms: a form learning and experimentation repo

I want to play around with hadling form submissions in Go, including:

- different request body content types
- single and multi file uploads
- dynamic form validation (no struct tags)
- ingnoring values stored in the URL query string
- protection against common attack vectors

This should help me form the basis around a good design for handling form submissions.

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

## Form requests

### Content-Type

Form post requests will either contain Content-Type "application/x-www-form-urlencoded" or "multipart/form-data", both need to be supported.

> if you have binary (non-alphanumeric) data (or a significantly sized payload) to transmit, use multipart/form-data. Otherwise, use application/x-www-form-urlencoded

You can also use JS to encode your forms into whatever you wanted, i.e. encode to JSON or XML and send that.

Sources:

- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type
- https://stackoverflow.com/questions/4007969/application-x-www-form-urlencoded-or-multipart-form-data

## Libraries

good general resources:
- https://awesome-go.com/#forms

Sanitizer (probs not required)
- https://github.com/leebenson/conform

Request to struct:
- https://github.com/gorilla/schema
- https://github.com/go-playground/form
- https://github.com/mholt/binding

validation:
- https://github.com/go-playground/validator
- https://github.com/asaskevich/govalidator
- https://github.com/go-ozzo/ozzo-validation

CSRF protection middleware
- https://github.com/gorilla/csrf
- https://github.com/justinas/nosurf
