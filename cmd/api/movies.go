package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jseow5177/greenlight/internal/data"
	"github.com/jseow5177/greenlight/internal/validator"
)

// Add a listMoviesHandler for "GET /V1/movies"
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// Define an input struct to hold the expected values from the request query string
	// This is consistent with other handlers
	var input struct {
		Title string
		Genres []string
		data.Filters
	}

	// Define a new validator instance
	v := validator.New()

	// To get the url.Values map containing the query string data
	qs := r.URL.Query()

	// Extract title from query string value
	// Defaults to empty string
	input.Title = app.readString(qs, "title", "")

	// Extract sort from query string value
	// Defaults to "id", which indicates an ascending sort by movie ID
	input.Filters.Sort = app.readString(qs, "sort", "id")

	// Extract genres from query string value
	// Defaults to empty slice
	input.Genres = app.readCSV(qs, "genres", []string{})

	// Extract page and page_size from query string values as integers
	// page defaults to 1, while page_size defaults to 20
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Add the supported sort values for this endpoint to sort safelist
	input.Filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Execute the validation checks on the Filters struct and send a response
	// containing the errors if necessary
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the movies, passing in the various filter parameters
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movies data
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

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

	// Call the Insert() method on the movies model.
	// This creates a record in the database and updates the movie struct with system-generated info.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Add a Location header to let the client know which URL they can find the newly-created resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Add a showMovieHandler for "GET /v1/movies/:id"
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Read id parameter from request url
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Check if error returned is data.ErrRecordNotFound
	// If yes, return a 404 Not Found response to the client
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// Add a updateMovieHandler for "PUT /v1/movies/:id"
func (app *application) updateMovieHandler (w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie record from the database
	// Send a 404 Not Found response to the client if we couldn't find a matching record
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Declare an input struct to hold the expected data from the client.
	// Fields in struct are pointers. Pointers have a zero-value of nil.
	// This makes it easy to differentiate between zero-value (which returns validation error)
	// and user not providing key/value pair (which 'skips' updating the field).
	var input struct {
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}
 
	// Read the JSON request body data into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Check if the pointers are nil.
	// If nil, the user did not provide any update to the key/value pair.
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres // No need to dereference a slice
	}

	// Validate the updated movie, sending the client a 422 Unprocessable Entity if any checks fail
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass the updated movie to the Update() method
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict): // Intercept conflict in data race
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated movie into JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Add a deleteMovieHandler for "DELETE /v1/movies/:id"
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract movie ID from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with status message
	// Optionally, can send a 204 No Content with an empty response body
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}