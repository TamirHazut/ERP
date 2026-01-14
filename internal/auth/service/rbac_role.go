package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	"erp.localhost/internal/infra/convertor"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/generated/infra/v1"
	"erp.localhost/internal/infra/proto/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RoleService implements the gRPC RoleService
type RoleService struct {
	roleAPI *api.RoleAPI
	logger  logger.Logger
	proto_auth.UnimplementedRoleServiceServer
}

// NewRoleService creates a new RoleService instance
func NewRoleService(roleAPI *api.RoleAPI) *RoleService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	return &RoleService{
		roleAPI: roleAPI,
		logger:  logger,
	}
}

// CreateRole creates a new role
func (rs *RoleService) CreateRole(ctx context.Context, req *proto_auth.CreateRoleRequest) (*proto_auth.CreateRoleResponse, error) {
	rs.logger.Debug("gRPC CreateRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetRole() == nil {
		return nil, status.Error(codes.InvalidArgument, "role data is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Convert proto → domain
	role, err := convertor.CreateRoleFromProto(req.GetRole())
	if err != nil {
		rs.logger.Error("Failed to convert proto to role", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3. Call API layer (with authorization)
	roleID, err := rs.roleAPI.CreateRole(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		role,
		req.GetTargetTenantId(),
	)
	if err != nil {
		rs.logger.Error("Failed to create role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.CreateRoleResponse{RoleId: roleID}, nil
}

// UpdateRole updates an existing role
func (rs *RoleService) UpdateRole(ctx context.Context, req *proto_auth.UpdateRoleRequest) (*proto_infra.Response, error) {
	rs.logger.Debug("gRPC UpdateRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		rs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetRole() == nil {
		return nil, status.Error(codes.InvalidArgument, "role data is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Get existing role
	existingRole, err := rs.roleAPI.GetRoleByID(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetRole().GetId(),
		req.GetTargetTenantId(),
	)
	if err != nil {
		rs.logger.Error("Failed to get existing role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// 3. Apply updates
	if err := convertor.UpdateRoleFromProto(existingRole, req.GetRole()); err != nil {
		rs.logger.Error("Failed to apply role updates", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 4. Call API layer (with authorization)
	if err := rs.roleAPI.UpdateRole(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		existingRole,
		req.GetTargetTenantId(),
	); err != nil {
		rs.logger.Error("Failed to update role", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_infra.Response{
		Success: true,
	}, nil
}

// GetRole retrieves a role by ID
func (rs *RoleService) GetRole(ctx context.Context, req *proto_auth.GetRoleRequest) (*proto_auth.GetRoleResponse, error) {
	rs.logger.Debug("gRPC GetRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
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

	// 3. Convert domain → proto
	roleProto := convertor.RoleToProto(role)
	if roleProto == nil {
		return nil, status.Error(codes.Internal, "failed to convert role to proto")
	}

	return &proto_auth.GetRoleResponse{Role: roleProto}, nil
}

// ListRoles retrieves all roles for a tenant
func (rs *RoleService) ListRoles(ctx context.Context, req *proto_auth.ListRolesRequest) (*proto_auth.ListRolesResponse, error) {
	rs.logger.Debug("gRPC ListRoles called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
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

	// 3. Convert domain → proto
	rolesProto := make([]*proto_auth.RoleData, 0, len(roles))
	for _, role := range roles {
		roleProto := convertor.RoleToProto(role)
		if roleProto != nil {
			rolesProto = append(rolesProto, roleProto)
		}
	}

	return &proto_auth.ListRolesResponse{
		Roles: rolesProto,
		// Pagination can be added later
	}, nil
}

// DeleteRole deletes a role
func (rs *RoleService) DeleteRole(ctx context.Context, req *proto_auth.DeleteRoleRequest) (*proto_infra.Response, error) {
	rs.logger.Debug("gRPC DeleteRole called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
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

	return &proto_infra.Response{
		Success: true,
	}, nil
}
