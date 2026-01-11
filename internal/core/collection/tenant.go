package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_core "erp.localhost/internal/infra/model/core"
)

type TenantCollection struct {
	collection collection.CollectionHandler[model_core.Tenant]
	logger     logger.Logger
}

func NewTenantCollection(collection collection.CollectionHandler[model_core.Tenant], logger logger.Logger) *TenantCollection {
	return &TenantCollection{
		collection: collection,
		logger:     logger,
	}
}

func (r *TenantCollection) CreateTenant(tenant *model_core.Tenant) (string, error) {
	if err := tenant.Validate(true); err != nil {
		return "", err
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	r.logger.Debug("Creating tenant", "tenant", tenant)
	return r.collection.Create(tenant)
}

func (r *TenantCollection) GetTenantByID(tenantID string) (*model_core.Tenant, error) {
	if tenantID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
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

func (r *TenantCollection) UpdateTenant(tenant *model_core.Tenant) error {
	if err := tenant.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"_id": tenant.ID,
	}
	r.logger.Debug("Updating tenant", "tenant", tenant)
	currentTenant, err := r.GetTenantByID(tenant.ID.Hex())
	if err != nil {
		return err
	}
	if tenant.CreatedAt != currentTenant.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	tenant.UpdatedAt = time.Now()
	return r.collection.Update(filter, tenant)
}

func (r *TenantCollection) DeleteTenant(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	r.logger.Debug("Deleting tenant", "filter", filter)
	return r.collection.Delete(filter)
}
