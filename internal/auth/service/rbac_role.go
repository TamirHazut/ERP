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

// RoleService implements the gRPC RoleService
type RoleService struct {
	roleAPI *api.RoleAPI
	logger  logger.Logger
	authv1.UnimplementedRoleServiceServer
}

// NewRoleService creates a new RoleService instance
func NewRoleService(roleAPI *api.RoleAPI, logger logger.Logger) *RoleService {
	return &RoleService{
		roleAPI: roleAPI,
		logger:  logger,
	}
}

// CreateRole creates a new role
func (rs *RoleService) CreateRole(ctx context.Context, req *authv1.CreateRoleRequest) (*authv1.CreateRoleResponse, error) {
	rs.logger.Debug("gRPC CreateRole called")

	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	role := req.GetRole()
	targetTenantID := req.GetRole().GetTenantId()

	roleID, err := rs.roleAPI.CreateRole(tenantID, userID, role, targetTenantID)
	if err != nil {
		rs.logger.Error("Failed to create role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &authv1.CreateRoleResponse{RoleId: roleID}, nil
}

// UpdateRole updates an existing role
func (rs *RoleService) UpdateRole(ctx context.Context, req *authv1.UpdateRoleRequest) (*infrav1.Response, error) {
	rs.logger.Debug("gRPC UpdateRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	role := req.GetRole()
	targetTenantID := req.GetRole().GetTenantId()

	// 2. Check if role exists
	existingRole, err := rs.roleAPI.GetRoleByID(tenantID, userID, role.GetId(), targetTenantID)
	if err != nil || existingRole == nil {
		rs.logger.Error("Failed to get existing role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// 4. Call API layer (with authorization)
	if err := rs.roleAPI.UpdateRole(tenantID, userID, role, targetTenantID); err != nil {
		rs.logger.Error("Failed to update role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &infrav1.Response{
		Success: true,
	}, nil
}

// GetRole retrieves a role by ID
func (rs *RoleService) GetRole(ctx context.Context, req *authv1.GetRoleRequest) (*authv1.Role, error) {
	rs.logger.Debug("gRPC GetRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetRoleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Call API layer (with authorization)
	role, err := rs.roleAPI.GetRoleByID(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetRoleId(),
		req.GetTargetTenantId(),
	)
	if err != nil {
		rs.logger.Error("Failed to get role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	return role, nil
}

// ListRoles retrieves all roles for a tenant
func (rs *RoleService) ListRoles(ctx context.Context, req *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	rs.logger.Debug("gRPC ListRoles called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Call API layer (with authorization)
	roles, err := rs.roleAPI.ListRoles(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetTargetTenantId(),
	)
	if err != nil {
		rs.logger.Error("Failed to list roles", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &authv1.ListRolesResponse{
		Roles: roles,
		// Pagination can be added later
	}, nil
}

// DeleteRole deletes a role
func (rs *RoleService) DeleteRole(ctx context.Context, req *authv1.DeleteRoleRequest) (*infrav1.Response, error) {
	rs.logger.Debug("gRPC DeleteRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetRoleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Call API layer (with authorization)
	if err := rs.roleAPI.DeleteRole(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetRoleId(),
		req.GetTargetTenantId(),
	); err != nil {
		rs.logger.Error("Failed to delete role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &infrav1.Response{
		Success: true,
	}, nil
}
