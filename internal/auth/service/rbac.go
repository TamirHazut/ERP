package service

import (
	"context"
	"errors"
	"strings"

	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/infra/convertor"
	erp_errors "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging"
	auth_models "erp.localhost/internal/infra/model/auth"
	shared_models "erp.localhost/internal/infra/model/shared"
	auth_proto "erp.localhost/internal/infra/proto/auth/v1"
	infra_proto "erp.localhost/internal/infra/proto/infra/v1"
	"erp.localhost/internal/infra/proto/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type RBACService struct {
	logger      *logging.Logger
	rbacManager *rbac.RBACManager
	auth_proto.UnimplementedRBACServiceServer
}

// TODO: add logs
func NewRBACService() *RBACService {
	logger := logging.NewLogger(shared_models.ModuleAuth)

	rbacManager := rbac.NewRBACManager()
	if rbacManager == nil {
		logger.Fatal("failed to create rbac manager")
		return nil
	}

	return &RBACService{
		logger:      logger,
		rbacManager: rbacManager,
	}
}

func (r *RBACService) CreateResource(ctx context.Context, req *auth_proto.CreateResourceRequest) (*auth_proto.CreateResourceResponse, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to validate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	// Get handler for resource type
	handler, err := getResourceHandler(req.GetResourceType())
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Invalid resource type", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	resourceType := handler.GetResourceType()
	resourceName := handler.GetResourceName()

	// Extract and convert resource from proto to domain model
	resource, err := handler.ExtractAndConvertCreate(req)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to extract/convert resource", "resourceType", resourceName, "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Create resource
	r.logger.Debug("Creating new resource", "resourceType", resourceName)
	resourceID, err := r.rbacManager.CreateResource(tenantID, userID, resourceType, resource)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to create resource", "resourceType", resourceName, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	r.logger.Debug("Resource created successfully", "resourceType", resourceName, "resourceID", resourceID)
	return &auth_proto.CreateResourceResponse{
		ResourceId: resourceID,
	}, nil
}

func (r *RBACService) UpdateResource(ctx context.Context, req *auth_proto.UpdateResourceRequest) (*infra_proto.Response, error) {
	// Validate input
	identifier := req.GetIdentifier()
	if err := validator.ValidateUserIdentifier(identifier); err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to validate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	// Get handler for resource type
	handler, err := getResourceHandler(req.GetResourceType())
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Invalid resource type", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	resourceType := handler.GetResourceType()
	resourceName := handler.GetResourceName()

	// Extract update data from proto
	updateData, err := handler.ExtractUpdateData(req)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to extract update data", "resourceType", resourceName, "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	// Get resource ID from update data
	resourceID := handler.GetResourceIDFromUpdate(updateData)

	// Get existing resource
	r.logger.Debug("Updating resource", "resourceType", resourceName, "resourceID", resourceID)
	existing, err := r.rbacManager.GetResource(tenantID, userID, resourceType, resourceID)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to get resource", "resourceType", resourceName, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	// Apply updates to existing resource
	if err := handler.ApplyUpdate(existing, updateData); err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to apply updates", "resourceType", resourceName, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	// Update resource in storage
	if err := r.rbacManager.UpdateResource(tenantID, userID, resourceType, existing); err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to update resource", "resourceType", resourceName, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	r.logger.Debug("Resource updated successfully", "resourceType", resourceName, "resourceID", resourceID)
	return &infra_proto.Response{
		Success: true,
	}, nil
}

func (r *RBACService) GetResource(ctx context.Context, req *auth_proto.GetResourceRequest) (*auth_proto.ResourceResponse, error) {
	var err error
	var resourceType string
	// Validate input
	identifier := req.GetIdentifier()
	err = validator.ValidateUserIdentifier(identifier)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to validate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()
	resourceID := req.GetId()

	var resource *auth_proto.Resource
	switch req.GetResourceType() {
	case auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = auth_models.ResourceTypePermission
	case auth_proto.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = auth_models.ResourceTypeRole
	default:
		return nil, status.Error(codes.Internal, erp_errors.Validation(erp_errors.ValidationInvalidValue, "resourceType").Error())
	}

	r.logger.Debug("Getting resource", "resourceType", resourceType, "resourceID", resourceID)
	res, err := r.rbacManager.GetResource(tenantID, userID, resourceType, resourceID)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to get resource", "resourceType", resourceType, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	resource, err = r.convertGetResultToResource(resourceType, res)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to convert resource", "resourceType", resourceType, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	return &auth_proto.ResourceResponse{
		Resource: resource,
	}, nil
}

func (r *RBACService) ListResources(ctx context.Context, req *auth_proto.ListResourcesRequest) (*auth_proto.ListResourcesResponse, error) {
	var err error
	var resourceType string
	// Validate input
	identifier := req.GetIdentifier()
	err = validator.ValidateUserIdentifier(identifier)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to validate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	var getResults []any
	switch req.GetResourceType() {
	case auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = auth_models.ResourceTypePermission
	case auth_proto.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = auth_models.ResourceTypeRole
	default:
		return nil, status.Error(codes.Internal, erp_errors.Validation(erp_errors.ValidationInvalidValue, "resourceType").Error())
	}

	r.logger.Debug("Getting resource", "resourceType", resourceType)
	getResults, err = r.rbacManager.GetResources(tenantID, userID, resourceType)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to get resources", "resourceType", resourceType, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	resources := make([]*auth_proto.Resource, 0)
	for _, res := range getResults {
		resource, err := r.convertGetResultToResource(resourceType, res)
		if err != nil {
			errMsg := err.Error()
			r.logger.Error("Failed to convert resources", "resourceType", resourceType, "error", errMsg)
			return nil, status.Error(codes.Internal, errMsg)
		}
		resources = append(resources, resource)
	}

	return &auth_proto.ListResourcesResponse{
		Resources: resources,
	}, nil
}

func (r *RBACService) DeleteResource(ctx context.Context, req *auth_proto.DeleteResourceRequest) (*infra_proto.Response, error) {
	var err error
	var resourceType string
	var resourceID string
	// Validate input
	identifier := req.GetIdentifier()
	err = validator.ValidateUserIdentifier(identifier)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to validate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	switch req.GetResourceType() {
	case auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = auth_models.ResourceTypePermission
	case auth_proto.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = auth_models.ResourceTypeRole
	default:
		err = erp_errors.Validation(erp_errors.ValidationInvalidValue, "resourceType")
	}

	if err == nil {
		r.logger.Debug("Deleting resource", "resourceType", resourceType, "resourceID", resourceID)
		err = r.rbacManager.DeleteResource(tenantID, userID, resourceType, resourceID)
	}

	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to delete resource", "resourceType", resourceType, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	r.logger.Debug("Resource deleted successfuly", "resourceType", resourceType, "resourceID", resourceID)
	return &infra_proto.Response{
		Success: true,
	}, nil
}

func (r *RBACService) VerifyUserResource(ctx context.Context, req *auth_proto.VerifyUserResourceRequest) (*auth_proto.VerifyResourceResponse, error) {
	var err error
	var resourceType string
	// Validate input
	identifier := req.GetIdentifier()
	err = validator.ValidateUserIdentifier(identifier)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to validate user", "error", errMsg)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	tenantID := identifier.GetTenantId()
	userID := identifier.GetUserId()

	resoucesAsStrings, err := r.getVerifyResources(req.GetResources(), req.GetResourceType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	verifyResources := make([]*auth_proto.VerifyResource, 0)
	switch req.GetResourceType() {
	case auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = auth_models.ResourceTypePermission
		// proccess request
		permissionsCheckResponse, opErr := r.rbacManager.CheckUserPermissions(tenantID, userID, resoucesAsStrings)
		if opErr != nil {
			err = opErr
			break
		}
		for permission, hasPermission := range permissionsCheckResponse {
			permissionSplt := strings.Split(permission, ":")
			if len(permissionSplt) != 2 || permissionSplt[0] == "" || permissionSplt[1] == "" {
				err = status.Error(codes.Internal, erp_errors.Internal(erp_errors.InternalUnexpectedError, errors.New("failed to check permissions")).Error())
				break
			}
			verifyPermission := r.createPermissionVerifyResource(permissionSplt[0], permissionSplt[1], hasPermission)
			verifyResources = append(verifyResources, verifyPermission)
		}
	case auth_proto.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = auth_models.ResourceTypeRole
		rolesCheckResponse, opErr := r.rbacManager.CheckUserRoles(tenantID, userID, resoucesAsStrings)
		if opErr != nil {
			err = opErr
			break
		}
		for role, hasRole := range rolesCheckResponse {
			if hasRole {
				verifyRole := r.createRolesVerifyResource(role)
				verifyResources = append(verifyResources, verifyRole)
			}
		}
	default:
		err = status.Error(codes.InvalidArgument, erp_errors.Validation(erp_errors.ValidationInvalidValue, "resourceType").Error())
	}

	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to verify resource", "resourceType", resourceType, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	r.logger.Debug("Verification completed", "resourceType", resourceType)
	return &auth_proto.VerifyResourceResponse{
		Resources: verifyResources,
	}, nil
}

func (r *RBACService) convertGetResultToResource(resourceType string, res any) (*auth_proto.Resource, error) {
	var resource *auth_proto.Resource
	switch resourceType {
	case auth_models.ResourceTypePermission:
		permission, ok := res.(*auth_models.Permission)
		if !ok {
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidType)
		}
		permissionProto := convertor.PermissionToProto(permission)
		if permissionProto == nil {
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue)
		}
		resource = &auth_proto.Resource{
			Resource: &auth_proto.Resource_Permission{
				Permission: permissionProto,
			},
		}
	case auth_models.ResourceTypeRole:
		role, ok := res.(*auth_models.Role)
		if !ok {
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidType)
		}
		roleProto := convertor.RoleToProto(role)
		if roleProto == nil {
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue)
		}
		resource = &auth_proto.Resource{
			Resource: &auth_proto.Resource_Role{
				Role: roleProto,
			},
		}
	default:
		return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "resourceType")
	}
	return resource, nil
}

func (r *RBACService) getVerifyResources(resources []*auth_proto.VerifyResource, resourceType auth_proto.ResourceType) ([]string, error) {
	if len(resources) == 0 {
		return nil, erp_errors.Validation(erp_errors.ValidationRequiredFields, "resources")
	}
	res := make([]string, len(resources))
	for i := range resources {
		resource := resources[i].GetResource()
		if resource == nil {
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "resource")
		}
		switch resourceType {
		case auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION:
			resourcePermission := resource.(*auth_proto.VerifyResource_Permission)
			permission, err := auth_models.CreatePermissionString(resourcePermission.Permission.GetPermission().GetResource(), resourcePermission.Permission.GetPermission().GetAction())
			if err != nil {
				return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "permission").WithError(err)
			}
			res[i] = permission
		case auth_proto.ResourceType_RESOURCE_TYPE_ROLE:
			resourceRole := resource.(*auth_proto.VerifyResource_Role)
			if resourceRole.Role.GetRoleName() == "" {
				return nil, erp_errors.Validation(erp_errors.ValidationInvalidValue, "role")
			}
			res[i] = resourceRole.Role.GetRoleName()
		default:
			return nil, erp_errors.Validation(erp_errors.ValidationInvalidType)
		}
	}
	return res, nil
}

/* Permissions */
func (r *RBACService) VerifyUserPermissions(ctx context.Context, req *auth_proto.VerifyUserResourceRequest) (*auth_proto.VerifyResourceResponse, error) {
	req.ResourceType = auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION
	return r.VerifyUserResource(ctx, req)
}

func (r *RBACService) CreateVerifyPermissionsResourceRequest(identifier *infra_proto.UserIdentifier, permissionIdentifiers ...string) *auth_proto.VerifyUserResourceRequest {
	if len(permissionIdentifiers) == 0 || len(permissionIdentifiers)%2 != 0 { // Permission identifiers come in pairs: [resource, action, resource, action, ...]
		return nil
	}
	resources := make([]*auth_proto.VerifyResource, 0)
	for i := 0; i < len(permissionIdentifiers); i += 2 {
		resources = append(resources, r.createPermissionVerifyResource(permissionIdentifiers[i], permissionIdentifiers[i+1]))
	}
	return &auth_proto.VerifyUserResourceRequest{
		Identifier:   identifier,
		ResourceType: auth_proto.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resources:    resources,
	}
}

func (r *RBACService) createPermissionVerifyResource(resource string, action string, hasPermission ...bool) *auth_proto.VerifyResource {
	var boolWrapper *wrapperspb.BoolValue
	if len(hasPermission) > 0 {
		boolWrapper = wrapperspb.Bool(hasPermission[0])
	}
	return &auth_proto.VerifyResource{
		Resource: &auth_proto.VerifyResource_Permission{
			Permission: &auth_proto.Permission{
				Identifier: &auth_proto.Permission_Permission{
					Permission: &auth_proto.PermissionIdentifier{
						Resource: resource,
						Action:   action,
					},
				},
				HasPermission: boolWrapper,
			},
		},
	}
}

/* Roles */
func (r *RBACService) VerifyUserRoles(ctx context.Context, req *auth_proto.VerifyUserResourceRequest) (*auth_proto.VerifyResourceResponse, error) {
	req.ResourceType = auth_proto.ResourceType_RESOURCE_TYPE_ROLE
	return r.VerifyUserResource(ctx, req)
}

func (r *RBACService) CreateVerifyRolesResourceRequest(identifier *infra_proto.UserIdentifier, roles ...string) *auth_proto.VerifyUserResourceRequest {
	if len(roles) == 0 {
		return nil
	}
	resources := make([]*auth_proto.VerifyResource, 0)
	for _, role := range roles {
		resources = append(resources, r.createRolesVerifyResource(role))
	}
	return &auth_proto.VerifyUserResourceRequest{
		Identifier:   identifier,
		ResourceType: auth_proto.ResourceType_RESOURCE_TYPE_ROLE,
		Resources:    resources,
	}
}

func (r *RBACService) createRolesVerifyResource(role string) *auth_proto.VerifyResource {
	return &auth_proto.VerifyResource{
		Resource: &auth_proto.VerifyResource_Role{
			Role: &auth_proto.Role{
				Identifier: &auth_proto.Role_RoleName{
					RoleName: role,
				},
			},
		},
	}
}
