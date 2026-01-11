package collection

import (
	"time"

	"erp.localhost/internal/infra/db/mongo/collection"
	infra_error "erp.localhost/internal/infra/error"
	"erp.localhost/internal/infra/logging/logger"
	model_event "erp.localhost/internal/infra/model/event"
)

// TODO: move this to Events service and consume from kafka topics
type AuditLogsCollection struct {
	collection collection.CollectionHandler[model_event.AuditLog]
	logger     logger.Logger
}

func NewAuditLogsCollection(collection collection.CollectionHandler[model_event.AuditLog], logger logger.Logger) *AuditLogsCollection {
	return &AuditLogsCollection{
		collection: collection,
		logger:     logger,
	}
}

func (c *AuditLogsCollection) CreateAuditLog(tenantID string, auditLog *model_event.AuditLog) error {
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID")
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
func (c *AuditLogsCollection) GetAuditLogsByFilter(tenantID string, filter map[string]any) ([]*model_event.AuditLog, error) {
	if tenantID == "" {
		return nil, infra_error.Validation(infra_error.ValidationRequiredFields, "tenantID")
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
