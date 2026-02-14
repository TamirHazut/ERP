package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAuditLogsIndexes returns all index definitions for the audit_logs collection
func GetAuditLogsIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "timestamp", Value: -1}, // Descending for recent-first queries
			},
			Options: options.Index().SetName("idx_tenant_timestamp"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "category", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_category"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "action", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_action"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "severity", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_severity"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "actor_id", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_actor"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "target_id", Value: 1},
			},
			Options: options.Index().SetName("idx_tenant_target"),
		},
	}
}
