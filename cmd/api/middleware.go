package main

import (
	"fmt"
	"net/http"
)

// recoverPanic() is a middleware that recovers a panic in a go routine to
// return a 500 Internal Server Error response to the client
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (will be run as Go unwinds the stack in the go routine)
		// The function then checks if a panic has occured using the recover() function
		defer func() {
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				// This acts as a trigger to make Go's HTTP server automatically close the current connection
				// after a response has been sent.
				w.Header().Set("Connection", "close")
				// Use fmt.Errorf() to normalize err into an error and call the serverErrorResponse() helper.
				// This will log the error at the ERROR level and send the client a 500 Internal Server Error.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}