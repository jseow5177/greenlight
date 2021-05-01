package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


func (app *application) routes() *httprouter.Router {
	// Initialize a new httprouter router instance
	// More options on how to customize the application behavior further:
	// https://pkg.go.dev/github.com/julienschmidt/httprouter#Router
	router := httprouter.New()

	// Register the relevant methods, URL patterns and handler functions for the endpoints
	// using the HandlerFunc() method.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	return router
}