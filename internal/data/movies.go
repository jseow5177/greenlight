package data

import (
	"context"
	"database/sql"
	"errors"
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
	// The bigserial type of primary key id is always positive (starts from 1)
	// Do an early check on negative integers to avoid unnecessary database calls
	if id < 1 {
		return nil, ErrRecordNotFound
	}


	// Declare the SQL query for retrieving a movie from the database
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1
	`

	// Declare a pointer to the Movie struct to hold the data returned by the query
	movie := new(Movie)

	// Use context.WithTimeout() function to create a context.Context which carries a 
	// 3-second timeout deadline. We use the empty context.Background() as the 'parent' context.
	// The countdown begins from the moment the context is created. Any time spent executing code between
	// creating the context and calling QueryRowContext() will count towards the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Make sure we cancel the context before the Get() method returns.
	// Calling the CancelFunc cancels ctx and its children, removes the parent's reference to ctx, and stops any associated timers.
	// This is important to prevent memory leak of childrent contexts.
	// Without it, the resources won't be released until either the timeout is hit or the parent context (Background) is canceled.
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&[]byte{},
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres), // Use pq.Array adapter to handler text[] array
		&movie.Version,
	)

	if err != nil {
		switch {
		// If no matching movie found, Scan() returns a sql.ErrNoRows error
		case errors.Is(err, sql.ErrNoRows):
				return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return movie, nil
}

// Update() updates a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {

	// Declare the SQL query for updating the record and returning the new version number
	// Filter by version to implement optimistic concurrency control
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`

	// Create an args slice containing the value for the placeholder parameters
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use QueryRowContext() to execute the query, passing in the args slice as a variadic parameter
	// Scan the new version value into the movie struct
	// If an error is returned, we check if it is ErrNoRows. If it is, this means that the movie version
	// has been changed (or the record is already deleted)
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}	
	}

	return nil
}

// Delete() deletes a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	// Declare the SQL query to delete the record
	query := `
		DELETE FROM movies
		WHERE id = $1
	`

	// Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query
	// This returns a sql.Request object and an error (if any)
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Get the numbr of rows affected by the query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	// If no rows affected, we know the movie does not exist
	// In that case, an ErrRecordNotFound is returned
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}