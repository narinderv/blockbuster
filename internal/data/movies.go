package data

import (
	"time"

	"github.com/narinderv/blockbuster/internal/validator"
)

// Structure with annotations which will be used for json naming
type Movies struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // The "-" will always hide this field from the output json
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"` // Omitempty will hide the field if it is empty or blank
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genre,omitempty"`
	Version   int32     `json:"info_version"`
}

func ValidateMovie(val *validator.Validator, movie *Movies) {

	// Title
	val.Check(movie.Title != "", "title", "must be provided")
	val.Check(len(movie.Title) <= validator.MAX_LEN, "title", "must not be more than 500 bytes")

	// Year
	val.Check(movie.Year != 0, "year", "must be provided")
	val.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	val.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// Runtime
	val.Check(movie.Runtime != 0, "runtime", "must be provided")
	val.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	// Genres
	val.Check(len(movie.Genres) != 0, "genres", "must be provided")
	val.Check(len(movie.Genres) >= 1, "genres", "must contain atleast 1 genre")
	val.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	val.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}
