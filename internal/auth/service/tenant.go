package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	validator_infra "erp.localhost/internal/infra/model/infra/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TenantService struct {
	logger    logger.Logger
	tenantAPI *api.TenantAPI

	authv1.UnimplementedTenantServiceServer
}

func NewTenantService(tenantAPI *api.TenantAPI, logger logger.Logger) *TenantService {
	return &TenantService{
		logger:    logger,
		tenantAPI: tenantAPI,
	}
}

func (t *TenantService) CreateTenant(ctx context.Context, req *authv1.CreateTenantRequest) (*authv1.CreateTenantResponse, error) {
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	tenant := req.GetTenant()
	if tenant == nil {
		t.logger.Error("tenant data is required")
		return nil, status.Error(codes.InvalidArgument, "tenant data is required")
	}

	t.logger.Info("creating tenant", "name", tenant.Name, "requested_by", identifier.UserId)

	tenantID, err := t.tenantAPI.CreateTenant(tenantID, userID, tenant)
	if err != nil {
		t.logger.Error("failed to create tenant", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	t.logger.Info("tenant created in database", "tenant_id", tenantID)

	return &authv1.CreateTenantResponse{TenantId: tenantID}, nil
}

func (t *TenantService) GetTenant(ctx context.Context, req *authv1.GetTenantRequest) (*authv1.Tenant, error) {
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	targetTenantID := req.GetTenantId()
	targetTenantName := req.GetName()

	tenant, err := t.tenantAPI.GetTenant(tenantID, userID, targetTenantID, targetTenantName)
	if err != nil {
		t.logger.Error("failed to get tenant", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	t.logger.Info("tenant retrieved", "tenant_id", tenant.Id)
	return tenant, nil
}

func (t *TenantService) ListTenants(ctx context.Context, req *authv1.ListTenantsRequest) (*authv1.ListTenantsResponse, error) {
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	status := req.GetStatus()

	tenants, err := t.tenantAPI.ListTenants(tenantID, userID, status)
	if err != nil {
		t.logger.Error("failed to get tenants", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	t.logger.Info("tenants retrieved", "count", len(tenants))
	return &authv1.ListTenantsResponse{
		Tenants: tenants,
	}, nil
}

func (t *TenantService) UpdateTenant(ctx context.Context, req *authv1.UpdateTenantRequest) (*authv1.UpdateTenantResponse, error) {
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	tenant := req.GetTenant()

	err := t.tenantAPI.UpdateTenant(tenantID, userID, tenant)
	if err != nil {
		t.logger.Error("failed to update tenant", "tenant_id", tenant.Id, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	t.logger.Info("tenant updated successfully", "tenant_id", tenant.Id)
	return &authv1.UpdateTenantResponse{Updated: true}, nil
}

func (t *TenantService) DeleteTenant(ctx context.Context, req *authv1.DeleteTenantRequest) (*authv1.DeleteTenantResponse, error) {
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	targetTenantID := req.GetTenantId()

	// STEP 8: Delete the tenant itself
	t.logger.Info("deleting tenant", "target_tenant_id", targetTenantID)
	if err := t.tenantAPI.DeleteTenant(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete tenant", "target_tenant_id", targetTenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	t.logger.Info("tenant deleted successfully", "target_tenant_id", targetTenantID)
	return &authv1.DeleteTenantResponse{Deleted: true}, nil
}
