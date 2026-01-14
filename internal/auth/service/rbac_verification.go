package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/generated/auth/v1"
	"erp.localhost/internal/infra/proto/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VerificationService implements the gRPC VerificationService
type VerificationService struct {
	verificationAPI *api.VerificationAPI
	logger          logger.Logger
	proto_auth.UnimplementedVerificationServiceServer
}

// NewVerificationService creates a new VerificationService instance
func NewVerificationService(
	verificationAPI *api.VerificationAPI,
) *VerificationService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	return &VerificationService{
		verificationAPI: verificationAPI,
		logger:          logger,
	}
}

// CheckPermissions checks if a user has specific permissions
func (vs *VerificationService) CheckPermissions(ctx context.Context, req *proto_auth.CheckPermissionsRequest) (*proto_auth.CheckPermissionsResponse, error) {
	vs.logger.Debug("gRPC CheckPermissions called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		vs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if len(req.GetPermissions()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "permissions list cannot be empty")
	}

	// 2. Call API layer (no authorization needed - verification service)
	permissions, err := vs.verificationAPI.CheckPermissions(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetPermissions(),
	)
	if err != nil {
		vs.logger.Error("Failed to check permissions", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.CheckPermissionsResponse{Permissions: permissions}, nil
}

// HasPermission checks if a user has a specific permission
func (vs *VerificationService) HasPermission(ctx context.Context, req *proto_auth.HasPermissionRequest) (*proto_auth.HasPermissionResponse, error) {
	vs.logger.Debug("gRPC HasPermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		vs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	if req.GetPermission() == "" {
		return nil, status.Error(codes.InvalidArgument, "permission is required")
	}
	if req.GetTargetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "target_tenant_id is required")
	}

	// 2. Call API layer (no authorization needed - verification service)
	err := vs.verificationAPI.HasPermission(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
		req.GetPermission(),
		req.GetTargetTenantId(),
	)

	// 3. Convert error to boolean response
	hasPermission := err == nil

	return &proto_auth.HasPermissionResponse{HasPermission: hasPermission}, nil
}

// GetUserPermissions retrieves all permissions for a user
func (vs *VerificationService) GetUserPermissions(ctx context.Context, req *proto_auth.GetUserPermissionsRequest) (*proto_auth.GetUserPermissionsResponse, error) {
	vs.logger.Debug("gRPC GetUserPermissions called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		vs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// 2. Call API layer (no authorization needed - verification service)
	permissions, err := vs.verificationAPI.GetUserPermissions(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
	)
	if err != nil {
		vs.logger.Error("Failed to get user permissions", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.GetUserPermissionsResponse{Permissions: permissions}, nil
}

// GetUserRoles retrieves all role IDs for a user
func (vs *VerificationService) GetUserRoles(ctx context.Context, req *proto_auth.GetUserRolesRequest) (*proto_auth.GetUserRolesResponse, error) {
	vs.logger.Debug("gRPC GetUserRoles called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		vs.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// 2. Call API layer (no authorization needed - verification service)
	roleIDs, err := vs.verificationAPI.GetUserRoles(
		req.GetIdentifier().GetTenantId(),
		req.GetIdentifier().GetUserId(),
	)
	if err != nil {
		vs.logger.Error("Failed to get user roles", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	return &proto_auth.GetUserRolesResponse{RoleIds: roleIDs}, nil
}

// IsSystemTenantUser checks if a tenant is the system tenant
func (vs *VerificationService) IsSystemTenantUser(ctx context.Context, req *proto_auth.IsSystemTenantUserRequest) (*proto_auth.IsSystemTenantUserResponse, error) {
	vs.logger.Debug("gRPC IsSystemTenantUser called")

	// 1. Validate request
	if req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	// 2. Call API layer (no authorization needed - verification service)
	isSystemTenant := vs.verificationAPI.IsSystemTenantUser(req.GetTenantId())

	return &proto_auth.IsSystemTenantUserResponse{IsSystemTenant: isSystemTenant}, nil
}
