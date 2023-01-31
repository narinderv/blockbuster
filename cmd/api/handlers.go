package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/narinderv/blockbuster/internal/data"
	"github.com/narinderv/blockbuster/internal/validator"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	// Create a map of the response
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	// Now write out the response
	err := app.writeJsonResponse(w, data, nil, http.StatusOK)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	// Structure to hold the request parameters
	var request struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Decode the input json
	err := app.readJsonRequest(w, r, &request)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Copy the input into the Movies structure
	movie := &data.Movies{
		Title:   request.Title,
		Year:    request.Year,
		Runtime: request.Runtime,
		Genres:  request.Genres,
	}

	// Validate the input fields
	val := validator.NewValidator()

	data.ValidateMovie(val, movie)

	// Check if there was any failure
	if !val.IsValid() {
		app.failedValidations(w, r, val.Errors)
		return
	}

	fmt.Fprintf(w, "New movie created\n%+v\n", request)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w, r)
		return
	}

	movie := &data.Movies{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Pathaan",
		Runtime:   135,
		Genres:    []string{"Action", "Thriller", "Romance"},
		Version:   1,
	}

	err = app.writeJsonResponse(w, envelope{"movie": movie}, nil, http.StatusOK)
	if err != nil {
		app.serverError(w, r, err)
	}
}
