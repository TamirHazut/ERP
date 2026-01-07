package password

import (
	infra_error "erp.localhost/internal/infra/error"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	minEntropyBits = 60.0
)

func HashPassword(password string) (string, error) {
	err := passwordvalidator.Validate(password, minEntropyBits)
	if err != nil {
		return "", infra_error.Validation(infra_error.ValidationPasswordTooWeak)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
