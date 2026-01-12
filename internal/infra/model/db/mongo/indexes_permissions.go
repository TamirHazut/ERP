package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetPermissionsIndexes returns all index definitions for the permissions collection
func GetPermissionsIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "permission_string", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_tenant_permission_string_unique"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "resource", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_resource"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "resource", Value: 1},
				{Key: "action", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_resource_action"),
		},
	}
}
