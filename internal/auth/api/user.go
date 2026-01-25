package api

import (
	"errors"
	"slices"

	"erp.localhost/internal/auth/handler"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
)

type FilterType int

const (
	filterTypeUnsupported FilterType = iota
	filterTypeID
	filterTypeEmail
	filterTypeUsername
)

type UserAPI struct {
	logger      logger.Logger
	userHandler *handler.UserHandler
	rbacAPI     *RBACAPI
}

func NewUserAPI(rbacAPI *RBACAPI, logger logger.Logger) (*UserAPI, error) {
	userHander, err := handler.NewUserHandler(logger)
	if err != nil {
		logger.Error("failed to create new user handler", "error", err)
		return nil, err
	}
	return &UserAPI{
		rbacAPI:     rbacAPI,
		userHandler: userHander,
		logger:      logger,
	}, nil
}

func (u *UserAPI) CreateUser(tenantID, userID string, newUser *authv1.User) (string, error) {
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("failed to create user", "error", err)
		return "", err
	}
	if err := validator_auth.ValidateUser(newUser, true); err != nil {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("failed to create user", "error", err)
		return "", err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionCreate, tenantID); err != nil {
		u.logger.Error("failed to create user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return "", err
	}

	user, err := u.getUser(tenantID, newUser.Email, filterTypeEmail)
	if err != nil {
		u.logger.Error("failed to get user for verification", "tenant_id", tenantID, "error", err)
		return "", err
	}
	if user != nil {
		err := infra_error.Validation(infra_error.ConflictDuplicateEmail)
		u.logger.Error("failed to create new account", "tenantID", tenantID, "error", err.Error())
		return "", err
	}

	// convert from proto user to model user
	return u.userHandler.CreateUser(newUser)
}

func (u *UserAPI) GetUser(tenantID, userID, targetTenantID, accountID string) (*authv1.User, error) {
	if tenantID == "" || userID == "" || targetTenantID == "" || accountID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id, account_id"))
		u.logger.Error("failed to get user", "error", err)
		return nil, err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionRead, targetTenantID); err != nil {
		u.logger.Error("failed to get user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return nil, err
	}
	return u.getUser(tenantID, accountID, filterTypeID)
}

func (u *UserAPI) GetUsers(tenantID, userID, targetTenantID, roleID string) ([]*authv1.User, error) {
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

// TODO: finish logic
func (u *UserAPI) UpdateUser(tenantID, userID string, newUserData *authv1.User) (bool, error) {
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

	oldUserData, err := u.getUser(tenantID, newUserData.Id, filterTypeID)
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
	return u.updateUser(newUserData)
}

func (u *UserAPI) DeleteUser(tenantID, userID, targetTenantID, accountID string) error {
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		u.logger.Error("failed to delete user", "error", err)
		return err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionDelete, targetTenantID); err != nil {
		u.logger.Error("failed to delete user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return err
	}

	if err := u.userHandler.DeleteUser(targetTenantID, accountID); err != nil {
		u.logger.Error("failed to delete user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return err
	}
	u.logger.Debug("user deleted successfuly", "tenant_id", tenantID, "user_id", userID, "target_tenant_id", targetTenantID)
	return nil
}

func (u *UserAPI) DeleteTenantUsers(tenantID, userID, targetTenantID string) error {
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
func (u *UserAPI) hasPermission(tenantID, userID, action, targetTenantID string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, action)
	if err != nil {
		return err
	}
	return u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, targetTenantID)
}

func (u *UserAPI) getUser(tenantID string, accountID string, filterType FilterType) (*authv1.User, error) {
	switch filterType {
	case filterTypeID:
		return u.userHandler.GetUserByID(tenantID, accountID)
	case filterTypeEmail:
		return u.userHandler.GetUserByEmail(tenantID, accountID)
	case filterTypeUsername:
		return u.userHandler.GetUserByUsername(tenantID, accountID)
	default:
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "account identifier")
	}
}

func (u *UserAPI) updateUser(user *authv1.User) (bool, error) {
	tenantID := user.GetTenantId()
	userID := user.GetId()
	err := u.userHandler.UpdateUser(user)
	success := err == nil
	if success {
		u.logger.Debug("user updated successfuly", "tenant_id", tenantID, "user_id", userID, "target_tenant_id")
	} else {
		u.logger.Error("failed to update user", "tenant_id", tenantID, "user_id", userID, "error", err)
	}
	return success, err
}

func (u *UserAPI) validateUserUpdateData(tenantID, userID string, old *authv1.User, new *authv1.User) error {
	if old.TenantId != new.TenantId ||
		old.Username != new.Username ||
		old.Email != new.Email ||
		old.CreatedBy != new.CreatedBy ||
		old.CreatedAt != new.CreatedAt {
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
