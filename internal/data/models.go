package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// A "Base" Model to encapsulate all Models
type Models struct {
	Movies MovieModel
	Users  UserModel
}

// Initializer for the Model
func NewModel(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
