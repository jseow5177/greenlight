package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jseow5177/greenlight/internal/data"
	"github.com/jseow5177/greenlight/internal/validator"
)


func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an annonymous struct to hold the expected data from the request body
	var input struct {
		Name string `json:"name"`
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data into a new user struct
	user := &data.User{
		Name: input.Name,
		Email: input.Email,
		Activated: false,
	}

	// Use the Set method to generate and store the hash and plaintext passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	fmt.Println(user)
	
	v := validator.New()

	// Validate user information
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write a JSON response containing the newly added user
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}