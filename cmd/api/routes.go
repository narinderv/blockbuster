package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	// Using httprouter. Limitation: Does NOT support confilicting routes
	// e.g. /foo/bar and /foo/:id will not be supported.
	// If such routes are required, we can use 'pat' instead
	router := httprouter.New()

	// Custom error handlers
	router.NotFound = http.HandlerFunc(app.notFound)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	// Route handlers
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)

	// Add a new movie
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)

	// View details of a particular movie
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// View list of Movies
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)

	// Using the PATCH method for partial update of a record
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.editMovieHandler)
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.editMovieHandler)

	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	return router
}
