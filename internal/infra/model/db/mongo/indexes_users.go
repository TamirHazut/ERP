package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetUsersIndexes returns all index definitions for the users collection
func GetUsersIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_tenant_email_unique"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "username", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_tenant_username_unique"),
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
			Options: options.Index().SetName("idx_tenant_id"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_status"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "roles.role_id", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_roles"),
		},
	}
}
