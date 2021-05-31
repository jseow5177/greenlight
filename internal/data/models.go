package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found") // To deal with missing record
	ErrEditConflict = errors.New("edit conflict") // To deal with race condition
)

// Create a Models struct that wraps all database models of this application.
type Models struct {
	Movies MovieModel
	Users UserModel
}

// The New() method returns a newly initialized Models struct 
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users: UserModel{DB: db},
	}
}