package service

import (
	"context"
	"errors"
	"strings"

	collection_auth "erp.localhost/internal/auth/collection"
	"erp.localhost/internal/auth/rbac"
	"erp.localhost/internal/infra/convertor"
	mongo_collection "erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_auth "erp.localhost/internal/infra/model/auth"
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	model_shared "erp.localhost/internal/infra/model/shared"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/infra/v1"
	"erp.localhost/internal/infra/proto/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type RBACService struct {
	logger      logger.Logger
	rbacManager *rbac.RBACManager
	proto_auth.UnimplementedRBACServiceServer
}

// TODO: add logs
// TODO: remove user retreive function calls and accept them as function aguments - This service only deals with roles and permissions not with users!
func NewRBACService() *RBACService {
	logger := logger.NewBaseLogger(model_shared.ModuleAuth)
	pc := mongo_collection.NewBaseCollectionHandler[model_auth.Permission](model_mongo.AuthDB, model_mongo.RolesCollection, logger)
	permissionsCollection := collection_auth.NewPermissionCollection(pc, logger)

	rc := mongo_collection.NewBaseCollectionHandler[model_auth.Role](model_mongo.AuthDB, model_mongo.RolesCollection, logger)
	rolesCollection := collection_auth.NewRoleCollection(rc, logger)

	rbacManager := rbac.NewRBACManager(permissionsCollection, rolesCollection)
	if rbacManager == nil {
		logger.Fatal("failed to create rbac manager")
		return nil
	}

	return &RBACService{
		logger:      logger,
		rbacManager: rbacManager,
	}
}

func (r *RBACService) CreateResource(ctx context.Context, req *proto_auth.CreateResourceRequest) (*proto_auth.CreateResourceResponse, error) {
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
	return &proto_auth.CreateResourceResponse{
		ResourceId: resourceID,
	}, nil
}

func (r *RBACService) UpdateResource(ctx context.Context, req *proto_auth.UpdateResourceRequest) (*proto_infra.Response, error) {
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
	return &proto_infra.Response{
		Success: true,
	}, nil
}

func (r *RBACService) GetResource(ctx context.Context, req *proto_auth.GetResourceRequest) (*proto_auth.ResourceResponse, error) {
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

	var resource *proto_auth.Resource
	switch req.GetResourceType() {
	case proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = model_auth.ResourceTypePermission
	case proto_auth.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = model_auth.ResourceTypeRole
	default:
		return nil, status.Error(codes.Internal, infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType").Error())
	}

	r.logger.Debug("Getting resource", "resourceType", resourceType, "resourceID", resourceID)
	res, err := r.rbacManager.GetResource(tenantID, userID, resourceType, resourceID)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to get resource", "resourceType", resourceType, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	resource, err = r.convertGetResultToResource(res)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to convert resource", "resourceType", resourceType, "resourceID", resourceID, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	return &proto_auth.ResourceResponse{
		Resource: resource,
	}, nil
}

func (r *RBACService) ListResources(ctx context.Context, req *proto_auth.ListResourcesRequest) (*proto_auth.ListResourcesResponse, error) {
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

	var getResults []model_auth.RBACResource
	switch req.GetResourceType() {
	case proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = model_auth.ResourceTypePermission
	case proto_auth.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = model_auth.ResourceTypeRole
	default:
		return nil, status.Error(codes.Internal, infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType").Error())
	}

	r.logger.Debug("Getting resource", "resourceType", resourceType)
	getResults, err = r.rbacManager.GetResources(tenantID, userID, resourceType)
	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to get resources", "resourceType", resourceType, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	resources := make([]*proto_auth.Resource, 0)
	for _, res := range getResults {
		resource, err := r.convertGetResultToResource(res)
		if err != nil {
			errMsg := err.Error()
			r.logger.Error("Failed to convert resources", "resourceType", resourceType, "error", errMsg)
			return nil, status.Error(codes.Internal, errMsg)
		}
		resources = append(resources, resource)
	}

	return &proto_auth.ListResourcesResponse{
		Resources: resources,
	}, nil
}

func (r *RBACService) DeleteResource(ctx context.Context, req *proto_auth.DeleteResourceRequest) (*proto_infra.Response, error) {
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
	case proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = model_auth.ResourceTypePermission
	case proto_auth.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = model_auth.ResourceTypeRole
	default:
		err = infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
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
	return &proto_infra.Response{
		Success: true,
	}, nil
}

func (r *RBACService) VerifyUserResource(ctx context.Context, req *proto_auth.VerifyUserResourceRequest) (*proto_auth.VerifyResourceResponse, error) {
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

	verifyResources := make([]*proto_auth.VerifyResource, 0)
	switch req.GetResourceType() {
	case proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION:
		resourceType = model_auth.ResourceTypePermission
		// proccess request
		permissionsCheckResponse, opErr := r.rbacManager.CheckUserPermissions(tenantID, userID, resoucesAsStrings)
		if opErr != nil {
			err = opErr
			break
		}
		for permission, hasPermission := range permissionsCheckResponse {
			permissionSplt := strings.Split(permission, ":")
			if len(permissionSplt) != 2 || permissionSplt[0] == "" || permissionSplt[1] == "" {
				err = status.Error(codes.Internal, infra_error.Internal(infra_error.InternalUnexpectedError, errors.New("failed to check permissions")).Error())
				break
			}
			verifyPermission := r.createPermissionVerifyResource(permissionSplt[0], permissionSplt[1], hasPermission)
			verifyResources = append(verifyResources, verifyPermission)
		}
	case proto_auth.ResourceType_RESOURCE_TYPE_ROLE:
		resourceType = model_auth.ResourceTypeRole
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
		err = status.Error(codes.InvalidArgument, infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType").Error())
	}

	if err != nil {
		errMsg := err.Error()
		r.logger.Error("Failed to verify resource", "resourceType", resourceType, "error", errMsg)
		return nil, status.Error(codes.Internal, errMsg)
	}

	r.logger.Debug("Verification completed", "resourceType", resourceType)
	return &proto_auth.VerifyResourceResponse{
		Resources: verifyResources,
	}, nil
}

func (r *RBACService) convertGetResultToResource(res model_auth.RBACResource) (*proto_auth.Resource, error) {
	var resource *proto_auth.Resource
	switch res.GetResourceType() {
	case model_auth.ResourceTypePermission:
		permission, ok := res.(*model_auth.Permission)
		if !ok {
			return nil, infra_error.Validation(infra_error.ValidationInvalidType)
		}
		permissionProto := convertor.PermissionToProto(permission)
		if permissionProto == nil {
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue)
		}
		resource = &proto_auth.Resource{
			Resource: &proto_auth.Resource_Permission{
				Permission: permissionProto,
			},
		}
	case model_auth.ResourceTypeRole:
		role, ok := res.(*model_auth.Role)
		if !ok {
			return nil, infra_error.Validation(infra_error.ValidationInvalidType)
		}
		roleProto := convertor.RoleToProto(role)
		if roleProto == nil {
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue)
		}
		resource = &proto_auth.Resource{
			Resource: &proto_auth.Resource_Role{
				Role: roleProto,
			},
		}
	default:
		return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "resourceType")
	}
	return resource, nil
}

func (r *RBACService) getVerifyResources(resources []*proto_auth.VerifyResource, resourceType proto_auth.ResourceType) ([]string, error) {
	if len(resources) == 0 {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "resources")
	}
	res := make([]string, len(resources))
	for i := range resources {
		resource := resources[i].GetResource()
		if resource == nil {
			return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "resource")
		}
		switch resourceType {
		case proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION:
			resourcePermission := resource.(*proto_auth.VerifyResource_Permission)
			permission, err := model_auth.CreatePermissionString(resourcePermission.Permission.GetPermission().GetResource(), resourcePermission.Permission.GetPermission().GetAction())
			if err != nil {
				return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "permission").WithError(err)
			}
			res[i] = permission
		case proto_auth.ResourceType_RESOURCE_TYPE_ROLE:
			resourceRole := resource.(*proto_auth.VerifyResource_Role)
			if resourceRole.Role.GetRoleName() == "" {
				return nil, infra_error.Validation(infra_error.ValidationInvalidValue, "role")
			}
			res[i] = resourceRole.Role.GetRoleName()
		default:
			return nil, infra_error.Validation(infra_error.ValidationInvalidType)
		}
	}
	return res, nil
}

/* Permissions */
func (r *RBACService) VerifyUserPermissions(ctx context.Context, req *proto_auth.VerifyUserResourceRequest) (*proto_auth.VerifyResourceResponse, error) {
	req.ResourceType = proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION
	return r.VerifyUserResource(ctx, req)
}

func (r *RBACService) CreateVerifyPermissionsResourceRequest(identifier *proto_infra.UserIdentifier, permissionIdentifiers ...string) *proto_auth.VerifyUserResourceRequest {
	if len(permissionIdentifiers) == 0 || len(permissionIdentifiers)%2 != 0 { // Permission identifiers come in pairs: [resource, action, resource, action, ...]
		return nil
	}
	resources := make([]*proto_auth.VerifyResource, 0)
	for i := 0; i < len(permissionIdentifiers); i += 2 {
		resources = append(resources, r.createPermissionVerifyResource(permissionIdentifiers[i], permissionIdentifiers[i+1]))
	}
	return &proto_auth.VerifyUserResourceRequest{
		Identifier:   identifier,
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_PERMISSION,
		Resources:    resources,
	}
}

func (r *RBACService) createPermissionVerifyResource(resource string, action string, hasPermission ...bool) *proto_auth.VerifyResource {
	var boolWrapper *wrapperspb.BoolValue
	if len(hasPermission) > 0 {
		boolWrapper = wrapperspb.Bool(hasPermission[0])
	}
	return &proto_auth.VerifyResource{
		Resource: &proto_auth.VerifyResource_Permission{
			Permission: &proto_auth.Permission{
				Identifier: &proto_auth.Permission_Permission{
					Permission: &proto_auth.PermissionIdentifier{
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
func (r *RBACService) VerifyUserRoles(ctx context.Context, req *proto_auth.VerifyUserResourceRequest) (*proto_auth.VerifyResourceResponse, error) {
	req.ResourceType = proto_auth.ResourceType_RESOURCE_TYPE_ROLE
	return r.VerifyUserResource(ctx, req)
}

func (r *RBACService) CreateVerifyRolesResourceRequest(identifier *proto_infra.UserIdentifier, roles ...string) *proto_auth.VerifyUserResourceRequest {
	if len(roles) == 0 {
		return nil
	}
	resources := make([]*proto_auth.VerifyResource, 0)
	for _, role := range roles {
		resources = append(resources, r.createRolesVerifyResource(role))
	}
	return &proto_auth.VerifyUserResourceRequest{
		Identifier:   identifier,
		ResourceType: proto_auth.ResourceType_RESOURCE_TYPE_ROLE,
		Resources:    resources,
	}
}

func (r *RBACService) createRolesVerifyResource(role string) *proto_auth.VerifyResource {
	return &proto_auth.VerifyResource{
		Resource: &proto_auth.VerifyResource_Role{
			Role: &proto_auth.Role{
				Identifier: &proto_auth.Role_RoleName{
					RoleName: role,
				},
			},
		},
	}
}
