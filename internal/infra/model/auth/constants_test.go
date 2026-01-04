package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePermissionString(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		action         string
		expectedResult string
		wantErr        bool
		expectedErrMsg string
	}{
		// Positive cases - valid resource and action combinations
		{
			name:           "valid user:create permission",
			resource:       "user",
			action:         "create",
			expectedResult: "user:create",
			wantErr:        false,
		},
		{
			name:           "valid role:read permission",
			resource:       "role",
			action:         "read",
			expectedResult: "role:read",
			wantErr:        false,
		},
		{
			name:           "valid permission:update permission",
			resource:       "permission",
			action:         "update",
			expectedResult: "permission:update",
			wantErr:        false,
		},
		{
			name:           "valid order:delete permission",
			resource:       "order",
			action:         "delete",
			expectedResult: "order:delete",
			wantErr:        false,
		},
		{
			name:           "valid product:create permission",
			resource:       "product",
			action:         "create",
			expectedResult: "product:create",
			wantErr:        false,
		},
		{
			name:           "valid vendor:read permission",
			resource:       "vendor",
			action:         "read",
			expectedResult: "vendor:read",
			wantErr:        false,
		},
		{
			name:           "valid customer:update permission",
			resource:       "customer",
			action:         "update",
			expectedResult: "customer:update",
			wantErr:        false,
		},
		{
			name:           "valid config:delete permission",
			resource:       "config",
			action:         "delete",
			expectedResult: "config:delete",
			wantErr:        false,
		},
		{
			name:           "valid tenant:create permission",
			resource:       "tenant",
			action:         "create",
			expectedResult: "tenant:create",
			wantErr:        false,
		},
		{
			name:           "valid refresh_token:read permission",
			resource:       "refresh_token",
			action:         "read",
			expectedResult: "refresh_token:read",
			wantErr:        false,
		},
		{
			name:           "valid access_token:update permission",
			resource:       "access_token",
			action:         "update",
			expectedResult: "access_token:update",
			wantErr:        false,
		},
		// Mixed case should be normalized to lowercase
		{
			name:           "mixed case resource - User",
			resource:       "User",
			action:         "create",
			expectedResult: "user:create",
			wantErr:        false,
		},
		{
			name:           "mixed case action - Create",
			resource:       "user",
			action:         "Create",
			expectedResult: "user:create",
			wantErr:        false,
		},
		{
			name:           "mixed case both - User:Create",
			resource:       "USER",
			action:         "CREATE",
			expectedResult: "user:create",
			wantErr:        false,
		},
		// Negative cases - invalid resource
		{
			name:           "invalid resource - empty string",
			resource:       "",
			action:         "create",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "resource",
		},
		{
			name:           "invalid resource - unknown type",
			resource:       "invalid_resource",
			action:         "create",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "resource",
		},
		{
			name:           "invalid resource - random string",
			resource:       "foobar",
			action:         "create",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "resource",
		},
		// Negative cases - invalid action
		{
			name:           "invalid action - empty string",
			resource:       "user",
			action:         "",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "action",
		},
		{
			name:           "invalid action - unknown type",
			resource:       "user",
			action:         "invalid_action",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "action",
		},
		{
			name:           "invalid action - random string",
			resource:       "user",
			action:         "foobar",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "action",
		},
		// Negative cases - both invalid
		{
			name:           "both invalid - empty strings",
			resource:       "",
			action:         "",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "resource",
		},
		{
			name:           "both invalid - unknown types",
			resource:       "invalid_resource",
			action:         "invalid_action",
			expectedResult: "",
			wantErr:        true,
			expectedErrMsg: "resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CreatePermissionString(tt.resource, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				assert.Equal(t, tt.expectedResult, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestIsValidPermissionFormat(t *testing.T) {
	tests := []struct {
		name             string
		permissionFormat string
		expected         bool
	}{
		// Positive cases - valid permission formats
		{
			name:             "valid user:create",
			permissionFormat: "user:create",
			expected:         true,
		},
		{
			name:             "valid role:read",
			permissionFormat: "role:read",
			expected:         true,
		},
		{
			name:             "valid permission:update",
			permissionFormat: "permission:update",
			expected:         true,
		},
		{
			name:             "valid order:delete",
			permissionFormat: "order:delete",
			expected:         true,
		},
		{
			name:             "valid product:create",
			permissionFormat: "product:create",
			expected:         true,
		},
		{
			name:             "valid vendor:read",
			permissionFormat: "vendor:read",
			expected:         true,
		},
		{
			name:             "valid customer:update",
			permissionFormat: "customer:update",
			expected:         true,
		},
		{
			name:             "valid config:delete",
			permissionFormat: "config:delete",
			expected:         true,
		},
		{
			name:             "valid tenant:create",
			permissionFormat: "tenant:create",
			expected:         true,
		},
		{
			name:             "valid refresh_token:read",
			permissionFormat: "refresh_token:read",
			expected:         true,
		},
		{
			name:             "valid access_token:update",
			permissionFormat: "access_token:update",
			expected:         true,
		},
		// Mixed case should be normalized to lowercase
		{
			name:             "mixed case User:Create",
			permissionFormat: "User:Create",
			expected:         true,
		},
		{
			name:             "uppercase USER:CREATE",
			permissionFormat: "USER:CREATE",
			expected:         true,
		},
		{
			name:             "mixed case RoLe:ReAd",
			permissionFormat: "RoLe:ReAd",
			expected:         true,
		},
		// Negative cases - invalid format
		{
			name:             "empty string",
			permissionFormat: "",
			expected:         false,
		},
		{
			name:             "missing colon - usercreate",
			permissionFormat: "usercreate",
			expected:         false,
		},
		{
			name:             "missing action - user:",
			permissionFormat: "user:",
			expected:         false,
		},
		{
			name:             "missing resource - :create",
			permissionFormat: ":create",
			expected:         false,
		},
		{
			name:             "too many colons - user:create:extra",
			permissionFormat: "user:create:extra",
			expected:         false,
		},
		{
			name:             "invalid resource - invalid:create",
			permissionFormat: "invalid:create",
			expected:         false,
		},
		{
			name:             "invalid action - user:invalid",
			permissionFormat: "user:invalid",
			expected:         false,
		},
		{
			name:             "both invalid - invalid:invalid",
			permissionFormat: "invalid:invalid",
			expected:         false,
		},
		{
			name:             "only colon - :",
			permissionFormat: ":",
			expected:         false,
		},
		{
			name:             "spaces in format - user : create",
			permissionFormat: "user : create",
			expected:         false,
		},
		{
			name:             "random string",
			permissionFormat: "foobar",
			expected:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPermissionFormat(tt.permissionFormat)
			assert.Equal(t, tt.expected, result)
		})
	}
}
