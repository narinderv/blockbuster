package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32 // Custom type for returning runtime as "<time> mins"

// Implement the interface MarshalJSON for Runtime type so that
// the json value can be sent as per our requirement above
func (r Runtime) MarshalJSON() ([]byte, error) {

	jsonValue := fmt.Sprintf("%d mins", r)

	// Need to enclose the value in quotes to ensure it is returned as a valid json string value
	valWithQuotes := strconv.Quote(jsonValue)

	return []byte(valWithQuotes), nil
}

// Implement the interface UnMarshalJSON for Runtime type so that
// the json value can be received from user as "<duration> mins" and then stored in our structure as int32
func (r *Runtime) UnmarshalJSON(request []byte) error {

	// request is received within quotes as "<duration> mins"
	// First try to unquote this to get the actual data
	unquoted, err := strconv.Unquote(string(request))
	if err != nil {
		return ErrInvalRuntimeFormat
	}

	// Now split the string into parts, breaking at " "
	parts := strings.Split(unquoted, " ")

	// Check if we have expected number of parts (2) and second part is 'mins'
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalRuntimeFormat
	}

	// Try to convert the duration into integer (int32)
	duration, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalRuntimeFormat
	}

	// Store the value back into the runtime variable
	*r = Runtime(duration)

	return nil
}
