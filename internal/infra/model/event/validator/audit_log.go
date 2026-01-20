package validator

import (
	"fmt"

	infra_error "erp.localhost/internal/infra/error"
	model_event "erp.localhost/internal/infra/model/event"
	eventv1 "erp.localhost/internal/infra/model/event/v1"
)

// Validate validates the audit log structure
func ValidateAuditLog(a *eventv1.AuditLog) error {
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

	if a.ActorId == "" {
		missingFields = append(missingFields, "ActorId")
	}

	if a.TargetId == "" {
		missingFields = append(missingFields, "TargetId")
	}

	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}

	// Validate category
	if !model_event.IsValidCategory(a.Category) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "Category", a.Category)
	}

	// Validate action (basic check - action should not be empty and should be reasonable length)
	if len(a.Action) > 100 {
		return infra_error.Validation(infra_error.ValidationOutOfRange, "Action", a.Action)
	}

	// Validate severity
	if !model_event.IsValidSeverity(a.Severity) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "Severity", a.Severity)
	}

	// Validate result
	if !model_event.IsValidResult(a.Result) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "Result", a.Result)
	}

	// Validate actor type if provided
	if !model_event.IsValidActorType(a.ActorType) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "ActorType", a.ActorType)
	}

	// Validate target type if provided
	if !model_event.IsValidTargetType(a.TargetType) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "TargetType", a.TargetType)
	}

	// Logical validations
	// Validate changes structure if present
	if a.Changes != nil {
		if err := ValidateChanges(a.Changes); err != nil {
			return fmt.Errorf("invalid changes: %w", err)
		}
	}

	// Validate context if present
	if a.Context != nil {
		if err := ValidateAuditContext(a.Context); err != nil {
			return fmt.Errorf("invalid context: %w", err)
		}
	}

	return nil
}

// Validate validates the Changes structure
func ValidateChanges(c *eventv1.Changes) error {
	missingFields := []string{}
	// If status change is specified, both from and to should be set
	if (c.StatusFrom != "" && c.StatusTo == "") || (c.StatusFrom == "" && c.StatusTo != "") {
		missingFields = append(missingFields, "StatusFrom", "StatusTo")
	}

	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}

	errors := []string{}

	// Validate field changes
	if c.Fields != nil {
		for fieldName, change := range c.Fields {
			if fieldName == "" {
				errors = append(errors, infra_error.Validation(infra_error.ValidationRequiredFields, "FieldName").Error())
			}

			if change == nil {
				errors = append(errors, infra_error.Validation(infra_error.ValidationRequiredFields, "FieldChange for "+fieldName).Error())
			}

			// At least one of old or new value should be set
			if change != nil && change.OldValue == nil && change.NewValue == nil {
				errors = append(errors, infra_error.Validation(infra_error.ValidationRequiredFields, "FieldChange for "+fieldName+" must have at least old_value or new_value").Error())
			}
		}
	}

	// StatusFrom and StatusTo should be different
	if c.StatusFrom != "" && c.StatusTo != "" && c.StatusFrom == c.StatusTo {
		errors = append(errors, infra_error.Validation(infra_error.ValidationInvalidValue, "StatusFrom and StatusTo must be different").Error())
	}

	if len(errors) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, errors...)
	}

	return nil
}

// Validate validates the AuditContext structure
func ValidateAuditContext(c *eventv1.AuditContext) error {
	missingFields := []string{}
	// IP address validation (basic)
	if c.IpAddress != "" && len(c.IpAddress) > 45 {
		// IPv6 max length is 45 characters
		missingFields = append(missingFields, "IPAddress")
	}

	// User agent validation (basic length check)
	if c.UserAgent != "" && len(c.UserAgent) > 500 {
		missingFields = append(missingFields, "UserAgent")
	}

	// Session Id validation
	if c.SessionId != "" && len(c.SessionId) > 100 {
		missingFields = append(missingFields, "SessionId")
	}

	// Request Id validation
	if c.RequestId != "" && len(c.RequestId) > 100 {
		missingFields = append(missingFields, "RequestId")
	}

	// API endpoint validation
	if c.ApiEndpoint != "" && len(c.ApiEndpoint) > 500 {
		missingFields = append(missingFields, "APIEndpoint")
	}

	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}

	return nil
}
