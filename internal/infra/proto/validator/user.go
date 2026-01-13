package validator

import (
	"regexp"
	"strings"

	infra_error "erp.localhost/internal/infra/error"
	model_auth "erp.localhost/internal/infra/model/auth"
	proto_auth "erp.localhost/internal/infra/proto/auth/v1"
	proto_infra "erp.localhost/internal/infra/proto/infra/v1"
)

var (
	// Email validation regex (basic RFC 5322 validation)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Username validation: 3-50 characters, alphanumeric, underscore, hyphen, dot
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._\-]{3,50}$`)

	// Phone validation: basic international format
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

func ValidateUserIdentifier(identifier *proto_infra.UserIdentifier) error {
	if identifier == nil {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "identifier")
	}
	tenantID := identifier.GetTenantId()
	if tenantID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	userID := identifier.GetUserId()
	if userID == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user_id")
	}
	return nil
}

// Helper validation functions

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	email = strings.TrimSpace(email)
	if len(email) > 254 { // RFC 5321
		return false
	}
	return emailRegex.MatchString(email)
}

func isValidUsername(username string) bool {
	if username == "" {
		return false
	}
	username = strings.TrimSpace(username)
	return usernameRegex.MatchString(username)
}

func isValidPhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}
	phone = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", ""))
	return phoneRegex.MatchString(phone)
}

func validateUserProfile(profile *proto_auth.UserProfileData) error {
	if profile == nil {
		return nil // Profile is optional
	}

	// Validate phone if provided
	if profile.Phone != "" && !isValidPhone(profile.Phone) {
		return infra_error.Validation(infra_error.ValidationInvalidPhone, "profile.phone")
	}

	// Validate field lengths
	if len(profile.FirstName) > 100 {
		return infra_error.Validation(infra_error.ValidationTooLong, "profile.first_name")
	}
	if len(profile.LastName) > 100 {
		return infra_error.Validation(infra_error.ValidationTooLong, "profile.last_name")
	}
	if len(profile.DisplayName) > 200 {
		return infra_error.Validation(infra_error.ValidationTooLong, "profile.display_name")
	}
	if len(profile.Title) > 100 {
		return infra_error.Validation(infra_error.ValidationTooLong, "profile.title")
	}
	if len(profile.Department) > 100 {
		return infra_error.Validation(infra_error.ValidationTooLong, "profile.department")
	}

	return nil
}

func validateUserPreferences(preferences *proto_auth.UserPreferencesData) error {
	if preferences == nil {
		return nil // Preferences are optional
	}

	// Validate timezone format (basic validation)
	if preferences.Timezone != "" && len(preferences.Timezone) > 100 {
		return infra_error.Validation(infra_error.ValidationTooLong, "preferences.timezone")
	}

	// Validate language code (basic validation)
	if preferences.Language != "" && len(preferences.Language) > 10 {
		return infra_error.Validation(infra_error.ValidationTooLong, "preferences.language")
	}

	// Validate theme
	if preferences.Theme != "" {
		theme := strings.ToLower(preferences.Theme)
		if theme != "light" && theme != "dark" && theme != "auto" {
			return infra_error.Validation(infra_error.ValidationInvalidValue, "preferences.theme")
		}
	}

	return nil
}

// ValidateUserData validates complete user data (used for responses/read operations)
func ValidateUserData(userData *proto_auth.UserData) error {
	if userData == nil {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user_data")
	}

	// Validate required fields
	if userData.Id == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "id")
	}
	if userData.TenantId == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	if userData.Email == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "email")
	}
	if userData.Username == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "username")
	}

	// Validate email format
	if !isValidEmail(userData.Email) {
		return infra_error.Validation(infra_error.ValidationInvalidEmail, "email")
	}

	// Validate username format
	if !isValidUsername(userData.Username) {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "username")
	}

	// Validate status
	if userData.Status != "" && !model_auth.IsValidUserStatus(userData.Status) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "status")
	}

	// Validate nested structures
	if err := validateUserProfile(userData.Profile); err != nil {
		return err
	}
	if err := validateUserPreferences(userData.Preferences); err != nil {
		return err
	}

	return nil
}

// ValidateCreateUserData validates data for creating a new user
func ValidateCreateUserData(userData *proto_auth.CreateUserData) error {
	if userData == nil {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user_data")
	}

	// Validate required fields
	if userData.TenantId == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}
	if userData.Email == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "email")
	}
	if userData.Username == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "username")
	}
	if userData.PasswordHash == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "password_hash")
	}
	if userData.CreatedBy == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "created_by")
	}

	// Validate email format
	if !isValidEmail(userData.Email) {
		return infra_error.Validation(infra_error.ValidationInvalidEmail, "email")
	}

	// Validate username format
	if !isValidUsername(userData.Username) {
		return infra_error.Validation(infra_error.ValidationInvalidFormat, "username")
	}

	// Validate status if provided
	if userData.Status != "" && !model_auth.IsValidUserStatus(userData.Status) {
		return infra_error.Validation(infra_error.ValidationInvalidValue, "status")
	}

	// Validate nested structures
	if err := validateUserProfile(userData.Profile); err != nil {
		return err
	}
	if err := validateUserPreferences(userData.Preferences); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateUserData validates data for updating an existing user
func ValidateUpdateUserData(userData *proto_auth.UpdateUserData) error {
	if userData == nil {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "user_data")
	}

	// Validate required identifier fields
	if userData.Id == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "id")
	}
	if userData.TenantId == "" {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "tenant_id")
	}

	// Validate email format if provided
	if userData.Email != nil && *userData.Email != "" {
		if !isValidEmail(*userData.Email) {
			return infra_error.Validation(infra_error.ValidationInvalidEmail, "email")
		}
	}

	// Validate username format if provided
	if userData.Username != nil && *userData.Username != "" {
		if !isValidUsername(*userData.Username) {
			return infra_error.Validation(infra_error.ValidationInvalidFormat, "username")
		}
	}

	// Validate status if provided
	if userData.Status != nil && *userData.Status != "" {
		if !model_auth.IsValidUserStatus(*userData.Status) {
			return infra_error.Validation(infra_error.ValidationInvalidValue, "status")
		}
	}

	// Validate nested structures if provided
	if userData.Profile != nil {
		if err := validateUserProfile(userData.Profile); err != nil {
			return err
		}
	}
	if userData.Preferences != nil {
		if err := validateUserPreferences(userData.Preferences); err != nil {
			return err
		}
	}

	// Validate that at least one field is being updated
	hasUpdates := userData.Email != nil ||
		userData.Username != nil ||
		userData.Profile != nil ||
		userData.Roles != nil ||
		userData.Permissions != nil ||
		userData.Status != nil ||
		userData.EmailVerified != nil ||
		userData.PhoneVerified != nil ||
		userData.Preferences != nil

	if !hasUpdates {
		return infra_error.Validation(infra_error.ValidationRequiredFields, "at least one field to update")
	}

	return nil
}
