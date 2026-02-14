package hash

import (
	infra_error "erp.localhost/infra/error"
	"golang.org/x/crypto/bcrypt"
)

func VerifyHash(obj, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(obj)) == nil
}

func Hash(obj string) (string, *infra_error.AppError) {
	hashedObj, err := bcrypt.GenerateFromPassword([]byte(obj), bcrypt.DefaultCost)
	if err != nil {
		return "", infra_error.Internal(infra_error.InternalUnexpectedError, err)
	}
	return string(hashedObj), nil
}
