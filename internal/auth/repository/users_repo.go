package repository

import (
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type UserRepository struct {
	repository *db.Repository[models.User]
	logger     *logging.Logger
}

func NewUserRepository(dbHandler db.DBHandler) *UserRepository {
	logger := logging.NewLogger(logging.ModuleAuth)
	repository := db.NewRepository[models.User](dbHandler, string(mongo.UsersCollection), logger)
	return &UserRepository{
		repository: repository,
		logger:     logger,
	}
}

func (r *UserRepository) CreateUser(user models.User) (string, error) {
	if err := user.Validate(true); err != nil {
		return "", err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.logger.Debug("Creating user", "user", user)
	return r.repository.Create(user)
}

func (r *UserRepository) GetUserByID(tenantID, userID string) (models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	r.logger.Debug("Getting user by id", "filter", filter)
	users, err := r.repository.Find(filter)
	if err != nil {
		return models.User{}, erp_errors.NotFound(erp_errors.NotFoundUser, "User", userID)
	}
	if len(users) == 0 {
		return models.User{}, erp_errors.NotFound(erp_errors.NotFoundUser, "User", userID)
	}
	return users[0], nil
}

func (r *UserRepository) GetUserByUsername(tenantID, username string) (models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"username":  username,
	}
	r.logger.Debug("Getting user by username", "filter", filter)
	users, err := r.repository.Find(filter)
	if err != nil {
		return models.User{}, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(users) == 0 {
		return models.User{}, erp_errors.NotFound(erp_errors.NotFoundUser, "User", username)
	}
	return users[0], nil
}

func (r *UserRepository) GetUsersByTenantID(tenantID string) ([]models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting users by tenant id", "filter", filter)
	users, err := r.repository.Find(filter)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return users, nil
}

func (r *UserRepository) GetUsersByRoleID(tenantID, roleID string) ([]models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"role_id":   roleID,
	}
	r.logger.Debug("Getting users by role id", "filter", filter)
	users, err := r.repository.Find(filter)
	if err != nil {
		return nil, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	return users, nil
}

func (r *UserRepository) UpdateUser(user models.User) error {
	if err := user.Validate(false); err != nil {
		return err
	}
	userID := user.ID.String()
	r.logger.Debug("Updating user", "user", user)
	currentUser, err := r.GetUserByID(user.TenantID, userID)
	if err != nil {
		return err
	}
	if user.Username != currentUser.Username ||
		user.CreatedAt != currentUser.CreatedAt {
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, "Username", "CreatedAt")
	}
	filter := map[string]any{
		"tenant_id": user.TenantID,
		"_id":       user.ID,
	}
	user.UpdatedAt = time.Now()
	return r.repository.Update(filter, user)
}

func (r *UserRepository) DeleteUser(tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID", "UserID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	r.logger.Debug("Deleting user", "filter", filter)
	return r.repository.Delete(filter)
}
