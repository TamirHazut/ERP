package rbac

// //go:generate mockgen -destination=mock/mock_rbac_handler.go -package=mock erp.localhost/internal/auth/rbac RBACHanlder
// //go:generate mockgen -destination=mock/mock_rbac_checker.go -package=mock erp.localhost/internal/auth/rbac RBACChecker

// // RBACHandler is for the CRUD operations of RBAC types (Roles, Permissions)
// type RBACHanlder[T any] interface {
// 	CreateResource(tenantID string, userID string, resource *T) (string, error)
// 	UpdateResource(tenantID string, userID string, resource *T) error
// 	DeleteResource(tenantID string, userID string, resourceID string) error
// 	GetResource(tenantID string, userID string, resourceID string) (*T, error)
// 	ListResources(tenantID string, userID string) ([]*T, error)
// }

// type RBACChecker interface {
// 	CheckUserPermissions(tenantID string, userID string, permissions []string) (map[string]bool, error)
// 	CheckUserRoles(tenantID string, userID string, roles []string) (map[string]bool, error)
// 	VerifyUserRole(tenantID string, userID string, roleID string) (bool, error)
// 	VerifyRolePermissions(tenantID string, roleID string, permissions []string) (map[string]bool, error)
// 	HasPermission(tenantID string, userID string, permission string) error
// }

// // RBACChecker - an interface for all the non-rbac components of the auth service who needs to verify permissions and roles
// type RBACChecker interface {
// 	VerifyUserPermissions(ctx context.Context, req *proto_auth.VerifyUserResourceRequest) (*proto_auth.VerifyResourceResponse, error)
// 	VerifyUserRoles(ctx context.Context, req *proto_auth.VerifyUserResourceRequest) (*proto_auth.VerifyResourceResponse, error)

// 	CreateVerifyPermissionsResourceRequest(identifier *proto_infra.UserIdentifier, permissionIdentifiers ...string) *proto_auth.VerifyUserResourceRequest
// 	CreateVerifyRolesResourceRequest(identifier *proto_infra.UserIdentifier, roles ...string) *proto_auth.VerifyUserResourceRequest
// }
