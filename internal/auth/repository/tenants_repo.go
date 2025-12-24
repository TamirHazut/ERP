package repository

import (
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type TenantRepository struct {
	repository *db.Repository[models.Tenant]
	logger     *logging.Logger
}

func NewTenantRepository(dbHandler db.DBHandler) *TenantRepository {
	logger := logging.NewLogger(logging.ModuleAuth)
	repository := db.NewRepository[models.Tenant](dbHandler, string(mongo.TenantsCollection), logger)
	return &TenantRepository{
		repository: repository,
		logger:     logger,
	}
}

func (r *TenantRepository) CreateTenant(tenant models.Tenant) (string, error) {
	if err := tenant.Validate(true); err != nil {
		return "", err
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	r.logger.Debug("Creating tenant", "tenant", tenant)
	return r.repository.Create(tenant)
}

func (r *TenantRepository) GetTenantByID(tenantID string) (models.Tenant, error) {
	if tenantID == "" {
		return models.Tenant{}, erp_errors.Validation(erp_errors.ValidationRequiredFields, []string{"TenantID"})
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	r.logger.Debug("Getting tenant by id", "filter", filter)
	tenants, err := r.repository.Find(filter)
	if err != nil {
		return models.Tenant{}, erp_errors.Internal(erp_errors.InternalDatabaseError, err)
	}
	if len(tenants) == 0 {
		return models.Tenant{}, erp_errors.NotFound(erp_errors.NotFoundTenant, "Tenant", tenantID)
	}
	return tenants[0], nil
}

func (r *TenantRepository) UpdateTenant(tenant models.Tenant) error {
	if err := tenant.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"_id": tenant.ID,
	}
	r.logger.Debug("Updating tenant", "tenant", tenant)
	currentTenant, err := r.GetTenantByID(tenant.ID.String())
	if err != nil {
		return err
	}
	if tenant.CreatedAt != currentTenant.CreatedAt {
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, []string{"CreatedAt"})
	}
	tenant.UpdatedAt = time.Now()
	return r.repository.Update(filter, tenant)
}

func (r *TenantRepository) DeleteTenant(tenantID string) error {
	if tenantID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, []string{"TenantID"})
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	r.logger.Debug("Deleting tenant", "filter", filter)
	return r.repository.Delete(filter)
}
