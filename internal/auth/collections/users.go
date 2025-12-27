package collection

import (
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type UserCollection struct {
	collection *mongo.CollectionHandler[models.User]
	logger     *logging.Logger
}

func NewUserCollection(dbHandler db.DBHandler) *UserCollection {
	logger := logging.NewLogger(logging.ModuleAuth)
	collection := mongo.NewCollectionHandler[models.User](dbHandler, string(mongo.UsersCollection), logger)
	return &UserCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *UserCollection) CreateUser(user models.User) (string, error) {
	if err := user.Validate(true); err != nil {
		return "", err
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.logger.Debug("Creating user", "user", user)
	return r.collection.Create(user)
}

func (r *UserCollection) GetUserByID(tenantID, userID string) (*models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	r.logger.Debug("Getting user by id", "filter", filter)
	return r.findUserByFilter(filter)
}

func (r *UserCollection) GetUserByEmail(tenantID, email string) (*models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"email":     email,
	}
	r.logger.Debug("Getting user by email", "filter", filter)
	return r.findUserByFilter(filter)
}

func (r *UserCollection) GetUserByUsername(tenantID, username string) (*models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"username":  username,
	}
	r.logger.Debug("Getting user by username", "filter", filter)
	return r.findUserByFilter(filter)
}

func (r *UserCollection) GetUsersByTenantID(tenantID string) ([]models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
	}
	r.logger.Debug("Getting users by tenant id", "filter", filter)
	return r.findUsersByFilter(filter)
}

func (r *UserCollection) GetUsersByRoleID(tenantID, roleID string) ([]models.User, error) {
	filter := map[string]any{
		"tenant_id": tenantID,
		"role_id":   roleID,
	}
	r.logger.Debug("Getting users by role id", "filter", filter)
	return r.findUsersByFilter(filter)
}

func (r *UserCollection) UpdateUser(user models.User) error {
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
	return r.collection.Update(filter, user)
}

func (r *UserCollection) DeleteUser(tenantID, userID string) error {
	if tenantID == "" || userID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID", "UserID")
	}
	filter := map[string]any{
		"tenant_id": tenantID,
		"_id":       userID,
	}
	r.logger.Debug("Deleting user", "filter", filter)
	return r.collection.Delete(filter)
}

func (r *UserCollection) findUserByFilter(filter map[string]any) (*models.User, error) {
	user, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserCollection) findUsersByFilter(filter map[string]any) ([]models.User, error) {
	users, err := r.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return users, nil
}
