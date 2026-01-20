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

// VerificationService implements the gRPC VerificationService
type VerificationService struct {
	verificationAPI *api.VerificationAPI
	logger          logger.Logger
	authv1.UnimplementedVerificationServiceServer
}

// NewVerificationService creates a new VerificationService instance
func NewVerificationService(verificationAPI *api.VerificationAPI, logger logger.Logger) *VerificationService {
	return &VerificationService{
		verificationAPI: verificationAPI,
		logger:          logger,
	}
}

// CheckPermissions checks if a user has specific permissions
func (vs *VerificationService) CheckPermissions(ctx context.Context, req *authv1.CheckPermissionsRequest) (*authv1.CheckPermissionsResponse, error) {
	vs.logger.Debug("gRPC CheckPermissions called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return &authv1.CheckPermissionsResponse{Permissions: permissions}, nil
}

// HasPermission checks if a user has a specific permission
func (vs *VerificationService) HasPermission(ctx context.Context, req *authv1.HasPermissionRequest) (*authv1.HasPermissionResponse, error) {
	vs.logger.Debug("gRPC HasPermission called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return &authv1.HasPermissionResponse{HasPermission: hasPermission}, nil
}

// GetUserPermissions retrieves all permissions for a user
func (vs *VerificationService) GetUserPermissions(ctx context.Context, req *authv1.GetUserPermissionsRequest) (*authv1.GetUserPermissionsResponse, error) {
	vs.logger.Debug("gRPC GetUserPermissions called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return &authv1.GetUserPermissionsResponse{Permissions: permissions}, nil
}

// GetUserRoles retrieves all role IDs for a user
func (vs *VerificationService) GetUserRoles(ctx context.Context, req *authv1.GetUserRolesRequest) (*authv1.GetUserRolesResponse, error) {
	vs.logger.Debug("gRPC GetUserRoles called")

	// 1. Validate request
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
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

	return &authv1.GetUserRolesResponse{RoleIds: roleIDs}, nil
}

// IsSystemTenantUser checks if a tenant is the system tenant
func (vs *VerificationService) IsSystemTenantUser(ctx context.Context, req *authv1.IsSystemTenantUserRequest) (*authv1.IsSystemTenantUserResponse, error) {
	vs.logger.Debug("gRPC IsSystemTenantUser called")

	// 1. Validate request
	if req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	// 2. Call API layer (no authorization needed - verification service)
	isSystemTenant := vs.verificationAPI.IsSystemTenantUser(req.GetTenantId())

	return &authv1.IsSystemTenantUserResponse{IsSystemTenant: isSystemTenant}, nil
}
