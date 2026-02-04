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
// BuildUserRolesPipeline
// ==========================================================
//
// Purpose:
//
//	Fetch all roles assigned to a user as full role documents.
//	The output includes the role's Permissions field (permission strings).
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
