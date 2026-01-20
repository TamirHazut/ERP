package validator

import (
	"regexp"
	"strings"

	infra_error "erp.localhost/internal/infra/error"
	authv1 "erp.localhost/internal/infra/model/auth/v1"
)

var (
	// Email validation regex (basic RFC 5322 validation)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Username validation: 3-50 characters, alphanumeric, underscore, hyphen, dot
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._\-]{3,50}$`)

	// Phone validation: basic international format
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

func ValidateUser(u *authv1.User, createOperation bool) error {
	missingFields := []string{}
	if !createOperation {
		if u.Id == "" {
			missingFields = append(missingFields, "Id")
		}
	}
	if u.TenantId == "" {
		missingFields = append(missingFields, "TenantId")
	}
	if (u.Email == "" || !IsValidEmail(u.Email)) && (u.Username == "" || !IsValidUsername(u.Username)) {
		missingFields = append(missingFields, "Email or Username")
	}
	if u.PasswordHash == "" {
		missingFields = append(missingFields, "PasswordHash")
	}
	if u.CreatedBy == "" {
		missingFields = append(missingFields, "CreatedBy")
	}
	if u.Status == authv1.UserStatus_USER_STATUS_UNSPECIFIED {
		missingFields = append(missingFields, "Status")
	}
	if len(u.Roles) > 0 {
		for _, role := range u.Roles {
			if err := ValidateUserRole(role); err != nil {
				missingFields = append(missingFields, "Roles")
				break
			}
		}
	}
	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}

	return nil
}

func ValidateUserRole(u *authv1.UserRole) error {
	missingFields := []string{}

	if u.RoleId == "" {
		missingFields = append(missingFields, "Id")
	}
	if u.TenantId == "" {
		missingFields = append(missingFields, "TenantId")
	}
	if u.AssignedBy == "" {
		missingFields = append(missingFields, "AssignedBy")
	}

	if len(missingFields) > 0 {
		return infra_error.Validation(infra_error.ValidationRequiredFields, missingFields...)
	}

	return nil
}

func ValidateUserProfile(profile *authv1.UserProfile) error {
	if profile == nil {
		return nil // Profile is optional
	}

	// Validate phone if provided
	if profile.Phone != "" && !IsValidPhone(profile.Phone) {
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

func ValidateUserPreferences(preferences *authv1.UserPreferences) error {
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

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	email = strings.TrimSpace(email)
	if len(email) > 254 { // RFC 5321
		return false
	}
	return emailRegex.MatchString(email)
}

func IsValidUsername(username string) bool {
	if username == "" {
		return false
	}
	username = strings.TrimSpace(username)
	return usernameRegex.MatchString(username)
}

func IsValidPhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}
	phone = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", ""))
	return phoneRegex.MatchString(phone)
}
