package redis

import (
	db "erp.localhost/internal/db"
	"erp.localhost/internal/logging"
)

// NewKeyHandlerWithMockForTest creates a KeyHandler with a mock DBHandler for testing
// This is a test helper function and should only be used in test files
func NewKeyHandlerWithMockForTest[T any](mockHandler db.DBHandler, logger *logging.Logger) *KeyHandler[T] {
	if logger == nil {
		logger = logging.NewLogger(logging.ModuleDB)
	}
	return &KeyHandler[T]{
		dbHandler: mockHandler,
		logger:    logger,
	}
}
