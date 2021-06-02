package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jseow5177/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// A custom password type.
// The plaintext field is a pointer to a string so that we can distinguish between an a plaintext password not
// present in the struct, versus a plaintext password which is an empty string.
type password struct {
	plaintext *string
	hash      []byte
}

// A User struct to represent an individual user.
type User struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name string `json:"name"`
	Email string `json:"emai"`
	Password password `json:"-"` // private field
	Activated bool `json:"activated"`
	Version int `json:"-"` // private field
}

// Define a UserModel that wraps around a sql.DB connection pool
type UserModel struct {
	DB *sql.DB
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// Validate email
	ValidateEmail(v, user.Email)

	// If plaintext password is not nil, validate password
	if user.Password.plaintext != nil {
		ValidatePassword(v, *user.Password.plaintext)
	}

	// If the password hash is ever nil, this will be due to a logic error in our code base. 
	// It's a useful sanity check to include here. Since it is not a problem with the data provided
	// by the client, we raise a panic instead of adding an error to the validation map.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// The Set() method calculates the bcrypt hash of the plaintext password, and stores both the hash
// and the plaintext in the struct
func (p *password) Set(plaintextPassword string) error {
	// Generates a bcrypt hash of a password with a cost parameter of 12.
	// The higher the cost, the more secure the password, but more expensive to generate the hash.
	// The hash string has a format: $2b$[cost]$[22-character salt][31-character hash]
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// The Matches() method checks whether the provided plaintext password matches the hashed password stored
// in the struct. Return true if matches, else false.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	// Re-hash the provided password using the same salt and cost parameter that is in the hash string
	// we are comparing against.
	// The re-hashed value is checked against the original hash string using the subtle.ConstantTimeCompare() function,
	// which performs a comparison in constant time (to mitigate the risks of timing attacks).
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Insert a new record in the database for the user. The id, created_at and version fields
// are generated by the database, so we use the RETURNING clause to read them into the user struct.
func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Activated)
	if err != nil {
		switch {
		// Check if there is a violation of the UNIQUE constraint of email field
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// Retrive the User details from the database based on the user's email address.
// The query is expected to return only one record, or none at all (which we will return ErrRecordNotFound)
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, email, password_hash, activated, version
		FROM users
		WHERE email = $1`
	
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	user := new(User)

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

// Update the details of a specific user. We check against the version field to help prevent race conditions.
// We also check for a violation of the "users_email_key" constraint when performing the update.
func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.Version,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 *time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
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