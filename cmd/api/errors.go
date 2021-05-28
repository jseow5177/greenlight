package main

import (
	"fmt"
	"net/http"
)

// logError() is a generic helper for logging an error message.
// Will be upgraded later to include structured logging.
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"method": r.Method,
		"url": r.URL.String(),
	})
}

// errorResponse() is a generic helper for sending JSON-formatted error
// messages to the client with a given status code. 
// The message has an interface{} type instead of string type to give more flexibility
// over the values that can be included in the response.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	// Write the response using the writeJSON() helper.
	// If an error is returned, an empty response is sent with a 500 Server Internal Error status code.
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorResponse() is used when the application encounters unexpected problem at runtime.
// It logs the detailed error message and uses the errorResponse() helper to send a 500 status code and JSON
// response to the client.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// notFoundResponse() is used to send a 404 Not Found status code and JSON response to the client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// methodNotAllowedResponse() is used to send a 405 Method Not Allowed status code and JSON response to the client.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// badRequestResponse() is used to send a 400 Bad Request status code and JSON response to the client.
// Deals with syntatic errors
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse() is used to send a 422 Unprocessable Entity status code and JSON response to the client
// Deals with semantic errors
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// editConflictResponse is used to send a 409 Conflict status code and JSON response to the client
// Used to handle conflict in race condition
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}