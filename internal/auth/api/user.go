package api

import (
	"errors"
	"slices"
	"time"

	"erp.localhost/internal/auth/collection"
	collection_auth "erp.localhost/internal/auth/collection"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
)

type FilterType int

const (
	filterTypeUnsupported FilterType = iota
	filterTypeID
	filterTypeEmail
	filterTypeUsername
)

type UserAPI struct {
	logger         logger.Logger
	userCollection *collection_auth.UserCollection
	rbacAPI        *RBACAPI
	authAPI        *AuthAPI
}

func NewUserAPI(rbacAPI *RBACAPI, authAPI *AuthAPI, logger logger.Logger) *UserAPI {
	userHandler := mongo_collection.NewBaseCollectionHandler[model_auth.User](model_mongo.AuthDB, model_mongo.UsersCollection, logger)
	userCollection := collection.NewUserCollection(userHandler, logger)
	if userCollection == nil {
		logger.Fatal("failed to create users collection")
		return nil
	}
	return &UserAPI{
		rbacAPI:        rbacAPI,
		authAPI:        authAPI,
		userCollection: userCollection,
		logger:         logger,
	}
}

func (u *UserAPI) CreateUser(tenantID, userID string, newUser *model_auth.User) (string, error) {
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("failed to create user", "error", err)
		return "", err
	}
	if err := newUser.Validate(true); err != nil {
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
	return u.userCollection.CreateUser(newUser)
}

func (u *UserAPI) GetUser(tenantID, userID, targetTenantID, accountID string) (*model_auth.User, error) {
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

func (u *UserAPI) GetUsers(tenantID, userID, targetTenantID, roleID string) ([]*model_auth.User, error) {
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
		return u.userCollection.GetUsersByRoleID(targetTenantID, roleID)
	}
	return u.userCollection.GetUsersByTenantID(targetTenantID)
}

// TODO: finish logic
func (u *UserAPI) UpdateUser(tenantID, userID string, newUserData *model_auth.User) (bool, error) {
	if tenantID == "" || userID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("failed to update user", "error", err)
		return false, err
	}
	if err := newUserData.Validate(true); err != nil {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id"))
		u.logger.Error("Failed to update user", "error", err)
		return false, err
	}

	targetTenantID := newUserData.TenantID

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionUpdate, targetTenantID); err != nil {
		u.logger.Error("failed to update user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return false, err
	}

	oldUserData, err := u.getUser(tenantID, newUserData.ID.Hex(), filterTypeID)
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

	err = u.userCollection.UpdateUser(newUserData)
	return err == nil, err
}

func (u *UserAPI) DeleteUser(tenantID, userID, targetTenantID, accountID string) (bool, error) {
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		u.logger.Error("failed to delete user", "error", err)
		return false, err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionDelete, targetTenantID); err != nil {
		u.logger.Error("failed to delete user", "tenant_id", tenantID, "user_id", userID, "error", err)
		return false, err
	}

	if err := u.userCollection.DeleteUser(targetTenantID, accountID); err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserAPI) DeleteTenantUsers(tenantID, userID, targetTenantID string) (bool, error) {
	if tenantID == "" || userID == "" || targetTenantID == "" {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, user_id, target_tenant_id"))
		u.logger.Error("failed to delete tenant users", "error", err)
		return false, err
	}

	if err := u.hasPermission(tenantID, userID, model_auth.PermissionActionDelete, targetTenantID); err != nil {
		u.logger.Error("failed to delete tenant users", "tenant_id", tenantID, "user_id", userID, "error", err)
		return false, err
	}

	if err := u.userCollection.DeleteTenantUsers(targetTenantID); err != nil {
		return false, err
	}
	return true, nil
}

func (u *UserAPI) Login(tenantID, email, username, password string) (*NewTokenResponse, error) {
	if tenantID == "" || password == "" || (email == "" && username == "") {
		err := infra_error.Validation(infra_error.ValidationInvalidValue).WithError(errors.New("missing one or more: tenant_id, email/username, password"))
		u.logger.Error("failed to delete tenant users", "error", err)
		return nil, err
	}

	var filterType FilterType
	if email != "" {
		filterType = filterTypeEmail
	} else if username != "" {
		filterType = filterTypeUsername
	} else {
		filterType = filterTypeUnsupported
	}
	user, err := u.getUser(tenantID, email, filterType)
	if err != nil {
		return nil, err
	}

	tokens, err := u.authAPI.Authenticate(tenantID, user.ID.Hex(), password, user.PasswordHash)
	if user.LoginHistory == nil {
		user.LoginHistory = make([]model_auth.LoginRecord, 0)
	}
	user.LoginHistory = append(user.LoginHistory, model_auth.LoginRecord{
		Timestamp: time.Now(),
		Success:   tokens != nil,
	})
	if updateErr := u.userCollection.UpdateUser(user); updateErr != nil {
		u.logger.Error("failed to update user login history", "error", err)
	}
	return tokens, err
}

func (u *UserAPI) Logout(tenantID, userID, accessToken, refreshToken, revokedBy string) (string, error) {
	err := u.authAPI.RevokeTokens(tenantID, userID, accessToken, refreshToken, revokedBy)
	if err != nil {
		return "logout failed", err
	}
	return "logout successful", err
}

/* Helper functions */
func (u *UserAPI) hasPermission(tenantID, userID, action, targetTenantID string) error {
	permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, action)
	if err != nil {
		return err
	}
	return u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, targetTenantID)
}

func (u *UserAPI) getUser(tenantID string, accountID string, filterType FilterType) (*model_auth.User, error) {
	switch filterType {
	case filterTypeID:
		return u.userCollection.GetUserByID(tenantID, accountID)
	case filterTypeEmail:
		return u.userCollection.GetUserByEmail(tenantID, accountID)
	case filterTypeUsername:
		return u.userCollection.GetUserByUsername(tenantID, accountID)
	default:
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "account identifier")
	}
}

func (u *UserAPI) validateUserUpdateData(tenantID, userID string, old *model_auth.User, new *model_auth.User) error {
	if old.TenantID != new.TenantID ||
		old.Username != new.Username ||
		old.Email != new.Email ||
		old.CreatedBy != new.CreatedBy ||
		old.CreatedAt != new.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields)
	}

	equal := slices.EqualFunc(old.Roles, new.Roles, func(a, b model_auth.UserRole) bool {
		return a.TenantID == b.TenantID &&
			a.RoleID == b.RoleID
	})
	if !equal {
		permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, model_auth.PermissionActionModifyRole)
		if err != nil {
			return err
		}
		if err := u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, new.TenantID); err != nil {
			return err
		}
	}

	if !slices.Equal(old.AdditionalPermissions, new.AdditionalPermissions) || !slices.Equal(old.RevokedPermissions, new.RevokedPermissions) {
		permission, err := model_auth.CreatePermissionString(model_auth.ResourceTypeUser, model_auth.PermissionActionModifyPermission)
		if err != nil {
			return err
		}
		if err := u.rbacAPI.Verification.HasPermission(tenantID, userID, permission, new.TenantID); err != nil {
			return err
		}
	}

	return nil
}
