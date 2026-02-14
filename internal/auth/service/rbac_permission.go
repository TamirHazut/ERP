package service

import (
	"context"

	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/logging/logger"
	authv1 "erp.localhost/infra/model/auth/v1"
	validator_infra "erp.localhost/infra/model/infra/validator"
	"erp.localhost/internal/auth/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PermissionService implements the gRPC PermissionService
type PermissionService struct {
	permissionAPI *api.PermissionAPI
	logger        logger.Logger
	authv1.UnimplementedPermissionServiceServer
}

// NewPermissionService creates a new PermissionService instance
func NewPermissionService(permissionAPI *api.PermissionAPI, logger logger.Logger) *PermissionService {
	return &PermissionService{
		permissionAPI: permissionAPI,
		logger:        logger,
	}
}

// GetPermission retrieves a permission by its permission string
func (ps *PermissionService) GetPermission(ctx context.Context, req *authv1.GetPermissionRequest) (*authv1.Permission, error) {
	ps.logger.Debug("gRPC GetPermission called")

	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetPermissionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "permission_id is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	permission, err := ps.permissionAPI.GetPermissionByID(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetPermissionId(),
		req.GetTargetTenantId(),
	)
	if err != nil {
		ps.logger.Error("Failed to get permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	return permission, nil
}

// ListPermissions retrieves all active permissions from the registry
func (ps *PermissionService) ListPermissions(ctx context.Context, req *authv1.ListPermissionsRequest) (*authv1.ListPermissionsResponse, error) {
	ps.logger.Debug("gRPC ListPermissions called")

	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	permissions, err := ps.permissionAPI.ListPermissions(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetTargetTenantId(),
	)
	if err != nil {
		ps.logger.Error("Failed to list permissions", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &authv1.ListPermissionsResponse{
		Permissions: permissions,
	}, nil
}
