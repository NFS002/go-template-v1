package models

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
