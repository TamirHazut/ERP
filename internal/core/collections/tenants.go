package collection

import (
	"time"

	"erp.localhost/internal/db/mongo"
	erp_errors "erp.localhost/internal/errors"
	"erp.localhost/internal/logging"
	shared_models "erp.localhost/internal/shared/models"
	core_models "erp.localhost/internal/shared/models/core"
	mongo_models "erp.localhost/internal/shared/models/db/mongo"
)

type TenantCollection struct {
	collection mongo.CollectionHandler[core_models.Tenant]
	logger     *logging.Logger
}

func NewTenantCollection(collection mongo.CollectionHandler[core_models.Tenant]) *TenantCollection {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	if collection == nil {
		collectionHandler := mongo.NewBaseCollectionHandler[core_models.Tenant](string(mongo_models.TenantsCollection), logger)
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

func (r *TenantCollection) CreateTenant(tenant core_models.Tenant) (string, error) {
	if err := tenant.Validate(true); err != nil {
		return "", err
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	r.logger.Debug("Creating tenant", "tenant", tenant)
	return r.collection.Create(tenant)
}

func (r *TenantCollection) GetTenantByID(tenantID string) (*core_models.Tenant, error) {
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

func (r *TenantCollection) UpdateTenant(tenant core_models.Tenant) error {
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
