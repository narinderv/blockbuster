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

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	/*
		router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
		router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.editMovieHandler)
		router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	*/

	return router
}