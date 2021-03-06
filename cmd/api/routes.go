package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance
	// More options on how to customize the application behavior further:
	// https://pkg.go.dev/github.com/julienschmidt/httprouter#Router
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using the 
	// http.HandlerFunc() adapter, and then set it as the custom error handler for 404
	// Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Convert the methodNotAllowed() helper to a http.Handler and set it as a custom error
	// handler for 405 Method Not Allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the relevant methods, URL patterns and handler functions for the endpoints using the HandlerFunc() method.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}