package main

import (
	"errors"
	"net/http"

	"github.com/narinderv/blockbuster/internal/data"
	"github.com/narinderv/blockbuster/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	// Request structure
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Read the user request
	err := app.readJsonRequest(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Create the user structure
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Generate the hashed password
	err = user.Password.SetPasswordHash(input.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Validate the user
	val := validator.NewValidator()
	data.ValidateUser(val, user)

	if !val.IsValid() {
		app.failedValidations(w, r, val.Errors)
		return
	}

	// Insert the user into the dataabse
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			val.AddError("email", "a user with this email already exists")
			app.failedValidations(w, r, val.Errors)
		default:
			app.serverError(w, r, err)
		}

		return
	}

	// Send response
	err = app.writeJsonResponse(w, envelope{"user": user}, nil, http.StatusCreated)
	if err != nil {
		app.serverError(w, r, err)
	}
}
