package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

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

// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a 
// header map containing any additional HTTP headers we want to include in the response.
func (app * application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON, return error if any.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// Loop through the header map and add each header to the http.ResponseWriter header map.
	for key, value := range(headers) {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header.
	w.Header().Set("Content-Type", "application/json")
	// Write status code.
	w.WriteHeader(status)
	w.Write(js)

	return nil
}