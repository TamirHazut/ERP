package repository

import (
	"errors"
	"testing"
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db/mock"
	erp_errors "erp.localhost/internal/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoleRepository(t *testing.T) {
	mockHandler := &mock.MockDBHandler{}
	repo := NewRoleRepository(mockHandler)

	require.NotNil(t, repo)
	assert.NotNil(t, repo.repository)
	assert.NotNil(t, repo.logger)
}

func TestRoleRepository_CreateRole(t *testing.T) {
	testCases := []struct {
		name      string
		role      models.Role
		mockFunc  func(collection string, data any) (string, error)
		wantID    string
		wantErr   bool
	}{
		{
			name: "successful create",
			role: models.Role{
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"read", "write"},
			},
			mockFunc: func(collection string, data any) (string, error) {
				return "role-id-123", nil
			},
			wantID:  "role-id-123",
			wantErr: false,
		},
		{
			name: "create with validation error - missing tenant ID",
			role: models.Role{
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"read", "write"},
			},
			mockFunc: func(collection string, data any) (string, error) {
				return "", nil
			},
			wantID:  "",
			wantErr: true,
		},
		{
			name: "create with database error",
			role: models.Role{
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"read", "write"},
			},
			mockFunc: func(collection string, data any) (string, error) {
				return "", errors.New("database connection failed")
			},
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				CreateFunc: tc.mockFunc,
			}
			repo := NewRoleRepository(mockHandler)

			id, err := repo.CreateRole(tc.role)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantID, id)
			}
		})
	}
}

func TestRoleRepository_GetRoleByID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		roleID    string
		mockFunc  func(collection string, filter map[string]any) ([]any, error)
		wantErr   bool
	}{
		{
			name:     "successful get by id",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				roleID, _ := primitive.ObjectIDFromHex("role-id-123")
				return []any{
					models.Role{
						ID:       roleID,
						TenantID: "tenant1",
						Name:     "Admin",
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "role not found",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindFunc: tc.mockFunc,
			}
			repo := NewRoleRepository(mockHandler)

			role, err := repo.GetRoleByID(tc.tenantID, tc.roleID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.name == "role not found" {
					appErr, ok := erp_errors.AsAppError(err)
					require.True(t, ok)
					assert.Equal(t, erp_errors.CategoryNotFound, appErr.Category)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, role.TenantID)
			}
		})
	}
}

func TestRoleRepository_GetRoleByName(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		roleName  string
		mockFunc  func(collection string, filter map[string]any) ([]any, error)
		wantErr   bool
	}{
		{
			name:     "successful get by name",
			tenantID: "tenant1",
			roleName: "Admin",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.Role{
						TenantID: "tenant1",
						Name:     "Admin",
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "role not found",
			tenantID: "tenant1",
			roleName: "Nonexistent",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			roleName: "Admin",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindFunc: tc.mockFunc,
			}
			repo := NewRoleRepository(mockHandler)

			role, err := repo.GetRoleByName(tc.tenantID, tc.roleName)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.roleName, role.Name)
			}
		})
	}
}

func TestRoleRepository_GetRolesByTenantID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		mockFunc  func(collection string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get roles by tenant",
			tenantID: "tenant1",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.Role{TenantID: "tenant1", Name: "Admin"},
					models.Role{TenantID: "tenant1", Name: "User"},
				}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "no roles found",
			tenantID: "tenant1",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindFunc: tc.mockFunc,
			}
			repo := NewRoleRepository(mockHandler)

			roles, err := repo.GetRolesByTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, roles, tc.wantCount)
			}
		})
	}
}

func TestRoleRepository_GetRolesByPermissionsIDs(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		permissionsIDs []string
		mockFunc       func(collection string, filter map[string]any) ([]any, error)
		wantCount      int
		wantErr        bool
	}{
		{
			name:           "successful get roles by permissions",
			tenantID:       "tenant1",
			permissionsIDs: []string{"perm1", "perm2"},
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.Role{TenantID: "tenant1", Name: "Admin"},
				}, nil
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:           "database error",
			tenantID:       "tenant1",
			permissionsIDs: []string{"perm1"},
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return nil, errors.New("database query failed")
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindFunc: tc.mockFunc,
			}
			repo := NewRoleRepository(mockHandler)

			roles, err := repo.GetRolesByPermissionsIDs(tc.tenantID, tc.permissionsIDs)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, roles, tc.wantCount)
			}
		})
	}
}

func TestRoleRepository_UpdateRole(t *testing.T) {
	roleID := primitive.NewObjectID()
	createdAt := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name        string
		role        models.Role
		mockFind    func(collection string, filter map[string]any) ([]any, error)
		mockUpdate  func(collection string, filter map[string]any, data any) error
		wantErr     bool
	}{
		{
			name: "successful update",
			role: models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Updated Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   createdAt,
				Permissions: []string{"read", "write"},
			},
			mockFind: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.Role{
						ID:        roleID,
						TenantID:  "tenant1",
						Name:      "Admin",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "update with validation error",
			role: models.Role{
				TenantID: "tenant1",
			},
			mockFind: func(collection string, filter map[string]any) ([]any, error) {
				return nil, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with role not found",
			role: models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   createdAt,
				Permissions: []string{"read", "write"},
			},
			mockFind: func(collection string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with restricted field change - CreatedAt",
			role: models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   time.Now(),
				Permissions: []string{"read", "write"},
			},
			mockFind: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.Role{
						ID:        roleID,
						TenantID:  "tenant1",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with database error",
			role: models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   createdAt,
				Permissions: []string{"read", "write"},
			},
			mockFind: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.Role{
						ID:        roleID,
						TenantID:  "tenant1",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return errors.New("update failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindFunc:   tc.mockFind,
				UpdateFunc: tc.mockUpdate,
			}
			repo := NewRoleRepository(mockHandler)

			err := repo.UpdateRole(tc.role)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoleRepository_DeleteRole(t *testing.T) {
	testCases := []struct {
		name       string
		tenantID   string
		roleID     string
		mockFunc   func(collection string, filter map[string]any) error
		wantErr    bool
	}{
		{
			name:     "successful delete",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with empty tenant ID",
			tenantID: "",
			roleID:   "role-id-123",
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with empty role ID",
			tenantID: "tenant1",
			roleID:   "",
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			mockFunc: func(collection string, filter map[string]any) error {
				return errors.New("delete failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				DeleteFunc: tc.mockFunc,
			}
			repo := NewRoleRepository(mockHandler)

			err := repo.DeleteRole(tc.tenantID, tc.roleID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

