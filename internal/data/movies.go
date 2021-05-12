package data

import (
	"database/sql"
	"time"

	"github.com/jseow5177/greenlight/internal/validator"
	"github.com/lib/pq"
)


type Movie struct {
	ID        int64 `json:"id"` // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"` // Timestamp for when the movie is added to our database
	Title 		string `json:"title"` // Movie title
	Year 			int32 `json:"year,omitempty"` // Movie release year
	Runtime		Runtime `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres		[]string `json:"genres,omitempty"` // Slice of genres for the movie (romance, comedy, etc)
	Version 	int32 `json:"version"` // The version number starts at 1 and will be incremented each time the movie info is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) < 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year < int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// Insert() inserts a new record in the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	// The SQL query for inserting a new record in the movies table and returning
	// the system-generated data
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters.
	// Declaring this slice immediately next to our SQL query helps to make it clear *what values are being 
	// used where* in the query.
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use Queryow() method to execute the SQL query passing in the args slice as variadic parameter.
	// Then, scan the system generated id, created_at and version values into the movie struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get() fetches a specific record from the movies table.
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update() updates a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// Delete() deletes a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	return nil
}