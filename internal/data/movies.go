package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
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

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movies) error {

	// Insert query
	query := "INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version"

	// Argumets to the query
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	// Execute the query and store the result
	return m.DB.QueryRowContext(ctxt, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movies, error) {

	// Validate if ID is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Get query
	query := `SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`

	// Response structure
	var movie Movies

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	// Use the context in the query
	// err := m.DB.QueryRow(query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	err := m.DB.QueryRowContext(ctxt, query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movies) error {

	// Update query
	query := `UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 and version = $6
	RETURNING version`

	// Argumets to the query
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctxt, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {

	// Validate if ID is valid
	if id < 1 {
		return ErrRecordNotFound
	}

	//Query
	query := "DELETE from MOVIES where id = $1"

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	res, err := m.DB.ExecContext(ctxt, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movies, Metadata, error) {

	// Basic Query
	/*query := `
	SELECT id, created_at, title, year, runtime, genres, version
	FROM movies WHERE (LOWER(title) =  LOWER($1) OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY id`
	*/

	// Query for full text search in PostGreSQL
	// count(*) OVER(), return total count of matching records returned by the query
	// 1st from clause below performs full text search
	// 2nd clause searches presenceof input in the genre list
	// Limit and Offset are used for pagination functionality
	query := fmt.Sprintf(`
			SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
			FROM movies WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
			AND (genres @> $2 OR $2 = '{}')
			ORDER BY %s %s, id
			LIMIT $3 OFFSET $4`, filters.getSortColumn(), filters.getSortDirection())

	// Create a context
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query
	rows, err := m.DB.QueryContext(ctxt, query, title, pq.Array(genres), filters.getLimit(), filters.getOffset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	// Variable to hold the data returned
	movies := []*Movies{}
	totalRecords := 0

	// Traverse the rows to get the data
	for rows.Next() {
		var movie Movies

		// Get the row data
		err = rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the moive into the return erray
		movies = append(movies, &movie)
	}

	// Check if any error occured while iterating
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate the metadata
	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
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
