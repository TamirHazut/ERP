package collection

import (
	"errors"
	"testing"
	"time"

	mongo_mocks "erp.localhost/internal/infra/db/mongo/mocks"
	auth_models "erp.localhost/internal/infra/models/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

// roleMatcher is a custom gomock matcher for Role objects
// It skips the CreatedAt and UpdatedAt fields which are set dynamically
type roleMatcher struct {
	expected auth_models.Role
}

func (m roleMatcher) Matches(x interface{}) bool {
	role, ok := x.(auth_models.Role)
	if !ok {
		return false
	}
	// Match all fields except CreatedAt and UpdatedAt which are set by the function
	return role.TenantID == m.expected.TenantID &&
		role.Name == m.expected.Name &&
		role.Status == m.expected.Status &&
		role.CreatedBy == m.expected.CreatedBy &&
		len(role.Permissions) == len(m.expected.Permissions)
}

func (m roleMatcher) String() string {
	return "matches role fields except CreatedAt and UpdatedAt"
}

func TestNewRoleCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
	collection := NewRoleCollection(mockHandler)

	require.NotNil(t, collection)
	assert.NotNil(t, collection.collection)
	assert.NotNil(t, collection.logger)
}

func TestRoleCollection_CreateRole(t *testing.T) {
	testCases := []struct {
		name              string
		role              auth_models.Role
		returnID          string
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name: "successful create",
			role: auth_models.Role{
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"read", "write"},
			},
			returnID:          "role-id-123",
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name: "create with validation error - missing tenant ID",
			role: auth_models.Role{
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"read", "write"},
			},
			returnID:          "",
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name: "create with database error",
			role: auth_models.Role{
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				Permissions: []string{"read", "write"},
			},
			returnID:          "",
			returnError:       errors.New("database connection failed"),
			wantErr:           true,
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Create(roleMatcher{expected: tc.role}).
					Return(tc.returnID, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewRoleCollection(mockHandler)
			id, err := collection.CreateRole(tc.role)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, id)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.returnID, id)
			}
		})
	}
}

func TestRoleCollection_GetRoleByID(t *testing.T) {
	roleID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")

	testCases := []struct {
		name           string
		tenantID       string
		roleID         string
		expectedFilter map[string]any
		returnRole     *auth_models.Role
		returnError    error
		wantErr        bool
	}{
		{
			name:     "successful get by id",
			tenantID: "tenant1",
			roleID:   "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnRole: &auth_models.Role{
				ID:       roleID,
				TenantID: "tenant1",
				Name:     "Admin",
			},
			returnError: nil,
			wantErr:     false,
		},
		{
			name:     "role not found",
			tenantID: "tenant1",
			roleID:   "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnRole:  nil,
			returnError: errors.New("role not found"),
			wantErr:     true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			roleID:   "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnRole:  nil,
			returnError: errors.New("database query failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			mockHandler.EXPECT().
				FindOne(tc.expectedFilter).
				Return(tc.returnRole, tc.returnError)

			collection := NewRoleCollection(mockHandler)
			role, err := collection.GetRoleByID(tc.tenantID, tc.roleID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, role.TenantID)
			}
		})
	}
}

func TestRoleCollection_GetRoleByName(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		roleName       string
		expectedFilter map[string]any
		returnRole     *auth_models.Role
		returnError    error
		wantErr        bool
	}{
		{
			name:     "successful get by name",
			tenantID: "tenant1",
			roleName: "Admin",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"name":      "Admin",
			},
			returnRole: &auth_models.Role{
				TenantID: "tenant1",
				Name:     "Admin",
			},
			returnError: nil,
			wantErr:     false,
		},
		{
			name:     "role not found",
			tenantID: "tenant1",
			roleName: "Nonexistent",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"name":      "Nonexistent",
			},
			returnRole:  nil,
			returnError: errors.New("role not found"),
			wantErr:     true,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			roleName: "Admin",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"name":      "Admin",
			},
			returnRole:  nil,
			returnError: errors.New("database query failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			mockHandler.EXPECT().
				FindOne(tc.expectedFilter).
				Return(tc.returnRole, tc.returnError)

			collection := NewRoleCollection(mockHandler)
			role, err := collection.GetRoleByName(tc.tenantID, tc.roleName)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.roleName, role.Name)
			}
		})
	}
}

func TestRoleCollection_GetRolesByTenantID(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		expectedFilter map[string]any
		returnRoles    []auth_models.Role
		returnError    error
		wantCount      int
		wantErr        bool
	}{
		{
			name:     "successful get roles by tenant",
			tenantID: "tenant1",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnRoles: []auth_models.Role{
				auth_models.Role{TenantID: "tenant1", Name: "Admin"},
				auth_models.Role{TenantID: "tenant1", Name: "User"},
			},
			returnError: nil,
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:     "no roles found",
			tenantID: "tenant1",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnRoles: []auth_models.Role{},
			returnError: nil,
			wantCount:   0,
			wantErr:     false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnRoles: []auth_models.Role{},
			returnError: errors.New("database query failed"),
			wantCount:   0,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnRoles, tc.returnError)

			collection := NewRoleCollection(mockHandler)
			roles, err := collection.GetRolesByTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, roles, tc.wantCount)
			}
		})
	}
}

func TestRoleCollection_GetRolesByPermissionsIDs(t *testing.T) {
	testCases := []struct {
		name           string
		tenantID       string
		permissionsIDs []string
		expectedFilter map[string]any
		returnRoles    []auth_models.Role
		returnError    error
		wantCount      int
		wantErr        bool
	}{
		{
			name:           "successful get roles by permissions",
			tenantID:       "tenant1",
			permissionsIDs: []string{"perm1", "perm2"},
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"permissions": map[string]any{
					"$all": []string{"perm1", "perm2"},
				},
			},
			returnRoles: []auth_models.Role{
				auth_models.Role{TenantID: "tenant1", Name: "Admin"},
			},
			returnError: nil,
			wantCount:   1,
			wantErr:     false,
		},
		{
			name:           "database error",
			tenantID:       "tenant1",
			permissionsIDs: []string{"perm1"},
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"permissions": map[string]any{
					"$all": []string{"perm1"},
				},
			},
			returnRoles: []auth_models.Role{},
			returnError: errors.New("database query failed"),
			wantCount:   0,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnRoles, tc.returnError)

			collection := NewRoleCollection(mockHandler)
			roles, err := collection.GetRolesByPermissionsIDs(tc.tenantID, tc.permissionsIDs)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, roles, tc.wantCount)
			}
		})
	}
}

func TestRoleCollection_UpdateRole(t *testing.T) {
	roleID := primitive.NewObjectID()
	createdAt := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name                 string
		role                 auth_models.Role
		expectedFindFilter   map[string]any
		expectedUpdateFilter map[string]any
		returnFindRole       *auth_models.Role
		returnFindError      error
		returnUpdateError    error
		wantErr              bool
		expectedFindCalls    int
		expectedUpdateCalls  int
	}{
		{
			name: "successful update",
			role: auth_models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Updated Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   createdAt,
				Permissions: []string{"read", "write"},
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       roleID.String(),
			},
			expectedUpdateFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       roleID,
			},
			returnFindRole: &auth_models.Role{
				ID:        roleID,
				TenantID:  "tenant1",
				Name:      "Admin",
				CreatedAt: createdAt,
			},
			returnFindError:     nil,
			returnUpdateError:   nil,
			wantErr:             false,
			expectedFindCalls:   1,
			expectedUpdateCalls: 1,
		},
		{
			name: "update with validation error",
			role: auth_models.Role{
				TenantID: "tenant1",
			},
			expectedFindFilter:   map[string]any{},
			expectedUpdateFilter: map[string]any{},
			returnFindRole:       nil,
			returnFindError:      nil,
			returnUpdateError:    nil,
			wantErr:              true,
			expectedFindCalls:    0,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with role not found",
			role: auth_models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   createdAt,
				Permissions: []string{"read", "write"},
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       roleID.String(),
			},
			expectedUpdateFilter: map[string]any{},
			returnFindRole:       nil,
			returnFindError:      errors.New("role not found"),
			returnUpdateError:    errors.New("validation error"),
			wantErr:              true,
			expectedFindCalls:    1,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with restricted field change - CreatedAt",
			role: auth_models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   time.Now(),
				Permissions: []string{"read", "write"},
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       roleID.String(),
			},
			expectedUpdateFilter: map[string]any{},
			returnFindRole: &auth_models.Role{
				ID:        roleID,
				TenantID:  "tenant1",
				CreatedAt: createdAt,
			},
			returnFindError:     nil,
			returnUpdateError:   nil,
			wantErr:             true,
			expectedFindCalls:   1,
			expectedUpdateCalls: 0,
		},
		{
			name: "update with database error",
			role: auth_models.Role{
				ID:          roleID,
				TenantID:    "tenant1",
				Name:        "Admin",
				Status:      "active",
				CreatedBy:   "admin",
				CreatedAt:   createdAt,
				Permissions: []string{"read", "write"},
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       roleID.String(),
			},
			expectedUpdateFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       roleID,
			},
			returnFindRole: &auth_models.Role{
				ID:        roleID,
				TenantID:  "tenant1",
				CreatedAt: createdAt,
			},
			returnFindError:     nil,
			returnUpdateError:   errors.New("update failed"),
			wantErr:             true,
			expectedFindCalls:   1,
			expectedUpdateCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			if tc.expectedFindCalls > 0 {
				mockHandler.EXPECT().
					FindOne(tc.expectedFindFilter).
					Return(tc.returnFindRole, tc.returnFindError).
					Times(tc.expectedFindCalls)
			}
			if tc.expectedUpdateCalls > 0 {
				mockHandler.EXPECT().
					Update(tc.expectedUpdateFilter, roleMatcher{expected: tc.role}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCalls)
			}

			collection := NewRoleCollection(mockHandler)
			err := collection.UpdateRole(tc.role)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRoleCollection_DeleteRole(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		roleID            string
		expectedFilter    map[string]any
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name:     "successful delete",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "role-id-123",
			},
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name:              "delete with empty tenant ID",
			tenantID:          "",
			roleID:            "role-id-123",
			expectedFilter:    map[string]any{},
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:              "delete with empty role ID",
			tenantID:          "tenant1",
			roleID:            "",
			expectedFilter:    map[string]any{},
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:     "delete with database error",
			tenantID: "tenant1",
			roleID:   "role-id-123",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "role-id-123",
			},
			returnError:       errors.New("delete failed"),
			wantErr:           true,
			expectedCallTimes: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mongo_mocks.NewMockCollectionHandler[auth_models.Role](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedFilter).
					Return(tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewRoleCollection(mockHandler)
			err := collection.DeleteRole(tc.tenantID, tc.roleID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
