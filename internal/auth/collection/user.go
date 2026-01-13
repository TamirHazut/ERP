package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
)

type UserCollection struct {
	collection collection.CollectionHandler[model_auth.User]
	logger     logger.Logger
}

func NewUserCollection(collection collection.CollectionHandler[model_auth.User], logger logger.Logger) *UserCollection {
	return &UserCollection{
		collection: collection,
		logger:     logger,
	}
}

func (u *UserCollection) CreateUser(user *model_auth.User) (string, error) {
	if err := user.Validate(true); err != nil {
		return "", err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	u.logger.Debug("Creating user", "user", user)
	return u.collection.Create(user)
}

func (u *UserCollection) GetUserByID(tenantID, userID string) (*model_auth.User, error) {
	if userID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "userID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	u.logger.Debug("Getting user by id", "filter", filter)
	return u.findUserByFilter(filter)
}

func (u *UserCollection) GetUserByEmail(tenantID, email string) (*model_auth.User, error) {
	if email == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "email")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"email":     email,
	}
	u.logger.Debug("Getting user by email", "filter", filter)
	return u.findUserByFilter(filter)
}

func (u *UserCollection) GetUserByUsername(tenantID, username string) (*model_auth.User, error) {
	if username == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "username")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"username":  username,
	}
	u.logger.Debug("Getting user by username", "filter", filter)
	return u.findUserByFilter(filter)
}

func (u *UserCollection) GetUsersByTenantID(tenantID string) ([]*model_auth.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Getting users by tenant id", "filter", filter)
	return u.findUsersByFilter(filter)
}

func (u *UserCollection) GetUsersByRoleID(tenantID, roleID string) ([]*model_auth.User, error) {
	if roleID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "roleID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"role_id":   roleID,
	}
	u.logger.Debug("Getting users by role id", "filter", filter)
	return u.findUsersByFilter(filter)
}

func (u *UserCollection) UpdateUser(user *model_auth.User) error {
	if err := user.Validate(false); err != nil {
		return err
	}
	userID := user.ID.Hex()
	u.logger.Debug("Updating user", "user", user)
	currentUser, err := u.GetUserByID(user.TenantID, userID)
	if err != nil {
		return err
	}
	if user.Username != currentUser.Username ||
		user.CreatedAt != currentUser.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "Username", "CreatedAt")
	}
	filter := map[string]any{
		"tenant_id": user.TenantID,
		"_id":       user.ID,
	}
	user.UpdatedAt = time.Now()
	return u.collection.Update(filter, user)
}

func (u *UserCollection) DeleteUser(tenantID, userID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID", "UserID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	if userID == "" {
		filter["_id"] = userID
	}
	u.logger.Debug("Deleting user", "filter", filter)
	return u.collection.Delete(filter)
}

func (u *UserCollection) findUserByFilter(filter map[string]any) (*model_auth.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	user, err := u.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserCollection) findUsersByFilter(filter map[string]any) ([]*model_auth.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	users, err := u.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}
