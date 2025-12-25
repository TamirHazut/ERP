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

func TestNewTenantRepository(t *testing.T) {
	mockHandler := &mock.MockDBHandler{}
	repo := NewTenantRepository(mockHandler)

	require.NotNil(t, repo)
	assert.NotNil(t, repo.repository)
	assert.NotNil(t, repo.logger)
}

func TestTenantRepository_CreateTenant(t *testing.T) {
	testCases := []struct {
		name      string
		tenant    models.Tenant
		mockFunc  func(db string, data any) (string, error)
		wantID    string
		wantErr   bool
	}{
		{
			name: "successful create",
			tenant: models.Tenant{
				Name:      "Test Company",
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
			},
			mockFunc: func(db string, data any) (string, error) {
				return "tenant-id-123", nil
			},
			wantID:  "tenant-id-123",
			wantErr: false,
		},
		{
			name: "create with validation error - missing name",
			tenant: models.Tenant{
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
			},
			mockFunc: func(db string, data any) (string, error) {
				return "", nil
			},
			wantID:  "",
			wantErr: true,
		},
		{
			name: "create with database error",
			tenant: models.Tenant{
				Name:      "Test Company",
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
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
			repo := NewTenantRepository(mockHandler)

			id, err := repo.CreateTenant(tc.tenant)
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

func TestTenantRepository_GetTenantByID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		mockFunc  func(db string, filter map[string]any) ([]any, error)
		wantErr   bool
	}{
		{
			name:     "successful get by id",
			tenantID: "507f1f77bcf86cd799439011",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				tenantID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
				return []any{
					models.Tenant{
						ID:     tenantID,
						Name:   "Test Company",
						Status: models.TenantStatusActive,
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "tenant not found",
			tenantID: "507f1f77bcf86cd799439011",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return []any{}, nil
			},
			wantErr: true,
		},
		{
			name:     "get with empty tenant ID",
			tenantID: "",
			mockFunc: func(db string, filter map[string]any) ([]any, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "507f1f77bcf86cd799439011",
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
			repo := NewTenantRepository(mockHandler)

			tenant, err := repo.GetTenantByID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.name == "tenant not found" {
					appErr, ok := erp_errors.AsAppError(err)
					require.True(t, ok)
					assert.Equal(t, erp_errors.CategoryNotFound, appErr.Category)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, tenant.ID.Hex())
			}
		})
	}
}

func TestTenantRepository_UpdateTenant(t *testing.T) {
	tenantID := primitive.NewObjectID()
	createdAt := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name        string
		tenant      models.Tenant
		mockFind    func(db string, filter map[string]any) ([]any, error)
		mockUpdate  func(db string, filter map[string]any, data any) error
		wantErr     bool
	}{
		{
			name: "successful update",
			tenant: models.Tenant{
				ID:        tenantID,
				Name:      "Updated Company",
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
				CreatedAt: createdAt,
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Tenant{
						ID:        tenantID,
						Name:      "Test Company",
						Status:    models.TenantStatusActive,
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
			tenant: models.Tenant{
				ID: tenantID,
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
			name: "update with tenant not found",
			tenant: models.Tenant{
				ID:        tenantID,
				Name:      "Test Company",
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
				CreatedAt: createdAt,
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
			tenant: models.Tenant{
				ID:        tenantID,
				Name:      "Test Company",
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
				CreatedAt: time.Now(),
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Tenant{
						ID:        tenantID,
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
			tenant: models.Tenant{
				ID:        tenantID,
				Name:      "Test Company",
				Status:    models.TenantStatusActive,
				CreatedBy: "admin",
				CreatedAt: createdAt,
			},
			mockFind: func(db string, filter map[string]any) ([]any, error) {
				return []any{
					models.Tenant{
						ID:        tenantID,
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
			repo := NewTenantRepository(mockHandler)

			err := repo.UpdateTenant(tc.tenant)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenantRepository_DeleteTenant(t *testing.T) {
	testCases := []struct {
		name       string
		tenantID   string
		mockFunc   func(db string, filter map[string]any) error
		wantErr    bool
	}{
		{
			name:     "successful delete",
			tenantID: "507f1f77bcf86cd799439011",
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with empty tenant ID",
			tenantID: "",
			mockFunc: func(db string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with database error",
			tenantID: "507f1f77bcf86cd799439011",
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
			repo := NewTenantRepository(mockHandler)

			err := repo.DeleteTenant(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

