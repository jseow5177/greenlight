package main

import (
	"errors"
	"net/http"

	"github.com/jseow5177/greenlight/internal/data"
	"github.com/jseow5177/greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an annonymous struct to hold the expected data from the request body
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data into a new user struct
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Use the Set method to generate and store the hash and plaintext passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

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

	// Launch a goroutine which runs an annonymous function that sends a welcome email
	app.runBackground(func() {
		// Call the Send() method on our Mailer, passing in the user's email address,
		// name of the template file, and the User struct containing the new user's data.
		err = app.mailer.Send(user.Email, "user_welcome.html", user)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// Write a JSON response containing the newly added user.
	// Notice that we send the client a 202 Accepted status code.
	// This indicates that the request has been accepted but still processing.
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
