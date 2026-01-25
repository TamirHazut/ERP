package pipeline

import (
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BuildUserPermissionsPipeline creates a pipeline to get all permissions for a user
// This replaces the N+1 query pattern: User → Roles → Permissions
// Pipeline: Match user → Lookup roles → Unwind → Lookup permissions from role → Unwind → Group unique permissions
func BuildUserPermissionsPipeline(tenantID, userID string) []bson.M {
	userObjectID, _ := primitive.ObjectIDFromHex(userID)

	return []bson.M{
		// Stage 1: Match the specific user
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       userObjectID,
			},
		},
		// Stage 2: Unwind user roles array to process each role
		{
			"$unwind": bson.M{
				"path":                       "$roles",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Stage 3: Lookup role details from roles collection
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "roles.role_id",
				"foreignField": "_id",
				"as":           "role_details",
			},
		},
		// Stage 4: Unwind role_details array
		{
			"$unwind": bson.M{
				"path":                       "$role_details",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Stage 5: Unwind permissions array from role
		{
			"$unwind": bson.M{
				"path":                       "$role_details.permissions",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Stage 6: Lookup permission details from permissions collection
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.PermissionsCollection),
				"localField":   "role_details.permissions",
				"foreignField": "_id",
				"as":           "permission_details",
			},
		},
		// Stage 7: Unwind permission_details array
		{
			"$unwind": bson.M{
				"path":                       "$permission_details",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Stage 8: Also handle additional_permissions from user
		{
			"$unionWith": bson.M{
				"coll": string(model_mongo.UsersCollection),
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"tenant_id": tenantID,
							"_id":       userObjectID,
						},
					},
					{
						"$unwind": bson.M{
							"path":                       "$additional_permissions",
							"preserveNullAndEmptyArrays": true,
						},
					},
					{
						"$lookup": bson.M{
							"from":         string(model_mongo.PermissionsCollection),
							"localField":   "additional_permissions",
							"foreignField": "_id",
							"as":           "permission_details",
						},
					},
					{
						"$unwind": "$permission_details",
					},
				},
			},
		},
		// Stage 9: Group by permission ID to get unique permissions
		{
			"$group": bson.M{
				"_id": "$permission_details._id",
				"permission": bson.M{
					"$first": "$permission_details",
				},
			},
		},
		// Stage 10: Replace root with permission document
		{
			"$replaceRoot": bson.M{
				"newRoot": "$permission",
			},
		},
	}
}

// BuildUserRolesPipeline creates a pipeline to get all roles for a user
// This replaces the N query pattern where we fetch each role individually
// Pipeline: Match user → Unwind roles → Lookup role details
func BuildUserRolesPipeline(tenantID, userID string) []bson.M {
	userObjectID, _ := primitive.ObjectIDFromHex(userID)

	return []bson.M{
		// Stage 1: Match the specific user
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       userObjectID,
			},
		},
		// Stage 2: Unwind user roles array
		{
			"$unwind": bson.M{
				"path":                       "$roles",
				"preserveNullAndEmptyArrays": false,
			},
		},
		// Stage 3: Lookup role details from roles collection
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "roles.role_id",
				"foreignField": "_id",
				"as":           "role_details",
			},
		},
		// Stage 4: Unwind role_details array
		{
			"$unwind": "$role_details",
		},
		// Stage 5: Replace root with role document
		{
			"$replaceRoot": bson.M{
				"newRoot": "$role_details",
			},
		},
	}
}

// BuildRolePermissionsPipeline creates a pipeline to get all permissions for a role
// Pipeline: Match role → Unwind permissions → Lookup permission details
func BuildRolePermissionsPipeline(tenantID, roleID string) []bson.M {
	roleObjectID, _ := primitive.ObjectIDFromHex(roleID)

	return []bson.M{
		// Stage 1: Match the specific role
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       roleObjectID,
			},
		},
		// Stage 2: Unwind permissions array
		{
			"$unwind": bson.M{
				"path":                       "$permissions",
				"preserveNullAndEmptyArrays": false,
			},
		},
		// Stage 3: Lookup permission details from permissions collection
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.PermissionsCollection),
				"localField":   "permissions",
				"foreignField": "_id",
				"as":           "permission_details",
			},
		},
		// Stage 4: Unwind permission_details array
		{
			"$unwind": "$permission_details",
		},
		// Stage 5: Replace root with permission document
		{
			"$replaceRoot": bson.M{
				"newRoot": "$permission_details",
			},
		},
	}
}

// BuildBatchGetByIDsPipeline creates a pipeline to get multiple documents by IDs
// Uses $in operator for efficient batch fetching
func BuildBatchGetByIDsPipeline(tenantID string, ids []string) []bson.M {
	// Convert string IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if objectID, err := primitive.ObjectIDFromHex(id); err == nil {
			objectIDs = append(objectIDs, objectID)
		}
	}

	return []bson.M{
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       bson.M{"$in": objectIDs},
			},
		},
	}
}

// BuildTenantWithUsersPipeline creates a pipeline to get a tenant with all its users
func BuildTenantWithUsersPipeline(tenantID string) []bson.M {
	tenantObjectID, _ := primitive.ObjectIDFromHex(tenantID)

	return []bson.M{
		// Stage 1: Match the specific tenant
		{
			"$match": bson.M{
				"_id": tenantObjectID,
			},
		},
		// Stage 2: Lookup all users for this tenant
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.UsersCollection),
				"localField":   "_id",
				"foreignField": "tenant_id",
				"as":           "users",
			},
		},
	}
}

// BuildTenantWithRolesPipeline creates a pipeline to get a tenant with all its roles
func BuildTenantWithRolesPipeline(tenantID string) []bson.M {
	tenantObjectID, _ := primitive.ObjectIDFromHex(tenantID)

	return []bson.M{
		// Stage 1: Match the specific tenant
		{
			"$match": bson.M{
				"_id": tenantObjectID,
			},
		},
		// Stage 2: Lookup all roles for this tenant
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "_id",
				"foreignField": "tenant_id",
				"as":           "roles",
			},
		},
	}
}

// BuildRolesWithUserCountsPipeline creates a pipeline to get all roles with user counts
func BuildRolesWithUserCountsPipeline(tenantID string) []bson.M {
	return []bson.M{
		// Stage 1: Match roles for tenant
		{
			"$match": bson.M{
				"tenant_id": tenantID,
			},
		},
		// Stage 2: Lookup users that have this role
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.UsersCollection),
				"localField":   "_id",
				"foreignField": "roles.role_id",
				"as":           "users_with_role",
			},
		},
		// Stage 3: Add user_count field
		{
			"$addFields": bson.M{
				"user_count": bson.M{"$size": "$users_with_role"},
			},
		},
		// Stage 4: Remove the users_with_role array (we only need the count)
		{
			"$project": bson.M{
				"users_with_role": 0,
			},
		},
	}
}

// BuildPermissionsWithRoleCountsPipeline creates a pipeline to get all permissions with role usage counts
func BuildPermissionsWithRoleCountsPipeline(tenantID string) []bson.M {
	return []bson.M{
		// Stage 1: Match permissions for tenant
		{
			"$match": bson.M{
				"tenant_id": tenantID,
			},
		},
		// Stage 2: Lookup roles that have this permission
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "_id",
				"foreignField": "permissions",
				"as":           "roles_with_permission",
			},
		},
		// Stage 3: Add role_count field
		{
			"$addFields": bson.M{
				"role_count": bson.M{"$size": "$roles_with_permission"},
			},
		},
		// Stage 4: Remove the roles_with_permission array (we only need the count)
		{
			"$project": bson.M{
				"roles_with_permission": 0,
			},
		},
	}
}

// BuildUnusedPermissionsPipeline creates a pipeline to get permissions not assigned to any role
func BuildUnusedPermissionsPipeline(tenantID string) []bson.M {
	return []bson.M{
		// Stage 1: Match permissions for tenant
		{
			"$match": bson.M{
				"tenant_id": tenantID,
			},
		},
		// Stage 2: Lookup roles that have this permission
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "_id",
				"foreignField": "permissions",
				"as":           "roles_with_permission",
			},
		},
		// Stage 3: Match only permissions with no roles
		{
			"$match": bson.M{
				"roles_with_permission": bson.M{"$size": 0},
			},
		},
		// Stage 4: Remove the roles_with_permission array
		{
			"$project": bson.M{
				"roles_with_permission": 0,
			},
		},
	}
}
