package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"math/rand"
	"time"

	"github.com/jseow5177/greenlight/internal/validator"
)

// Define constants for the token scope.
const (
	ScopeActivation = "activation"
)

const (
	TokenByteLength = 26
)

// Define a Token struct to hold the data for an individual token.
// This includes the plaintext and hashed versions of the token, associated user ID, expiry time and scope.
type Token struct {
	Plaintext string // To be sent to client
	Hash      []byte // To be stored in DB
	UserID    int64
	Expiry    time.Time
	Scope     string
}

// Define the TokenModel struct
type TokenModel struct {
	DB *sql.DB
}

// Check that the plaintext token is provided and is exactly 26 bytes long
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == TokenByteLength, "token", "must be 26 bytes long")
}

// New() is a shortcut method to create a new Token struct and then insert the data
// in the tokens table.
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

// Insert() adds the data for a specific token to the tokens table.
func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)

	return err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create a Token instance containing the user ID, expiry, and scope information.
	// We add the provided ttl (time-to-live) duration parameter to the current time to get the expiry time.
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	// This gives our tokens 128-bits (16 bytes) of entropy (randomness).
	randomBytes := make([]byte, 16)

	// Use the Read() function from the crypto/rand package to fill the byte slice with random bytes from the OS's
	// cryptographically secure random number generator (CSPRNG). This returns an error if CSPRNG fails to function properly.
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encode the byte slice to a base-32-encoded string and assign it to the token Plaintext field.
	// This is then token string we send to the user in their welcome email. It looks like this:
	//
	// Y3QMGX3PJ3WLRL2YRTQGQ6KRHU
	//
	// The following uses base32.StdEncoding, which uses the standard base32 encoding, as defined in RFC 4648.
	// Note that by default, base-32 strings may be padded at the end with the = character.
	// We don't need this padding character for the purpose of our tokens, so we use WithPadding(base32.NoPadding)
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Generate a SHA-256 hash of the plaintext token string.
	// sha256.Sum256() returns an *array* of length 32. Hence, to make it easier to work with, we convert the array to a slice.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}
