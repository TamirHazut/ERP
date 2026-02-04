package handler

import (
	aggregation_auth "erp.localhost/internal/auth/aggregation"
	collection_auth "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/hash"
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

func NewUserHandler(logger logger.Logger) (*UserHandler, *infra_error.AppError) {
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

func (u *UserHandler) CreateUser(user *authv1.User) (string, *infra_error.AppError) {
	if err := validator_auth.ValidateUser(user, true); err != nil {
		return "", err
	}

	u.logger.Debug("Creating user", "user", user)
	passwordHash, err := hash.Hash(user.Password)
	if err != nil {
		return "", err
	}
	user.Password = passwordHash
	user.CreatedAt = timestamppb.Now()
	user.UpdatedAt = timestamppb.Now()

	return u.collection.Create(user)
}

func (u *UserHandler) GetUserByID(tenantID, userID string) (*authv1.User, *infra_error.AppError) {
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

func (u *UserHandler) GetUserByEmailOrUsername(tenantID, email, username string) (*authv1.User, *infra_error.AppError) {
	if email == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "email")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"$or": []map[string]any{
			{"email": email},
			{"username": username},
		},
	}
	u.logger.Debug("Getting user by email", "filter", filter)
	return u.findUserByFilter(filter)
}

func (u *UserHandler) GetUsersByTenantID(tenantID string) ([]*authv1.User, *infra_error.AppError) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Getting users by tenant id", "filter", filter)
	return u.findUsersByFilter(filter)
}

func (u *UserHandler) GetUsersByRoleID(tenantID, roleID string) ([]*authv1.User, *infra_error.AppError) {
	if roleID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "roleID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"roles": map[string]any{
			"$elemMatch": map[string]any{
				"tenant_id": tenantID,
				"role_id":   roleID,
			},
		},
	}
	u.logger.Debug("Getting users by role id", "filter", filter)
	return u.findUsersByFilter(filter)
}

func (u *UserHandler) GetUsersWithRoles(tenantID string) ([]*authv1.User, *infra_error.AppError) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"roles":     map[string]any{"$ne": []any{}},
	}
	users, err := u.findUsersByFilter(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserHandler) UpdateUser(user *authv1.User) *infra_error.AppError {
	if err := validator_auth.ValidateUser(user, false); err != nil {
		return err
	}
	currentUser, err := u.GetUserByID(user.TenantId, user.Id)
	if err != nil {
		return err
	}
	user.Protected = currentUser.Protected
	if currentUser.Protected {
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}
	u.logger.Debug("Updating user", "user", user)
	filter := map[string]any{
		"tenant_id": user.TenantId,
		"_id":       user.Id,
	}
	user.UpdatedAt = timestamppb.Now()
	return u.collection.Update(filter, user)
}

func (u *UserHandler) DeleteUser(tenantID, userID string) *infra_error.AppError {
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

func (u *UserHandler) DeleteTenantUsers(tenantID string) *infra_error.AppError {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId", "UserId")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	u.logger.Debug("Deleting user", "filter", filter)
	return u.collection.Delete(filter)
}

func (u *UserHandler) findUserByFilter(filter map[string]any) (*authv1.User, *infra_error.AppError) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	user, err := u.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserHandler) findUsersByFilter(filter map[string]any) ([]*authv1.User, *infra_error.AppError) {
	if _, ok := filter["tenant_id"]; !ok {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	users, err := u.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}
