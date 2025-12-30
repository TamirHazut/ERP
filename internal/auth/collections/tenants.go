package collection

import (
	"time"

	"erp.localhost/internal/auth/models"
	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
)

type TenantCollection struct {
	collection mongo.CollectionHandler[models.Tenant]
	logger     *logging.Logger
}

func NewTenantCollection(collection mongo.CollectionHandler[models.Tenant]) *TenantCollection {
	logger := logging.NewLogger(logging.ModuleAuth)
	if collection == nil {
		collectionHandler := mongo.NewBaseCollectionHandler[models.Tenant](string(mongo.TenantsCollection), logger)
		if collectionHandler == nil {
			logger.Fatal("failed to create tenants collection handler")
			return nil
		}
		collection = collectionHandler
	}
	return &TenantCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *TenantCollection) CreateTenant(tenant models.Tenant) (string, error) {
	if err := tenant.Validate(true); err != nil {
		return "", err
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	r.logger.Debug("Creating tenant", "tenant", tenant)
	return r.collection.Create(tenant)
}

func (r *TenantCollection) GetTenantByID(tenantID string) (*models.Tenant, error) {
	if tenantID == "" {
		return nil, erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	r.logger.Debug("Getting tenant by id", "filter", filter)
	tenant, err := r.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (r *TenantCollection) UpdateTenant(tenant models.Tenant) error {
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
		return erp_errors.Validation(erp_errors.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	tenant.UpdatedAt = time.Now()
	return r.collection.Update(filter, tenant)
}

func (r *TenantCollection) DeleteTenant(tenantID string) error {
	if tenantID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	r.logger.Debug("Deleting tenant", "filter", filter)
	return r.collection.Delete(filter)
}
