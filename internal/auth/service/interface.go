package service

//go:generate mockgen -destination=mock/mock_rbac_checker.go -package=mock erp.localhost/internal/auth/service RBACChecker

import (
	"context"

	auth_proto "erp.localhost/internal/infra/proto/auth/v1"
	infra_proto "erp.localhost/internal/infra/proto/infra/v1"
)

// RBACChecker - an interface for all the non-rbac components of the auth service who needs to verify permissions and roles
type RBACChecker interface {
	VerifyUserPermissions(ctx context.Context, req *auth_proto.VerifyUserResourceRequest) (*auth_proto.VerifyResourceResponse, error)
	VerifyUserRoles(ctx context.Context, req *auth_proto.VerifyUserResourceRequest) (*auth_proto.VerifyResourceResponse, error)

	CreateVerifyPermissionsResourceRequest(identifier *infra_proto.UserIdentifier, permissionIdentifiers ...string) *auth_proto.VerifyUserResourceRequest
	CreateVerifyRolesResourceRequest(identifier *infra_proto.UserIdentifier, roles ...string) *auth_proto.VerifyUserResourceRequest
}
