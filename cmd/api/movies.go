package main

import (
	"fmt"
	"net/http"
)

// Add a createMovieHandler for "POST /v1/movies"
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create movie..."))
}

// Add a showMovieHandler for "GET /v1/movies/:id"
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read id parameter from request url
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Simple response for now
	fmt.Fprintf(w, "Show the details of the movie %d\n", id)
}