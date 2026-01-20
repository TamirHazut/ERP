package collection

import (
	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserCollection struct {
	collection collection.CollectionHandler[authv1.User]
	logger     logger.Logger
}

func NewUserCollection(collection collection.CollectionHandler[authv1.User], logger logger.Logger) *UserCollection {
	if collection == nil {
		return nil
	}
	return &UserCollection{
		collection: collection,
		logger:     logger,
	}
}

func (u *UserCollection) CreateUser(user *authv1.User) (string, error) {
	if err := validator_auth.ValidateUser(user, true); err != nil {
		return "", err
	}
	user.CreatedAt = timestamppb.Now()
	user.UpdatedAt = timestamppb.Now()
	u.logger.Debug("Creating user", "user", user)
	return u.collection.Create(user)
}

func (u *UserCollection) GetUserByID(tenantID, userID string) (*authv1.User, error) {
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

func (u *UserCollection) GetUserByEmail(tenantID, email string) (*authv1.User, error) {
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

func (u *UserCollection) GetUserByUsername(tenantID, username string) (*authv1.User, error) {
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

func (u *UserCollection) GetUsersByTenantID(tenantID string) ([]*authv1.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Getting users by tenant id", "filter", filter)
	return u.findUsersByFilter(filter)
}

func (u *UserCollection) GetUsersByRoleID(tenantID, roleID string) ([]*authv1.User, error) {
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

func (u *UserCollection) UpdateUser(user *authv1.User) error {
	if err := validator_auth.ValidateUser(user, false); err != nil {
		return err
	}
	u.logger.Debug("Updating user", "user", user)
	filter := map[string]any{
		"tenant_id": user.TenantId,
		"_id":       user.Id,
	}
	user.UpdatedAt = timestamppb.Now()
	return u.collection.Update(filter, user)
}

func (u *UserCollection) DeleteUser(tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "UserId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	u.logger.Debug("Deleting user", "filter", filter)
	return u.collection.Delete(filter)
}

func (u *UserCollection) DeleteTenantUsers(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "UserId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Deleting user", "filter", filter)
	return u.collection.Delete(filter)
}

func (u *UserCollection) findUserByFilter(filter map[string]any) (*authv1.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	user, err := u.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserCollection) findUsersByFilter(filter map[string]any) ([]*authv1.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	users, err := u.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}
