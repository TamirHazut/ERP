package service

import (
	"context"
	"fmt"

	collection "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/infra/convertor"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/infra/v1"
	"erp.localhost/internal/infra/proto/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TenantService struct {
	logger           logger.Logger
	tenantCollection *collection.TenantCollection
	tenantSeeder     *TenantSeeder
	authGRPCClient   proto_auth.AuthServiceClient
	rbacGRPCClient   proto_auth.RBACServiceClient
	proto_auth.UnimplementedTenantServiceServer
}

func NewTenantService(authGRPCClient proto_auth.AuthServiceClient, rbacGRPCClient proto_auth.RBACServiceClient) *TenantService {
	logger := logger.NewBaseLogger(model_shared.ModuleCore)

	tc := mongo_collection.NewBaseCollectionHandler[model_auth.Tenant](model_mongo.CoreDB, model_mongo.TenantsCollection, logger)
	if tc == nil {
		logger.Fatal("failed to create tenants collection handler")
		return nil
	}

	tenantCollection := collection.NewTenantCollection(tc, logger)
	if tenantCollection == nil {
		logger.Fatal("failed to create tenants collection")
		return nil
	}

	tenantSeeder := NewTenantSeeder(rbacGRPCClient, logger)
	if tenantSeeder == nil {
		logger.Fatal("failed to create tenant seeder")
		return nil
	}

	return &TenantService{
		logger:           logger,
		tenantCollection: tenantCollection,
		tenantSeeder:     tenantSeeder,
		authGRPCClient:   authGRPCClient,
		rbacGRPCClient:   rbacGRPCClient,
	}
}

// checkPermission verifies if a user has the required permission
func (t *TenantService) checkPermission(ctx context.Context, tenantID, userID, resource, action string) error {
	// Create permission string using helper
	permString, err := model_auth.CreatePermissionString(resource, action)
	if err != nil {
		t.logger.Error("invalid permission format", "resource", resource, "action", action, "error", err)
		return status.Error(codes.InvalidArgument, "invalid permission format")
	}

	// Build RBAC verification request
	req := &proto_auth.VerifyUserResourceRequest{
		Identifier: &proto_infra.UserIdentifier{
			TenantId: tenantID,
			UserId:   userID,
		},
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resources: []*proto_auth.VerifyResource{
			{
				Resource: &proto_auth.VerifyResource_Permission{
					Permission: &proto_auth.Permission{
						Identifier: &proto_auth.Permission_Permission{
							Permission: &proto_auth.PermissionIdentifier{
								Resource: resource,
								Action:   action,
							},
						},
					},
				},
			},
		},
	}

	// Call RBAC service
	resp, err := t.rbacGRPCClient.VerifyUserResource(ctx, req)
	if err != nil {
		t.logger.Error("RBAC verification failed", "error", err, "user_id", userID, "permission", permString)
		return status.Error(codes.Internal, "failed to verify permissions")
	}

	// Check result
	if len(resp.Resources) == 0 || !resp.Resources[0].GetPermission().GetHasPermission().GetValue() {
		t.logger.Warn("permission denied", "user_id", userID, "tenant_id", tenantID, "permission", permString)
		return status.Error(codes.PermissionDenied, fmt.Sprintf("missing required permission: %s", permString))
	}

	t.logger.Debug("permission check passed", "user_id", userID, "permission", permString)
	return nil
}

func (t *TenantService) CreateTenant(ctx context.Context, req *proto_auth.CreateTenantRequest) (*proto_auth.CreateTenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		t.logger.Error("invalid identifier", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionCreate); err != nil {
		return nil, err
	}

	// Step 3: Validate tenant data
	tenantData := req.GetTenant()
	if tenantData == nil {
		t.logger.Error("tenant data is required")
		return nil, status.Error(codes.InvalidArgument, "tenant data is required")
	}

	t.logger.Info("creating tenant", "name", tenantData.Name, "requested_by", identifier.UserId)

	// Step 4: Convert proto → model
	tenant, err := convertor.CreateTenantFromProto(tenantData)
	if err != nil {
		t.logger.Error("failed to convert proto to tenant model", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Step 5: Create tenant in MongoDB
	tenantID, err := t.tenantCollection.CreateTenant(tenant)
	if err != nil {
		t.logger.Error("failed to create tenant", "error", err)
		return nil, status.Error(codes.Internal, "failed to create tenant")
	}
	t.logger.Info("tenant created in database", "tenant_id", tenantID)

	// Step 6: Seed defaults (permission, role, admin user)
	defaults, err := t.tenantSeeder.SeedDefaults(ctx, tenantID, tenantData.AdminEmail, tenantData.AdminPassword, identifier.UserId)
	if err != nil {
		t.logger.Error("failed to seed tenant defaults", "tenant_id", tenantID, "error", err)

		// Rollback: Delete tenant
		if deleteErr := t.tenantCollection.DeleteTenant(tenantID); deleteErr != nil {
			t.logger.Error("failed to rollback tenant creation", "tenant_id", tenantID, "error", deleteErr)
		}

		return nil, status.Error(codes.Internal, "failed to seed tenant defaults")
	}
	t.logger.Info("tenant defaults seeded", "tenant_id", tenantID, "permission_id", defaults.PermissionID, "role_id", defaults.RoleID, "user_id", defaults.UserID)

	return &proto_auth.CreateTenantResponse{TenantId: tenantID}, nil
}

func (t *TenantService) GetTenant(ctx context.Context, req *proto_auth.GetTenantRequest) (*proto_auth.TenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		t.logger.Error("invalid identifier", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionRead); err != nil {
		return nil, err
	}

	// Step 3: Extract tenant identifier from oneof
	var tenant *model_auth.Tenant
	var err error

	switch v := req.Tenant.(type) {
	case *proto_auth.GetTenantRequest_TenantId:
		t.logger.Debug("getting tenant by id", "tenant_id", v.TenantId)
		tenant, err = t.tenantCollection.GetTenantByID(v.TenantId)
	case *proto_auth.GetTenantRequest_Name:
		t.logger.Debug("getting tenant by name", "name", v.Name)
		tenant, err = t.tenantCollection.GetTenantByName(v.Name)
	default:
		t.logger.Error("tenant identifier not provided")
		return nil, status.Error(codes.InvalidArgument, "tenant_id or name is required")
	}

	// Step 4: Handle errors
	if err != nil {
		t.logger.Error("failed to get tenant", "error", err)
		return nil, status.Error(codes.NotFound, "tenant not found")
	}

	// Step 5: Convert model → proto
	tenantProto := convertor.TenantToProto(tenant)
	if tenantProto == nil {
		t.logger.Error("failed to convert tenant to proto")
		return nil, status.Error(codes.Internal, "failed to convert tenant")
	}

	t.logger.Info("tenant retrieved", "tenant_id", tenant.ID.Hex())
	return &proto_auth.TenantResponse{Tenant: tenantProto}, nil
}

func (t *TenantService) GetTenants(ctx context.Context, req *proto_auth.GetTenantsRequest) (*proto_auth.TenantsResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		t.logger.Error("invalid identifier", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionRead); err != nil {
		return nil, err
	}

	// Step 3: Get tenants with optional status filter
	var tenants []*model_auth.Tenant
	var err error

	if req.Status != nil && *req.Status != "" {
		t.logger.Debug("getting tenants by status", "status", *req.Status)
		tenants, err = t.tenantCollection.GetTenantsByStatus(*req.Status)
	} else {
		t.logger.Debug("getting all tenants")
		tenants, err = t.tenantCollection.GetTenants()
	}

	if err != nil {
		t.logger.Error("failed to get tenants", "error", err)
		return nil, status.Error(codes.Internal, "failed to get tenants")
	}

	// Step 4: Convert models → proto
	tenantsProto := convertor.TenantsToProto(tenants)

	// Step 5: Build pagination response (TODO: implement actual pagination)
	// For now, return all results without pagination
	paginationResp := &proto_infra.PaginationResponse{
		Page:       1,
		PageSize:   int32(len(tenantsProto)),
		TotalPages: 1,
		TotalItems: int64(len(tenantsProto)),
		HasNext:    false,
		HasPrev:    false,
	}

	t.logger.Info("tenants retrieved", "count", len(tenantsProto))
	return &proto_auth.TenantsResponse{
		Tenants:    tenantsProto,
		Pagination: paginationResp,
	}, nil
}

func (t *TenantService) UpdateTenant(ctx context.Context, req *proto_auth.UpdateTenantRequest) (*proto_auth.UpdateTenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		t.logger.Error("invalid identifier", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionUpdate); err != nil {
		return nil, err
	}

	// Step 3: Validate tenant data
	tenantData := req.GetTenant()
	if tenantData == nil {
		t.logger.Error("tenant data is required")
		return nil, status.Error(codes.InvalidArgument, "tenant data is required")
	}

	if tenantData.Id == "" {
		t.logger.Error("tenant id is required")
		return nil, status.Error(codes.InvalidArgument, "tenant id is required")
	}

	t.logger.Info("updating tenant", "tenant_id", tenantData.Id, "requested_by", identifier.UserId)

	// Step 4: Get existing tenant
	existingTenant, err := t.tenantCollection.GetTenantByID(tenantData.Id)
	if err != nil {
		t.logger.Error("failed to get existing tenant", "tenant_id", tenantData.Id, "error", err)
		return nil, status.Error(codes.NotFound, "tenant not found")
	}

	// Step 5: Apply updates from proto
	if err := convertor.UpdateTenantFromProto(existingTenant, tenantData); err != nil {
		t.logger.Error("failed to apply updates", "tenant_id", tenantData.Id, "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Step 6: Update in MongoDB
	if err := t.tenantCollection.UpdateTenant(existingTenant); err != nil {
		t.logger.Error("failed to update tenant", "tenant_id", tenantData.Id, "error", err)
		return nil, status.Error(codes.Internal, "failed to update tenant")
	}

	t.logger.Info("tenant updated successfully", "tenant_id", tenantData.Id)
	return &proto_auth.UpdateTenantResponse{Updated: true}, nil
}

func (t *TenantService) DeleteTenant(ctx context.Context, req *proto_auth.DeleteTenantRequest) (*proto_auth.DeleteTenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		t.logger.Error("invalid identifier", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionDelete); err != nil {
		return nil, err
	}

	// Step 3: Validate tenant_id
	tenantID := req.TenantId
	if tenantID == "" {
		t.logger.Error("tenant_id is required")
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	t.logger.Info("starting tenant deletion cascade", "tenant_id", tenantID, "requested_by", identifier.UserId)

	// Step 4: Verify tenant exists
	_, err := t.tenantCollection.GetTenantByID(tenantID)
	if err != nil {
		t.logger.Error("tenant not found", "tenant_id", tenantID, "error", err)
		return nil, status.Error(codes.NotFound, "tenant not found")
	}

	// STEP 5: Revoke all tokens for ALL users in this tenant (bulk operation)
	// This will revoke all access tokens and refresh tokens for every user in the tenant
	t.logger.Info("revoking all tokens for tenant", "tenant_id", tenantID)
	revokeReq := &proto_auth.RevokeAllTenantTokensRequest{
		TenantId:  tenantID,
		RevokedBy: identifier.UserId,
	}
	if _, err := t.authGRPCClient.RevokeAllTenantTokens(ctx, revokeReq); err != nil {
		t.logger.Error("failed to revoke tokens for tenant", "tenant_id", tenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("revoked all tokens for tenant", "tenant_id", tenantID)
	}

	// STEP 6: Delete ALL users for this tenant (bulk operation)
	// TODO: Uncomment once UserCollection supports bulk delete by tenant_id
	// This deletes all user documents with matching tenant_id in one operation
	// t.logger.Info("deleting all users for tenant", "tenant_id", tenantID)
	// if err := t.userCollection.DeleteAllUsersByTenantID(tenantID); err != nil {
	//     t.logger.Error("failed to delete users for tenant", "tenant_id", tenantID, "error", err)
	//     return nil, status.Error(codes.Internal, "failed to delete users")
	// } else {
	//     t.logger.Info("deleted all users for tenant", "tenant_id", tenantID)
	// }

	// STEP 7: Delete ALL roles for this tenant (bulk operation)
	// TODO: Uncomment once RBAC service supports bulk delete by tenant_id filter
	// This deletes all role documents with matching tenant_id in one operation
	// t.logger.Info("deleting all roles for tenant", "tenant_id", tenantID)
	// deleteRolesReq := &proto_auth.DeleteResourcesByFilterRequest{
	//     TenantId:     tenantID,
	//     ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
	// }
	// if _, err := t.rbacGRPCClient.DeleteResourcesByFilter(ctx, deleteRolesReq); err != nil {
	//     t.logger.Error("failed to delete roles for tenant", "tenant_id", tenantID, "error", err)
	//     // Continue with deletion even if this fails
	// } else {
	//     t.logger.Info("deleted all roles for tenant", "tenant_id", tenantID)
	// }

	// STEP 8: Delete ALL permissions for this tenant (bulk operation)
	// TODO: Uncomment once RBAC service supports bulk delete by tenant_id filter
	// This deletes all permission documents with matching tenant_id in one operation
	// t.logger.Info("deleting all permissions for tenant", "tenant_id", tenantID)
	// deletePermsReq := &proto_auth.DeleteResourcesByFilterRequest{
	//     TenantId:     tenantID,
	//     ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
	// }
	// if _, err := t.rbacGRPCClient.DeleteResourcesByFilter(ctx, deletePermsReq); err != nil {
	//     t.logger.Error("failed to delete permissions for tenant", "tenant_id", tenantID, "error", err)
	//     // Continue with deletion even if this fails
	// } else {
	//     t.logger.Info("deleted all permissions for tenant", "tenant_id", tenantID)
	// }

	// STEP 9: Delete the tenant itself
	t.logger.Info("deleting tenant", "tenant_id", tenantID)
	if err := t.tenantCollection.DeleteTenant(tenantID); err != nil {
		t.logger.Error("failed to delete tenant", "tenant_id", tenantID, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete tenant")
	}

	t.logger.Info("tenant deleted successfully", "tenant_id", tenantID)
	return &proto_auth.DeleteTenantResponse{Deleted: true}, nil
}
