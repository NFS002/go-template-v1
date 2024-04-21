package utils

import (
	"github.com/go-playground/validator/v10"
)

func ValidateScope(fl validator.FieldLevel) bool {
	var value = fl.Field().String()
	for _, scope := range ValidScopes {
		if value == scope {
			return true
		}
	}
	return false
}
