package convertor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
)

// =============================================================================
// Test Nested Converters (Domain → Proto)
// =============================================================================

func TestUserProfileToProto(t *testing.T) {
	t.Run("valid profile with all fields", func(t *testing.T) {
		profile := &model_auth.UserProfile{
			FirstName:   "John",
			LastName:    "Doe",
			DisplayName: "John Doe",
			AvatarURL:   "https://example.com/avatar.jpg",
			Phone:       "+1234567890",
			Title:       "Software Engineer",
			Department:  "Engineering",
		}

		proto := UserProfileToProto(profile)

		require.NotNil(t, proto)
		assert.Equal(t, "John", proto.FirstName)
		assert.Equal(t, "Doe", proto.LastName)
		assert.Equal(t, "John Doe", proto.DisplayName)
		assert.Equal(t, "https://example.com/avatar.jpg", proto.AvatarUrl)
		assert.Equal(t, "+1234567890", proto.Phone)
		assert.Equal(t, "Software Engineer", proto.Title)
		assert.Equal(t, "Engineering", proto.Department)
	})

	t.Run("nil profile", func(t *testing.T) {
		proto := UserProfileToProto(nil)
		assert.Nil(t, proto)
	})

	t.Run("profile with optional fields empty", func(t *testing.T) {
		profile := &model_auth.UserProfile{
			FirstName:   "John",
			LastName:    "Doe",
			DisplayName: "John Doe",
		}

		proto := UserProfileToProto(profile)

		require.NotNil(t, proto)
		assert.Equal(t, "John", proto.FirstName)
		assert.Equal(t, "Doe", proto.LastName)
		assert.Equal(t, "John Doe", proto.DisplayName)
		assert.Empty(t, proto.AvatarUrl)
		assert.Empty(t, proto.Phone)
		assert.Empty(t, proto.Title)
		assert.Empty(t, proto.Department)
	})
}

func TestUserRoleToProto(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("valid role with ExpiresAt", func(t *testing.T) {
		role := &model_auth.UserRole{
			RoleID:     "role-123",
			TenantID:   "tenant-123",
			AssignedAt: now,
			AssignedBy: "admin",
			ExpiresAt:  &expiresAt,
		}

		proto := UserRoleToProto(role)

		require.NotNil(t, proto)
		assert.Equal(t, "role-123", proto.RoleId)
		assert.Equal(t, "tenant-123", proto.TenantId)
		assert.NotNil(t, proto.AssignedAt)
		assert.Equal(t, "admin", proto.AssignedBy)
		assert.NotNil(t, proto.ExpiresAt)
	})

	t.Run("valid role without ExpiresAt", func(t *testing.T) {
		role := &model_auth.UserRole{
			RoleID:     "role-123",
			TenantID:   "tenant-123",
			AssignedAt: now,
			AssignedBy: "admin",
			ExpiresAt:  nil,
		}

		proto := UserRoleToProto(role)

		require.NotNil(t, proto)
		assert.Equal(t, "role-123", proto.RoleId)
		assert.Equal(t, "tenant-123", proto.TenantId)
		assert.NotNil(t, proto.AssignedAt)
		assert.Equal(t, "admin", proto.AssignedBy)
		assert.Nil(t, proto.ExpiresAt)
	})

	t.Run("nil role", func(t *testing.T) {
		proto := UserRoleToProto(nil)
		assert.Nil(t, proto)
	})
}

func TestUserRolesToProto(t *testing.T) {
	now := time.Now()

	t.Run("valid slice with multiple roles", func(t *testing.T) {
		roles := []model_auth.UserRole{
			{
				RoleID:     "role-1",
				TenantID:   "tenant-123",
				AssignedAt: now,
				AssignedBy: "admin",
			},
			{
				RoleID:     "role-2",
				TenantID:   "tenant-123",
				AssignedAt: now,
				AssignedBy: "admin",
			},
		}

		protoRoles := UserRolesToProto(roles)

		require.Len(t, protoRoles, 2)
		assert.Equal(t, "role-1", protoRoles[0].RoleId)
		assert.Equal(t, "role-2", protoRoles[1].RoleId)
	})

	t.Run("empty slice", func(t *testing.T) {
		roles := []model_auth.UserRole{}
		protoRoles := UserRolesToProto(roles)
		assert.Empty(t, protoRoles)
	})

	t.Run("nil slice", func(t *testing.T) {
		protoRoles := UserRolesToProto(nil)
		assert.Empty(t, protoRoles)
	})
}

func TestNotificationSettingsToProto(t *testing.T) {
	t.Run("valid settings", func(t *testing.T) {
		settings := &model_auth.NotificationSettings{
			Email: true,
			Push:  false,
			SMS:   true,
		}

		proto := NotificationSettingsToProto(settings)

		require.NotNil(t, proto)
		assert.True(t, proto.Email)
		assert.False(t, proto.Push)
		assert.True(t, proto.Sms)
	})

	t.Run("nil settings return defaults", func(t *testing.T) {
		proto := NotificationSettingsToProto(nil)

		require.NotNil(t, proto)
		assert.False(t, proto.Email)
		assert.False(t, proto.Push)
		assert.False(t, proto.Sms)
	})
}

func TestUserPreferencesToProto(t *testing.T) {
	t.Run("valid preferences with all fields", func(t *testing.T) {
		prefs := &model_auth.UserPreferences{
			Language: "en",
			Timezone: "UTC",
			Theme:    "dark",
			Notifications: model_auth.NotificationSettings{
				Email: true,
				Push:  true,
				SMS:   false,
			},
		}

		proto := UserPreferencesToProto(prefs)

		require.NotNil(t, proto)
		assert.Equal(t, "en", proto.Language)
		assert.Equal(t, "UTC", proto.Timezone)
		assert.Equal(t, "dark", proto.Theme)
		require.NotNil(t, proto.Notifications)
		assert.True(t, proto.Notifications.Email)
		assert.True(t, proto.Notifications.Push)
		assert.False(t, proto.Notifications.Sms)
	})

	t.Run("nil preferences return defaults", func(t *testing.T) {
		proto := UserPreferencesToProto(nil)

		require.NotNil(t, proto)
		assert.Equal(t, "en", proto.Language)
		assert.Equal(t, "UTC", proto.Timezone)
		assert.Equal(t, "light", proto.Theme)
		require.NotNil(t, proto.Notifications)
	})
}

// =============================================================================
// Test Main User Converter (Domain → Proto)
// =============================================================================

func TestUserToProto_ValidUser(t *testing.T) {
	now := time.Now()
	objectID := primitive.NewObjectID()
	lastLogin := now.Add(-1 * time.Hour)

	user := &model_auth.User{
		ID:           objectID,
		TenantID:     "tenant-123",
		PasswordHash: "x",
		Email:        "user@example.com",
		Username:     "testuser",
		Profile: model_auth.UserProfile{
			FirstName:   "John",
			LastName:    "Doe",
			DisplayName: "John Doe",
		},
		Roles: []model_auth.UserRole{
			{
				RoleID:     "role-123",
				TenantID:   "tenant-123",
				AssignedAt: now,
				AssignedBy: "admin",
			},
		},
		AdditionalPermissions: []string{"perm1", "perm2"},
		RevokedPermissions:    []string{"perm3"},
		Status:                model_auth.UserStatusActive,
		EmailVerified:         true,
		PhoneVerified:         false,
		MFAEnabled:            true,
		LastLogin:             &lastLogin,
		LastPasswordChange:    now,
		Preferences: model_auth.UserPreferences{
			Language: "en",
			Timezone: "UTC",
			Theme:    "dark",
			Notifications: model_auth.NotificationSettings{
				Email: true,
				Push:  false,
				SMS:   false,
			},
		},
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    "admin",
		LastActivity: now,
		LoginHistory: []model_auth.LoginRecord{
			{
				Timestamp: now,
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
				Success:   true,
			},
		},
	}

	proto := UserToProto(user)

	require.NotNil(t, proto)
	assert.Equal(t, objectID.Hex(), proto.Id)
	assert.Equal(t, "tenant-123", proto.TenantId)
	assert.Equal(t, "user@example.com", proto.Email)
	assert.Equal(t, "testuser", proto.Username)
	assert.NotNil(t, proto.Profile)
	assert.Len(t, proto.Roles, 1)
	assert.Equal(t, []string{"perm1", "perm2"}, proto.AdditionalPermissions)
	assert.Equal(t, []string{"perm3"}, proto.RevokedPermissions)
	assert.Equal(t, model_auth.UserStatusActive, proto.Status)
	assert.True(t, proto.EmailVerified)
	assert.False(t, proto.PhoneVerified)
	assert.True(t, proto.MfaEnabled)
	assert.NotNil(t, proto.LastLogin)
	assert.NotNil(t, proto.LastPasswordChange)
	assert.NotNil(t, proto.Preferences)
	assert.NotNil(t, proto.CreatedAt)
	assert.NotNil(t, proto.UpdatedAt)
	assert.Equal(t, "admin", proto.CreatedBy)
	assert.NotNil(t, proto.LastActivity)
}

func TestUserToProto_NilUser(t *testing.T) {
	proto := UserToProto(nil)
	assert.Nil(t, proto)
}

func TestUserToProto_InvalidUser(t *testing.T) {
	// User with missing required fields (will fail Validate())
	user := &model_auth.User{
		ID: primitive.NewObjectID(),
		// Missing TenantID, Email/Username
	}

	proto := UserToProto(user)
	assert.Nil(t, proto)
}

func TestUserToProto_MinimalUser(t *testing.T) {
	now := time.Now()
	objectID := primitive.NewObjectID()

	user := &model_auth.User{
		ID:                 objectID,
		TenantID:           "tenant-123",
		Email:              "user@example.com",
		PasswordHash:       "hash",
		Status:             model_auth.UserStatusInvited,
		LastPasswordChange: now,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          "admin",
		LastActivity:       now,
	}

	proto := UserToProto(user)

	require.NotNil(t, proto)
	assert.Equal(t, objectID.Hex(), proto.Id)
	assert.Equal(t, "tenant-123", proto.TenantId)
	assert.Equal(t, "user@example.com", proto.Email)
	assert.Empty(t, proto.Username)
	assert.Nil(t, proto.LastLogin)
}

func TestUsersToProto(t *testing.T) {
	now := time.Now()
	objectID1 := primitive.NewObjectID()
	objectID2 := primitive.NewObjectID()

	t.Run("valid slice with multiple users", func(t *testing.T) {
		users := []*model_auth.User{
			{
				ID:                 objectID1,
				TenantID:           "tenant-123",
				Email:              "user1@example.com",
				PasswordHash:       "hash",
				Status:             model_auth.UserStatusActive,
				LastPasswordChange: now,
				CreatedAt:          now,
				UpdatedAt:          now,
				CreatedBy:          "admin",
				LastActivity:       now,
			},
			{
				ID:                 objectID2,
				TenantID:           "tenant-123",
				Email:              "user2@example.com",
				PasswordHash:       "hash",
				Status:             model_auth.UserStatusActive,
				LastPasswordChange: now,
				CreatedAt:          now,
				UpdatedAt:          now,
				CreatedBy:          "admin",
				LastActivity:       now,
			},
		}

		protoUsers := UsersToProto(users)

		require.Len(t, protoUsers, 2)
		assert.Equal(t, objectID1.Hex(), protoUsers[0].Id)
		assert.Equal(t, objectID2.Hex(), protoUsers[1].Id)
	})

	t.Run("empty slice", func(t *testing.T) {
		users := []*model_auth.User{}
		protoUsers := UsersToProto(users)
		assert.Empty(t, protoUsers)
	})

	t.Run("nil slice", func(t *testing.T) {
		protoUsers := UsersToProto(nil)
		assert.Empty(t, protoUsers)
	})

	t.Run("slice with nil user skips nil", func(t *testing.T) {
		users := []*model_auth.User{
			{
				ID:                 objectID1,
				TenantID:           "tenant-123",
				Email:              "user1@example.com",
				PasswordHash:       "hash",
				Status:             model_auth.UserStatusActive,
				LastPasswordChange: now,
				CreatedAt:          now,
				UpdatedAt:          now,
				CreatedBy:          "admin",
				LastActivity:       now,
			},
			nil,
		}

		protoUsers := UsersToProto(users)

		require.Len(t, protoUsers, 1)
		assert.Equal(t, objectID1.Hex(), protoUsers[0].Id)
	})
}

// =============================================================================
// Test Create Converter (Proto → Domain)
// =============================================================================

func TestCreateUserFromProto_ValidProto(t *testing.T) {
	proto := &proto_auth.CreateUserData{
		TenantId:     "tenant-123",
		Email:        "user@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		RoleIds:      []string{"role-1", "role-2"},
		Status:       model_auth.UserStatusActive,
		CreatedBy:    "admin",
	}

	user, err := CreateUserFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "tenant-123", user.TenantID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Len(t, user.Roles, 2)
	assert.Equal(t, "role-1", user.Roles[0].RoleID)
	assert.Equal(t, "role-2", user.Roles[1].RoleID)
	assert.Equal(t, model_auth.UserStatusActive, user.Status)
	assert.False(t, user.EmailVerified)
	assert.False(t, user.PhoneVerified)
	assert.False(t, user.MFAEnabled)
	assert.Empty(t, user.RevokedPermissions)
	assert.NotNil(t, user.CreatedAt)
	assert.NotNil(t, user.UpdatedAt)
	assert.Equal(t, "admin", user.CreatedBy)
}

func TestCreateUserFromProto_WithOptionalFields(t *testing.T) {
	proto := &proto_auth.CreateUserData{
		TenantId:     "tenant-123",
		Email:        "user@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Profile: &proto_auth.UserProfileData{
			FirstName:   "John",
			LastName:    "Doe",
			DisplayName: "John Doe",
		},
		Preferences: &proto_auth.UserPreferencesData{
			Language: "es",
			Timezone: "America/New_York",
			Theme:    "dark",
		},
		AdditionalPermissions: []string{"perm1", "perm2"},
		Status:                model_auth.UserStatusActive,
		CreatedBy:             "admin",
	}

	user, err := CreateUserFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "John", user.Profile.FirstName)
	assert.Equal(t, "Doe", user.Profile.LastName)
	assert.Equal(t, "es", user.Preferences.Language)
	assert.Equal(t, "America/New_York", user.Preferences.Timezone)
	assert.Equal(t, "dark", user.Preferences.Theme)
	assert.Equal(t, []string{"perm1", "perm2"}, user.AdditionalPermissions)
}

func TestCreateUserFromProto_WithoutOptionalFields(t *testing.T) {
	proto := &proto_auth.CreateUserData{
		TenantId:     "tenant-123",
		Email:        "user@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		CreatedBy:    "admin",
	}

	user, err := CreateUserFromProto(proto)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Empty(t, user.Profile.FirstName)
	assert.Equal(t, "en", user.Preferences.Language)
	assert.Equal(t, "UTC", user.Preferences.Timezone)
	assert.Equal(t, "light", user.Preferences.Theme)
	assert.Equal(t, model_auth.UserStatusInvited, user.Status) // default
}

func TestCreateUserFromProto_NilProto(t *testing.T) {
	user, err := CreateUserFromProto(nil)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "proto")
}

// =============================================================================
// Test Update Converter (Proto → Domain)
// =============================================================================

func TestUpdateUserFromProto_FullUpdate(t *testing.T) {
	now := time.Now()
	existing := &model_auth.User{
		ID:                 primitive.NewObjectID(),
		TenantID:           "tenant-123",
		PasswordHash:       "x",
		Email:              "old@example.com",
		Username:           "olduser",
		Status:             model_auth.UserStatusInvited,
		EmailVerified:      false,
		PhoneVerified:      false,
		LastPasswordChange: now,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          "admin",
		LastActivity:       now,
	}

	newStatus := model_auth.UserStatusActive
	emailVerified := true
	phoneVerified := true

	proto := &proto_auth.UpdateUserData{
		Id:            existing.ID.Hex(),
		TenantId:      existing.TenantID,
		Status:        &newStatus,
		EmailVerified: &emailVerified,
		PhoneVerified: &phoneVerified,
	}

	// sleep before test to let time pass for UpdateAt check
	time.Sleep(time.Second)

	err := UpdateUserFromProto(existing, proto)

	require.NoError(t, err)
	assert.Equal(t, model_auth.UserStatusActive, existing.Status)
	assert.True(t, existing.EmailVerified)
	assert.True(t, existing.PhoneVerified)
	assert.True(t, existing.UpdatedAt.After(now))
}

func TestUpdateUserFromProto_PartialUpdate(t *testing.T) {
	now := time.Now()
	existing := &model_auth.User{
		ID:                 primitive.NewObjectID(),
		TenantID:           "tenant-123",
		Email:              "old@example.com",
		Username:           "olduser",
		Status:             model_auth.UserStatusInvited,
		LastPasswordChange: now,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          "admin",
		LastActivity:       now,
	}

	proto := &proto_auth.UpdateUserData{
		Id:       existing.ID.Hex(),
		TenantId: existing.TenantID,
		Profile:  &proto_auth.UserProfileData{},
	}

	err := UpdateUserFromProto(existing, proto)

	require.NoError(t, err)
	assert.Equal(t, "old@example.com", existing.Email)
	assert.Equal(t, "olduser", existing.Username)                  // Unchanged
	assert.Equal(t, model_auth.UserStatusInvited, existing.Status) // Unchanged
	assert.NotNil(t, existing.Profile)
}

func TestUpdateUserFromProto_UpdateRoles(t *testing.T) {
	now := time.Now()
	existing := &model_auth.User{
		ID:       primitive.NewObjectID(),
		TenantID: "tenant-123",
		Email:    "user@example.com",
		Roles: []model_auth.UserRole{
			{RoleID: "role-1", TenantID: "tenant-123", AssignedAt: now, AssignedBy: "admin"},
			{RoleID: "role-2", TenantID: "tenant-123", AssignedAt: now, AssignedBy: "admin"},
		},
		LastPasswordChange: now,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          "admin",
		LastActivity:       now,
	}

	proto := &proto_auth.UpdateUserData{
		Id:       existing.ID.Hex(),
		TenantId: existing.TenantID,
		Roles: &proto_auth.UpdateRolesData{
			AddRoleIds:    []string{"role-3"},
			RemoveRoleIds: []string{"role-1"},
		},
	}

	err := UpdateUserFromProto(existing, proto)

	require.NoError(t, err)
	assert.Len(t, existing.Roles, 2) // 2 original - 1 removed + 1 added = 2
	roleIDs := make([]string, len(existing.Roles))
	for i, role := range existing.Roles {
		roleIDs[i] = role.RoleID
	}
	assert.Contains(t, roleIDs, "role-2")
	assert.Contains(t, roleIDs, "role-3")
	assert.NotContains(t, roleIDs, "role-1")
}

func TestUpdateUserFromProto_UpdatePermissions(t *testing.T) {
	now := time.Now()
	existing := &model_auth.User{
		ID:                    primitive.NewObjectID(),
		TenantID:              "tenant-123",
		Email:                 "user@example.com",
		AdditionalPermissions: []string{"perm1", "perm2"},
		RevokedPermissions:    []string{"perm3"},
		LastPasswordChange:    now,
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             "admin",
		LastActivity:          now,
	}

	proto := &proto_auth.UpdateUserData{
		Id:       existing.ID.Hex(),
		TenantId: existing.TenantID,
		Permissions: &proto_auth.UpdatePermissionsData{
			AddPermissions:      []string{"perm4"},
			RemovePermissions:   []string{"perm1"},
			RevokePermissions:   []string{"perm5"},
			UnrevokePermissions: []string{"perm3"},
		},
	}

	err := UpdateUserFromProto(existing, proto)

	require.NoError(t, err)
	assert.Contains(t, existing.AdditionalPermissions, "perm2")
	assert.Contains(t, existing.AdditionalPermissions, "perm4")
	assert.NotContains(t, existing.AdditionalPermissions, "perm1")
	assert.Contains(t, existing.RevokedPermissions, "perm5")
	assert.NotContains(t, existing.RevokedPermissions, "perm3")
}

func TestUpdateUserFromProto_NilExisting(t *testing.T) {
	proto := &proto_auth.UpdateUserData{
		Id:       primitive.NewObjectID().Hex(),
		TenantId: "tenant-123",
	}

	err := UpdateUserFromProto(nil, proto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing")
}

func TestUpdateUserFromProto_NilProto(t *testing.T) {
	now := time.Now()
	existing := &model_auth.User{
		ID:                 primitive.NewObjectID(),
		TenantID:           "tenant-123",
		Email:              "user@example.com",
		LastPasswordChange: now,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          "admin",
		LastActivity:       now,
	}

	err := UpdateUserFromProto(existing, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "proto")
}

func TestUpdateUserFromProto_BothNil(t *testing.T) {
	err := UpdateUserFromProto(nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "existing")
}

// =============================================================================
// Test Helper Functions
// =============================================================================

func TestUserObjectIDFromString_ValidHex(t *testing.T) {
	objectID := primitive.NewObjectID()
	hexString := objectID.Hex()

	parsedID, err := UserObjectIDFromString(hexString)

	require.NoError(t, err)
	assert.Equal(t, objectID, parsedID)
	assert.Equal(t, hexString, parsedID.Hex())
}

func TestUserObjectIDFromString_EmptyString(t *testing.T) {
	parsedID, err := UserObjectIDFromString("")

	assert.Error(t, err)
	assert.Equal(t, primitive.NilObjectID, parsedID)
	assert.Contains(t, err.Error(), "id")
}

func TestUserObjectIDFromString_InvalidHex(t *testing.T) {
	_, err := UserObjectIDFromString("invalid-hex-string")

	assert.Error(t, err)
}
