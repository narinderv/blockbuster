package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {

	// Get the parameters from the request URL.
	// Fur httprouter, the parameters are obtained from the request context
	params := httprouter.ParamsFromContext(r.Context())
	if params == nil {
		return 0, errors.New("ID parameter not found")
	}

	// Parameter values can be retrieved by using the names of the parameters.
	// This will always return string values and need to be converted to appropriate types before use.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid ID parameter")
	}

	return id, nil
}
