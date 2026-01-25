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

type TenantHandler struct {
	collection  collection_mongo.CollectionHandler[authv1.Tenant]
	aggregation aggregation_mongo.AggregationHandler[authv1.Tenant]
	logger      logger.Logger
}

func NewTenantHandler(logger logger.Logger) (*TenantHandler, error) {
	collection, err := collection_auth.NewTenantCollection(logger)
	if err != nil {
		logger.Error("failed to create user collection handler", "error", err)
		return nil, err
	}
	aggregation, err := aggregation_auth.NewTenantAggregationHandler(logger)
	if err != nil {
		logger.Error("failed to create user aggregation handler", "error", err)
		return nil, err
	}
	return &TenantHandler{
		collection:  collection,
		aggregation: aggregation,
		logger:      logger,
	}, nil
}

func (t TenantHandler) CreateTenant(tenant *authv1.Tenant) (string, error) {
	if err := validator_auth.ValidateTenant(tenant, true); err != nil {
		return "", err
	}
	tenant.CreatedAt = timestamppb.Now()
	tenant.UpdatedAt = timestamppb.Now()
	t.logger.Debug("Creating tenant", "tenant", tenant)
	tenant.Name = strings.ToLower(tenant.Name)
	return t.collection.Create(tenant)
}

func (t TenantHandler) GetTenantByID(tenantID string) (*authv1.Tenant, error) {
	if tenantID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	t.logger.Debug("Getting tenant by id", "filter", filter)
	return t.findTenantByFilter(filter)
}

func (t TenantHandler) GetTenantByName(name string) (*authv1.Tenant, error) {
	if name == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"name": strings.ToLower(name),
	}
	t.logger.Debug("Getting tenant by id", "filter", filter)
	return t.findTenantByFilter(filter)
}

func (t TenantHandler) GetTenants() ([]*authv1.Tenant, error) {
	t.logger.Debug("Getting all tenants")
	return t.findTenantsByFilter(nil)
}

func (t TenantHandler) GetTenantsByStatus(status string) ([]*authv1.Tenant, error) {
	if status == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "status")
	}
	filter := map[string]any{
		"status": status,
	}
	t.logger.Debug("Getting all tenants by status")
	return t.findTenantsByFilter(filter)
}

func (t TenantHandler) UpdateTenant(tenant *authv1.Tenant) error {
	if err := validator_auth.ValidateTenant(tenant, false); err != nil {
		return err
	}
	filter := map[string]any{
		"_id": tenant.Id,
	}
	t.logger.Debug("Updating tenant", "tenant", tenant)
	currentTenant, err := t.GetTenantByID(tenant.Id)
	if err != nil {
		return err
	}
	if tenant.Id != currentTenant.Id ||
		tenant.Name != currentTenant.Name ||
		tenant.CreatedAt != currentTenant.CreatedAt ||
		tenant.CreatedBy != currentTenant.CreatedBy {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields)
	}
	tenant.UpdatedAt = timestamppb.Now()
	return t.collection.Update(filter, tenant)
}

func (t TenantHandler) DeleteTenant(tenantID string) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	t.logger.Debug("Deleting tenant", "filter", filter)
	return t.collection.Delete(filter)
}

func (t TenantHandler) findTenantByFilter(filter map[string]any) (*authv1.Tenant, error) {
	if len(filter) == 0 {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "filter")
	}
	tenant, err := t.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}
func (t TenantHandler) findTenantsByFilter(filter map[string]any) ([]*authv1.Tenant, error) {
	tenants, err := t.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}
