package handler

import (
	"strings"

	aggregation_auth "erp.localhost/internal/auth/aggregation"
	collection_auth "erp.localhost/internal/auth/collection"
	aggregation_mongo "erp.localhost/internal/infra/db/mongo/aggregation"
	collection_mongo "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_auth "erp.localhost/internal/infra/model/auth/validator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	collection  collection_mongo.CollectionHandler[authv1.User]
	aggregation aggregation_mongo.AggregationHandler[authv1.User]
	logger      logger.Logger
}

func NewUserHandler(logger logger.Logger) (*UserHandler, error) {
	collection, err := collection_auth.NewUserCollection(logger)
	if err != nil {
		logger.Error("failed to create user collection handler", "error", err)
		return nil, err
	}
	aggregation, err := aggregation_auth.NewUserAggregationHandler(logger)
	if err != nil {
		logger.Error("failed to create user aggregation handler", "error", err)
		return nil, err
	}
	return &UserHandler{
		collection:  collection,
		aggregation: aggregation,
		logger:      logger,
	}, nil
}

func (u *UserHandler) CreateUser(user *authv1.User) (string, error) {
	if err := validator_auth.ValidateUser(user, true); err != nil {
		return "", err
	}
	user.CreatedAt = timestamppb.Now()
	user.UpdatedAt = timestamppb.Now()
	u.logger.Debug("Creating user", "user", user)
	if user.GetUsername() != "" {
		user.Username = strings.ToLower(user.Username)
	}
	if user.GetEmail() != "" {
		user.Email = strings.ToLower(user.Email)
	}
	return u.collection.Create(user)
}

func (u *UserHandler) GetUserByID(tenantID, userID string) (*authv1.User, error) {
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

func (u *UserHandler) GetUserByEmail(tenantID, email string) (*authv1.User, error) {
	if email == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "email")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"email":     strings.ToLower(email),
	}
	u.logger.Debug("Getting user by email", "filter", filter)
	return u.findUserByFilter(filter)
}

func (u *UserHandler) GetUserByUsername(tenantID, username string) (*authv1.User, error) {
	if username == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "username")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"username":  strings.ToLower(username),
	}
	u.logger.Debug("Getting user by username", "filter", filter)
	return u.findUserByFilter(filter)
}

func (u *UserHandler) GetUsersByTenantID(tenantID string) ([]*authv1.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Getting users by tenant id", "filter", filter)
	return u.findUsersByFilter(filter)
}

func (u *UserHandler) GetUsersByRoleID(tenantID, roleID string) ([]*authv1.User, error) {
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

func (u *UserHandler) UpdateUser(user *authv1.User) error {
	if err := validator_auth.ValidateUser(user, false); err != nil {
		return err
	}
	u.logger.Debug("Updating user", "user", user)
	filter := map[string]any{
		"tenant_id": user.TenantId,
		"_id":       user.Id,
	}
	user.UpdatedAt = timestamppb.Now()
	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)
	return u.collection.Update(filter, user)
}

func (u *UserHandler) DeleteUser(tenantID, userID string) error {
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

func (u *UserHandler) DeleteTenantUsers(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "UserId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Deleting user", "filter", filter)
	return u.collection.Delete(filter)
}

func (u *UserHandler) findUserByFilter(filter map[string]any) (*authv1.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	user, err := u.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserHandler) findUsersByFilter(filter map[string]any) ([]*authv1.User, error) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	users, err := u.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}
