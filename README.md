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
- all valid input types
- `application/x-www-form-urlencoded` and `multipart/form-data` form enctypes
- exclusively non-`GET` forms
- simple server side form validation
- dynamic server side form validation
- forms uploading a single file
- forms uploading multiple files via a single input using the "multiple" tag
- forms uploading a multiple files via multiple inputs
- forms input files using "accept" tag for certain file types
- form post success responses (i.e. a "thank you" page)
- restricting submissions to a given request domain
- protect against: posting massive form body
- protect against: large amounts of files via "multiple" file inputs
- protect against: CSRF
- protect against: Slowloris
- protect against: injection via form names and values

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
- https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html


## Form rules

- when dealing with file uploads, require Content-Type `multipart/form-data`
- only accept non `GET` method on form element

Open Questions:

- allow `multi` tag on input file elements?
    - it's doable but invites more complexity
- what happens when you don't define a form "method" tag? Is there a default?

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

## Input types

[source](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input)

### Email

	<input type="email" id="email" name="email">

### Search XXX

	<input type="search" id="search" name="search">

### Phone number field

	<input type="tel" id="tel" name="tel">

### URL

	<input type="url" id="url" name="url">

### Numeric

	<input type="number" name="age" id="age" min="1" max="10" step="2">
	<input type="number" name="change" id="pennies" min="0" max="1" step="0.01">

### Slider

	<label for="price">Choose a maximum house price: </label>
	<input type="range" name="price" id="price" min="50000" max="500000" step="100" value="250000">
	<output class="price-output" for="price"></output>

### Datetime-local

	<input type="datetime-local" name="datetime" id="datetime">

### Datetime

	[Shouldn't be used](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/datetime)

### Date

	<input type="date" id="start" name="trip-start" value="2018-07-22" min="2018-01-01" max="2018-12-31">

### Month

	<input type="month" name="month" id="month">

### Time

	<input type="time" name="time" id="time">

### Week

	<input type="week" name="week" id="week">

### Color

	<input type="color" name="color" id="color">

### Multiline text

	<textarea cols="30" rows="8"></textarea>

### Drop down

	<select id="simple" name="simple">
	  <option>Banana</option>
	  <option selected>Cherry</option>
	  <option>Lemon</option>
	</select>

with multiple choice

	<select id="multi" name="multi" multiple size="2">
	  <optgroup label="fruits">
	     <option>Banana</option>
	     <option selected>Cherry</option>
	     <option>Lemon</option>
	   </optgroup>
	   <optgroup label="vegetables">
	     <option>Carrot</option>
	     <option>Eggplant</option>
	     <option>Potato</option>
	   </optgroup>
	</select>

### File

	<input type="file" id="avatar" name="avatar" accept="image/png, image/jpeg">


### Password

	<input type="password" id="pass" name="password" minlength="8" required>

### Radio

	<input type="radio" id="contactChoice1" name="contact" value="email" checked>
	<label for="contactChoice1">Email</label>
	<input type="radio" id="contactChoice2" name="contact" value="phone">
	<label for="contactChoice2">Phone</label>
	<input type="radio" id="contactChoice3" name="contact" value="mail">
	<label for="contactChoice3">Mail</label>

### Range

	<input type="range" id="volume" name="volume" min="0" max="11">

### Checkbox

	<input type="checkbox" id="coding" name="interest" value="coding">
	<label for="coding">Coding</label>
	<input type="checkbox" id="music" name="interest" value="music">
	<label for="music">Music</label>
