package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	infrav1 "erp.localhost/internal/infra/model/infra/v1"
	validator_infra "erp.localhost/internal/infra/model/infra/validator"
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

// CreatePermission creates a new permission
func (ps *PermissionService) CreatePermission(ctx context.Context, req *authv1.CreatePermissionRequest) (*authv1.CreatePermissionResponse, error) {
	ps.logger.Debug("gRPC CreatePermission called")

	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	permission := req.GetPermission()
	targetTenantID := req.GetPermission().GetTenantId()

	permissionID, err := ps.permissionAPI.CreatePermission(tenantID, userID, permission, targetTenantID)
	if err != nil {
		ps.logger.Error("Failed to create permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &authv1.CreatePermissionResponse{PermissionId: permissionID}, nil
}

// UpdatePermission updates an existing permission
func (ps *PermissionService) UpdatePermission(ctx context.Context, req *authv1.UpdatePermissionRequest) (*infrav1.Response, error) {
	ps.logger.Debug("gRPC UpdatePermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	permission := req.GetPermission()
	targetTenantID := req.GetPermission().GetTenantId()

	// 2. Get existing permission
	existingPermission, err := ps.permissionAPI.GetPermissionByID(tenantID, userID, permission.GetId(), targetTenantID)
	if err != nil || existingPermission == nil {
		ps.logger.Error("Failed to get existing permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// 4. Call API layer (with authorization)
	if err := ps.permissionAPI.UpdatePermission(tenantID, userID, permission, targetTenantID); err != nil {
		ps.logger.Error("Failed to update permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &infrav1.Response{
		Success: true,
	}, nil
}

// GetPermission retrieves a permission by ID
func (ps *PermissionService) GetPermission(ctx context.Context, req *authv1.GetPermissionRequest) (*authv1.Permission, error) {
	ps.logger.Debug("gRPC GetPermission called")

	// 1. Validate request
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

	// 2. Call API layer (with authorization)
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

// ListPermissions retrieves all permissions for a tenant
func (ps *PermissionService) ListPermissions(ctx context.Context, req *authv1.ListPermissionsRequest) (*authv1.ListPermissionsResponse, error) {
	ps.logger.Debug("gRPC ListPermissions called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Call API layer (with authorization)
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
		// Pagination can be added later
	}, nil
}

// DeletePermission deletes a permission
func (ps *PermissionService) DeletePermission(ctx context.Context, req *authv1.DeletePermissionRequest) (*infrav1.Response, error) {
	ps.logger.Debug("gRPC DeletePermission called")

	// 1. Validate request
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

	// 2. Call API layer (with authorization)
	if err := ps.permissionAPI.DeletePermission(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetPermissionId(),
		req.GetTargetTenantId(),
	); err != nil {
		ps.logger.Error("Failed to delete permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &infrav1.Response{
		Success: true,
	}, nil
}
