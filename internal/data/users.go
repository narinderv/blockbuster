package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/narinderv/blockbuster/internal/validator"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type password struct {
	plaintext *string
	hash      []byte
}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// User Model
type UserModel struct {
	DB *sql.DB
}

// Generate and save the password hash from the plaintext password
func (pass *password) SetPasswordHash(passwrd string) error {

	// Get the hash of the password
	hash, err := bcrypt.GenerateFromPassword([]byte(passwrd), 12)
	if err != nil {
		return err
	}

	// Store the plaintext and Hash in the password structure
	pass.hash = hash
	pass.plaintext = &passwrd

	return nil
}

// Match the input password with the stored password by  comparing the hash
func (pass *password) MatchPassword(passwrd string) (bool, error) {

	err := bcrypt.CompareHashAndPassword(pass.hash, []byte(passwrd))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, err
		default:
			return false, err
		}
	}

	return true, err
}

// Validation functions

// Validate Email
func ValidateEmail(val *validator.Validator, email string) {

	// Email should not be empty
	val.Check(email != "", "email", "must be provided")

	// Check if valid email address is provided
	val.Check(validator.MatchPattern(email, validator.EmailRX), "email", "must be a vaild email address")
}

// Validate plaintext password
func ValidatePassword(val *validator.Validator, passwrd string) {

	// Password should not be empty
	val.Check(passwrd != "", "password", "must be provided")

	// Check length (min 8 bytes; max. 72 bytes)
	val.Check(len(passwrd) >= 8, "password", "must be atleast 8 bytes long")
	val.Check(len(passwrd) <= 72, "password", "must not be more than 72	 bytes long")
}

// Validate input user details
func ValidateUser(val *validator.Validator, user *User) {

	// Name
	val.Check(user.Name != "", "name", "must be provided")
	val.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes")

	// Email
	ValidateEmail(val, user.Email)

	// Plaintext password, if not nil
	if user.Password.plaintext != nil {
		ValidatePassword(val, *user.Password.plaintext)
	}

	// Password hash can never be nil
	if user.Password.hash == nil {
		panic("missing password hash for the user")
	}
}

// User Model functions
// Insert
func (userModel *UserModel) Insert(user *User) error {

	// Insert query
	query := `INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	// Argumets to the query
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	// Execute the query and store the result
	err := userModel.DB.QueryRowContext(ctxt, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// Get By Email
func (userModel *UserModel) GetByEmail(email string) (*User, error) {

	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
		WHERE email = $1`

	// Response structure
	var user User

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the context in the query
	err := userModel.DB.QueryRowContext(ctxt, query, email).Scan(&user.ID, &user.CreatedAt, &user.Name,
		&user.Email, &user.Password.hash, &user.Activated, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update user details
func (userModel *UserModel) Update(user *User) error {

	// Update query
	query := `UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	// Argumets to the query
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated, user.ID, user.Version}

	// Create a DB context to timeout the query if it exceeds a certian duration
	ctxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := userModel.DB.QueryRowContext(ctxt, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
