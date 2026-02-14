package auth

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPermissionString(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		expected   bool
	}{
		// Valid infra permissions
		{"user:create", "user:create", true},
		{"user:read", "user:read", true},
		{"user:update", "user:update", true},
		{"user:delete", "user:delete", true},
		{"user:permission", "user:permission", true},
		{"user:role", "user:role", true},
		{"user:*", "user:*", true},
		{"role:create", "role:create", true},
		{"role:read", "role:read", true},
		{"role:*", "role:*", true},
		{"permission:read", "permission:read", true},
		{"permission:*", "permission:*", true},
		{"tenant:create", "tenant:create", true},
		{"tenant:delete", "tenant:delete", true},
		{"tenant:*", "tenant:*", true},
		{"config:read", "config:read", true},
		{"config:update", "config:update", true},
		{"config:*", "config:*", true},
		// Wildcard resource permissions
		{"*:*", "*:*", true},
		{"*:create", "*:create", true},
		{"*:read", "*:read", true},
		{"*:update", "*:update", true},
		{"*:delete", "*:delete", true},
		// Case insensitive
		{"USER:READ uppercase", "USER:READ", true},
		{"User:Read mixed", "User:Read", true},
		// Invalid permissions
		{"token:read - token not a resource", "token:read", false},
		{"token:delete - token not a resource", "token:delete", false},
		{"permission:create - no CRUD on permission", "permission:create", false},
		{"permission:update - no CRUD on permission", "permission:update", false},
		{"permission:delete - no CRUD on permission", "permission:delete", false},
		{"config:create - no create on config", "config:create", false},
		{"config:delete - no delete on config", "config:delete", false},
		{"*:permission - no *:permission wildcard", "*:permission", false},
		{"*:role - no *:role wildcard", "*:role", false},
		{"garbage:garbage", "garbage:garbage", false},
		{"empty string", "", false},
		{"no colon", "userread", false},
		{"only colon", ":", false},
		{"role:permission - not valid on role", "role:permission", false},
		{"tenant:role - not valid on tenant", "tenant:role", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPermissionString(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAllPermissions_ContainsExpectedInfraPermissions(t *testing.T) {
	perms := GetAllPermissions()
	permStrings := make(map[string]bool, len(perms))
	for _, p := range perms {
		permStrings[p.String] = true
	}

	expected := []string{
		"*:*", "*:create", "*:read", "*:update", "*:delete",
		"user:create", "user:read", "user:update", "user:delete", "user:permission", "user:role", "user:*",
		"role:create", "role:read", "role:update", "role:delete", "role:*",
		"permission:read", "permission:*",
		"tenant:create", "tenant:read", "tenant:update", "tenant:delete", "tenant:*",
		"config:read", "config:update", "config:*",
	}

	for _, e := range expected {
		assert.True(t, permStrings[e], "expected permission %q to be in registry", e)
	}
}

func TestGetAllPermissions_DoesNotContainInvalidPermissions(t *testing.T) {
	perms := GetAllPermissions()
	permStrings := make(map[string]bool, len(perms))
	for _, p := range perms {
		permStrings[p.String] = true
	}

	invalid := []string{
		"token:read", "token:delete", "token:*",
		"permission:create", "permission:update", "permission:delete",
		"config:create", "config:delete",
		"*:permission", "*:role",
	}

	for _, inv := range invalid {
		assert.False(t, permStrings[inv], "permission %q should NOT be in registry", inv)
	}
}

func TestGetActivePermissions_AllFlagsEnabled(t *testing.T) {
	// With no flags registered (base infra only), all flags enabled returns everything
	allPerms := GetAllPermissions()
	activePerms := GetActivePermissions(map[string]bool{})

	// All infra perms have no flag, so they are always active
	assert.Equal(t, len(allPerms), len(activePerms))
}

func TestGetActivePermissions_FlagDisabled(t *testing.T) {
	// Register a test resource with a feature flag
	RegisterResourcePermissions("testresource", []string{"create", "read"}, "test_module_flag")
	defer func() {
		// Cleanup: remove the test resource from the registry
		mu.Lock()
		defer mu.Unlock()
		cleaned := make([]PermissionDef, 0, len(allPermissions))
		for _, p := range allPermissions {
			if p.Resource != "testresource" {
				cleaned = append(cleaned, p)
			}
		}
		allPermissions = cleaned
		delete(permissionSet, "testresource:create")
		delete(permissionSet, "testresource:read")
		delete(permissionSet, "testresource:*")
		delete(resourceFlags, "testresource")
	}()

	// Flag disabled → testresource permissions excluded
	active := GetActivePermissions(map[string]bool{"test_module_flag": false})
	for _, p := range active {
		assert.NotEqual(t, "testresource", p.Resource, "testresource should be excluded when flag is disabled")
	}

	// Flag enabled → testresource permissions included
	active = GetActivePermissions(map[string]bool{"test_module_flag": true})
	found := false
	for _, p := range active {
		if p.Resource == "testresource" {
			found = true
			break
		}
	}
	assert.True(t, found, "testresource should be included when flag is enabled")
}

func TestNormalizePermissions_Empty(t *testing.T) {
	result := NormalizePermissions([]string{})
	assert.Empty(t, result)
}

func TestNormalizePermissions_AdminWildcard(t *testing.T) {
	// *:* collapses everything
	result := NormalizePermissions([]string{"user:read", "role:create", "*:*", "tenant:delete"})
	assert.Equal(t, []string{"*:*"}, result)
}

func TestNormalizePermissions_ActionWildcard(t *testing.T) {
	// *:read removes all {resource}:read
	input := []string{"user:read", "role:read", "role:create", "*:read"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	expected := []string{"*:read", "role:create"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_ResourceWildcard(t *testing.T) {
	// user:* removes all user:{action}
	input := []string{"user:read", "user:create", "user:*", "role:read"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	expected := []string{"role:read", "user:*"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_BothWildcardsNoOverlap(t *testing.T) {
	// user:* and *:read overlap on user:read but neither subsumes the other
	input := []string{"user:*", "*:read", "role:create"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	expected := []string{"*:read", "role:create", "user:*"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_BothWildcardsRemovesConcrete(t *testing.T) {
	// user:read is covered by both user:* and *:read — removed either way
	input := []string{"user:read", "user:*", "*:read", "role:create"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	expected := []string{"*:read", "role:create", "user:*"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_NoDuplicates(t *testing.T) {
	input := []string{"user:read", "user:read", "role:create", "role:create"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	expected := []string{"role:create", "user:read"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_CaseInsensitive(t *testing.T) {
	input := []string{"User:Read", "ROLE:CREATE", "*:READ"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	// user:read removed by *:read; role:create kept
	expected := []string{"*:read", "role:create"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_NoRedundancy(t *testing.T) {
	// All permissions are distinct, no wildcards — nothing removed
	input := []string{"user:read", "role:create", "tenant:delete"}
	result := NormalizePermissions(input)
	sort.Strings(result)
	expected := []string{"role:create", "tenant:delete", "user:read"}
	sort.Strings(expected)
	assert.Equal(t, expected, result)
}

func TestNormalizePermissions_SinglePermission(t *testing.T) {
	result := NormalizePermissions([]string{"user:read"})
	assert.Equal(t, []string{"user:read"}, result)
}
