package main

import (
	"net/http"
	"time"

	"github.com/jseow5177/greenlight/internal/data"
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
		app.logger.Println(err)
		http.Error(w, "The server encountered an error and could not process your request", http.StatusInternalServerError)
		return
	}
}