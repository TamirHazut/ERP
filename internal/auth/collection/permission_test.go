package collection

import (
	"errors"
	"testing"
	"time"

	mock_collection "erp.localhost/internal/infra/db/mongo/collection/mock"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_shared "erp.localhost/internal/infra/model/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

var (
	basePermissionLogger = logger.NewBaseLogger(model_shared.ModuleAuth)
)

// permissionMatcher is a custom gomock matcher for Permission objects
type permissionMatcher struct {
	expected *model_auth.Permission
}

func (m permissionMatcher) Matches(x interface{}) bool {
	perm, ok := x.(*model_auth.Permission)
	if !ok {
		return false
	}
	// Match fields except timestamps which are set by the function
	return perm.TenantID == m.expected.TenantID &&
		perm.Resource == m.expected.Resource &&
		perm.Action == m.expected.Action &&
		perm.PermissionString == m.expected.PermissionString &&
		perm.DisplayName == m.expected.DisplayName &&
		perm.CreatedBy == m.expected.CreatedBy
}

func (m permissionMatcher) String() string {
	return "matches permission fields"
}

func TestNewPermissionCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
	collection := NewPermissionCollection(mockHandler, basePermissionLogger)

	require.NotNil(t, collection)
	assert.NotNil(t, collection.collection)
	assert.NotNil(t, collection.logger)
}

func TestPermissionCollection_CreatePermission(t *testing.T) {
	testCases := []struct {
		name              string
		permission        *model_auth.Permission
		returnID          string
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name: "successful create",
			permission: &model_auth.Permission{
				TenantID:         "tenant1",
				Resource:         "products",
				Action:           "read",
				PermissionString: "products:read",
				DisplayName:      "Read Products",
				CreatedBy:        "admin",
			},
			returnID:          "permission-id-123",
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name: "create with validation error - missing tenant ID",
			permission: &model_auth.Permission{
				Resource:         "products",
				Action:           "read",
				PermissionString: "products:read",
				DisplayName:      "Read Products",
				CreatedBy:        "admin",
			},
			returnID:          "",
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name: "create with database error",
			permission: &model_auth.Permission{
				TenantID:         "tenant1",
				Resource:         "products",
				Action:           "read",
				PermissionString: "products:read",
				DisplayName:      "Read Products",
				CreatedBy:        "admin",
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

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Create(permissionMatcher{expected: tc.permission}).
					Return(tc.returnID, tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			id, err := collection.CreatePermission(tc.permission)
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

func TestPermissionCollection_GetPermissionByID(t *testing.T) {
	permissionID, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")

	testCases := []struct {
		name             string
		tenantID         string
		permissionID     string
		expectedFilter   map[string]any
		returnPermission *model_auth.Permission
		returnError      error
		wantErr          bool
	}{
		{
			name:         "successful get by id",
			tenantID:     "tenant1",
			permissionID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnPermission: &model_auth.Permission{
				ID:       permissionID,
				TenantID: "tenant1",
				Resource: "products",
				Action:   "read",
			},
			returnError: nil,
			wantErr:     false,
		},
		{
			name:         "permission not found",
			tenantID:     "tenant1",
			permissionID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnPermission: nil,
			returnError:      errors.New("permission not found"),
			wantErr:          true,
		},
		{
			name:         "database error",
			tenantID:     "tenant1",
			permissionID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnPermission: nil,
			returnError:      errors.New("database query failed"),
			wantErr:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			mockHandler.EXPECT().
				FindOne(tc.expectedFilter).
				Return(tc.returnPermission, tc.returnError)

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			permission, err := collection.GetPermissionByID(tc.tenantID, tc.permissionID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tenantID, permission.TenantID)
			}
		})
	}
}

func TestPermissionCollection_GetPermissionByName(t *testing.T) {
	testCases := []struct {
		name             string
		tenantID         string
		permissionName   string
		expectedFilter   map[string]any
		returnPermission *model_auth.Permission
		returnError      error
		wantErr          bool
	}{
		{
			name:           "successful get by name",
			tenantID:       "tenant1",
			permissionName: "Read Products",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"name":      "Read Products",
			},
			returnPermission: &model_auth.Permission{
				TenantID:    "tenant1",
				DisplayName: "Read Products",
				Resource:    "products",
				Action:      "read",
			},
			returnError: nil,
			wantErr:     false,
		},
		{
			name:           "permission not found",
			tenantID:       "tenant1",
			permissionName: "Nonexistent",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"name":      "Nonexistent",
			},
			returnPermission: nil,
			returnError:      errors.New("permission not found"),
			wantErr:          true,
		},
		{
			name:           "database error",
			tenantID:       "tenant1",
			permissionName: "Read Products",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"name":      "Read Products",
			},
			returnPermission: nil,
			returnError:      errors.New("database query failed"),
			wantErr:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			mockHandler.EXPECT().
				FindOne(tc.expectedFilter).
				Return(tc.returnPermission, tc.returnError)

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			permission, err := collection.GetPermissionByName(tc.tenantID, tc.permissionName)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.permissionName, permission.DisplayName)
			}
		})
	}
}

func TestPermissionCollection_GetPermissionsByTenantID(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		expectedFilter    map[string]any
		returnPermissions []*model_auth.Permission
		returnError       error
		wantCount         int
		wantErr           bool
	}{
		{
			name:     "successful get permissions by tenant",
			tenantID: "tenant1",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnPermissions: []*model_auth.Permission{
				&model_auth.Permission{TenantID: "tenant1", DisplayName: "Read Products"},
				&model_auth.Permission{TenantID: "tenant1", DisplayName: "Write Products"},
			},
			returnError: nil,
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:     "no permissions found",
			tenantID: "tenant1",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnPermissions: []*model_auth.Permission{},
			returnError:       nil,
			wantCount:         0,
			wantErr:           false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
			},
			returnPermissions: []*model_auth.Permission{},
			returnError:       errors.New("database query failed"),
			wantCount:         0,
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnPermissions, tc.returnError)

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			permissions, err := collection.GetPermissionsByTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionCollection_GetPermissionsByResource(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		resource          string
		expectedFilter    map[string]any
		returnPermissions []*model_auth.Permission
		returnError       error
		wantCount         int
		wantErr           bool
	}{
		{
			name:     "successful get by resource",
			tenantID: "tenant1",
			resource: "products",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"resource":  "products",
			},
			returnPermissions: []*model_auth.Permission{
				&model_auth.Permission{TenantID: "tenant1", Resource: "products", Action: "write"},
				&model_auth.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
			},
			returnError: nil,
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			resource: "products",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"resource":  "products",
			},
			returnPermissions: []*model_auth.Permission{},
			returnError:       errors.New("database query failed"),
			wantCount:         0,
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnPermissions, tc.returnError)

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			permissions, err := collection.GetPermissionsByResource(tc.tenantID, tc.resource)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionCollection_GetPermissionsByAction(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		action            string
		expectedFilter    map[string]any
		returnPermissions []*model_auth.Permission
		returnError       error
		wantCount         int
		wantErr           bool
	}{
		{
			name:     "successful get by action",
			tenantID: "tenant1",
			action:   "read",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"action":    "read",
			},
			returnPermissions: []*model_auth.Permission{
				&model_auth.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
				&model_auth.Permission{TenantID: "tenant1", Resource: "orders", Action: "read"},
			},
			returnError: nil,
			wantCount:   2,
			wantErr:     false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			action:   "read",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"action":    "read",
			},
			returnPermissions: []*model_auth.Permission{},
			returnError:       errors.New("database query failed"),
			wantCount:         0,
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnPermissions, tc.returnError)

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			permissions, err := collection.GetPermissionsByAction(tc.tenantID, tc.action)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionCollection_GetPermissionsByResourceAndAction(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		resource          string
		action            string
		expectedFilter    map[string]any
		returnPermissions []*model_auth.Permission
		returnError       error
		wantCount         int
		wantErr           bool
	}{
		{
			name:     "successful get by resource and action",
			tenantID: "tenant1",
			resource: "products",
			action:   "read",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"resource":  "products",
				"action":    "read",
			},
			returnPermissions: []*model_auth.Permission{
				&model_auth.Permission{TenantID: "tenant1", Resource: "products", Action: "read"},
			},
			returnError: nil,
			wantCount:   1,
			wantErr:     false,
		},
		{
			name:     "database error",
			tenantID: "tenant1",
			resource: "products",
			action:   "read",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"resource":  "products",
				"action":    "read",
			},
			returnPermissions: []*model_auth.Permission{},
			returnError:       errors.New("database query failed"),
			wantCount:         0,
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			mockHandler.EXPECT().
				FindAll(tc.expectedFilter).
				Return(tc.returnPermissions, tc.returnError)

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			permissions, err := collection.GetPermissionsByResourceAndAction(tc.tenantID, tc.resource, tc.action)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, permissions, tc.wantCount)
			}
		})
	}
}

func TestPermissionCollection_UpdatePermission(t *testing.T) {
	permissionID := primitive.NewObjectID()
	createdAt := time.Now().Add(-24 * time.Hour)

	testCases := []struct {
		name                 string
		permission           *model_auth.Permission
		expectedFindFilter   map[string]any
		returnFindPermission *model_auth.Permission
		returnFindError      error
		expectedUpdateFilter map[string]any
		returnUpdateError    error
		wantErr              bool
		expectedFindCalls    int
		expectedUpdateCalls  int
	}{
		{
			name: "successful update",
			permission: &model_auth.Permission{
				ID:               permissionID,
				TenantID:         "tenant1",
				Resource:         "products",
				Action:           "read",
				PermissionString: "products:read",
				DisplayName:      "Read Products Updated",
				CreatedBy:        "admin",
				CreatedAt:        createdAt,
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       permissionID.Hex(),
			},
			returnFindPermission: &model_auth.Permission{
				ID:        permissionID,
				TenantID:  "tenant1",
				CreatedAt: createdAt,
			},
			returnFindError: nil,
			expectedUpdateFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       permissionID,
			},
			returnUpdateError:   nil,
			wantErr:             false,
			expectedFindCalls:   1,
			expectedUpdateCalls: 1,
		},
		{
			name: "update with validation error",
			permission: &model_auth.Permission{
				TenantID: "tenant1",
			},
			expectedFindFilter:   nil,
			returnFindPermission: nil,
			returnFindError:      nil,
			expectedUpdateFilter: nil,
			returnUpdateError:    errors.New("validation error"),
			wantErr:              true,
			expectedFindCalls:    0,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with permission not found",
			permission: &model_auth.Permission{
				ID:               permissionID,
				TenantID:         "tenant1",
				Resource:         "products",
				Action:           "read",
				PermissionString: "products:read",
				DisplayName:      "Read Products",
				CreatedBy:        "admin",
				CreatedAt:        createdAt,
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       permissionID.Hex(),
			},
			returnFindPermission: nil,
			returnFindError:      errors.New("permission not found"),
			expectedUpdateFilter: nil,
			returnUpdateError:    errors.New("validation error"),
			wantErr:              true,
			expectedFindCalls:    1,
			expectedUpdateCalls:  0,
		},
		{
			name: "update with database error",
			permission: &model_auth.Permission{
				ID:               permissionID,
				TenantID:         "tenant1",
				Resource:         "products",
				Action:           "read",
				PermissionString: "products:read",
				DisplayName:      "Read Products",
				CreatedBy:        "admin",
				CreatedAt:        createdAt,
			},
			expectedFindFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       permissionID.Hex(),
			},
			returnFindPermission: &model_auth.Permission{
				ID:        permissionID,
				TenantID:  "tenant1",
				CreatedAt: createdAt,
			},
			returnFindError: nil,
			expectedUpdateFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       permissionID,
			},
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

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			if tc.expectedFindCalls > 0 {
				mockHandler.EXPECT().
					FindOne(tc.expectedFindFilter).
					Return(tc.returnFindPermission, tc.returnFindError).
					Times(tc.expectedFindCalls)
			}
			if tc.expectedUpdateCalls > 0 {
				mockHandler.EXPECT().
					Update(tc.expectedUpdateFilter, permissionMatcher{expected: tc.permission}).
					Return(tc.returnUpdateError).
					Times(tc.expectedUpdateCalls)
			}

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			err := collection.UpdatePermission(tc.permission)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPermissionCollection_DeletePermission(t *testing.T) {
	testCases := []struct {
		name              string
		tenantID          string
		permissionID      string
		expectedFilter    map[string]any
		returnError       error
		wantErr           bool
		expectedCallTimes int
	}{
		{
			name:         "successful delete",
			tenantID:     "tenant1",
			permissionID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
			},
			returnError:       nil,
			wantErr:           false,
			expectedCallTimes: 1,
		},
		{
			name:              "delete with empty tenant ID",
			tenantID:          "",
			permissionID:      "507f1f77bcf86cd799439011",
			expectedFilter:    nil,
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:              "delete with empty permission ID",
			tenantID:          "tenant1",
			permissionID:      "",
			expectedFilter:    nil,
			returnError:       nil,
			wantErr:           true,
			expectedCallTimes: 0,
		},
		{
			name:         "delete with database error",
			tenantID:     "tenant1",
			permissionID: "507f1f77bcf86cd799439011",
			expectedFilter: map[string]any{
				"tenant_id": "tenant1",
				"_id":       "507f1f77bcf86cd799439011",
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

			mockHandler := mock_collection.NewMockCollectionHandler[model_auth.Permission](ctrl)
			if tc.expectedCallTimes > 0 {
				mockHandler.EXPECT().
					Delete(tc.expectedFilter).
					Return(tc.returnError).
					Times(tc.expectedCallTimes)
			}

			collection := NewPermissionCollection(mockHandler, basePermissionLogger)
			err := collection.DeletePermission(tc.tenantID, tc.permissionID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
