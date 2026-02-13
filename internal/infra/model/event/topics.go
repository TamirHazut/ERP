package event

type Topic string

const (
	AuthPermissionCreated   Topic = "auth.permission.created"
	AuthPermissionUpdated   Topic = "auth.permission.updated"
	AuthPermissionDeleted   Topic = "auth.permission.deleted"
	AuthRoleCreated         Topic = "auth.role.created"
	AuthRoleUpdated         Topic = "auth.role.updated"
	AuthRoleDeleted         Topic = "auth.role.deleted"
	AuthRoleAssigned        Topic = "auth.role.assigned"
	AuthRoleRevoked         Topic = "auth.role.revoked"
	AuthTenantCreated       Topic = "auth.tenant.created"
	AuthTenantUpdated       Topic = "auth.tenant.updated"
	AuthTenantDeleted       Topic = "auth.tenant.deleted"
	AuthTenantTokensRevoked Topic = "auth.tenant.tokens_revoked"
	AuthTokenRefreshed      Topic = "auth.token.refreshed"
	AuthTokenRevoked        Topic = "auth.token.revoked"
	AuthUserCreated         Topic = "auth.user.created"
	AuthUserUpdated         Topic = "auth.user.updated"
	AuthUserDeleted         Topic = "auth.user.deleted"
	AuthUserLogin           Topic = "auth.user.login"
	AuthUserLogout          Topic = "auth.user.logout"
)
