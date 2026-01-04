package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo"
	erp_errors "erp.localhost/internal/infra/errors"
	"erp.localhost/internal/infra/logging"
	mongo_models "erp.localhost/internal/infra/models/db/mongo"
	events_models "erp.localhost/internal/infra/models/events"
	shared_models "erp.localhost/internal/infra/models/shared"
)

// TODO: move this to Events service and consume from kafka topics
type AuditLogsCollection struct {
	collection mongo.CollectionHandler[events_models.AuditLog]
	logger     *logging.Logger
}

func NewAuditLogsCollection(collection mongo.CollectionHandler[events_models.AuditLog]) *AuditLogsCollection {
	logger := logging.NewLogger(shared_models.ModuleAuth)
	if collection == nil {
		collectionHandler := mongo.NewBaseCollectionHandler[events_models.AuditLog](string(mongo_models.AuditLogsCollection), logger)
		if collectionHandler == nil {
			logger.Fatal("failed to create audit logs collection handler")
			return nil
		}
		collection = collectionHandler
	}
	return &AuditLogsCollection{
		collection: collection,
		logger:     logger,
	}
}

func (c *AuditLogsCollection) CreateAuditLog(tenantID string, auditLog events_models.AuditLog) error {
	if tenantID == "" {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenantID")
	}
	auditLog.TenantID = tenantID
	if err := auditLog.Validate(); err != nil {
		return err
	}

	auditLog.Timestamp = time.Now()
	c.logger.Debug("Creating audit log", "auditLog", auditLog)
	_, err := c.collection.Create(auditLog)
	if err != nil {
		return err
	}
	return nil
}

// GetAuditLogsByFilter gets audit logs by filter
// tenantID is the tenant ID
// filter is the filter to apply to the audit logs
// filter can be:
// - category
// - action
// - severity
// - result
// - actor_type
// - actor_name
// - target_type
// - target_name
// - target_id
// - resource_type
// - resource_id
// - resource_name
func (c *AuditLogsCollection) GetAuditLogsByFilter(tenantID string, filter map[string]any) ([]events_models.AuditLog, error) {
	if tenantID == "" {
		return nil, erp_errors.Validation(erp_errors.ValidationRequiredFields, "tenantID")
	}
	if filter == nil {
		filter = make(map[string]any)
	}
	filter["tenant_id"] = tenantID
	auditLogs, err := c.collection.FindAll(filter)
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}
