package rbac

import (
	"errors"
	"testing"

	collection "erp.localhost/internal/auth/collections"
	"erp.localhost/internal/auth/models"
	common_models "erp.localhost/internal/common/models"
	mongo_mocks "erp.localhost/internal/db/mongo/mocks"
	"erp.localhost/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

// Tests for GetUserPermissions

func TestRBACManager_GetUserPermissions(t *testing.T) {
	tenantID := "tenant-1"
	userIDObjectID := primitive.NewObjectID()
	roleIDObjectID := primitive.NewObjectID()

	userID := userIDObjectID.String()
	roleID := roleIDObjectID.String()

	testCases := []struct {
		name                string
		tenantID            string
		userID              string
		mockUser            *models.User
		mockUserError       error
		mockRole            *models.Role
		mockRoleError       error
		expectedPermissions map[string]bool
		wantErr             bool
	}{
		{
			name:     "successful get user permissions with role permissions",
			tenantID: tenantID,
			userID:   userID,
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			expectedPermissions: map[string]bool{
				"user:create": true,
				"user:read":   true,
			},
			wantErr: false,
		},
		{
			name:     "user permissions with additional permissions",
			tenantID: tenantID,
			userID:   userID,
			mockUser: &models.User{
				ID:                    userIDObjectID,
				TenantID:              tenantID,
				Roles:                 []models.UserRole{},
				AdditionalPermissions: []string{"user:delete"},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			expectedPermissions: map[string]bool{
				"user:delete": true,
			},
			wantErr: false,
		},
		{
			name:     "user permissions with revoked permissions",
			tenantID: tenantID,
			userID:   userID,
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{"user:create"},
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			expectedPermissions: map[string]bool{
				"user:create": false,
				"user:read":   true,
			},
			wantErr: false,
		},
		{
			name:          "user not found",
			tenantID:      tenantID,
			userID:        userID,
			mockUserError: errors.New("user not found"),
			wantErr:       true,
		},
		{
			name:     "role not found",
			tenantID: tenantID,
			userID:   userID,
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
			},
			mockUserError: nil,
			mockRoleError: errors.New("role not found"),
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserHandler := mongo_mocks.NewMockCollectionHandler[models.User](ctrl)
			mockRoleHandler := mongo_mocks.NewMockCollectionHandler[models.Role](ctrl)

			userFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.userID,
			}
			mockUserHandler.EXPECT().
				FindOne(userFilter).
				Return(tc.mockUser, tc.mockUserError).
				Times(1)

			if tc.mockUserError == nil && tc.mockUser != nil && len(tc.mockUser.Roles) > 0 {
				roleFilter := map[string]any{
					"tenant_id": tc.tenantID,
					"_id":       tc.mockUser.Roles[0].RoleID,
				}
				mockRoleHandler.EXPECT().
					FindOne(roleFilter).
					Return(tc.mockRole, tc.mockRoleError).
					Times(1)
			}

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        collection.NewUserCollection(mockUserHandler),
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			permissions, err := rbacManager.GetUserPermissions(tc.tenantID, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, permissions)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedPermissions, permissions)
			}
		})
	}
}

// Tests for GetUserRoles

func TestRBACManager_GetUserRoles(t *testing.T) {
	tenantID := "tenant-1"
	userIDObjectID := primitive.NewObjectID()
	roleID1ObjectID := primitive.NewObjectID()
	roleID2ObjectID := primitive.NewObjectID()

	userID := userIDObjectID.String()
	roleID1 := roleID1ObjectID.String()
	roleID2 := roleID2ObjectID.String()

	testCases := []struct {
		name          string
		tenantID      string
		userID        string
		mockUser      *models.User
		mockUserError error
		expectedRoles []string
		wantErr       bool
	}{
		{
			name:     "successful get user roles",
			tenantID: tenantID,
			userID:   userID,
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID1},
					{RoleID: roleID2},
				},
			},
			mockUserError: nil,
			expectedRoles: []string{roleID1, roleID2},
			wantErr:       false,
		},
		{
			name:     "user with no roles",
			tenantID: tenantID,
			userID:   userID,
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles:    []models.UserRole{},
			},
			mockUserError: nil,
			expectedRoles: []string{},
			wantErr:       false,
		},
		{
			name:          "user not found",
			tenantID:      tenantID,
			userID:        userID,
			mockUserError: errors.New("user not found"),
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserHandler := mongo_mocks.NewMockCollectionHandler[models.User](ctrl)

			userFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.userID,
			}
			mockUserHandler.EXPECT().
				FindOne(userFilter).
				Return(tc.mockUser, tc.mockUserError).
				Times(1)

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        collection.NewUserCollection(mockUserHandler),
				rolesCollection:       nil,
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			roles, err := rbacManager.GetUserRoles(tc.tenantID, tc.userID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, roles)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRoles, roles)
			}
		})
	}
}

// Tests for GetRolePermissions

func TestRBACManager_GetRolePermissions(t *testing.T) {
	tenantID := "tenant-1"
	roleIDObjectID := primitive.NewObjectID()
	roleID := roleIDObjectID.String()

	testCases := []struct {
		name                string
		tenantID            string
		roleID              string
		mockRole            *models.Role
		mockRoleError       error
		expectedPermissions []string
		wantErr             bool
	}{
		{
			name:     "successful get role permissions",
			tenantID: tenantID,
			roleID:   roleID,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read", "user:update"},
			},
			mockRoleError:       nil,
			expectedPermissions: []string{"user:create", "user:read", "user:update"},
			wantErr:             false,
		},
		{
			name:     "role with no permissions",
			tenantID: tenantID,
			roleID:   roleID,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{},
			},
			mockRoleError:       nil,
			expectedPermissions: []string{},
			wantErr:             false,
		},
		{
			name:          "role not found",
			tenantID:      tenantID,
			roleID:        roleID,
			mockRoleError: errors.New("role not found"),
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleHandler := mongo_mocks.NewMockCollectionHandler[models.Role](ctrl)

			roleFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.roleID,
			}
			mockRoleHandler.EXPECT().
				FindOne(roleFilter).
				Return(tc.mockRole, tc.mockRoleError).
				Times(1)

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        nil,
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			permissions, err := rbacManager.GetRolePermissions(tc.tenantID, tc.roleID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, permissions)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedPermissions, permissions)
			}
		})
	}
}

// Tests for CheckUserPermissions

func TestRBACManager_CheckUserPermissions(t *testing.T) {
	tenantID := "tenant-1"
	userIDObjectID := primitive.NewObjectID()
	userID := userIDObjectID.String()
	roleIDObjectID := primitive.NewObjectID()
	roleID := roleIDObjectID.String()

	testCases := []struct {
		name           string
		tenantID       string
		userID         string
		permissions    []string
		mockUser       *models.User
		mockUserError  error
		mockRole       *models.Role
		mockRoleError  error
		expectedResult map[string]bool
		wantErr        bool
	}{
		{
			name:        "check existing permissions",
			tenantID:    tenantID,
			userID:      userID,
			permissions: []string{"user:create", "user:delete"},
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			expectedResult: map[string]bool{
				"user:create": true,
				"user:delete": false,
			},
			wantErr: false,
		},
		{
			name:        "check permissions with invalid format",
			tenantID:    tenantID,
			userID:      userID,
			permissions: []string{"invalid_permission_format"},
			wantErr:     true,
		},
		{
			name:          "user not found",
			tenantID:      tenantID,
			userID:        userID,
			permissions:   []string{"user:create"},
			mockUserError: errors.New("user not found"),
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserHandler := mongo_mocks.NewMockCollectionHandler[models.User](ctrl)
			mockRoleHandler := mongo_mocks.NewMockCollectionHandler[models.Role](ctrl)

			if tc.mockUserError != nil || (tc.mockUser != nil && models.IsValidPermissionFormat(tc.permissions[0])) {
				if models.IsValidPermissionFormat(tc.permissions[0]) {
					userFilter := map[string]any{
						"tenant_id": tc.tenantID,
						"_id":       tc.userID,
					}
					mockUserHandler.EXPECT().
						FindOne(userFilter).
						Return(tc.mockUser, tc.mockUserError).
						Times(1)

					if tc.mockUserError == nil && tc.mockUser != nil && len(tc.mockUser.Roles) > 0 {
						roleFilter := map[string]any{
							"tenant_id": tc.tenantID,
							"_id":       tc.mockUser.Roles[0].RoleID,
						}
						mockRoleHandler.EXPECT().
							FindOne(roleFilter).
							Return(tc.mockRole, tc.mockRoleError).
							Times(1)
					}
				}
			}

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        collection.NewUserCollection(mockUserHandler),
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			result, err := rbacManager.CheckUserPermissions(tc.tenantID, tc.userID, tc.permissions)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// Tests for VerifyUserRole

func TestRBACManager_VerifyUserRole(t *testing.T) {
	tenantID := "tenant-1"
	userIDObjectID := primitive.NewObjectID()
	userID := userIDObjectID.String()
	roleID1ObjectID := primitive.NewObjectID()
	roleID1 := roleID1ObjectID.String()
	roleID2ObjectID := primitive.NewObjectID()
	roleID2 := roleID2ObjectID.String()

	testCases := []struct {
		name           string
		tenantID       string
		userID         string
		roleID         string
		mockUser       *models.User
		mockUserError  error
		expectedResult bool
		wantErr        bool
	}{
		{
			name:     "user has role",
			tenantID: tenantID,
			userID:   userID,
			roleID:   roleID1,
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID1},
					{RoleID: roleID2},
				},
			},
			mockUserError:  nil,
			expectedResult: true,
			wantErr:        false,
		},
		{
			name:     "user does not have role",
			tenantID: tenantID,
			userID:   userID,
			roleID:   "00000000-0000-0000-0000-000000000099",
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID1},
				},
			},
			mockUserError:  nil,
			expectedResult: false,
			wantErr:        false,
		},
		{
			name:          "user not found",
			tenantID:      tenantID,
			userID:        userID,
			roleID:        roleID1,
			mockUserError: errors.New("user not found"),
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserHandler := mongo_mocks.NewMockCollectionHandler[models.User](ctrl)

			userFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.userID,
			}
			mockUserHandler.EXPECT().
				FindOne(userFilter).
				Return(tc.mockUser, tc.mockUserError).
				Times(1)

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        collection.NewUserCollection(mockUserHandler),
				rolesCollection:       nil,
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			result, err := rbacManager.VerifyUserRole(tc.tenantID, tc.userID, tc.roleID)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// Tests for VerifyRolePermissions

func TestRBACManager_VerifyRolePermissions(t *testing.T) {
	tenantID := "tenant-1"
	roleIDObjectID := primitive.NewObjectID()
	roleID := roleIDObjectID.String()

	testCases := []struct {
		name           string
		tenantID       string
		roleID         string
		permissions    []string
		mockRole       *models.Role
		mockRoleError  error
		expectedResult map[string]bool
		wantErr        bool
	}{
		{
			name:        "verify existing permissions",
			tenantID:    tenantID,
			roleID:      roleID,
			permissions: []string{"user:create", "user:delete"},
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			expectedResult: map[string]bool{
				"user:create": true,
				"user:delete": false,
			},
			wantErr: false,
		},
		{
			name:          "role not found",
			tenantID:      tenantID,
			roleID:        roleID,
			permissions:   []string{"user:create"},
			mockRoleError: errors.New("role not found"),
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleHandler := mongo_mocks.NewMockCollectionHandler[models.Role](ctrl)

			roleFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.roleID,
			}
			mockRoleHandler.EXPECT().
				FindOne(roleFilter).
				Return(tc.mockRole, tc.mockRoleError).
				Times(1)

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        nil,
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			result, err := rbacManager.VerifyRolePermissions(tc.tenantID, tc.roleID, tc.permissions)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// Tests for HasPermission

func TestRBACManager_HasPermission(t *testing.T) {
	tenantID := "tenant-1"
	userIDObjectID := primitive.NewObjectID()
	userID := userIDObjectID.String()
	roleIDObjectID := primitive.NewObjectID()
	roleID := roleIDObjectID.String()

	testCases := []struct {
		name          string
		tenantID      string
		userID        string
		permission    string
		mockUser      *models.User
		mockUserError error
		mockRole      *models.Role
		mockRoleError error
		wantErr       bool
		expectedErr   string
	}{
		{
			name:       "user has permission - returns nil",
			tenantID:   tenantID,
			userID:     userID,
			permission: "user:create",
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			wantErr:       false,
		},
		{
			name:       "user has permission via additional permissions - returns nil",
			tenantID:   tenantID,
			userID:     userID,
			permission: "user:delete",
			mockUser: &models.User{
				ID:                    userIDObjectID,
				TenantID:              tenantID,
				Roles:                 []models.UserRole{},
				AdditionalPermissions: []string{"user:delete"},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			wantErr:       false,
		},
		{
			name:       "user does not have permission - returns AuthPermissionDenied",
			tenantID:   tenantID,
			userID:     userID,
			permission: "user:delete",
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			wantErr:       true,
			expectedErr:   "don't have permission",
		},
		{
			name:       "user has revoked permission - returns AuthPermissionDenied",
			tenantID:   tenantID,
			userID:     userID,
			permission: "user:create",
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{"user:create"},
			},
			mockUserError: nil,
			mockRole: &models.Role{
				ID:          roleIDObjectID,
				TenantID:    tenantID,
				Permissions: []string{"user:create", "user:read"},
			},
			mockRoleError: nil,
			wantErr:       true,
			expectedErr:   "don't have permission",
		},
		{
			name:        "empty tenantID - returns validation error",
			tenantID:    "",
			userID:      userID,
			permission:  "user:create",
			wantErr:     true,
			expectedErr: "required",
		},
		{
			name:        "empty userID - returns validation error",
			tenantID:    tenantID,
			userID:      "",
			permission:  "user:create",
			wantErr:     true,
			expectedErr: "required",
		},
		{
			name:        "empty permission - returns validation error",
			tenantID:    tenantID,
			userID:      userID,
			permission:  "",
			wantErr:     true,
			expectedErr: "invalid permission format",
		},
		{
			name:        "invalid permission format - returns validation error",
			tenantID:    tenantID,
			userID:      userID,
			permission:  "invalid_format",
			wantErr:     true,
			expectedErr: "invalid permission format",
		},
		{
			name:          "user not found - returns error",
			tenantID:      tenantID,
			userID:        userID,
			permission:    "user:create",
			mockUserError: errors.New("user not found"),
			wantErr:       true,
			expectedErr:   "user not found",
		},
		{
			name:       "role not found - returns error",
			tenantID:   tenantID,
			userID:     userID,
			permission: "user:create",
			mockUser: &models.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []models.UserRole{
					{RoleID: roleID},
				},
			},
			mockUserError: nil,
			mockRoleError: errors.New("role not found"),
			wantErr:       true,
			expectedErr:   "role not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserHandler := mongo_mocks.NewMockCollectionHandler[models.User](ctrl)
			mockRoleHandler := mongo_mocks.NewMockCollectionHandler[models.Role](ctrl)

			// Setup mock expectations based on test case
			if tc.tenantID != "" && tc.userID != "" && tc.permission != "" && models.IsValidPermissionFormat(tc.permission) {
				userFilter := map[string]any{
					"tenant_id": tc.tenantID,
					"_id":       tc.userID,
				}
				mockUserHandler.EXPECT().
					FindOne(userFilter).
					Return(tc.mockUser, tc.mockUserError).
					Times(1)

				if tc.mockUserError == nil && tc.mockUser != nil && len(tc.mockUser.Roles) > 0 {
					roleFilter := map[string]any{
						"tenant_id": tc.tenantID,
						"_id":       tc.mockUser.Roles[0].RoleID,
					}
					mockRoleHandler.EXPECT().
						FindOne(roleFilter).
						Return(tc.mockRole, tc.mockRoleError).
						Times(1)
				}
			}

			rbacManager := &RBACManager{
				logger:                logging.NewLogger(common_models.ModuleAuth),
				userCollection:        collection.NewUserCollection(mockUserHandler),
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
				auditLogsCollection:   nil,
			}

			err := rbacManager.HasPermission(tc.tenantID, tc.userID, tc.permission)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErr != "" {
					assert.Contains(t, err.Error(), tc.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
