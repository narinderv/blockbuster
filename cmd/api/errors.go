package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {

	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})

}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {

	resp := envelope{
		"error": message,
	}

	err := app.writeJsonResponse(w, resp, nil, status)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {

	app.logError(r, err)

	msg := "the server encountered an internal error and could not process your request."

	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {

	msg := "the requested resource could not be found"

	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {

	msg := fmt.Sprintf("the %s method is not supported for this request.", r.Method)

	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {

	app.logError(r, err)

	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidations(w http.ResponseWriter, r *http.Request, errors map[string]string) {

	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictError(w http.ResponseWriter, r *http.Request) {

	message := "unable to edit record due to an edit conflict. please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *application) tpsExceedResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}
