package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test error definitions for testing
var (
	testAuthError = ErrorDef{
		Code:     "TEST_AUTH_ERROR",
		Message:  "Test auth error message",
		Category: CategoryAuth,
	}
	testValidationError = ErrorDef{
		Code:     "TEST_VALIDATION_ERROR",
		Message:  "Test validation error message",
		Category: CategoryValidation,
	}
	testNotFoundError = ErrorDef{
		Code:     "TEST_NOT_FOUND_ERROR",
		Message:  "Test not found error message",
		Category: CategoryNotFound,
	}
	testConflictError = ErrorDef{
		Code:     "TEST_CONFLICT_ERROR",
		Message:  "Test conflict error message",
		Category: CategoryConflict,
	}
	testBusinessError = ErrorDef{
		Code:     "TEST_BUSINESS_ERROR",
		Message:  "Test business error message",
		Category: CategoryBusiness,
	}
	testInternalError = ErrorDef{
		Code:     "TEST_INTERNAL_ERROR",
		Message:  "Test internal error message",
		Category: CategoryInternal,
	}
	testNoCategoryError = ErrorDef{
		Code:    "TEST_NO_CATEGORY_ERROR",
		Message: "Test error without category",
	}
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name         string
		def          ErrorDef
		wantCode     string
		wantMessage  string
		wantCategory ErrorCategory
	}{
		{
			name:         "auth error",
			def:          testAuthError,
			wantCode:     "TEST_AUTH_ERROR",
			wantMessage:  "Test auth error message",
			wantCategory: CategoryAuth,
		},
		{
			name:         "validation error",
			def:          testValidationError,
			wantCode:     "TEST_VALIDATION_ERROR",
			wantMessage:  "Test validation error message",
			wantCategory: CategoryValidation,
		},
		{
			name:         "not found error",
			def:          testNotFoundError,
			wantCode:     "TEST_NOT_FOUND_ERROR",
			wantMessage:  "Test not found error message",
			wantCategory: CategoryNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(tc.def)
			require.NotNil(t, err)
			assert.Equal(t, tc.wantCode, err.Code)
			assert.Equal(t, tc.wantMessage, err.Message)
			assert.Equal(t, tc.wantCategory, err.Category)
			assert.NotNil(t, err.Details)
			assert.Nil(t, err.Err)
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	testCases := []struct {
		name         string
		def          ErrorDef
		wrappedErr   error
		wantContains string
	}{
		{
			name:         "wrap with error",
			def:          testInternalError,
			wrappedErr:   originalErr,
			wantContains: "original error",
		},
		{
			name:         "wrap with nil error",
			def:          testInternalError,
			wrappedErr:   nil,
			wantContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Wrap(tc.def, tc.wrappedErr)
			require.NotNil(t, err)
			assert.Equal(t, tc.def.Code, err.Code)
			assert.Equal(t, tc.def.Message, err.Message)
			assert.Equal(t, tc.wrappedErr, err.Err)
			if tc.wrappedErr != nil {
				assert.Contains(t, err.Error(), tc.wantContains)
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	testCases := []struct {
		name         string
		appErr       *AppError
		wantContains []string
	}{
		{
			name: "error without wrapped error",
			appErr: &AppError{
				Code:     "TEST_CODE",
				Message:  "Test message",
				Category: CategoryAuth,
			},
			wantContains: []string{"TEST_CODE", "Test message", "AUTH"},
		},
		{
			name: "error with wrapped error",
			appErr: &AppError{
				Code:     "TEST_CODE",
				Message:  "Test message",
				Category: CategoryInternal,
				Err:      errors.New("wrapped error"),
			},
			wantContains: []string{"TEST_CODE", "Test message", "INTERNAL", "wrapped error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errStr := tc.appErr.Error()
			for _, want := range tc.wantContains {
				assert.Contains(t, errStr, want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")

	testCases := []struct {
		name       string
		appErr     *AppError
		wantUnwrap error
	}{
		{
			name: "with wrapped error",
			appErr: &AppError{
				Code: "TEST",
				Err:  originalErr,
			},
			wantUnwrap: originalErr,
		},
		{
			name: "without wrapped error",
			appErr: &AppError{
				Code: "TEST",
				Err:  nil,
			},
			wantUnwrap: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped := tc.appErr.Unwrap()
			assert.Equal(t, tc.wantUnwrap, unwrapped)
		})
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Code: "ERROR_A"}
	err2 := &AppError{Code: "ERROR_A"}
	err3 := &AppError{Code: "ERROR_B"}
	stdErr := errors.New("standard error")

	testCases := []struct {
		name   string
		err    *AppError
		target error
		want   bool
	}{
		{
			name:   "same code",
			err:    err1,
			target: err2,
			want:   true,
		},
		{
			name:   "different code",
			err:    err1,
			target: err3,
			want:   false,
		},
		{
			name:   "non-AppError target",
			err:    err1,
			target: stdErr,
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.Is(tc.target)
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestAppError_WithDetails(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		value     any
		wantKey   string
		wantValue any
	}{
		{
			name:      "string value",
			key:       "field",
			value:     "username",
			wantKey:   "field",
			wantValue: "username",
		},
		{
			name:      "int value",
			key:       "count",
			value:     42,
			wantKey:   "count",
			wantValue: 42,
		},
		{
			name:      "slice value",
			key:       "fields",
			value:     []string{"a", "b"},
			wantKey:   "fields",
			wantValue: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(testAuthError)
			err.WithDetails(tc.key, tc.value)
			assert.Equal(t, tc.wantValue, err.Details[tc.wantKey])
		})
	}
}

func TestAppError_WithError(t *testing.T) {
	originalErr := errors.New("original error")

	err := New(testAuthError)
	assert.Nil(t, err.Err)

	err.WithError(originalErr)
	assert.Equal(t, originalErr, err.Err)
}

func TestAuth(t *testing.T) {
	testCases := []struct {
		name         string
		def          ErrorDef
		wantCategory ErrorCategory
	}{
		{
			name:         "with auth category",
			def:          testAuthError,
			wantCategory: CategoryAuth,
		},
		{
			name:         "without category - defaults to auth",
			def:          testNoCategoryError,
			wantCategory: CategoryAuth,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Auth(tc.def)
			require.NotNil(t, err)
			assert.Equal(t, tc.wantCategory, err.Category)
		})
	}
}

func TestValidation(t *testing.T) {
	testCases := []struct {
		name       string
		def        ErrorDef
		fields     []string
		wantFields bool
	}{
		{
			name:       "with fields",
			def:        testValidationError,
			fields:     []string{"email", "password"},
			wantFields: true,
		},
		{
			name:       "without fields",
			def:        testValidationError,
			fields:     nil,
			wantFields: false,
		},
		{
			name:       "empty fields slice",
			def:        testValidationError,
			fields:     []string{},
			wantFields: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validation(tc.def, tc.fields...)
			require.NotNil(t, err)
			assert.Equal(t, CategoryValidation, err.Category)
			if tc.wantFields {
				assert.Contains(t, err.Details, "fields")
				assert.Equal(t, tc.fields, err.Details["fields"])
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	testCases := []struct {
		name             string
		def              ErrorDef
		resourceType     string
		resourceID       any
		wantResourceType bool
		wantResourceID   bool
	}{
		{
			name:             "with both resource type and ID",
			def:              testNotFoundError,
			resourceType:     "User",
			resourceID:       "123",
			wantResourceType: true,
			wantResourceID:   true,
		},
		{
			name:             "only resource type",
			def:              testNotFoundError,
			resourceType:     "User",
			resourceID:       nil,
			wantResourceType: true,
			wantResourceID:   false,
		},
		{
			name:             "only resource ID",
			def:              testNotFoundError,
			resourceType:     "",
			resourceID:       "123",
			wantResourceType: false,
			wantResourceID:   true,
		},
		{
			name:             "neither",
			def:              testNotFoundError,
			resourceType:     "",
			resourceID:       nil,
			wantResourceType: false,
			wantResourceID:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NotFound(tc.def, tc.resourceType, tc.resourceID)
			require.NotNil(t, err)
			assert.Equal(t, CategoryNotFound, err.Category)

			if tc.wantResourceType {
				assert.Equal(t, tc.resourceType, err.Details["resource_type"])
			} else {
				assert.NotContains(t, err.Details, "resource_type")
			}

			if tc.wantResourceID {
				assert.Equal(t, tc.resourceID, err.Details["resource_id"])
			} else {
				assert.NotContains(t, err.Details, "resource_id")
			}
		})
	}
}

func TestConflict(t *testing.T) {
	err := Conflict(testConflictError)
	require.NotNil(t, err)
	assert.Equal(t, CategoryConflict, err.Category)
	assert.Equal(t, testConflictError.Code, err.Code)
}

func TestBusiness(t *testing.T) {
	err := Business(testBusinessError)
	require.NotNil(t, err)
	assert.Equal(t, CategoryBusiness, err.Category)
	assert.Equal(t, testBusinessError.Code, err.Code)
}

func TestInternal(t *testing.T) {
	originalErr := errors.New("database connection failed")

	testCases := []struct {
		name       string
		def        ErrorDef
		wrappedErr error
	}{
		{
			name:       "with wrapped error",
			def:        testInternalError,
			wrappedErr: originalErr,
		},
		{
			name:       "with nil error",
			def:        testInternalError,
			wrappedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Internal(tc.def, tc.wrappedErr)
			require.NotNil(t, err)
			assert.Equal(t, CategoryInternal, err.Category)
			assert.Equal(t, tc.wrappedErr, err.Err)
		})
	}
}

func TestIsAppError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is AppError",
			err:  New(testAuthError),
			want: true,
		},
		{
			name: "is standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "is nil",
			err:  nil,
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAppError(tc.err)
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestAsAppError(t *testing.T) {
	appErr := New(testAuthError)

	testCases := []struct {
		name    string
		err     error
		wantErr *AppError
		wantOk  bool
	}{
		{
			name:    "is AppError",
			err:     appErr,
			wantErr: appErr,
			wantOk:  true,
		},
		{
			name:    "is standard error",
			err:     errors.New("standard error"),
			wantErr: nil,
			wantOk:  false,
		},
		{
			name:    "is nil",
			err:     nil,
			wantErr: nil,
			wantOk:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, ok := AsAppError(tc.err)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantErr, result)
		})
	}
}

func TestIsCategory(t *testing.T) {
	authErr := Auth(testAuthError)
	validationErr := Validation(testValidationError)
	stdErr := errors.New("standard error")

	testCases := []struct {
		name     string
		err      error
		category ErrorCategory
		want     bool
	}{
		{
			name:     "auth error - auth category",
			err:      authErr,
			category: CategoryAuth,
			want:     true,
		},
		{
			name:     "auth error - wrong category",
			err:      authErr,
			category: CategoryValidation,
			want:     false,
		},
		{
			name:     "validation error - validation category",
			err:      validationErr,
			category: CategoryValidation,
			want:     true,
		},
		{
			name:     "standard error",
			err:      stdErr,
			category: CategoryAuth,
			want:     false,
		},
		{
			name:     "nil error",
			err:      nil,
			category: CategoryAuth,
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsCategory(tc.err, tc.category)
			assert.Equal(t, tc.want, result)
		})
	}
}

