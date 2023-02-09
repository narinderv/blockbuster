package main

import (
	"errors"
	"fmt"
	"net/http"

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

	// Insert the record into the database
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Send response containing the header with the location of the movie added
	header := make(http.Header)
	header.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Send the response
	app.writeJsonResponse(w, envelope{"movie": movie}, header, http.StatusCreated)

}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}

		return
	}

	err = app.writeJsonResponse(w, envelope{"movie": movie}, nil, http.StatusOK)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) editMovieHandler(w http.ResponseWriter, r *http.Request) {

	// Get the ID to be updated
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w, r)
		return
	}

	// Get the existing record from the database
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}

		return
	}

	// Now parse the input parameters
	// Structure to hold the request parameters.
	// All members are declared as pointers to check later whether values have been provided by user or not
	var request struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Decode the input json
	err = app.readJsonRequest(w, r, &request)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Copy the input into the Movies structure
	// Copy only values that are porvided in the request
	if request.Title != nil {
		movie.Title = *request.Title
	}
	if request.Year != nil {
		movie.Year = *request.Year
	}

	if request.Runtime != nil {
		movie.Runtime = *request.Runtime
	}

	if request.Genres != nil {
		movie.Genres = request.Genres
	}

	// Validate the input fields
	val := validator.NewValidator()

	data.ValidateMovie(val, movie)

	// Check if there was any failure
	if !val.IsValid() {
		app.failedValidations(w, r, val.Errors)
		return
	}

	// Update the data into the database
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictError(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	// Send the updated record as the response
	err = app.writeJsonResponse(w, envelope{"movie": movie}, nil, http.StatusOK)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {

	// Get the ID to be updated
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w, r)
		return
	}

	// Delete the record
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}

		return
	}

	err = app.writeJsonResponse(w, envelope{"message": "movie successfully deleted"}, nil, http.StatusOK)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {

	// Define a structure to hold the query string values
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	// initalize a Validator for tracking any errors
	val := validator.NewValidator()

	//get the query string from the request
	queryString := r.URL.Query()

	// Get the query string values
	input.Title = app.readString(queryString, "title", "")
	input.Genres = app.readCSV(queryString, "genres", []string{})

	input.Filters.Page = app.readInt(queryString, "page", 1, val)
	input.Filters.PageSize = app.readInt(queryString, "page_size", 10, val)

	input.Filters.Sort = app.readString(queryString, "sort", "id")

	// Supported sort values
	input.Filters.SortList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Validate the provided filter values
	data.ValidateFilters(val, &input.Filters)

	// Check if app values were read without any error
	if !val.IsValid() {
		app.failedValidations(w, r, val.Errors)
		return
	}

	// Get all the data from the database
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Send response
	app.writeJsonResponse(w, envelope{"metadata": metadata, "movies": movies}, nil, http.StatusOK)
	if err != nil {
		app.serverError(w, r, err)
	}
}
