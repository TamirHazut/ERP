package convertor

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	infra_error "erp.localhost/internal/infra/error"
	model_core "erp.localhost/internal/infra/model/core"
	proto_core "erp.localhost/internal/infra/proto/core/v1"
)

// =============================================================================
// Domain Model → Proto (for responses)
// =============================================================================

// UserToProto converts a User model to UserData proto message
func UserToProto(user *model_core.User) *proto_core.UserData {
	if user == nil || user.Validate(false) != nil {
		return nil
	}

	// Handle optional last_login timestamp
	var lastLogin *timestamppb.Timestamp
	if user.LastLogin != nil {
		lastLogin = timestamppb.New(*user.LastLogin)
	}

	return &proto_core.UserData{
		Id:                    user.ID.Hex(),
		TenantId:              user.TenantID,
		Email:                 user.Email,
		Username:              user.Username,
		Profile:               UserProfileToProto(&user.Profile),
		Roles:                 UserRolesToProto(user.Roles),
		AdditionalPermissions: user.AdditionalPermissions,
		RevokedPermissions:    user.RevokedPermissions,
		Status:                user.Status,
		EmailVerified:         user.EmailVerified,
		PhoneVerified:         user.PhoneVerified,
		MfaEnabled:            user.MFAEnabled,
		LastLogin:             lastLogin,
		LastPasswordChange:    timestamppb.New(user.LastPasswordChange),
		Preferences:           UserPreferencesToProto(&user.Preferences),
		CreatedAt:             timestamppb.New(user.CreatedAt),
		UpdatedAt:             timestamppb.New(user.UpdatedAt),
		CreatedBy:             user.CreatedBy,
		LastActivity:          timestamppb.New(user.LastActivity),
	}
}

// UsersToProto converts a slice of User models to UserData proto messages
func UsersToProto(users []*model_core.User) []*proto_core.UserData {
	if users == nil {
		return []*proto_core.UserData{}
	}

	protoUsers := make([]*proto_core.UserData, 0, len(users))
	for _, user := range users {
		if protoUser := UserToProto(user); protoUser != nil {
			protoUsers = append(protoUsers, protoUser)
		}
	}
	return protoUsers
}

// UserProfileToProto converts a UserProfile model to UserProfileData proto message
func UserProfileToProto(profile *model_core.UserProfile) *proto_core.UserProfileData {
	if profile == nil {
		return nil
	}

	return &proto_core.UserProfileData{
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		DisplayName: profile.DisplayName,
		AvatarUrl:   profile.AvatarURL,
		Phone:       profile.Phone,
		Title:       profile.Title,
		Department:  profile.Department,
	}
}

// UserRoleToProto converts a UserRole model to UserRoleData proto message
func UserRoleToProto(role *model_core.UserRole) *proto_core.UserRoleData {
	if role == nil {
		return nil
	}

	// Handle optional expires_at timestamp
	var expiresAt *timestamppb.Timestamp
	if role.ExpiresAt != nil {
		expiresAt = timestamppb.New(*role.ExpiresAt)
	}

	return &proto_core.UserRoleData{
		RoleId:     role.RoleID,
		TenantId:   role.TenantID,
		AssignedAt: timestamppb.New(role.AssignedAt),
		AssignedBy: role.AssignedBy,
		ExpiresAt:  expiresAt,
	}
}

// UserRolesToProto converts a slice of UserRole models to UserRoleData proto messages
func UserRolesToProto(roles []model_core.UserRole) []*proto_core.UserRoleData {
	if roles == nil {
		return []*proto_core.UserRoleData{}
	}

	protoRoles := make([]*proto_core.UserRoleData, 0, len(roles))
	for i := range roles {
		if protoRole := UserRoleToProto(&roles[i]); protoRole != nil {
			protoRoles = append(protoRoles, protoRole)
		}
	}
	return protoRoles
}

// NotificationSettingsToProto converts a NotificationSettings model to NotificationSettingsData proto message
func NotificationSettingsToProto(settings *model_core.NotificationSettings) *proto_core.NotificationSettingsData {
	if settings == nil {
		return &proto_core.NotificationSettingsData{
			Email: false,
			Push:  false,
			Sms:   false,
		}
	}

	return &proto_core.NotificationSettingsData{
		Email: settings.Email,
		Push:  settings.Push,
		Sms:   settings.SMS,
	}
}

// UserPreferencesToProto converts a UserPreferences model to UserPreferencesData proto message
func UserPreferencesToProto(prefs *model_core.UserPreferences) *proto_core.UserPreferencesData {
	if prefs == nil {
		return &proto_core.UserPreferencesData{
			Language:      "en",
			Timezone:      "UTC",
			Theme:         "light",
			Notifications: NotificationSettingsToProto(nil),
		}
	}

	return &proto_core.UserPreferencesData{
		Language:      prefs.Language,
		Timezone:      prefs.Timezone,
		Theme:         prefs.Theme,
		Notifications: NotificationSettingsToProto(&prefs.Notifications),
	}
}

// UserObjectIDFromString converts a hex string to a MongoDB ObjectID
func UserObjectIDFromString(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, infra_error.Validation(infra_error.ValidationInvalidValue, "id")
	}
	return primitive.ObjectIDFromHex(id)
}

// =============================================================================
// Proto → Domain Model (for create operations)
// =============================================================================

// CreateUserFromProto converts a CreateUserData proto message to a User model
func CreateUserFromProto(proto *proto_core.CreateUserData) (*model_core.User, error) {
	if proto == nil {
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
	}

	// Convert role_ids to UserRole structs
	roles := make([]model_core.UserRole, 0, len(proto.RoleIds))
	now := time.Now()
	for _, roleID := range proto.RoleIds {
		roles = append(roles, model_core.UserRole{
			RoleID:     roleID,
			TenantID:   proto.TenantId,
			AssignedAt: now,
			AssignedBy: proto.CreatedBy,
			ExpiresAt:  nil,
		})
	}

	// Set default status if not provided
	status := proto.Status
	if status == "" {
		status = model_core.UserStatusInvited
	}

	user := &model_core.User{
		TenantID:              proto.TenantId,
		Email:                 proto.Email,
		Username:              proto.Username,
		PasswordHash:          proto.PasswordHash,
		Profile:               CreateUserProfileFromProto(proto.Profile),
		Roles:                 roles,
		AdditionalPermissions: proto.AdditionalPermissions,
		RevokedPermissions:    []string{},
		Status:                status,
		EmailVerified:         false,
		PhoneVerified:         false,
		MFAEnabled:            false,
		MFASecret:             "",
		LastLogin:             nil,
		LastPasswordChange:    now,
		PasswordResetToken:    "",
		PasswordResetExpires:  nil,
		Preferences:           CreateUserPreferencesFromProto(proto.Preferences),
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             proto.CreatedBy,
		LastActivity:          now,
		LoginHistory:          []model_core.LoginRecord{},
	}

	return user, nil
}

// CreateUserProfileFromProto converts a UserProfileData proto message to a UserProfile model
func CreateUserProfileFromProto(proto *proto_core.UserProfileData) model_core.UserProfile {
	if proto == nil {
		return model_core.UserProfile{}
	}

	return model_core.UserProfile{
		FirstName:   proto.FirstName,
		LastName:    proto.LastName,
		DisplayName: proto.DisplayName,
		AvatarURL:   proto.AvatarUrl,
		Phone:       proto.Phone,
		Title:       proto.Title,
		Department:  proto.Department,
	}
}

// CreateUserPreferencesFromProto converts a UserPreferencesData proto message to a UserPreferences model
func CreateUserPreferencesFromProto(proto *proto_core.UserPreferencesData) model_core.UserPreferences {
	if proto == nil {
		return model_core.UserPreferences{
			Language: "en",
			Timezone: "UTC",
			Theme:    "light",
			Notifications: model_core.NotificationSettings{
				Email: true,
				Push:  false,
				SMS:   false,
			},
			DashboardLayout: make(map[string]interface{}),
		}
	}

	return model_core.UserPreferences{
		Language:        proto.Language,
		Timezone:        proto.Timezone,
		Theme:           proto.Theme,
		Notifications:   CreateNotificationSettingsFromProto(proto.Notifications),
		DashboardLayout: make(map[string]interface{}),
	}
}

// CreateNotificationSettingsFromProto converts a NotificationSettingsData proto message to a NotificationSettings model
func CreateNotificationSettingsFromProto(proto *proto_core.NotificationSettingsData) model_core.NotificationSettings {
	if proto == nil {
		return model_core.NotificationSettings{
			Email: true,
			Push:  false,
			SMS:   false,
		}
	}

	return model_core.NotificationSettings{
		Email: proto.Email,
		Push:  proto.Push,
		SMS:   proto.Sms,
	}
}

// =============================================================================
// Proto → Domain Model (for update operations)
// =============================================================================

// UpdateUserFromProto applies updates from UpdateUserData proto to an existing User model
func UpdateUserFromProto(existing *model_core.User, proto *proto_core.UpdateUserData) error {
	if existing == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "existing")
	}
	if proto == nil {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "proto")
	}

	// Update simple fields if provided
	// if proto.Email != nil {
	// 	existing.Email = *proto.Email
	// }
	// if proto.Username != nil {
	// 	existing.Username = *proto.Username
	// }
	if proto.Status != nil {
		existing.Status = *proto.Status
	}
	if proto.EmailVerified != nil {
		existing.EmailVerified = *proto.EmailVerified
	}
	if proto.PhoneVerified != nil {
		existing.PhoneVerified = *proto.PhoneVerified
	}

	// Update profile if provided
	if proto.Profile != nil {
		existing.Profile = UpdateUserProfileFromProto(proto.Profile)
	}

	// Update preferences if provided
	if proto.Preferences != nil {
		existing.Preferences = UpdateUserPreferencesFromProto(proto.Preferences)
	}

	// Handle role updates
	if proto.Roles != nil {
		now := time.Now()

		// Add new roles
		for _, roleID := range proto.Roles.AddRoleIds {
			existing.Roles = append(existing.Roles, model_core.UserRole{
				RoleID:     roleID,
				TenantID:   existing.TenantID,
				AssignedAt: now,
				AssignedBy: existing.CreatedBy, // TODO: Should be current user, but we don't have that context
				ExpiresAt:  nil,
			})
		}

		// Remove roles
		if len(proto.Roles.RemoveRoleIds) > 0 {
			removeMap := make(map[string]bool)
			for _, roleID := range proto.Roles.RemoveRoleIds {
				removeMap[roleID] = true
			}

			filteredRoles := make([]model_core.UserRole, 0, len(existing.Roles))
			for _, role := range existing.Roles {
				if !removeMap[role.RoleID] {
					filteredRoles = append(filteredRoles, role)
				}
			}
			existing.Roles = filteredRoles
		}
	}

	// Handle permission updates
	if proto.Permissions != nil {
		// Add permissions to AdditionalPermissions
		if len(proto.Permissions.AddPermissions) > 0 {
			existing.AdditionalPermissions = append(existing.AdditionalPermissions, proto.Permissions.AddPermissions...)
		}

		// Remove permissions from AdditionalPermissions
		if len(proto.Permissions.RemovePermissions) > 0 {
			removeMap := make(map[string]bool)
			for _, perm := range proto.Permissions.RemovePermissions {
				removeMap[perm] = true
			}

			filteredPerms := make([]string, 0, len(existing.AdditionalPermissions))
			for _, perm := range existing.AdditionalPermissions {
				if !removeMap[perm] {
					filteredPerms = append(filteredPerms, perm)
				}
			}
			existing.AdditionalPermissions = filteredPerms
		}

		// Add permissions to RevokedPermissions
		if len(proto.Permissions.RevokePermissions) > 0 {
			existing.RevokedPermissions = append(existing.RevokedPermissions, proto.Permissions.RevokePermissions...)
		}

		// Remove permissions from RevokedPermissions (unrevoke)
		if len(proto.Permissions.UnrevokePermissions) > 0 {
			unrevokeMap := make(map[string]bool)
			for _, perm := range proto.Permissions.UnrevokePermissions {
				unrevokeMap[perm] = true
			}

			filteredRevoked := make([]string, 0, len(existing.RevokedPermissions))
			for _, perm := range existing.RevokedPermissions {
				if !unrevokeMap[perm] {
					filteredRevoked = append(filteredRevoked, perm)
				}
			}
			existing.RevokedPermissions = filteredRevoked
		}
	}

	// Always update timestamp
	existing.UpdatedAt = time.Now()

	return nil
}

// UpdateUserProfileFromProto converts a UserProfileData proto message to a UserProfile model
func UpdateUserProfileFromProto(proto *proto_core.UserProfileData) model_core.UserProfile {
	if proto == nil {
		return model_core.UserProfile{}
	}

	return model_core.UserProfile{
		FirstName:   proto.FirstName,
		LastName:    proto.LastName,
		DisplayName: proto.DisplayName,
		AvatarURL:   proto.AvatarUrl,
		Phone:       proto.Phone,
		Title:       proto.Title,
		Department:  proto.Department,
	}
}

// UpdateUserPreferencesFromProto converts a UserPreferencesData proto message to a UserPreferences model
func UpdateUserPreferencesFromProto(proto *proto_core.UserPreferencesData) model_core.UserPreferences {
	if proto == nil {
		return model_core.UserPreferences{
			Language: "en",
			Timezone: "UTC",
			Theme:    "light",
			Notifications: model_core.NotificationSettings{
				Email: true,
				Push:  false,
				SMS:   false,
			},
			DashboardLayout: make(map[string]interface{}),
		}
	}

	return model_core.UserPreferences{
		Language:        proto.Language,
		Timezone:        proto.Timezone,
		Theme:           proto.Theme,
		Notifications:   CreateNotificationSettingsFromProto(proto.Notifications),
		DashboardLayout: make(map[string]interface{}),
	}
}
