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

func TestNewPermissionRepository(t *testing.T) {
	mockHandler := &mock.MockDBHandler{}
	repo := NewPermissionRepository(mockHandler)

	require.NotNil(t, repo)
	assert.NotNil(t, repo.repository)
	assert.NotNil(t, repo.logger)
}

func TestPermissionRepository_CreatePermission(t *testing.T) {
	testCases := []struct {
		name      string
		permission models.Permission
		mockFunc  func(db string, data any) (string, error)
		wantID    string
		wantErr   bool
	}{
		{
			name: "successful create",
			permission: models.Permission{
				TenantID:        "tenant1",
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products",
				CreatedBy:       "admin",
			},
			mockFunc: func(db string, data any) (string, error) {
				return "permission-id-123", nil
			},
			wantID:  "permission-id-123",
			wantErr: false,
		},
		{
			name: "create with validation error - missing tenant ID",
			permission: models.Permission{
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products",
				CreatedBy:       "admin",
			},
			mockFunc: func(db string, data any) (string, error) {
				return "", nil
			},
			wantID:  "",
			wantErr: true,
		},
		{
			name: "create with database error",
			permission: models.Permission{
				TenantID:        "tenant1",
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products",
				CreatedBy:       "admin",
			},
			mockFunc: func(db string, data any) (string, error) {
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
			repo := NewPermissionRepository(mockHandler)

			id, err := repo.CreatePermission(tc.permission)
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

func TestPermissionRepository_GetPermissionByID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		permissionID string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantErr   bool
	}{
		{
			name:     "successful get by id",
			tenantID: "tenant1",
			permissionID: "permission-id-123",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				permissionID, _ := primitive.ObjectIDFromHex("permission-id-123")
				return []any{
					models.Permission{
						ID:       permissionID,
						TenantID: "tenant1",
						Resource: "products",
						Action:   "read",
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "permission not found",
			tenantID: "tenant1",
			permissionID: "permission-id-123",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			permissionID: "permission-id-123",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
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
			repo := NewPermissionRepository(mockHandler)

			permission, err := repo.GetPermissionByID(tc.tenantID, tc.permissionID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.name == "permission not found" {
					appErr, ok := erp_errors.AsAppError(err)
					require.True(t, ok)
					assert.Equal(t, erp_errors.CategoryNotFound, appErr.Category)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, permission.TenantID)
			}
		})
	}
}

func TestPermissionRepository_GetPermissionByName(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		permissionName string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantErr   bool
	}{
		{
			name:     "successful get by name",
			tenantID: "tenant1",
			permissionName: "Read Products",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{
						TenantID:    "tenant1",
						DisplayName: "Read Products",
						Resource:    "products",
						Action:      "read",
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "permission not found",
			tenantID: "tenant1",
			permissionName: "Nonexistent",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			permissionName: "Read Products",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
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
			repo := NewPermissionRepository(mockHandler)

			permission, err := repo.GetPermissionByName(tc.tenantID, tc.permissionName)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.permissionName, permission.DisplayName)
			}
		})
	}
}

func TestPermissionRepository_GetPermissionsByTenantID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get permissions by tenant",
			tenantID: "tenant1",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
					models.Permission{TenantID: "tenant1", Resource: "products", Action: "write"},
				}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "no permissions found",
			tenantID: "tenant1",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
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
			repo := NewPermissionRepository(mockHandler)

			permissions, err := repo.GetPermissionsByTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionRepository_GetPermissionsByResource(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		resource  string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get permissions by resource",
			tenantID: "tenant1",
			resource: "products",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
					models.Permission{TenantID: "tenant1", Resource: "products", Action: "write"},
				}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			resource: "products",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
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
			repo := NewPermissionRepository(mockHandler)

			permissions, err := repo.GetPermissionsByResource(tc.tenantID, tc.resource)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionRepository_GetPermissionsByAction(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		action    string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get permissions by action",
			tenantID: "tenant1",
			action:   "read",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
					models.Permission{TenantID: "tenant1", Resource: "orders", Action: "read"},
				}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			action:   "read",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
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
			repo := NewPermissionRepository(mockHandler)

			permissions, err := repo.GetPermissionsByAction(tc.tenantID, tc.action)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionRepository_GetPermissionsByResourceAndAction(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		resource  string
		action    string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get permissions by resource and action",
			tenantID: "tenant1",
			resource: "products",
			action:   "read",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
				}, nil
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			resource: "products",
			action:   "read",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
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
			repo := NewPermissionRepository(mockHandler)

			permissions, err := repo.GetPermissionsByResourceAndAction(tc.tenantID, tc.resource, tc.action)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionRepository_UpdatePermission(t *testing.T) {
	permissionID := primitive.NewObjectID()
	createdAt := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name        string
		permission  models.Permission
		mockFind    func(db string, filter map[string]any) ([]any, error)
		mockUpdate  func(db string, filter map[string]any, data any) error
		wantErr     bool
	}{
		{
			name: "successful update",
			permission: models.Permission{
				ID:              permissionID,
				TenantID:        "tenant1",
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products Updated",
				CreatedBy:       "admin",
				CreatedAt:       createdAt,
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{
						ID:        permissionID,
						TenantID:  "tenant1",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockUpdate: func(db string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "update with validation error",
			permission: models.Permission{
				TenantID: "tenant1",
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return nil, nil
			},
			mockUpdate: func(db string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with permission not found",
			permission: models.Permission{
				ID:              permissionID,
				TenantID:        "tenant1",
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products",
				CreatedBy:       "admin",
				CreatedAt:       createdAt,
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			mockUpdate: func(db string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with restricted field change - CreatedAt",
			permission: models.Permission{
				ID:              permissionID,
				TenantID:        "tenant1",
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products",
				CreatedBy:       "admin",
				CreatedAt:       time.Now(),
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{
						ID:        permissionID,
						TenantID:  "tenant1",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockUpdate: func(db string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with database error",
			permission: models.Permission{
				ID:              permissionID,
				TenantID:        "tenant1",
				Resource:        "products",
				Action:          "read",
				PermissionString: "products:read",
				DisplayName:     "Read Products",
				CreatedBy:       "admin",
				CreatedAt:       createdAt,
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Permission{
						ID:        permissionID,
						TenantID:  "tenant1",
						CreatedAt: createdAt,
					},
				}, nil
			},
			mockUpdate: func(db string, filter map[string]any, data any) error {
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
			repo := NewPermissionRepository(mockHandler)

			err := repo.UpdatePermission(tc.permission)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPermissionRepository_DeletePermission(t *testing.T) {
	testCases := []struct {
		name       string
		tenantID   string
		permissionID string
		mockFunc   func(db string, filter map[string]any) error
		wantErr    bool
	}{
		{
			name:     "successful delete",
			tenantID: "tenant1",
			permissionID: "permission-id-123",
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with empty tenant ID",
			tenantID: "",
			permissionID: "permission-id-123",
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with empty permission ID",
			tenantID: "tenant1",
			permissionID: "",
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant1",
			permissionID: "permission-id-123",
			mockFunc: func(db string, filter map[string]any) error {
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
			repo := NewPermissionRepository(mockHandler)

			err := repo.DeletePermission(tc.tenantID, tc.permissionID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

