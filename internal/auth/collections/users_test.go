package collection

import (
	"errors"
	"testing"
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db/mock"
	erp_errors "erp.localhost/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewUserCollection(t *testing.T) {
	mockHandler := &mock.MockDBHandler{}
	repo := NewUserCollection(mockHandler)

	require.NotNil(t, repo)
	assert.NotNil(t, repo.collection)
	assert.NotNil(t, repo.logger)
}

func TestUserCollection_CreateUser(t *testing.T) {
	testCases := []struct {
		name      string
		user      models.User
		mockFunc  func(collection string, data any) (string, error)
		wantID    string
		wantErr   bool
		wantError error
	}{
		{
			name: "successful create",
			user: models.User{
				TenantID:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []models.UserRole{},
			},
			mockFunc: func(collection string, data any) (string, error) {
				return "user-id-123", nil
			},
			wantID:  "user-id-123",
			wantErr: false,
		},
		{
			name: "create with validation error - missing tenant ID",
			user: models.User{
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []models.UserRole{},
			},
			mockFunc: func(collection string, data any) (string, error) {
				return "", nil
			},
			wantID:  "",
			wantErr: true,
		},
		{
			name: "create with database error",
			user: models.User{
				TenantID:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				Roles:        []models.UserRole{},
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
			repo := NewUserCollection(mockHandler)

			id, err := repo.CreateUser(tc.user)
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

func TestUserCollection_GetUserByID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		userID    string
		mockFunc  func(collection string, filter map[string]any) (any, error)
		wantUser  models.User
		wantErr   bool
		wantError error
	}{
		{
			name:     "successful get by id",
			tenantID: "tenant1",
			userID:   "user-id-123",
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				userID, _ := primitive.ObjectIDFromHex("user-id-123")
				return models.User{
					ID:       userID,
					TenantID: "tenant1",
					Email:    "test@example.com",
					Username: "testuser",
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			tenantID: "tenant1",
			userID:   "user-id-123",
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			userID:   "user-id-123",
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return nil, errors.New("database query failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.mockFunc,
			}
			repo := NewUserCollection(mockHandler)

			user, err := repo.GetUserByID(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.name == "user not found" {
					appErr, ok := erp_errors.AsAppError(err)
					require.True(t, ok)
					assert.Equal(t, erp_errors.CategoryNotFound, appErr.Category)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, user.TenantID)
			}
		})
	}
}

func TestUserCollection_GetUserByUsername(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		username  string
		mockFunc  func(collection string, filter map[string]any) (any, error)
		wantErr   bool
		wantError error
	}{
		{
			name:     "successful get by username",
			tenantID: "tenant1",
			username: "testuser",
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return models.User{
					TenantID: "tenant1",
					Username: "testuser",
					Email:    "test@example.com",
				}, nil
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			tenantID: "tenant1",
			username: "nonexistent",
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return nil, nil
			},
			wantErr: true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			username: "testuser",
			mockFunc: func(collection string, filter map[string]any) (any, error) {
				return nil, errors.New("database query failed")
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandler := &mock.MockDBHandler{
				FindOneFunc: tc.mockFunc,
			}
			repo := NewUserCollection(mockHandler)

			user, err := repo.GetUserByUsername(tc.tenantID, tc.username)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.username, user.Username)
			}
		})
	}
}

func TestUserCollection_GetUsersByTenantID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		mockFunc  func(collection string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get users by tenant",
			tenantID: "tenant1",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.User{TenantID: "tenant1", Username: "user1"},
					models.User{TenantID: "tenant1", Username: "user2"},
				}, nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:     "no users found",
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
				FindAllFunc: tc.mockFunc,
			}
			repo := NewUserCollection(mockHandler)

			users, err := repo.GetUsersByTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				if tc.wantCount == 0 {
					assert.Empty(t, users)
				} else {
					require.NoError(t, err)
					assert.Len(t, users, tc.wantCount)
				}
			}
		})
	}
}

func TestUserCollection_GetUsersByRoleID(t *testing.T) {
	testCases := []struct {
		name      string
		tenantID  string
		roleID    string
		mockFunc  func(collection string, filter map[string]any) ([]any, error)
		wantCount int
		wantErr   bool
	}{
		{
			name:     "successful get users by role",
			tenantID: "tenant1",
			roleID:   "role1",
			mockFunc: func(collection string, filter map[string]any) ([]any, error) {
				return []any{
					models.User{TenantID: "tenant1", Username: "user1"},
				}, nil
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			roleID:   "role1",
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
				FindAllFunc: tc.mockFunc,
			}
			repo := NewUserCollection(mockHandler)

			users, err := repo.GetUsersByRoleID(tc.tenantID, tc.roleID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, users, tc.wantCount)
			}
		})
	}
}

func TestUserCollection_UpdateUser(t *testing.T) {
	userID := primitive.NewObjectID()
	createdAt := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name       string
		user       models.User
		mockFind   func(collection string, filter map[string]any) (any, error)
		mockUpdate func(collection string, filter map[string]any, data any) error
		wantErr    bool
	}{
		{
			name: "successful update",
			user: models.User{
				ID:           userID,
				TenantID:     "tenant1",
				Email:        "updated@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				CreatedAt:    createdAt,
				Roles:        []models.UserRole{},
			},
			mockFind: func(collection string, filter map[string]any) (any, error) {
				return models.User{
					ID:        userID,
					TenantID:  "tenant1",
					Username:  "testuser",
					CreatedAt: createdAt,
				}, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "update with validation error",
			user: models.User{
				TenantID: "tenant1",
			},
			mockFind: func(collection string, filter map[string]any) (any, error) {
				return nil, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with user not found",
			user: models.User{
				ID:           userID,
				TenantID:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				CreatedAt:    createdAt,
				Roles:        []models.UserRole{},
			},
			mockFind: func(collection string, filter map[string]any) (any, error) {
				return nil, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with restricted field change - username",
			user: models.User{
				ID:           userID,
				TenantID:     "tenant1",
				Email:        "test@example.com",
				Username:     "newusername",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				CreatedAt:    createdAt,
				Roles:        []models.UserRole{},
			},
			mockFind: func(collection string, filter map[string]any) (any, error) {
				return models.User{
					ID:        userID,
					TenantID:  "tenant1",
					Username:  "testuser",
					CreatedAt: createdAt,
				}, nil
			},
			mockUpdate: func(collection string, filter map[string]any, data any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update with database error",
			user: models.User{
				ID:           userID,
				TenantID:     "tenant1",
				Email:        "test@example.com",
				Username:     "testuser",
				PasswordHash: "hashed_password",
				Status:       models.UserStatusActive,
				CreatedBy:    "admin",
				CreatedAt:    createdAt,
				Roles:        []models.UserRole{},
			},
			mockFind: func(collection string, filter map[string]any) (any, error) {
				return models.User{
					ID:        userID,
					TenantID:  "tenant1",
					Username:  "testuser",
					CreatedAt: createdAt,
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
				FindOneFunc: tc.mockFind,
				UpdateFunc:  tc.mockUpdate,
			}
			repo := NewUserCollection(mockHandler)

			err := repo.UpdateUser(tc.user)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserCollection_DeleteUser(t *testing.T) {
	testCases := []struct {
		name     string
		tenantID string
		userID   string
		mockFunc func(collection string, filter map[string]any) error
		wantErr  bool
	}{
		{
			name:     "successful delete",
			tenantID: "tenant1",
			userID:   "user-id-123",
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:     "delete with empty tenant ID",
			tenantID: "",
			userID:   "user-id-123",
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with empty user ID",
			tenantID: "tenant1",
			userID:   "",
			mockFunc: func(collection string, filter map[string]any) error {
				return nil
			},
			wantErr: true,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant1",
			userID:   "user-id-123",
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
			repo := NewUserCollection(mockHandler)

			err := repo.DeleteUser(tc.tenantID, tc.userID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
