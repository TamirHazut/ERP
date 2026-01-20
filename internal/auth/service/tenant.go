package service

import (
	"context"

	"erp.localhost/internal/auth/api"
	collection "erp.localhost/internal/auth/collection"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	validator_infra "erp.localhost/internal/infra/model/infra/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TenantService struct {
	logger           logger.Logger
	tenantCollection *collection.TenantCollection
	tenantSeeder     *TenantSeeder
	authAPI          *api.AuthAPI
	rbacAPI          *api.RBACAPI
	userAPI          *api.UserAPI
	authv1.UnimplementedTenantServiceServer
}

func NewTenantService(authAPI *api.AuthAPI, rbacAPI *api.RBACAPI, userAPI *api.UserAPI, logger logger.Logger) *TenantService {

	tenantHandler := mongo_collection.NewBaseCollectionHandler[authv1.Tenant](model_mongo.AuthDB, model_mongo.TenantsCollection, logger)
	tenantCollection := collection.NewTenantCollection(tenantHandler, logger)
	if tenantCollection == nil {
		logger.Fatal("failed to create tenants collection")
		return nil
	}

	tenantSeeder := NewTenantSeeder(rbacAPI, logger)
	if tenantSeeder == nil {
		logger.Fatal("failed to create tenant seeder")
		return nil
	}

	return &TenantService{
		logger:           logger,
		tenantCollection: tenantCollection,
		tenantSeeder:     tenantSeeder,
		authAPI:          authAPI,
		rbacAPI:          rbacAPI,
		userAPI:          userAPI,
	}
}

// checkPermission verifies if a user has the required permission
func (t *TenantService) checkPermission(ctx context.Context, tenantID, userID, resource, action string) error {
	// Create permission string using helper
	permString, err := model_auth.CreatePermissionString(resource, action)
	if err != nil {
		t.logger.Error("invalid permission format", "resource", resource, "action", action, "error", err)
		return err
	}

	permissions := []string{permString}
	res, err := t.rbacAPI.Verification.CheckPermissions(tenantID, userID, permissions)
	if err != nil {
		return err
	}
	// Check result
	if !res[permString] {
		t.logger.Warn("permission denied", "user_id", userID, "tenant_id", tenantID, "permission", permString)
		return infra_error.Auth(infra_error.AuthPermissionDenied)
	}

	t.logger.Debug("permission check passed", "user_id", userID, "permission", permString)
	return nil
}

func (t *TenantService) CreateTenant(ctx context.Context, req *authv1.CreateTenantRequest) (*authv1.CreateTenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionCreate); err != nil {
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 3: Validate tenant data
	tenant := req.GetTenant()
	if tenant == nil {
		t.logger.Error("tenant data is required")
		return nil, status.Error(codes.InvalidArgument, "tenant data is required")
	}

	adminEmail := tenant.GetContact().GetEmail()
	if adminEmail == "" {
		t.logger.Error("email is required")
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	t.logger.Info("creating tenant", "name", tenant.Name, "requested_by", identifier.UserId)

	// Step 5: Create tenant in MongoDB
	tenantID, err := t.tenantCollection.CreateTenant(tenant)
	if err != nil {
		t.logger.Error("failed to create tenant", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}
	t.logger.Info("tenant created in database", "tenant_id", tenantID)

	// Step 6: Seed defaults (permission, role, admin user)
	defaults, err := t.tenantSeeder.SeedDefaults(ctx, tenantID, adminEmail, identifier.UserId)
	if err != nil {
		t.logger.Error("failed to seed tenant defaults", "tenant_id", tenantID, "error", err)

		// Rollback: Delete tenant
		if deleteErr := t.tenantCollection.DeleteTenant(tenantID); deleteErr != nil {
			t.logger.Error("failed to rollback tenant creation", "tenant_id", tenantID, "error", deleteErr)
		}

		return nil, status.Error(codes.Internal, "failed to seed tenant defaults")
	}
	t.logger.Info("tenant defaults seeded", "tenant_id", tenantID, "permission_id", defaults.PermissionID, "role_id", defaults.RoleId, "user_id", defaults.UserId)

	return &authv1.CreateTenantResponse{TenantId: tenantID}, nil
}

func (t *TenantService) GetTenant(ctx context.Context, req *authv1.GetTenantRequest) (*authv1.Tenant, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionRead); err != nil {
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 3: Extract tenant identifier from oneof
	var tenant *authv1.Tenant
	var err error

	switch v := req.Tenant.(type) {
	case *authv1.GetTenantRequest_TenantId:
		t.logger.Debug("getting tenant by id", "tenant_id", v.TenantId)
		tenant, err = t.tenantCollection.GetTenantByID(v.TenantId)
	case *authv1.GetTenantRequest_Name:
		t.logger.Debug("getting tenant by name", "name", v.Name)
		tenant, err = t.tenantCollection.GetTenantByName(v.Name)
	default:
		t.logger.Error("tenant identifier not provided")
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id or name is required")
	}

	// Step 4: Handle errors
	if err != nil {
		t.logger.Error("failed to get tenant", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	t.logger.Info("tenant retrieved", "tenant_id", tenant.Id)
	return tenant, nil
}

func (t *TenantService) ListTenants(ctx context.Context, req *authv1.ListTenantsRequest) (*authv1.ListTenantsResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionRead); err != nil {
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 3: Get tenants with optional status filter
	var tenants []*authv1.Tenant
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
		return nil, infra_error.ToGRPCError(err)
	}

	// // Step 5: Build pagination response (TODO: implement actual pagination)
	// // For now, return all results without pagination
	// paginationResp := &proto_infra.PaginationResponse{
	// 	Page:       1,
	// 	PageSize:   int32(len(tenants)),
	// 	TotalPages: 1,
	// 	TotalItems: int64(len(tenants)),
	// 	HasNext:    false,
	// 	HasPrev:    false,
	// }

	t.logger.Info("tenants retrieved", "count", len(tenants))
	return &authv1.ListTenantsResponse{
		Tenants: tenants,
	}, nil
}

func (t *TenantService) UpdateTenant(ctx context.Context, req *authv1.UpdateTenantRequest) (*authv1.UpdateTenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 2: Check RBAC permission
	if err := t.checkPermission(ctx, identifier.TenantId, identifier.UserId, model_auth.ResourceTypeTenant, model_auth.PermissionActionUpdate); err != nil {
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 3: Validate tenant data
	tenant := req.GetTenant()
	if tenant == nil {
		t.logger.Error("tenant data is required")
		return nil, status.Error(codes.InvalidArgument, "tenant data is required")
	}

	if tenant.Id == "" {
		t.logger.Error("tenant id is required")
		return nil, status.Error(codes.InvalidArgument, "tenant id is required")
	}

	t.logger.Info("updating tenant", "tenant_id", tenant.Id, "requested_by", identifier.UserId)

	// Step 4: Get existing tenant
	existingTenant, err := t.tenantCollection.GetTenantByID(tenant.Id)
	if err != nil || existingTenant == nil {
		t.logger.Error("failed to get existing tenant", "tenant_id", tenant.Id, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 6: Update in MongoDB
	if err := t.tenantCollection.UpdateTenant(tenant); err != nil {
		t.logger.Error("failed to update tenant", "tenant_id", tenant.Id, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	t.logger.Info("tenant updated successfully", "tenant_id", tenant.Id)
	return &authv1.UpdateTenantResponse{Updated: true}, nil
}

func (t *TenantService) DeleteTenant(ctx context.Context, req *authv1.DeleteTenantRequest) (*authv1.DeleteTenantResponse, error) {
	// Step 1: Validate identifier
	identifier := req.GetIdentifier()
	if err := validator_infra.ValidateUserIdentifier(identifier); err != nil {
		t.logger.Error("invalid identifier", "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 2: Validate tenant_id
	tenantID := req.GetIdentifier().GetTenantId()
	userID := req.GetIdentifier().GetUserId()
	targetTenantID := req.GetTenantId()

	// Step 3: Verify tenant exists
	_, err := t.tenantCollection.GetTenantByID(targetTenantID)
	if err != nil {
		t.logger.Error("tenant not found", "target_tenant_id", targetTenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	}

	// Step 4: Revoke all tenant users tokens
	t.logger.Info("starting tenant deletion cascade", "tenant_id", tenantID, "requested_by", userID, "target_tenant_id", targetTenantID)
	if _, _, err := t.authAPI.RevokeAllTenantTokens(tenantID, identifier.GetUserId(), targetTenantID); err != nil {
		t.logger.Error("failed to revoke tokens for tenant", "tenant_id", tenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("revoked all tokens for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 5: Delete ALL users for this tenant (bulk operation)
	// This deletes all user documents with matching tenant_id in one operation
	t.logger.Info("deleting all users for tenant", "target_tenant_id", targetTenantID)
	if err := t.userAPI.DeleteTenantUsers(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete roles for tenant", "target_tenant_id", targetTenantID, "error", err)
		return nil, infra_error.ToGRPCError(err)
	} else {
		t.logger.Info("deleted all roles for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 6: Delete ALL roles for this tenant (bulk operation)
	// This deletes all role documents with matching tenant_id in one operation
	t.logger.Info("deleting all roles for tenant", "target_tenant_id", targetTenantID)
	if err := t.rbacAPI.Roles.DeleteTenantRoles(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete roles for tenant", "target_tenant_id", targetTenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("deleted all roles for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 7: Delete ALL permissions for this tenant (bulk operation)
	// This deletes all permission documents with matching tenant_id in one operation
	t.logger.Info("deleting all permissions for tenant", "target_tenant_id", targetTenantID)
	if err := t.rbacAPI.Permissions.DeleteTenantPermissions(tenantID, userID, targetTenantID); err != nil {
		t.logger.Error("failed to delete permissions for tenant", "target_tenant_id", targetTenantID, "error", err)
		// Continue with deletion even if this fails
	} else {
		t.logger.Info("deleted all permissions for tenant", "target_tenant_id", targetTenantID)
	}

	// STEP 8: Delete the tenant itself
	t.logger.Info("deleting tenant", "target_tenant_id", targetTenantID)
	if err := t.tenantCollection.DeleteTenant(targetTenantID); err != nil {
		t.logger.Error("failed to delete tenant", "target_tenant_id", targetTenantID, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete tenant")
	}

	t.logger.Info("tenant deleted successfully", "target_tenant_id", targetTenantID)
	return &authv1.DeleteTenantResponse{Deleted: true}, nil
}
