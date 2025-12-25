package models

import (
	"testing"

	erp_errors "erp.localhost/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =============================================================================
// TENANT VALIDATION TESTS
// =============================================================================

func TestTenant_Validate(t *testing.T) {
	validID := primitive.NewObjectID()

	testCases := []struct {
		name            string
		tenant          Tenant
		createOperation bool
		wantErr         bool
		wantFields      []string
	}{
		{
			name: "valid tenant - create operation",
			tenant: Tenant{
				Name:      "Test Company",
				Status:    TenantStatusActive,
				CreatedBy: "admin-user",
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid tenant - update operation",
			tenant: Tenant{
				ID:        validID,
				Name:      "Test Company",
				Status:    TenantStatusActive,
				CreatedBy: "admin-user",
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			tenant: Tenant{
				Name:      "Test Company",
				Status:    TenantStatusActive,
				CreatedBy: "admin-user",
			},
			createOperation: false,
			wantErr:         true,
			wantFields:      []string{"ID"},
		},
		{
			name: "missing name",
			tenant: Tenant{
				Status:    TenantStatusActive,
				CreatedBy: "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Name"},
		},
		{
			name: "missing status",
			tenant: Tenant{
				Name:      "Test Company",
				CreatedBy: "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Status"},
		},
		{
			name: "missing createdBy",
			tenant: Tenant{
				Name:   "Test Company",
				Status: TenantStatusActive,
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"CreatedBy"},
		},
		{
			name:            "empty tenant - create",
			tenant:          Tenant{},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Name", "Status", "CreatedBy"},
		},
		{
			name:            "empty tenant - update",
			tenant:          Tenant{},
			createOperation: false,
			wantErr:         true,
			wantFields:      []string{"ID", "Name", "Status", "CreatedBy"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tenant.Validate(tc.createOperation)
			if tc.wantErr {
				require.Error(t, err)
				// Check that it's an AppError with validation category
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryValidation, appErr.Category)
				// Check fields if specified
				if len(tc.wantFields) > 0 {
					fields, exists := appErr.Details["fields"]
					require.True(t, exists, "Expected 'fields' in error details")
					for _, wantField := range tc.wantFields {
						assert.Contains(t, fields, wantField)
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// USER VALIDATION TESTS
// =============================================================================

func TestUser_Validate(t *testing.T) {
	validID := primitive.NewObjectID()

	testCases := []struct {
		name            string
		user            User
		createOperation bool
		wantErr         bool
		wantFields      []string
	}{
		{
			name: "valid user - create operation",
			user: User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid user - update operation",
			user: User{
				ID:           validID,
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			user: User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: false,
			wantErr:         true,
			wantFields:      []string{"ID"},
		},
		{
			name: "missing tenantID",
			user: User{
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        []UserRole{{RoleID: "role-1"}},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"TenantID"},
		},
		{
			name: "missing email",
			user: User{
				TenantID:     "tenant-123",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Email"},
		},
		{
			name: "missing passwordHash",
			user: User{
				TenantID:  "tenant-123",
				Email:     "test@example.com",
				Status:    UserStatusActive,
				CreatedBy: "admin-user",
				Roles:     []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"PasswordHash"},
		},
		{
			name: "missing status",
			user: User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				CreatedBy:    "admin-user",
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Status"},
		},
		{
			name: "missing createdBy",
			user: User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				Roles:        []UserRole{{RoleID: "role-1", TenantID: "tenant-123"}},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"CreatedBy"},
		},
		{
			name: "missing roles (nil)",
			user: User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        nil,
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Roles"},
		},
		{
			name: "empty roles slice is valid",
			user: User{
				TenantID:     "tenant-123",
				Email:        "test@example.com",
				PasswordHash: "$2a$10$hashedpassword",
				Status:       UserStatusActive,
				CreatedBy:    "admin-user",
				Roles:        []UserRole{},
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name:            "empty user - create",
			user:            User{},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"TenantID", "Email", "PasswordHash", "Status", "CreatedBy", "Roles"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.user.Validate(tc.createOperation)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryValidation, appErr.Category)
				if len(tc.wantFields) > 0 {
					fields, exists := appErr.Details["fields"]
					require.True(t, exists, "Expected 'fields' in error details")
					for _, wantField := range tc.wantFields {
						assert.Contains(t, fields, wantField)
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// ROLE VALIDATION TESTS
// =============================================================================

func TestRole_Validate(t *testing.T) {
	validID := primitive.NewObjectID()

	testCases := []struct {
		name            string
		role            Role
		createOperation bool
		wantErr         bool
		wantFields      []string
	}{
		{
			name: "valid role - create operation",
			role: Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: []string{"read:users", "write:users"},
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid role - update operation",
			role: Role{
				ID:          validID,
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: []string{"read:users", "write:users"},
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			role: Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: []string{"read:users"},
			},
			createOperation: false,
			wantErr:         true,
			wantFields:      []string{"ID"},
		},
		{
			name: "missing tenantID",
			role: Role{
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: []string{"read:users"},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"TenantID"},
		},
		{
			name: "missing name",
			role: Role{
				TenantID:    "tenant-123",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: []string{"read:users"},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Name"},
		},
		{
			name: "missing status",
			role: Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				CreatedBy:   "admin-user",
				Permissions: []string{"read:users"},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Status"},
		},
		{
			name: "missing createdBy",
			role: Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				Permissions: []string{"read:users"},
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"CreatedBy"},
		},
		{
			name: "missing permissions (nil)",
			role: Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: nil,
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Permissions"},
		},
		{
			name: "empty permissions slice is valid",
			role: Role{
				TenantID:    "tenant-123",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin-user",
				Permissions: []string{},
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name:            "empty role - create",
			role:            Role{},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"TenantID", "Name", "Status", "CreatedBy", "Permissions"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.role.Validate(tc.createOperation)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryValidation, appErr.Category)
				if len(tc.wantFields) > 0 {
					fields, exists := appErr.Details["fields"]
					require.True(t, exists, "Expected 'fields' in error details")
					for _, wantField := range tc.wantFields {
						assert.Contains(t, fields, wantField)
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// PERMISSION VALIDATION TESTS
// =============================================================================

func TestPermission_Validate(t *testing.T) {
	validID := primitive.NewObjectID()

	testCases := []struct {
		name            string
		permission      Permission
		createOperation bool
		wantErr         bool
		wantFields      []string
	}{
		{
			name: "valid permission - create operation",
			permission: Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin-user",
			},
			createOperation: true,
			wantErr:         false,
		},
		{
			name: "valid permission - update operation",
			permission: Permission{
				ID:               validID,
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin-user",
			},
			createOperation: false,
			wantErr:         false,
		},
		{
			name: "missing ID on update",
			permission: Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin-user",
			},
			createOperation: false,
			wantErr:         true,
			wantFields:      []string{"ID"},
		},
		{
			name: "missing tenantID",
			permission: Permission{
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"TenantID"},
		},
		{
			name: "missing resource",
			permission: Permission{
				TenantID:         "tenant-123",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Resource"},
		},
		{
			name: "missing action",
			permission: Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
				CreatedBy:        "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"Action"},
		},
		{
			name: "missing permissionString",
			permission: Permission{
				TenantID:    "tenant-123",
				Resource:    "users",
				Action:      "read",
				DisplayName: "Read Users",
				CreatedBy:   "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"PermissionString"},
		},
		{
			name: "missing displayName",
			permission: Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				CreatedBy:        "admin-user",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"DisplayName"},
		},
		{
			name: "missing createdBy",
			permission: Permission{
				TenantID:         "tenant-123",
				Resource:         "users",
				Action:           "read",
				PermissionString: "users:read",
				DisplayName:      "Read Users",
			},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"CreatedBy"},
		},
		{
			name:            "empty permission - create",
			permission:      Permission{},
			createOperation: true,
			wantErr:         true,
			wantFields:      []string{"TenantID", "Resource", "Action", "PermissionString", "DisplayName", "CreatedBy"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.permission.Validate(tc.createOperation)
			if tc.wantErr {
				require.Error(t, err)
				appErr, ok := erp_errors.AsAppError(err)
				require.True(t, ok, "Expected AppError")
				assert.Equal(t, erp_errors.CategoryValidation, appErr.Category)
				if len(tc.wantFields) > 0 {
					fields, exists := appErr.Details["fields"]
					require.True(t, exists, "Expected 'fields' in error details")
					for _, wantField := range tc.wantFields {
						assert.Contains(t, fields, wantField)
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// CONSTANTS TESTS
// =============================================================================

func TestUserStatusConstants(t *testing.T) {
	// Verify all user status constants are defined correctly
	assert.Equal(t, "active", UserStatusActive)
	assert.Equal(t, "inactive", UserStatusInactive)
	assert.Equal(t, "suspended", UserStatusSuspended)
	assert.Equal(t, "invited", UserStatusInvited)
}

func TestTenantStatusConstants(t *testing.T) {
	// Verify all tenant status constants are defined correctly
	assert.Equal(t, "active", TenantStatusActive)
	assert.Equal(t, "suspended", TenantStatusSuspended)
	assert.Equal(t, "inactive", TenantStatusInactive)
	assert.Equal(t, "trial", TenantStatusTrial)
}

func TestRoleTypeConstants(t *testing.T) {
	// Verify all role type constants are defined correctly
	assert.Equal(t, "system_admin", RoleSystemAdmin)
	assert.Equal(t, "tenant_admin", RoleTenantAdmin)
}

func TestPermissionConstants(t *testing.T) {
	// Verify permission constants are defined correctly
	assert.Equal(t, "*:*", PermissionWildcard)
	assert.Equal(t, "resource:action[:scope]", PermissionFormat)
}

