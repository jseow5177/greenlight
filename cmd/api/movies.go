package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jseow5177/greenlight/internal/data"
	"github.com/jseow5177/greenlight/internal/validator"
)

// Add a createMovieHandler for "POST /v1/movies"
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the HTTP request body.
	// This struct will be our *target decode destination*.
	// The struct fields must start with a capital letter so that they are exported.
	var input struct {
		Title string `json:"title"`
		Year int32 `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}

	// Initialize a new json.Decoder instance which reads from the request body.
	// Use the Decode() method to decode the body contents into the input struct.
	// A pointer to the input struct is passed into Decode().
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to the new Movie struct
	movie := &data.Movie{
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

// Add a showMovieHandler for "GET /v1/movies/:id"
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read id parameter from request url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie := &data.Movie{
		ID: id,
		CreatedAt: time.Now(),
		Title: "Casablance",
		Runtime: 102,
		Genres: []string{"drama", "romance", "war"},
		Version: 1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}