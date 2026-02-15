package api

import (
	"errors"
	"slices"
	"strings"

	"erp.localhost/auth/handler"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	model_auth "erp.localhost/infra/model/auth"
	authv1 "erp.localhost/infra/model/auth/v1"
	validator_auth "erp.localhost/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// type FilterType int

// const (
// 	filterTypeUnsupported FilterType = iota
// 	filterTypeID
// 	filterTypeEmail
// 	filterTypeUsername
// )

type UserAPI struct {
	logger      logger.Logger
	userHandler *handler.UserHandler
	roleHandler *handler.RoleHandler
	rbacAPI     *RBACAPI
}

func NewUserAPI(
	rbacAPI *RBACAPI,
	roleHandler *handler.RoleHandler,
	logger logger.Logger,
) (*UserAPI, *infra_error.AppError) {
	userHander, err := handler.NewUserHandler(logger)
	if err != nil {
		logger.Error("failed to create new user handler", "error", err)
		return nil, err
	}
	return &UserAPI{
		rbacAPI:     rbacAPI,
		userHandler: userHander,
		roleHandler: roleHandler,
		logger:      logger,
	}, nil
}

func (u *UserAPI) CreateUser(tenantID, userID string, newUser *authv1.User) (string, *infra_error.AppError) {
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("failed to create user", "error", err)
		return "", err
	}

	newUser.Email = strings.TrimSpace(newUser.GetEmail())
	newUser.Username = strings.TrimSpace(newUser.GetUsername())
	newUser.Password = strings.TrimSpace(newUser.GetPassword())

	if err := validator_auth.ValidateUser(newUser, true); err != nil {
		u.logger.Error("failed to create user", "error", err)
		return "", err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionCreate, newUser.GetTenantId()); err != nil {
		u.logger.Error("failed to create user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return "", err
	}

	if newUser.Protected && !u.rbacAPI.Verification.IsSystemTenantUser(tenantID) {
		newUser.Protected = false
	}

	// Validate role IDs exist
	if err := u.validateRoleIDs(newUser.TenantId, newUser.Roles); err != nil {
		u.logger.Warn("Invalid role IDs in user", "tenant_id", newUser.TenantId, "email", newUser.Email, "error", err)
		return "", err
	}

	// Validate and normalize additional permission strings
	if err := validator_auth.ValidatePermissionStrings(newUser.AdditionalPermissions); err != nil {
		u.logger.Warn("Invalid additional permission strings in user", "tenant_id", newUser.TenantId, "email", newUser.Email, "error", err)
		return "", err
	}
	newUser.AdditionalPermissions = model_auth.NormalizePermissions(newUser.AdditionalPermissions)

	// Validate and normalize revoked permission strings
	if err := validator_auth.ValidatePermissionStrings(newUser.RevokedPermissions); err != nil {
		u.logger.Warn("Invalid revoked permission strings in user", "tenant_id", newUser.TenantId, "email", newUser.Email, "error", err)
		return "", err
	}
	newUser.RevokedPermissions = model_auth.NormalizePermissions(newUser.RevokedPermissions)

	return u.userHandler.CreateUser(newUser)
}

func (u *UserAPI) GetUserByID(tenantID, userID, targetTenantID, accountID string) (*authv1.User, *infra_error.AppError) {
	if tenantID == "" || userID == "" || targetTenantID == "" || accountID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id, account_id"))
		u.logger.Error("failed to get user", "error", err)
		return nil, err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionRead, targetTenantID); err != nil {
		u.logger.Error("failed to get user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return nil, err
	}
	return u.userHandler.GetUserByID(targetTenantID, accountID)
}

func (u *UserAPI) GetUserByEmailOrUsername(tenantID, userID, targetTenantID, email, username string) (*authv1.User, *infra_error.AppError) {
	if tenantID == "" || userID == "" || targetTenantID == "" || (email == "" && username == "") {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id, email, username"))
		u.logger.Error("failed to get user", "error", err)
		return nil, err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionRead, targetTenantID); err != nil {
		u.logger.Error("failed to get user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return nil, err
	}
	return u.userHandler.GetUserByEmailOrUsername(targetTenantID, email, username)
}

func (u *UserAPI) GetUsers(tenantID, userID, targetTenantID, roleID string) ([]*authv1.User, *infra_error.AppError) {
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		u.logger.Error("failed to get users", "error", err)
		return nil, err
	}
	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionRead, targetTenantID); err != nil {
		u.logger.Error("failed to get users", "tenant_id", tenantID, "user_id", userID, "error", err)
		return nil, err
	}

	if roleID != "" {
		return u.userHandler.GetUsersByRoleID(targetTenantID, roleID)
	}
	return u.userHandler.GetUsersByTenantID(targetTenantID)
}

func (u *UserAPI) UpdateUser(tenantID, userID string, newUserData *authv1.User) (bool, *infra_error.AppError) {
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("failed to update user", "error", err)
		return false, err
	}
	if err := validator_auth.ValidateUser(newUserData, true); err != nil {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("Failed to update user", "error", err)
		return false, err
	}

	targetTenantID := newUserData.TenantId

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionUpdate, targetTenantID); err != nil {
		u.logger.Error("failed to update user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return false, err
	}

	oldUserData, err := u.userHandler.GetUserByID(tenantID, newUserData.Id)
	if err != nil {
		u.logger.Error("failed to update user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return false, err
	}

	// Do diff and validate
	err = u.validateUserUpdateData(tenantID, userID, oldUserData, newUserData)
	if err != nil {
		u.logger.Error("failed to update user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return false, err
	}

	// Validate role IDs exist
	if err := u.validateRoleIDs(newUserData.TenantId, newUserData.Roles); err != nil {
		u.logger.Warn("Invalid role IDs in user", "tenant_id", newUserData.TenantId, "user_id", newUserData.Id, "error", err)
		return false, err
	}

	// Validate and normalize additional permission strings
	if err := validator_auth.ValidatePermissionStrings(newUserData.AdditionalPermissions); err != nil {
		u.logger.Warn("Invalid additional permission strings in user", "tenant_id", newUserData.TenantId, "user_id", newUserData.Id, "error", err)
		return false, err
	}
	newUserData.AdditionalPermissions = model_auth.NormalizePermissions(newUserData.AdditionalPermissions)

	// Validate and normalize revoked permission strings
	if err := validator_auth.ValidatePermissionStrings(newUserData.RevokedPermissions); err != nil {
		u.logger.Warn("Invalid revoked permission strings in user", "tenant_id", newUserData.TenantId, "user_id", newUserData.Id, "error", err)
		return false, err
	}
	newUserData.RevokedPermissions = model_auth.NormalizePermissions(newUserData.RevokedPermissions)

	newUserData.CreatedBy = oldUserData.CreatedBy
	newUserData.CreatedAt = oldUserData.CreatedAt
	newUserData.UpdatedAt = timestamppb.Now()

	return u.updateUser(newUserData)
}

func (u *UserAPI) DeleteUser(tenantID, userID, targetTenantID, accountID string) *infra_error.AppError {
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		u.logger.Error("failed to delete user", "error", err)
		return err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionDelete, targetTenantID); err != nil {
		u.logger.Error("failed to delete user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return err
	}

	if err := u.checkProtectedUser(tenantID, userID, accountID); err != nil {
		return err
	}

	if err := u.userHandler.DeleteUser(targetTenantID, accountID); err != nil {
		u.logger.Error("failed to delete user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return err
	}

	u.logger.Debug("user deleted successfuly", "tenant_id", tenantID, "user_id", userID, "target_tenant_id", targetTenantID)
	return nil
}

func (u *UserAPI) DeleteTenantUsers(tenantID, userID, targetTenantID string) *infra_error.AppError {
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		u.logger.Error("failed to delete tenant users", "error", err)
		return err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionDelete, targetTenantID); err != nil {
		u.logger.Error("failed to delete tenant users", "tenant_id", tenantID, "user_id", userID, "error", err)
		return err
	}

	if err := u.userHandler.DeleteTenantUsers(targetTenantID); err != nil {
		u.logger.Error("failed to delete tenant users", "tenant_id", tenantID, "user_id", userID, "error", err)
		return err
	}

	u.logger.Debug("tenant users deleted successfuly", "tenant_id", tenantID, "user_id", userID, "target_tenant_id", targetTenantID)
	return nil
}

/* Helper functions */
func (u *UserAPI) hasPermission(tenantID, userID, action, targetTenantID string) *infra_error.AppError {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, action)
	if err != nil {
		return err
	}
	return u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, targetTenantID)
}

// func (u *UserAPI) getUser(tenantID string, accountID string, filterType FilterType) (*authv1.User, *infra_error.AppError) {
// 	switch filterType {
// 	case filterTypeID:
// 		return u.userHandler.GetUserByID(tenantID, accountID)
// 	case filterTypeEmail:
// 		return u.userHandler.GetUserByEmail(tenantID, accountID)
// 	case filterTypeUsername:
// 		return u.userHandler.GetUserByUsername(tenantID, accountID)
// 	default:
// 		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "account identifier")
// 	}
// }

func (u *UserAPI) updateUser(user *authv1.User) (bool, *infra_error.AppError) {
	tenantID := user.GetTenantId()
	userID := user.GetId()
	currentUser, err := u.userHandler.GetUserByID(user.TenantId, user.Id)
	if err != nil {
		return false, err
	}
	user.Protected = currentUser.Protected
	if currentUser.Protected {
		return false, infra_error.Auth(infra_error.AuthPermissionDenied)
	}
	err = u.userHandler.UpdateUser(user)
	success := err == nil
	if success {
		u.logger.Debug("user updated successfuly", "tenant_id", tenantID, "user_id", userID, "target_tenant_id")
	} else {
		u.logger.Error("failed to update user", "tenant_id", tenantID, "user_id", userID, "error", err)
	}
	return success, err
}

func (u *UserAPI) validateUserUpdateData(tenantID, userID string, old *authv1.User, new *authv1.User) *infra_error.AppError {
	if old.TenantId != new.TenantId ||
		old.Username != new.Username ||
		old.Email != new.Email {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields)
	}

	equal := slices.EqualFunc(old.Roles, new.Roles, func(a, b *authv1.UserRole) bool {
		return a.TenantId == b.TenantId &&
			a.RoleId == b.RoleId
	})
	if !equal {
		permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, model_auth.PermissionActionModifyRole)
		if err != nil {
			return err
		}
		if err := u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, new.TenantId); err != nil {
			return err
		}
	}

	if !slices.Equal(old.AdditionalPermissions, new.AdditionalPermissions) || !slices.Equal(old.RevokedPermissions, new.RevokedPermissions) {
		permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, model_auth.PermissionActionModifyPermission)
		if err != nil {
			return err
		}
		if err := u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, new.TenantId); err != nil {
			return err
		}
	}

	return nil
}

func (u *UserAPI) checkProtectedUser(tenantID, requestorUserID, userID string) *infra_error.AppError {
	user, err := u.userHandler.GetUserByID(tenantID, userID)
	if err != nil {
		return err
	}
	if user.Protected {
		if !u.rbacAPI.Verification.IsSystemTenantUser(tenantID) ||
			!u.rbacAPI.Verification.IsSystemAdminUser(requestorUserID) ||
			u.rbacAPI.Verification.IsSystemAdminUser(userID) {
			return infra_error.Auth(infra_error.AuthPermissionDenied)
		}
	}
	return nil
}

// validateRoleIDs verifies that all role IDs exist in the database
// Uses MongoDB aggregation for efficient batch validation (single query instead of N queries)
func (u *UserAPI) validateRoleIDs(tenantID string, roles []*authv1.UserRole) *infra_error.AppError {
	if len(roles) == 0 {
		return nil // Empty roles allowed
	}

	// Extract role IDs from UserRole objects
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		if role.TenantId != tenantID {
			return infra_error.Auth(infra_error.AuthPermissionDenied).WithError(errors.New("trying to assign roles from different tenant"))
		}
		roleIDs[i] = role.RoleId
	}

	// Use aggregation to batch validate (single database query)
	foundRoles, err := u.roleHandler.GetRolesByIDsAggregation(
		tenantID,
		roleIDs,
		[]string{"_id"}, // Only fetch IDs for efficiency
	)
	if err != nil {
		u.logger.Error("Failed to validate role IDs", "tenant_id", tenantID, "error", err)
		return err
	}

	// Check if all IDs were found
	if len(foundRoles) != len(roleIDs) {
		// Find which IDs are missing for detailed error message
		foundIDs := make(map[string]bool)
		for _, role := range foundRoles {
			foundIDs[role.Id] = true
		}

		missingIDs := []string{}
		for _, id := range roleIDs {
			if !foundIDs[id] {
				missingIDs = append(missingIDs, id)
			}
		}

		u.logger.Warn("Invalid role IDs in user", "tenant_id", tenantID, "missing_ids", missingIDs)
		return infra_error.NotFound(
			infra_error.NotFoundRole,
			model_auth.ResourceTypeRole,
			strings.Join(missingIDs, ", "),
		)
	}

	return nil
}
