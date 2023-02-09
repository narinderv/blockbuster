package validator

import (
	"regexp"
)

const MAX_LEN = 500
const MAX_PAGE_LEN = 10_000_000

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator structure to hold map of errors
type Validator struct {
	Errors map[string]string
}

// Initialize a new Validator instance with empty error map
func NewValidator() *Validator {
	return &Validator{
		Errors: map[string]string{},
	}
}

// Returns true if no errors are there in the error map
func (val *Validator) IsValid() bool {
	return len(val.Errors) == 0
}

// Add an error to the error map, if it is not already present
func (val *Validator) AddError(err, message string) {

	if _, exists := val.Errors[err]; !exists {
		val.Errors[err] = message
	}
}

// Add an error message if validation check is not ok
func (val *Validator) Check(ok bool, err, message string) {
	if !ok {
		val.AddError(err, message)
	}
}

// Check if values are among permitted values
func Permittedvalues(value string, validVals ...string) bool {

	for _, validVal := range validVals {
		if value == validVal {
			return true
		}
	}

	return false
}

// Check if value matches the defined regex
func MatchPattern(field string, pattern *regexp.Regexp) bool {

	return pattern.MatchString(field)
}

// Check if all strings in the slice are unique
func Unique(values []string) bool {

	uniqueVals := make(map[string]bool)

	// Create a map from the list of values in the slice.
	// A map always stores unique values.
	for _, value := range values {
		uniqueVals[value] = true
	}

	return len(uniqueVals) == len(values)
}

/*
// Check Required fields
func (form *FormInfo) Required(fields ...string) {
	for _, field := range fields {
		value := form.Get(field)
		if strings.TrimSpace(value) == "" {
			form.Errors.Add(field, "This field cannot be empty")
		}
	}
}

// Check Max length
func (form *FormInfo) MaxLength(field string, maxLen int) {
	value := form.Get(field)

	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) > maxLen {
		form.Errors.Add(field, fmt.Sprintf("This field is too long (max. length: %d)", maxLen))
	}
}

// Check Max length
func (form *FormInfo) MinLength(field string, minLen int) {
	value := form.Get(field)

	if value == "" {
		return
	}

	if utf8.RuneCountInString(value) < minLen {
		form.Errors.Add(field, fmt.Sprintf("This field is too short (min. length: %d)", minLen))
	}
}
*/
