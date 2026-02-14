package handler

import (
	aggregation_mongo "erp.localhost/infra/db/mongo/aggregation"
	collection_mongo "erp.localhost/infra/db/mongo/collection"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	authv1 "erp.localhost/infra/model/auth/v1"
	validator_auth "erp.localhost/infra/model/auth/validator"
	aggregation_auth "erp.localhost/internal/auth/aggregation"
	collection_auth "erp.localhost/internal/auth/collection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TenantHandler struct {
	collection  collection_mongo.CollectionHandler[authv1.Tenant]
	aggregation aggregation_mongo.AggregationHandler[authv1.Tenant]
	logger      logger.Logger
}

func NewTenantHandler(logger logger.Logger) (*TenantHandler, *infra_error.AppError) {
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

func (t TenantHandler) CreateTenant(tenant *authv1.Tenant) (string, *infra_error.AppError) {
	if err := validator_auth.ValidateTenant(tenant, true); err != nil {
		return "", err
	}
	tenant.CreatedAt = timestamppb.Now()
	tenant.UpdatedAt = timestamppb.Now()
	t.logger.Debug("Creating tenant", "tenant", tenant)
	return t.collection.Create(tenant)
}

func (t TenantHandler) GetTenantByID(tenantID string) (*authv1.Tenant, *infra_error.AppError) {
	if tenantID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	t.logger.Debug("Getting tenant by id", "filter", filter)
	return t.findTenantByFilter(filter)
}

func (t TenantHandler) GetTenantByName(name string) (*authv1.Tenant, *infra_error.AppError) {
	if name == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"name": name,
	}
	t.logger.Debug("Getting tenant by id", "filter", filter)
	return t.findTenantByFilter(filter)
}

func (t TenantHandler) GetTenants() ([]*authv1.Tenant, *infra_error.AppError) {
	t.logger.Debug("Getting all tenants")
	return t.findTenantsByFilter(nil)
}

func (t TenantHandler) GetTenantsByStatus(status string) ([]*authv1.Tenant, *infra_error.AppError) {
	if status == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "status")
	}
	filter := map[string]any{
		"status": status,
	}
	t.logger.Debug("Getting all tenants by status")
	return t.findTenantsByFilter(filter)
}

func (t TenantHandler) UpdateTenant(newTenant *authv1.Tenant) *infra_error.AppError {
	if err := validator_auth.ValidateTenant(newTenant, false); err != nil {
		return err
	}
	filter := map[string]any{
		"_id": newTenant.Id,
	}
	t.logger.Debug("Updating tenant", "tenant", newTenant)
	oldTenant, err := t.GetTenantByID(newTenant.Id)
	if err != nil {
		return err
	}
	newTenant.Protected = oldTenant.Protected
	if oldTenant.Protected {
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}
	if newTenant.Id != oldTenant.Id ||
		newTenant.Name != oldTenant.Name {
		return infra_error.Validation(infra_error.ValidationTryToChangeRestrictedFields)
	}
	newTenant.CreatedAt = oldTenant.CreatedAt
	newTenant.CreatedBy = oldTenant.CreatedBy
	newTenant.UpdatedAt = timestamppb.Now()
	return t.collection.Update(filter, newTenant)
}

func (t TenantHandler) DeleteTenant(tenantID string) *infra_error.AppError {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "TenantId")
	}
	filter := map[string]any{
		"_id": tenantID,
	}
	t.logger.Debug("Deleting tenant", "filter", filter)
	return t.collection.Delete(filter)
}

func (t TenantHandler) findTenantByFilter(filter map[string]any) (*authv1.Tenant, *infra_error.AppError) {
	if len(filter) == 0 {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "filter")
	}
	tenant, err := t.collection.FindOne(filter)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}
func (t TenantHandler) findTenantsByFilter(filter map[string]any) ([]*authv1.Tenant, *infra_error.AppError) {
	tenants, err := t.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}
