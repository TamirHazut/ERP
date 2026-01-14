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

// PermissionService implements the gRPC PermissionService
type PermissionService struct {
	permissionAPI *api.PermissionAPI
	logger        logger.Logger
	proto_auth.UnimplementedPermissionServiceServer
}

// NewPermissionService creates a new PermissionService instance
func NewPermissionService(permissionAPI *api.PermissionAPI) *PermissionService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	return &PermissionService{
		permissionAPI: permissionAPI,
		logger:        logger,
	}
}

// CreatePermission creates a new permission
func (ps *PermissionService) CreatePermission(ctx context.Context, req *proto_auth.CreatePermissionRequest) (*proto_auth.CreatePermissionResponse, error) {
	ps.logger.Debug("gRPC CreatePermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetPermission() == nil {
		return nil, status.Error(codes.InvalidArgument, "permission data is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Convert proto → domain
	permission, err := convertor.CreatePermissionFromProto(req.GetPermission())
	if err != nil {
		ps.logger.Error("Failed to convert proto to permission", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3. Call API layer (with authorization)
	permissionID, err := ps.permissionAPI.CreatePermission(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		permission,
		req.GetTargetTenantId(),
	)
	if err != nil {
		ps.logger.Error("Failed to create permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.CreatePermissionResponse{PermissionId: permissionID}, nil
}

// UpdatePermission updates an existing permission
func (ps *PermissionService) UpdatePermission(ctx context.Context, req *proto_auth.UpdatePermissionRequest) (*proto_infra.Response, error) {
	ps.logger.Debug("gRPC UpdatePermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		ps.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetPermission() == nil {
		return nil, status.Error(codes.InvalidArgument, "permission data is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Get existing permission
	existingPermission, err := ps.permissionAPI.GetPermissionByID(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetPermission().GetId(),
		req.GetTargetTenantId(),
	)
	if err != nil {
		ps.logger.Error("Failed to get existing permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// 3. Apply updates
	if err := convertor.UpdatePermissionFromProto(existingPermission, req.GetPermission()); err != nil {
		ps.logger.Error("Failed to apply permission updates", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 4. Call API layer (with authorization)
	if err := ps.permissionAPI.UpdatePermission(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		existingPermission,
		req.GetTargetTenantId(),
	); err != nil {
		ps.logger.Error("Failed to update permission", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_infra.Response{
		Success: true,
	}, nil
}

// GetPermission retrieves a permission by ID
func (ps *PermissionService) GetPermission(ctx context.Context, req *proto_auth.GetPermissionRequest) (*proto_auth.GetPermissionResponse, error) {
	ps.logger.Debug("gRPC GetPermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
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

	// 3. Convert domain → proto
	permissionProto := convertor.PermissionToProto(permission)
	if permissionProto == nil {
		return nil, status.Error(codes.Internal, "failed to convert permission to proto")
	}

	return &proto_auth.GetPermissionResponse{Permission: permissionProto}, nil
}

// ListPermissions retrieves all permissions for a tenant
func (ps *PermissionService) ListPermissions(ctx context.Context, req *proto_auth.ListPermissionsRequest) (*proto_auth.ListPermissionsResponse, error) {
	ps.logger.Debug("gRPC ListPermissions called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
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

	// 3. Convert domain → proto
	permissionsProto := make([]*proto_auth.PermissionData, 0, len(permissions))
	for _, permission := range permissions {
		permissionProto := convertor.PermissionToProto(permission)
		if permissionProto != nil {
			permissionsProto = append(permissionsProto, permissionProto)
		}
	}

	return &proto_auth.ListPermissionsResponse{
		Permissions: permissionsProto,
		// Pagination can be added later
	}, nil
}

// DeletePermission deletes a permission
func (ps *PermissionService) DeletePermission(ctx context.Context, req *proto_auth.DeletePermissionRequest) (*proto_infra.Response, error) {
	ps.logger.Debug("gRPC DeletePermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
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

	return &proto_infra.Response{
		Success: true,
	}, nil
}
