package models

import (
	"strings"
)

// JSON Unmarshall will already validate the type when parsing
type TokenRequest struct {
	Email    string   `json:"email"  validate:"required,email"`
	Password string   `json:"password" validate:"required"`
	Scope    []string `json:"scope" validate:"dive,scope"`
	Expiry   int      `json:"expiry" validate:"gte=-55,lte=1380"`
}

func (t *TokenRequest) Defaults() {
	if t.Scope == nil {
		t.Scope = []string{}
	}

	// Default value for expiry (int) is already set by JSON Unmarshall as 0
}

// Request body for updating a user record
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty" validate:"len=0|email"`
	Password  string `json:"password,omitempty"`
}

// Trim leading and trailing whitespace from all fields
func (u *UpdateUserRequest) Trim() {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)
	u.Email = strings.TrimSpace(u.Email)
	u.Password = strings.TrimSpace(u.Password)
}

func (u *UpdateUserRequest) IsEmpty() bool {
	return *u == UpdateUserRequest{}
}
