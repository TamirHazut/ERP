package auth

// Define an interface that all RBAC resources must implement
type RBACResource interface {
	GetResourceType() string
}
