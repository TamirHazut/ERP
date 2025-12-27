package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LogID        string             `bson:"log_id" json:"log_id"`
	TenantID     string             `bson:"tenant_id" json:"tenant_id"`
	Actor        Actor              `bson:"actor" json:"actor"`
	Action       string             `bson:"action" json:"action"`
	ResourceType string             `bson:"resource_type" json:"resource_type"`
	ResourceID   string             `bson:"resource_id" json:"resource_id"`
	Changes      Changes            `bson:"changes,omitempty" json:"changes,omitempty"`
	Status       string             `bson:"status" json:"status"` // success, failure
	ErrorMessage string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
	Metadata     AuditMetadata      `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
}

type Actor struct {
	UserID    string `bson:"user_id" json:"user_id"`
	Username  string `bson:"username" json:"username"`
	IPAddress string `bson:"ip_address" json:"ip_address"`
	UserAgent string `bson:"user_agent" json:"user_agent"`
}

type Changes struct {
	Before map[string]interface{} `bson:"before,omitempty" json:"before,omitempty"`
	After  map[string]interface{} `bson:"after,omitempty" json:"after,omitempty"`
}

type AuditMetadata struct {
	RequestID  string `bson:"request_id,omitempty" json:"request_id,omitempty"`
	SessionID  string `bson:"session_id,omitempty" json:"session_id,omitempty"`
	APIVersion string `bson:"api_version,omitempty" json:"api_version,omitempty"`
}
