package pipeline

import (
	model_mongo "erp.localhost/internal/infra/model/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//
// ---------- Helpers ----------
//

// safeObjectIdConvert builds a $convert expression that safely converts a value to ObjectId.
// We use this everywhere IDs are stored as strings to avoid lookup mismatches and crashes.
func safeObjectIdConvert(field string) bson.M {
	return bson.M{
		"$convert": bson.M{
			"input":   field,
			"to":      "objectId",
			"onError": nil,
			"onNull":  nil,
		},
	}
}

// ==========================================================
// BuildUserPermissionsPipeline
// ==========================================================
//
// Purpose:
//
//	Resolve ALL permissions a user effectively has.
//	This includes:
//	  - permissions inherited via roles
//	  - permissions directly assigned to the user
//
// Why this pipeline exists:
//
//	Avoids N+1 queries:
//	  user → roles → permissions
func BuildUserPermissionsPipeline(tenantID, userID string) []bson.M {
	userObjectID, _ := primitive.ObjectIDFromHex(userID)

	return []bson.M{
		// Select the single user within the tenant.
		// Everything else in this pipeline operates on this user only.
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       userObjectID,
			},
		},

		// Expand the user's roles array so each role can be processed independently.
		// preserveNullAndEmptyArrays allows users with no roles to still continue
		// (they may still have additional_permissions).
		{
			"$unwind": bson.M{
				"path":                       "$roles",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// Convert roles.role_id from string → ObjectId so it can be joined
		// against roles._id in the roles collection.
		{
			"$addFields": bson.M{
				"roles.role_id": safeObjectIdConvert("$roles.role_id"),
			},
		},

		// Join the full role document for each user role.
		// This gives us access to role.permissions.
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "roles.role_id",
				"foreignField": "_id",
				"as":           "role_details",
			},
		},

		// Flatten the joined role document.
		{
			"$unwind": bson.M{
				"path":                       "$role_details",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// Expand the permissions array inside each role so permissions
		// can be resolved individually.
		{
			"$unwind": bson.M{
				"path":                       "$role_details.permissions",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// Convert role permission IDs from string → ObjectId
		// so they can be joined against permissions._id.
		{
			"$addFields": bson.M{
				"role_details.permissions": safeObjectIdConvert("$role_details.permissions"),
			},
		},

		// Join the permission documents referenced by the role.
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.PermissionsCollection),
				"localField":   "role_details.permissions",
				"foreignField": "_id",
				"as":           "permission_details",
			},
		},

		// Flatten the permission document.
		{
			"$unwind": bson.M{
				"path":                       "$permission_details",
				"preserveNullAndEmptyArrays": true,
			},
		},

		// UNION additional_permissions directly assigned to the user.
		// These bypass roles entirely.
		{
			"$unionWith": bson.M{
				"coll": string(model_mongo.UsersCollection),
				"pipeline": []bson.M{
					// Re-select the same user.
					{
						"$match": bson.M{
							"tenant_id": tenantID,
							"_id":       userObjectID,
						},
					},

					// Expand additional_permissions array.
					{
						"$unwind": bson.M{
							"path":                       "$additional_permissions",
							"preserveNullAndEmptyArrays": true,
						},
					},

					// Convert permission ID to ObjectId for lookup.
					{
						"$addFields": bson.M{
							"additional_permissions": safeObjectIdConvert("$additional_permissions"),
						},
					},

					// Join permission documents.
					{
						"$lookup": bson.M{
							"from":         string(model_mongo.PermissionsCollection),
							"localField":   "additional_permissions",
							"foreignField": "_id",
							"as":           "permission_details",
						},
					},

					// Flatten permission document.
					{
						"$unwind": "$permission_details",
					},
				},
			},
		},

		// Deduplicate permissions coming from multiple roles
		// or both role-based and direct assignment.
		{
			"$group": bson.M{
				"_id": "$permission_details._id",
				"permission": bson.M{
					"$first": "$permission_details",
				},
			},
		},

		// Output clean permission documents as the final result.
		{
			"$replaceRoot": bson.M{
				"newRoot": "$permission",
			},
		},
	}
}

// ==========================================================
// BuildUserRolesPipeline
// ==========================================================
//
// Purpose:
//
//	Fetch all roles assigned to a user as full role documents.
func BuildUserRolesPipeline(tenantID, userID string) []bson.M {
	userObjectID, _ := primitive.ObjectIDFromHex(userID)

	return []bson.M{
		// Select the user within the tenant.
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       userObjectID,
			},
		},

		// Expand roles array so each role can be resolved.
		{
			"$unwind": "$roles",
		},

		// Normalize role_id for lookup compatibility.
		{
			"$addFields": bson.M{
				"roles.role_id": safeObjectIdConvert("$roles.role_id"),
			},
		},

		// Join role documents.
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.RolesCollection),
				"localField":   "roles.role_id",
				"foreignField": "_id",
				"as":           "role_details",
			},
		},

		// Flatten joined role.
		{
			"$unwind": "$role_details",
		},

		// Output role document directly.
		{
			"$replaceRoot": bson.M{
				"newRoot": "$role_details",
			},
		},
	}
}

// ==========================================================
// BuildRolePermissionsPipeline
// ==========================================================
//
// Purpose:
//
//	Resolve all permissions belonging to a single role.
func BuildRolePermissionsPipeline(tenantID, roleID string) []bson.M {
	roleObjectID, _ := primitive.ObjectIDFromHex(roleID)

	return []bson.M{
		// Select the role within the tenant.
		{
			"$match": bson.M{
				"tenant_id": tenantID,
				"_id":       roleObjectID,
			},
		},

		// Expand permissions array so each permission can be resolved.
		{
			"$unwind": "$permissions",
		},

		// Normalize permission ID for lookup.
		{
			"$addFields": bson.M{
				"permissions": safeObjectIdConvert("$permissions"),
			},
		},

		// Join permission documents.
		{
			"$lookup": bson.M{
				"from":         string(model_mongo.PermissionsCollection),
				"localField":   "permissions",
				"foreignField": "_id",
				"as":           "permission_details",
			},
		},

		// Flatten permission.
		{
			"$unwind": "$permission_details",
		},

		// Output permission document.
		{
			"$replaceRoot": bson.M{
				"newRoot": "$permission_details",
			},
		},
	}
}
