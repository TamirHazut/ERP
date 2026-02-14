package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetRolesIndexes returns all index definitions for the roles collection
func GetRolesIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_tenant_name_unique"),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}},
			Options: options.Index().SetName("idx_tenant_id"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "permissions", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_permissions"),
		},
	}
}
