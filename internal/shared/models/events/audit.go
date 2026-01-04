package events_models

import (
	"fmt"
	"time"

	erp_errors "erp.localhost/internal/errors"
	auth_models "erp.localhost/internal/shared/models/auth"
)

type AuditLog struct {
	// Identity
	ID        string    `bson:"_id" json:"id"`
	TenantID  string    `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"` // Empty for system-wide
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`

	// Event Classification
	Category string `bson:"category" json:"category"` // auth, user_mgmt, order, product, security, config, etc.
	Action   string `bson:"action" json:"action"`     // login_success, order_created, role_assigned, etc.
	Severity string `bson:"severity" json:"severity"` // info, warning, error, critical

	// Actor (who did it)
	ActorID   string `bson:"actor_id,omitempty" json:"actor_id,omitempty"`     // User ID
	ActorType string `bson:"actor_type,omitempty" json:"actor_type,omitempty"` // user, system, api_key, cron
	ActorName string `bson:"actor_name,omitempty" json:"actor_name,omitempty"` // For display

	// Target (what was affected)
	TargetID   string `bson:"target_id,omitempty" json:"target_id,omitempty"`     // Resource ID
	TargetType string `bson:"target_type,omitempty" json:"target_type,omitempty"` // user, order, product, role, etc.
	TargetName string `bson:"target_name,omitempty" json:"target_name,omitempty"` // For display

	// Changes (what changed)
	Changes *Changes `bson:"changes,omitempty" json:"changes,omitempty"`

	// Context
	Context *AuditContext `bson:"context,omitempty" json:"context,omitempty"`

	// Result
	Result  string `bson:"result" json:"result"`                       // success, failure, partial
	Message string `bson:"message,omitempty" json:"message,omitempty"` // Human-readable description
	Error   string `bson:"error,omitempty" json:"error,omitempty"`     // Error message if failed

	// Flexible metadata for type-specific data
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

type Changes struct {
	// Field-level changes
	Fields map[string]*FieldChange `bson:"fields,omitempty" json:"fields,omitempty"`

	// Collection changes (for arrays)
	Added   []string `bson:"added,omitempty" json:"added,omitempty"`
	Removed []string `bson:"removed,omitempty" json:"removed,omitempty"`

	// State transition
	StatusFrom string `bson:"status_from,omitempty" json:"status_from,omitempty"`
	StatusTo   string `bson:"status_to,omitempty" json:"status_to,omitempty"`
}

type FieldChange struct {
	OldValue interface{} `bson:"old_value,omitempty" json:"old_value,omitempty"`
	NewValue interface{} `bson:"new_value,omitempty" json:"new_value,omitempty"`
}

type AuditContext struct {
	IPAddress   string `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	UserAgent   string `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	Location    string `bson:"location,omitempty" json:"location,omitempty"` // Geo-IP
	SessionID   string `bson:"session_id,omitempty" json:"session_id,omitempty"`
	RequestID   string `bson:"request_id,omitempty" json:"request_id,omitempty"`
	APIEndpoint string `bson:"api_endpoint,omitempty" json:"api_endpoint,omitempty"`

	// Additional context fields as needed
	Extra map[string]interface{} `bson:"extra,omitempty" json:"extra,omitempty"`
}

// Validate validates the audit log structure
func (a *AuditLog) Validate() error {
	missingFields := []string{}
	// Required fields
	if a.Category == "" {
		missingFields = append(missingFields, "Category")
	}

	if a.Action == "" {
		missingFields = append(missingFields, "Action")
	}

	if a.Severity == "" {
		missingFields = append(missingFields, "Severity")
	}

	if a.Result == "" {
		missingFields = append(missingFields, "Result")
	}

	if a.ActorType == "" {
		missingFields = append(missingFields, "ActorType")
	}

	if a.TargetType == "" {
		missingFields = append(missingFields, "TargetType")
	}

	if a.ActorID == "" {
		missingFields = append(missingFields, "ActorID")
	}

	if a.TargetID == "" {
		missingFields = append(missingFields, "TargetID")
	}

	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}

	// Validate category
	if !auth_models.IsValidCategory(a.Category) {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "Category", a.Category)
	}

	// Validate action (basic check - action should not be empty and should be reasonable length)
	if len(a.Action) > 100 {
		return erp_errors.Validation(erp_errors.ValidationOutOfRange, "Action", a.Action)
	}

	// Validate severity
	if !auth_models.IsValidSeverity(a.Severity) {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "Severity", a.Severity)
	}

	// Validate result
	if !auth_models.IsValidResult(a.Result) {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "Result", a.Result)
	}

	// Validate actor type if provided
	if !auth_models.IsValidActorType(a.ActorType) {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "ActorType", a.ActorType)
	}

	// Validate target type if provided
	if !auth_models.IsValidTargetType(a.TargetType) {
		return erp_errors.Validation(erp_errors.ValidationInvalidValue, "TargetType", a.TargetType)
	}

	// Logical validations
	// Validate changes structure if present
	if a.Changes != nil {
		if err := a.Changes.Validate(); err != nil {
			return fmt.Errorf("invalid changes: %w", err)
		}
	}

	// Validate context if present
	if a.Context != nil {
		if err := a.Context.Validate(); err != nil {
			return fmt.Errorf("invalid context: %w", err)
		}
	}

	return nil
}

// Validate validates the Changes structure
func (c *Changes) Validate() error {
	missingFields := []string{}
	// If status change is specified, both from and to should be set
	if (c.StatusFrom != "" && c.StatusTo == "") || (c.StatusFrom == "" && c.StatusTo != "") {
		missingFields = append(missingFields, "StatusFrom", "StatusTo")
	}

	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}

	errors := []string{}

	// Validate field changes
	if c.Fields != nil {
		for fieldName, change := range c.Fields {
			if fieldName == "" {
				errors = append(errors, erp_errors.Validation(erp_errors.ValidationRequiredFields, "FieldName").Error())
			}

			if change == nil {
				errors = append(errors, erp_errors.Validation(erp_errors.ValidationRequiredFields, "FieldChange for "+fieldName).Error())
			}

			// At least one of old or new value should be set
			if change != nil && change.OldValue == nil && change.NewValue == nil {
				errors = append(errors, erp_errors.Validation(erp_errors.ValidationRequiredFields, "FieldChange for "+fieldName+" must have at least old_value or new_value").Error())
			}
		}
	}

	// StatusFrom and StatusTo should be different
	if c.StatusFrom != "" && c.StatusTo != "" && c.StatusFrom == c.StatusTo {
		errors = append(errors, erp_errors.Validation(erp_errors.ValidationInvalidValue, "StatusFrom and StatusTo must be different").Error())
	}

	if len(errors) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, errors...)
	}

	return nil
}

// Validate validates the AuditContext structure
func (c *AuditContext) Validate() error {
	missingFields := []string{}
	// IP address validation (basic)
	if c.IPAddress != "" && len(c.IPAddress) > 45 {
		// IPv6 max length is 45 characters
		missingFields = append(missingFields, "IPAddress")
	}

	// User agent validation (basic length check)
	if c.UserAgent != "" && len(c.UserAgent) > 500 {
		missingFields = append(missingFields, "UserAgent")
	}

	// Session ID validation
	if c.SessionID != "" && len(c.SessionID) > 100 {
		missingFields = append(missingFields, "SessionID")
	}

	// Request ID validation
	if c.RequestID != "" && len(c.RequestID) > 100 {
		missingFields = append(missingFields, "RequestID")
	}

	// API endpoint validation
	if c.APIEndpoint != "" && len(c.APIEndpoint) > 500 {
		missingFields = append(missingFields, "APIEndpoint")
	}

	if len(missingFields) > 0 {
		return erp_errors.Validation(erp_errors.ValidationRequiredFields, missingFields...)
	}

	return nil
}
