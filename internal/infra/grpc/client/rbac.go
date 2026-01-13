package client

import (
	"context"
	"fmt"

	"erp.localhost/internal/infra/convertor"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	infrav1 "erp.localhost/internal/infra/proto/infra/v1"
)

//go:generate mockgen -destination=mock/rbac_go -package=mock erp.localhost/internal/infra/grpc/client RBACClient

// RBACClient defines the interface for calling Auth RBAC service
type RBACClient interface {
	// Role operations
	CreateRole(ctx context.Context, tenantID, userID string, role *model_auth.Role) (string, error)
	UpdateRole(ctx context.Context, tenantID, userID string, role *model_auth.Role) error
	GetRole(ctx context.Context, tenantID, userID, roleID string) (*model_auth.Role, error)
	ListRoles(ctx context.Context, tenantID, userID string, status *string, page, pageSize int32) ([]*model_auth.Role, *PaginationResponse, error)
	DeleteRole(ctx context.Context, tenantID, userID, roleID string) error

	// Permission operations
	CreatePermission(ctx context.Context, tenantID, userID string, permission *model_auth.Permission) (string, error)
	UpdatePermission(ctx context.Context, tenantID, userID string, permission *model_auth.Permission) error
	GetPermission(ctx context.Context, tenantID, userID, permissionID string) (*model_auth.Permission, error)
	ListPermissions(ctx context.Context, tenantID, userID string, status *string, page, pageSize int32) ([]*model_auth.Permission, *PaginationResponse, error)
	DeletePermission(ctx context.Context, tenantID, userID, permissionID string) error

	// Verification operations
	VerifyUserPermissions(ctx context.Context, tenantID, userID string, permissions []PermissionCheck) ([]PermissionResult, error)
	VerifyUserRoles(ctx context.Context, tenantID, userID string, roles []RoleCheck) ([]RoleResult, error)

	Close() error
}

// Helper types for verification
type PermissionCheck struct {
	PermissionID *string // Either PermissionID or Resource+Action must be set
	Resource     *string
	Action       *string
}

type PermissionResult struct {
	PermissionID  *string
	Resource      *string
	Action        *string
	HasPermission bool
}

type RoleCheck struct {
	RoleID   *string // Either RoleID or RoleName must be set
	RoleName *string
}

type RoleResult struct {
	RoleID   *string
	RoleName *string
	HasRole  bool
}

type PaginationResponse struct {
	CurrentPage  int32
	PageSize     int32
	TotalPages   int32
	TotalRecords int64
}

// rbacClient implements RBACClient
type rbacClient struct {
	grpcClient *GRPCClient
	logger     logger.Logger
	stub       proto_auth.RBACServiceClient
}

// NewRBACClient creates a new RBAC service client
func NewRBACClient(ctx context.Context, config *Config, logger logger.Logger) (RBACClient, error) {
	grpcClient, err := NewGRPCClient(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	stub := proto_auth.NewRBACServiceClient(grpcClient.Conn())

	return &rbacClient{
		grpcClient: grpcClient,
		logger:     logger,
		stub:       stub,
	}, nil
}

// ============================================================================
// Role Operations
// ============================================================================

func (c *rbacClient) CreateRole(ctx context.Context, tenantID, userID string, role *model_auth.Role) (string, error) {
	createRoleData, err := convertor.RoleToCreateProto(role)
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}
	req := &proto_auth.CreateResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Resource: &proto_auth.CreateResourceRequest_Role{
			Role: createRoleData,
		},
	}

	resp, err := c.stub.CreateResource(ctx, req)
	if err != nil {
		return "", mapGRPCError(err)
	}

	return resp.ResourceId, nil
}

func (c *rbacClient) UpdateRole(ctx context.Context, tenantID, userID string, role *model_auth.Role) error {
	updateRoleData, err := convertor.RoleToUpdateProto(role)
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	req := &proto_auth.UpdateResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Resource: &proto_auth.UpdateResourceRequest_Role{
			Role: updateRoleData,
		},
	}

	_, err = c.stub.UpdateResource(ctx, req)
	if err != nil {
		return mapGRPCError(err)
	}

	return nil
}

func (c *rbacClient) GetRole(ctx context.Context, tenantID, userID, roleID string) (*model_auth.Role, error) {
	req := &proto_auth.GetResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Id:           roleID,
	}

	resp, err := c.stub.GetResource(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	if resp.Resource == nil || resp.Resource.GetRole() == nil {
		return nil, fmt.Errorf("invalid response: role not found")
	}

	return convertor.ProtoToRole(resp.Resource.GetRole())
}

func (c *rbacClient) ListRoles(ctx context.Context, tenantID, userID string, status *string, page, pageSize int32) ([]*model_auth.Role, *PaginationResponse, error) {
	req := &proto_auth.ListResourcesRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
	}

	if status != nil {
		req.Status = status
	}

	if page > 0 && pageSize > 0 {
		req.Pagination = &infrav1.PaginationRequest{
			Page:     page,
			PageSize: pageSize,
		}
	}

	resp, err := c.stub.ListResources(ctx, req)
	if err != nil {
		return nil, nil, mapGRPCError(err)
	}

	roles := make([]*model_auth.Role, 0, len(resp.Resources))
	for _, resource := range resp.Resources {
		if roleProto := resource.GetRole(); roleProto != nil {
			role, err := convertor.ProtoToRole(roleProto)
			if err != nil {
				c.logger.Error(err.Error(), nil)
				continue
			}
			roles = append(roles, role)
		}
	}

	var pagination *PaginationResponse
	if resp.Pagination != nil {
		pagination = &PaginationResponse{
			CurrentPage:  resp.Pagination.Page,
			PageSize:     resp.Pagination.PageSize,
			TotalPages:   resp.Pagination.TotalPages,
			TotalRecords: resp.Pagination.TotalItems,
		}
	}

	return roles, pagination, nil
}

func (c *rbacClient) DeleteRole(ctx context.Context, tenantID, userID, roleID string) error {
	req := &proto_auth.DeleteResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Resource: &proto_auth.DeleteResourceRequest_ResourceId{
			ResourceId: roleID,
		},
	}

	_, err := c.stub.DeleteResource(ctx, req)
	if err != nil {
		return mapGRPCError(err)
	}

	return nil
}

// ============================================================================
// Permission Operations (similar pattern to roles)
// ============================================================================

func (c *rbacClient) CreatePermission(ctx context.Context, tenantID, userID string, permission *model_auth.Permission) (string, error) {
	createPermissionData, err := convertor.PermissionToCreateProto(permission)
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}
	req := &proto_auth.CreateResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resource: &proto_auth.CreateResourceRequest_Permission{
			Permission: createPermissionData,
		},
	}

	resp, err := c.stub.CreateResource(ctx, req)
	if err != nil {
		return "", mapGRPCError(err)
	}

	return resp.ResourceId, nil
}

func (c *rbacClient) UpdatePermission(ctx context.Context, tenantID, userID string, permission *model_auth.Permission) error {
	updatePermissionData, err := convertor.PermissionToUpdateProto(permission)
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}
	req := &proto_auth.UpdateResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resource: &proto_auth.UpdateResourceRequest_Permission{
			Permission: updatePermissionData,
		},
	}

	_, err = c.stub.UpdateResource(ctx, req)
	if err != nil {
		return mapGRPCError(err)
	}

	return nil
}

func (c *rbacClient) GetPermission(ctx context.Context, tenantID, userID, permissionID string) (*model_auth.Permission, error) {
	req := &proto_auth.GetResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Id:           permissionID,
	}

	resp, err := c.stub.GetResource(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	if resp.Resource == nil || resp.Resource.GetPermission() == nil {
		return nil, fmt.Errorf("invalid response: permission not found")
	}

	return convertor.ProtoToPermission(resp.Resource.GetPermission())
}

func (c *rbacClient) ListPermissions(ctx context.Context, tenantID, userID string, status *string, page, pageSize int32) ([]*model_auth.Permission, *PaginationResponse, error) {
	req := &proto_auth.ListResourcesRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
	}

	if status != nil {
		req.Status = status
	}

	if page > 0 && pageSize > 0 {
		req.Pagination = &infrav1.PaginationRequest{
			Page:     page,
			PageSize: pageSize,
		}
	}

	resp, err := c.stub.ListResources(ctx, req)
	if err != nil {
		return nil, nil, mapGRPCError(err)
	}

	permissions := make([]*model_auth.Permission, 0, len(resp.Resources))
	for _, resource := range resp.Resources {
		if permProto := resource.GetPermission(); permProto != nil {
			perm, err := convertor.ProtoToPermission(permProto)
			if err != nil {
				c.logger.Error(err.Error())
				continue
			}
			permissions = append(permissions, perm)
		}
	}

	var pagination *PaginationResponse
	if resp.Pagination != nil {
		pagination = &PaginationResponse{
			CurrentPage:  resp.Pagination.Page,
			PageSize:     resp.Pagination.PageSize,
			TotalPages:   resp.Pagination.TotalPages,
			TotalRecords: resp.Pagination.TotalItems,
		}
	}

	return permissions, pagination, nil
}

func (c *rbacClient) DeletePermission(ctx context.Context, tenantID, userID, permissionID string) error {
	req := &proto_auth.DeleteResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resource: &proto_auth.DeleteResourceRequest_ResourceId{
			ResourceId: permissionID,
		},
	}

	_, err := c.stub.DeleteResource(ctx, req)
	if err != nil {
		return mapGRPCError(err)
	}

	return nil
}

// ============================================================================
// Verification Operations
// ============================================================================

func (c *rbacClient) VerifyUserPermissions(ctx context.Context, tenantID, userID string, permissions []PermissionCheck) ([]PermissionResult, error) {
	verifyResources := make([]*proto_auth.VerifyResource, 0, len(permissions))

	for _, perm := range permissions {
		var permission *proto_auth.Permission

		if perm.PermissionID != nil {
			permission = &proto_auth.Permission{
				Identifier: &proto_auth.Permission_PermissionId{
					PermissionId: *perm.PermissionID,
				},
			}
		} else if perm.Resource != nil && perm.Action != nil {
			permission = &proto_auth.Permission{
				Identifier: &proto_auth.Permission_Permission{
					Permission: &proto_auth.PermissionIdentifier{
						Resource: *perm.Resource,
						Action:   *perm.Action,
					},
				},
			}
		} else {
			return nil, fmt.Errorf("invalid permission check: must specify either permission_id or resource+action")
		}

		verifyResources = append(verifyResources, &proto_auth.VerifyResource{
			Resource: &proto_auth.VerifyResource_Permission{
				Permission: permission,
			},
		})
	}

	req := &proto_auth.VerifyUserResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resources:    verifyResources,
	}

	resp, err := c.stub.VerifyUserResource(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	results := make([]PermissionResult, 0, len(resp.Resources))
	for _, resource := range resp.Resources {
		if perm := resource.GetPermission(); perm != nil {
			result := PermissionResult{}

			if permID := perm.GetPermissionId(); permID != "" {
				result.PermissionID = &permID
			} else if permIdent := perm.GetPermission(); permIdent != nil {
				result.Resource = &permIdent.Resource
				result.Action = &permIdent.Action
			}

			if perm.HasPermission != nil {
				result.HasPermission = perm.HasPermission.Value
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func (c *rbacClient) VerifyUserRoles(ctx context.Context, tenantID, userID string, roles []RoleCheck) ([]RoleResult, error) {
	verifyResources := make([]*proto_auth.VerifyResource, 0, len(roles))

	for _, roleCheck := range roles {
		var role *proto_auth.Role

		if roleCheck.RoleID != nil {
			role = &proto_auth.Role{
				Identifier: &proto_auth.Role_RoleId{
					RoleId: *roleCheck.RoleID,
				},
			}
		} else if roleCheck.RoleName != nil {
			role = &proto_auth.Role{
				Identifier: &proto_auth.Role_RoleName{
					RoleName: *roleCheck.RoleName,
				},
			}
		} else {
			return nil, fmt.Errorf("invalid role check: must specify either role_id or role_name")
		}

		verifyResources = append(verifyResources, &proto_auth.VerifyResource{
			Resource: &proto_auth.VerifyResource_Role{
				Role: role,
			},
		})
	}

	req := &proto_auth.VerifyUserResourceRequest{
		Identifier: &infrav1.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Resources:    verifyResources,
	}

	resp, err := c.stub.VerifyUserResource(ctx, req)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	results := make([]RoleResult, 0, len(resp.Resources))
	for _, resource := range resp.Resources {
		if role := resource.GetRole(); role != nil {
			result := RoleResult{}

			if roleID := role.GetRoleId(); roleID != "" {
				result.RoleID = &roleID
			} else if roleName := role.GetRoleName(); roleName != "" {
				result.RoleName = &roleName
			}

			if role.HasRole != nil {
				result.HasRole = role.HasRole.Value
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func (c *rbacClient) Close() error {
	return c.grpcClient.Close()
}
