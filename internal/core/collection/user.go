package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_core "erp.localhost/internal/infra/model/core"
)

type UserCollection struct {
	collection collection.CollectionHandler[model_core.User]
	logger     logger.Logger
}

func NewUserCollection(collection collection.CollectionHandler[model_core.User], logger logger.Logger) *UserCollection {
	return &UserCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *UserCollection) CreateUser(user *model_core.User) (string, error) {
	if err := user.Validate(true); err != nil {
		return "", err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.logger.Debug("Creating user", "user", user)
	return r.collection.Create(user)
}

func (r *UserCollection) GetUserByID(tenantID, userID string) (*model_core.User, error) {
	if userID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "userID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	r.logger.Debug("Getting user by id", "filter", filter)
	return r.findUserByFilter(filter)
}

func (r *UserCollection) GetUserByEmail(tenantID, email string) (*model_core.User, error) {
	if email == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "email")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"email":     email,
	}
	r.logger.Debug("Getting user by email", "filter", filter)
	return r.findUserByFilter(filter)
}

func (r *UserCollection) GetUserByUsername(tenantID, username string) (*model_core.User, error) {
	if username == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "username")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"username":  username,
	}
	r.logger.Debug("Getting user by username", "filter", filter)
	return r.findUserByFilter(filter)
}

func (r *UserCollection) GetUsersByTenantID(tenantID string) ([]*model_core.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting users by tenant id", "filter", filter)
	return r.findUsersByFilter(filter)
}

func (r *UserCollection) GetUsersByRoleID(tenantID, roleID string) ([]*model_core.User, error) {
	if roleID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "roleID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"role_id":   roleID,
	}
	r.logger.Debug("Getting users by role id", "filter", filter)
	return r.findUsersByFilter(filter)
}

func (r *UserCollection) UpdateUser(user *model_core.User) error {
	if err := user.Validate(false); err != nil {
		return err
	}
	userID := user.ID.Hex()
	r.logger.Debug("Updating user", "user", user)
	currentUser, err := r.GetUserByID(user.TenantID, userID)
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
	return r.collection.Update(filter, user)
}

func (r *UserCollection) DeleteUser(tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID", "UserID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	r.logger.Debug("Deleting user", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *UserCollection) findUserByFilter(filter map[string]any) (*model_core.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	user, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserCollection) findUsersByFilter(filter map[string]any) ([]*model_core.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	users, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}
