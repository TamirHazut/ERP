package auth

import (
	"time"

	"erp.localhost/internal/logging"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
	logger        *logging.Logger
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	logger := logging.NewLogger(logging.ModuleAuth)
	if secretKey == "" || tokenDuration <= 0 {
		logger.Fatal("secret key and token duration are required")
	}
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
		logger:        logger,
	}
}

func (m *JWTManager) GenerateToken(userID string, tenantID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       userID,
		"tenant_id": tenantID,
		"exp":       time.Now().Add(m.tokenDuration).Unix(),
	})
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) VerifyToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		m.logger.Error("failed to parse token", "error", err)
		return false, err
	}
	return token.Valid, nil
}

func (m *JWTManager) RefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		m.logger.Error("failed to parse token", "error", err)
		return "", err
	}
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) RevokeToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		m.logger.Error("failed to parse token", "error", err)
		return "", err
	}
	token.Valid = false
	token.Signature = []byte{}
	newToken, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		m.logger.Error("failed to sign token", "error", err)
		return "", err
	}
	return newToken, nil
}
