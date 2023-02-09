package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/narinderv/blockbuster/internal/validator"
)

// Constants for handling incoming requests
const maxBytes int64 = 1_048_576
const reqBodyTooLarge = "http: request body too large"
const unknownField = "json: unknown field "

type envelope map[string]interface{}

func (app *application) writeJsonResponse(w http.ResponseWriter, data envelope, headers http.Header, status int) error {

	// Marshal the input data into a JSON byte array
	// respJson, err := json.Marshal(data)

	// For pretty, formatted output we can use teh below function, but this is generally little slower and
	// larger than the response from json.Marshal
	respJson, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Set all the passed in headers
	for head, val := range headers {
		w.Header()[head] = val
	}

	// Set content type header as json
	w.Header().Set("Content-Type", "application/json")

	// set theinput request status
	w.WriteHeader(status)

	// Send the response
	w.Write(respJson)

	return nil
}

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

func (app *application) readJsonRequest(w http.ResponseWriter, r *http.Request, data interface{}) error {

	// Limit the reading of body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	// Initialize the Decoder
	dec := json.NewDecoder(r.Body)

	// Explicitily set the decoder to read only known fields and raise error in case of unknown fields
	dec.DisallowUnknownFields()

	// Try to read the json request parameters
	err := dec.Decode(&data)
	if err != nil {
		// Check the different types of errors that can occur and respond accordingly
		var syntaxErr *json.SyntaxError
		var unMarshalTypeErr *json.UnmarshalTypeError
		var invalUnmarshalErr *json.InvalidUnmarshalError

		switch {
		// Syntax error
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("request contains badly formed JSON (at character %d)", syntaxErr.Offset)

		// Incomplete or badly formed JSON
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")

		// Wrong JSON field value for a type
		case errors.As(err, &unMarshalTypeErr):
			if unMarshalTypeErr.Field != "" {
				return fmt.Errorf("request contains incorrect JSON type for field: %q", unMarshalTypeErr.Field)
			}

			return fmt.Errorf("request contains badly incorrect JSON type (at character %d)", unMarshalTypeErr.Offset)

		// Empty body
		case errors.Is(err, io.EOF):
			return errors.New("body cannot be empty")

			// Unknown field in the request JSON
		case strings.HasPrefix(err.Error(), unknownField):
			field := strings.TrimPrefix(err.Error(), unknownField)
			return fmt.Errorf("body contains unknown key: %q", field)

		// Check if body size exceeds allowed 1MB limit
		case err.Error() == reqBodyTooLarge:
			return fmt.Errorf("body must not be larger than %d bytes,", maxBytes)

		// Error at server end where invalid parameter is passed to the Decode() function
		case errors.As(err, &invalUnmarshalErr):
			panic(err)

		default:
			return err
		}
	}

	// Check if the body contains more than 1 JSON. We will permit only 1 JSON per request
	// The Decode() function only processes 1 JSON request and ignores any following data
	// in the request. However, we will flag this condition as error instead of ignoring it.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain only single JSON request")
	}

	return nil
}

func (app *application) readString(queryString url.Values, key string, defString string) string {

	// Get the value corresponding to the given key
	val := queryString.Get(key)

	// If key is not present, return the default value
	if val == "" {
		return defString
	}

	return val
}

func (app *application) readInt(queryString url.Values, key string, defValue int, val *validator.Validator) int {

	// Get the value corresponding to the given key
	v := queryString.Get(key)

	// If key is not present, return the default value
	if v == "" {
		return defValue
	}

	// Key is present, try to convert the value into an Integer
	// In case of failure, add error into the error map and return the default value
	value, err := strconv.Atoi(v)
	if err != nil {
		val.AddError(key, "value must be an integer")
		return defValue
	}
	return value
}

func (app *application) readCSV(queryString url.Values, key string, defVal []string) []string {

	// Get the value corresponding to the given key
	v := queryString.Get(key)

	// If key is not present, return the default value
	if v == "" {
		return defVal
	}

	//Key is present. Split the value at "," to get a slice of strings and return it
	return strings.Split(v, ",")
}
