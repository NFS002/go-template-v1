package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
)

// Token is the type for authentication tokens
type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      string    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     []string  `json:"scope"`
}

func (t Token) HasScope(scope []string) error {
	if len(scope) > 0 {
		for _, s := range scope {
			if !slices.Contains(t.Scope, s) {
				return fmt.Errorf("insufficent scope: %s", s)
			}
		}
	}
	return nil
}

// GenerateToken generates a token that lasts for ttl, and returns it
func GenerateToken(userID int, ttl time.Duration, scope []string) (*Token, error) {
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hex.EncodeToString(hash[:])
	return token, nil
}

func (m *DBModel) InsertToken(t *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// delete existing tokens
	stmt := `delete from tokens where user_id = $1`
	_, err := m.DB.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	// insert new token
	stmt = `
		insert into tokens 
			(user_id, token_hash, scope, expiry)
		values ($1, $2, $3, $4)
	`

	scope := strings.Join(t.Scope, ",")
	_, err = m.DB.ExecContext(ctx, stmt, u.ID, t.Hash, scope, t.Expiry)

	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetUserForToken(tokenStr string) (*User, *Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// convert plain text token into hash
	tokenHash := sha256.Sum256([]byte(tokenStr))

	var user User
	var token Token = Token{PlainText: tokenStr}
	var scope string

	query := `
	SELECT
		u.id, u.first_name, u.last_name, u.email, t.expiry, t.scope
	FROM
		users u
		INNER JOIN tokens t ON (u.id = t.user_id)
	WHERE
		t.token_hash = $1 AND t.expiry > NOW();

	`

	err := m.DB.QueryRowContext(ctx, query, hex.EncodeToString(tokenHash[:])).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&token.Expiry,
		&scope)

	token.Scope = strings.Split(scope, ",")

	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	return &user, &token, nil
}
