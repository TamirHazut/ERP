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

func (t *TenantCollection) CreateTenant(tenant *model_core.Tenant) (string, error) {
	if err := tenant.Validate(true); err != nil {
		return "", err
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	t.logger.Debug("Creating tenant", "tenant", tenant)
	return t.collection.Create(tenant)
}

func (t *TenantCollection) GetTenantByID(tenantID string) (*model_core.Tenant, error) {
	if tenantID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	t.logger.Debug("Getting tenant by id", "filter", filter)
	return t.findTenantByFilter(filter)
}

func (t *TenantCollection) GetTenantByName(name string) (*model_core.Tenant, error) {
	if name == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"name": name,
	}
	t.logger.Debug("Getting tenant by id", "filter", filter)
	return t.findTenantByFilter(filter)
}

func (t *TenantCollection) GetTenants() ([]*model_core.Tenant, error) {
	t.logger.Debug("Getting all tenants")
	return t.findTenantsByFilter(nil)
}

func (t *TenantCollection) GetTenantsByStatus(status string) ([]*model_core.Tenant, error) {
	if status == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "status")
	}
	filter := map[string]any{
		"status": status,
	}
	t.logger.Debug("Getting all tenants by status")
	return t.findTenantsByFilter(filter)
}

func (t *TenantCollection) UpdateTenant(tenant *model_core.Tenant) error {
	if err := tenant.Validate(false); err != nil {
		return err
	}
	filter := map[string]any{
		"_id": tenant.ID,
	}
	t.logger.Debug("Updating tenant", "tenant", tenant)
	currentTenant, err := t.GetTenantByID(tenant.ID.Hex())
	if err != nil {
		return err
	}
	if tenant.CreatedAt != currentTenant.CreatedAt {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields, "CreatedAt")
	}
	tenant.UpdatedAt = time.Now()
	return t.collection.Update(filter, tenant)
}

func (t *TenantCollection) DeleteTenant(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantID")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	t.logger.Debug("Deleting tenant", "filter", filter)
	return t.collection.Delete(filter)
}

func (t *TenantCollection) findTenantByFilter(filter map[string]any) (*model_core.Tenant, error) {
	if len(filter) == 0 {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "filter")
	}
	tenant, err := t.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}
func (t *TenantCollection) findTenantsByFilter(filter map[string]any) ([]*model_core.Tenant, error) {
	tenants, err := t.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}
