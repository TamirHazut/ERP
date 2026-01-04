package password

import (
	erp_errors "erp.localhost/internal/infra/error"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	minEntropyBits = 60.0
)

func HashPassword(password string) (string, error) {
	err := passwordvalidator.Validate(password, minEntropyBits)
	if err != nil {
		return "", erp_errors.Validation(erp_errors.ValidationPasswordTooWeak)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", erp_errors.Internal(erp_errors.InternalUnexpectedError, err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
