package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jseow5177/greenlight/internal/validator"
	"github.com/julienschmidt/httprouter"
)

const RequestBodyTooLargeMessage = "http: request body too large"

// Define an envelope type
type envelope map[string]interface{}

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it.
// The integer is base10 with type int (machine-dependent bit size).
// If the conversion is not successful, return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// Define a writeJSON() helper for sending JSON responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON, return error if any.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// Loop through the header map and add each header to the http.ResponseWriter header map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header.
	w.Header().Set("Content-Type", "application/json")
	// Write status code.
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// readJSON() helper reads JSON data in the request body into a destination dst.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.MaxBytesReader() to limit the size of request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, and call the DisallowUnknownFields() method on it before
	// decoding. This means that if the JSON from the client includes any field that cannot
	// be mapped to the target destination, the decoder will return an error instead of just
	// ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// Catch syntax error with JSON being decoded.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some cases, Decode() may return an io.ErrUnexpectedEOF error for syntax errors in JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains-badly formed JSON")

		// A json.UnmarshalTypeError error is returned when the JSON value is the wrong type for the target destination.
		// If the error relates to a specific field, we include that in the error message.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// A json.InvalidUnmarshalError error will be returned if we pass a non-nil pointer to Decode().
		// This is a problem with the application code, not the JSON itself.
		// We catch this error and panic.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// An io.EOF error is returned if the request body is empty.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field that cannot be mapped to the target destination then Decode()
		// will now return an error message in the format "json: unknown field <name>".
		// We check for this, extract the field name from the error, and interpolate it into our custom error message.
		// There is an open issue to turn this into a distinct error type: https://github.com/golang/go/issues/29035
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the request body exceeds 1MB in size, the decode will now fail with the error
		// "http: request body too large". Currently, the error checking is done with string comparison.
		// There is an open issue to turn it into a distinct error type: https://github.com/golang/go/issues/30715.
		case err.Error() == RequestBodyTooLargeMessage:
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct as the destination.
	// It should return an io.EOF error if the request body contains a single JSON value.
	// If there is anything else, there must be additional data (another JSON, dirty value, etc) in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain a single JSON value")
	}

	return nil
}

// readString() helper returns a string value from the query string, or the provided default value if no matching
// key could be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string
	// If no key exists, this returns an empty string ""
	s := qs.Get(key)

	// If no key exists (or the value is empty), return the default string value
	if s == "" {
		return defaultValue
	}

	// Else, return the extracted value
	return s
}

// readCSV() helper reads a comma-separated string value from the query string and then splits it
// into a slice on the comma character. If no matching key is found, it returns the provided default value
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// Extract the value from query string
	csv := qs.Get(key)

	// If no key exists (or the value is empty), return the default string value
	if csv == "" {
		return defaultValue
	}

	// Else, parse the value into a []string slice and return it
	return strings.Split(csv, ",")
}

// readInt() helper reads a string value from the query string and converts it to an integer before returning.
// If no matching key is found, it returns the provided default value.
// If the value could not be converted to an integer, then we record the error message in the provided Validator instance.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// Extract the value from query string
	s := qs.Get(key)

	// If no key exists (or the value is empty), return the default string value
	if s == "" {
		return defaultValue
	}

	// Try to convert the value to an int.
	// If it fails, add an error message to the validator instance and return the default value
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	// Otherwise, return the converted integer value
	return i
}

// runBackground() accepts and executes an arbitrary function in a new goroutine.
// It catches and logs any error as a result of a panic.
func (app *application) runBackground(fn func()) {
	// Increment the WaitGroup counter
	app.wg.Add(1)

	go func() {
		// Use defer to decrement the WaitGroup
		defer app.wg.Done()

		// Recover from any panic in background routine else will crash application in panic.
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// Executes the function in a background routine.
		fn()
	}()
}
