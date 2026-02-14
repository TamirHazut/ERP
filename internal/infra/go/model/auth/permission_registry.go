package auth

import (
	"strings"
	"sync"
)

// PermissionDef represents a single valid permission in the system.
type PermissionDef struct {
	Resource   string
	Action     string
	String     string // "resource:action"
	FeatureFlag string // empty = always active; non-empty = gated by this flag
}

var (
	mu              sync.RWMutex
	allPermissions  []PermissionDef
	permissionSet   map[string]bool // for O(1) IsValidPermissionString
	resourceFlags   map[string]string // resource -> feature flag name
)

func init() {
	allPermissions = make([]PermissionDef, 0)
	permissionSet = make(map[string]bool)
	resourceFlags = make(map[string]string)

	// Register the wildcard resource permissions as special cases.
	// *:* is the full admin wildcard.
	// *:action covers all resources for that action.
	registerSpecialWildcards()

	// Register base infra resources (always active, no feature flag).
	RegisterResourcePermissions("user", []string{"create", "read", "update", "delete", "permission", "role"}, "")
	RegisterResourcePermissions("role", []string{"create", "read", "update", "delete"}, "")
	RegisterResourcePermissions("permission", []string{"read"}, "")
	RegisterResourcePermissions("tenant", []string{"create", "read", "update", "delete"}, "")
	RegisterResourcePermissions("config", []string{"read", "update"}, "")
}

// registerSpecialWildcards adds the *:* and *:{action} permissions.
// These are not registered via RegisterResourcePermissions because
// the * resource is not a concrete resource — it's a wildcard.
func registerSpecialWildcards() {
	wildcards := []string{"*", "create", "read", "update", "delete"}
	for _, action := range wildcards {
		def := PermissionDef{
			Resource: ResourceTypeAll,
			Action:   action,
			String:   ResourceTypeAll + ":" + action,
		}
		allPermissions = append(allPermissions, def)
		permissionSet[def.String] = true
	}
}

// RegisterResourcePermissions registers a resource type with its valid actions
// and an optional feature flag gate. Call from module init().
//
// actions: subset of the standard actions (create, read, update, delete).
// featureFlag: empty string = always active; non-empty = gated by that flag.
//
// A resource:* wildcard is automatically registered in addition to the
// explicitly listed actions.
func RegisterResourcePermissions(resource string, actions []string, featureFlag string) {
	mu.Lock()
	defer mu.Unlock()

	resource = strings.ToLower(resource)

	if featureFlag != "" {
		resourceFlags[resource] = featureFlag
	}

	// Register each explicit action
	for _, action := range actions {
		action = strings.ToLower(action)
		s := resource + ":" + action
		if permissionSet[s] {
			continue // already registered
		}
		def := PermissionDef{
			Resource:    resource,
			Action:      action,
			String:      s,
			FeatureFlag: featureFlag,
		}
		allPermissions = append(allPermissions, def)
		permissionSet[s] = true
	}

	// Auto-register resource:* wildcard
	starPerm := resource + ":*"
	if !permissionSet[starPerm] {
		def := PermissionDef{
			Resource:    resource,
			Action:      PermissionActionAll,
			String:      starPerm,
			FeatureFlag: featureFlag,
		}
		allPermissions = append(allPermissions, def)
		permissionSet[starPerm] = true
	}
}

// GetAllPermissions returns the full set of registered permissions.
func GetAllPermissions() []PermissionDef {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]PermissionDef, len(allPermissions))
	copy(result, allPermissions)
	return result
}

// IsValidPermissionString checks if a permission string is in the registry.
func IsValidPermissionString(s string) bool {
	mu.RLock()
	defer mu.RUnlock()
	return permissionSet[strings.ToLower(s)]
}

// GetActivePermissions returns permissions filtered by active feature flags.
// featureFlags maps flag name -> enabled. Permissions whose resource is gated
// by a disabled flag are excluded. Permissions with no flag are always included.
func GetActivePermissions(featureFlags map[string]bool) []PermissionDef {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]PermissionDef, 0, len(allPermissions))
	for _, p := range allPermissions {
		flag, gated := resourceFlags[p.Resource]
		if gated && !featureFlags[flag] {
			continue
		}
		result = append(result, p)
	}
	return result
}

// NormalizePermissions collapses redundant permissions based on wildcard rules.
//
// Rules:
//  1. If *:* is present, the result is ["*:*"].
//  2. A permission resource:action is removed if *:action OR resource:* is also present.
//  3. Wildcards themselves (*:action, resource:*) are kept unless *:* is present.
func NormalizePermissions(permissions []string) []string {
	if len(permissions) == 0 {
		return permissions
	}

	// Normalize to lowercase and deduplicate
	seen := make(map[string]bool, len(permissions))
	for _, p := range permissions {
		seen[strings.ToLower(p)] = true
	}

	// Rule 1: *:* present → everything collapses
	if seen[PermissionAdminString] {
		return []string{PermissionAdminString}
	}

	// Collect which action-wildcards and resource-wildcards are present
	actionWildcards := make(map[string]bool) // action -> true if *:action is present
	resourceWildcards := make(map[string]bool) // resource -> true if resource:* is present

	for p := range seen {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 {
			continue
		}
		resource, action := parts[0], parts[1]
		if resource == ResourceTypeAll && action != PermissionActionAll {
			actionWildcards[action] = true
		}
		if action == PermissionActionAll && resource != ResourceTypeAll {
			resourceWildcards[resource] = true
		}
	}

	// Rule 2 & 3: filter out non-wildcard permissions covered by a wildcard
	result := make([]string, 0, len(seen))
	for p := range seen {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 {
			result = append(result, p)
			continue
		}
		resource, action := parts[0], parts[1]

		// Keep wildcards themselves
		if resource == ResourceTypeAll || action == PermissionActionAll {
			result = append(result, p)
			continue
		}

		// A concrete permission is redundant if covered by either wildcard type
		if actionWildcards[action] || resourceWildcards[resource] {
			continue
		}

		result = append(result, p)
	}

	return result
}
