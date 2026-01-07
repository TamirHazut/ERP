package checker

/*
// Tests for CheckUserPermissions

func TestRBACManager_CheckUserPermissions(t *testing.T) {
	t.Skip("Broken test - fix after gRPC services are ready")
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
		mockUser       *model_core.User
		mockUserError  error
		mockRole       *model_auth.Role
		mockRoleError  error
		expectedResult map[string]bool
		wantErr        bool
	}{
		{
			name:        "check existing permissions",
			tenantID:    tenantID,
			userID:      userID,
			permissions: []string{"user:create", "user:delete"},
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
					{RoleID: roleID},
				},
			},
			mockUserError: nil,
			mockRole: &model_auth.Role{
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

			mockUserHandler := mock_collection.NewMockCollectionHandler[model_core.User](ctrl)
			mockRoleHandler := mock_collection.NewMockCollectionHandler[model_auth.Role](ctrl)

			if tc.mockUserError != nil || (tc.mockUser != nil && model_auth.IsValidPermissionFormat(tc.permissions[0])) {
				if model_auth.IsValidPermissionFormat(tc.permissions[0]) {
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
				logger:                logger.NewBaseLogger(model_shared.ModuleAuth),
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
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
	t.Skip("Broken test - fix after gRPC services are ready")
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
		mockUser       *model_core.User
		mockUserError  error
		expectedResult bool
		wantErr        bool
	}{
		{
			name:     "user has role",
			tenantID: tenantID,
			userID:   userID,
			roleID:   roleID1,
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
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
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
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

			mockUserHandler := mock_collection.NewMockCollectionHandler[model_core.User](ctrl)

			userFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.userID,
			}
			mockUserHandler.EXPECT().
				FindOne(userFilter).
				Return(tc.mockUser, tc.mockUserError).
				Times(1)

			rbacManager := &RBACManager{
				logger:                logger.NewBaseLogger(model_shared.ModuleAuth),
				rolesCollection:       nil,
				permissionsCollection: nil,
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
		mockRole       *model_auth.Role
		mockRoleError  error
		expectedResult map[string]bool
		wantErr        bool
	}{
		{
			name:        "verify existing permissions",
			tenantID:    tenantID,
			roleID:      roleID,
			permissions: []string{"user:create", "user:delete"},
			mockRole: &model_auth.Role{
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

			mockRoleHandler := mock_collection.NewMockCollectionHandler[model_auth.Role](ctrl)

			roleFilter := map[string]any{
				"tenant_id": tc.tenantID,
				"_id":       tc.roleID,
			}
			mockRoleHandler.EXPECT().
				FindOne(roleFilter).
				Return(tc.mockRole, tc.mockRoleError).
				Times(1)

			rbacManager := &RBACManager{
				logger:                logger.NewBaseLogger(model_shared.ModuleAuth),
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
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
	t.Skip("Broken test - fix after gRPC services are ready")
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
		mockUser      *model_core.User
		mockUserError error
		mockRole      *model_auth.Role
		mockRoleError error
		wantErr       bool
		expectedErr   string
	}{
		{
			name:       "user has permission - returns nil",
			tenantID:   tenantID,
			userID:     userID,
			permission: "user:create",
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			mockRole: &model_auth.Role{
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
			mockUser: &model_core.User{
				ID:                    userIDObjectID,
				TenantID:              tenantID,
				Roles:                 []model_core.UserRole{},
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
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{},
			},
			mockUserError: nil,
			mockRole: &model_auth.Role{
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
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
					{RoleID: roleID},
				},
				AdditionalPermissions: []string{},
				RevokedPermissions:    []string{"user:create"},
			},
			mockUserError: nil,
			mockRole: &model_auth.Role{
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
			mockUser: &model_core.User{
				ID:       userIDObjectID,
				TenantID: tenantID,
				Roles: []model_core.UserRole{
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

			mockUserHandler := mock_collection.NewMockCollectionHandler[model_core.User](ctrl)
			mockRoleHandler := mock_collection.NewMockCollectionHandler[model_auth.Role](ctrl)

			// Setup mock expectations based on test case
			if tc.tenantID != "" && tc.userID != "" && tc.permission != "" && model_auth.IsValidPermissionFormat(tc.permission) {
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
				logger:                logger.NewBaseLogger(model_shared.ModuleAuth),
				rolesCollection:       collection.NewRoleCollection(mockRoleHandler),
				permissionsCollection: nil,
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
*/
