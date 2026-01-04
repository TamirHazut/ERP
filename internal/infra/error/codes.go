package error

// ErrorDef defines an error code with its default message and category
type ErrorDef struct {
	Code     string
	Message  string
	Category ErrorCategory
}

// ============================================================================
// AUTH ERRORS
// ============================================================================

var (
	// Authentication errors
	AuthInvalidCredentials = ErrorDef{
		Code:     "AUTH_INVALID_CREDENTIALS",
		Message:  "Invalid email or password",
		Category: CategoryAuth,
	}
	AuthTokenExpired = ErrorDef{
		Code:     "AUTH_TOKEN_EXPIRED",
		Message:  "Your session has expired. Please log in again",
		Category: CategoryAuth,
	}
	AuthTokenRevoked = ErrorDef{
		Code:     "AUTH_TOKEN_REVOKED",
		Message:  "Your session has been revoked. Please log in again",
		Category: CategoryAuth,
	}
	AuthTokenInvalid = ErrorDef{
		Code:     "AUTH_TOKEN_INVALID",
		Message:  "Invalid authentication token",
		Category: CategoryAuth,
	}
	AuthTokenMissing = ErrorDef{
		Code:     "AUTH_TOKEN_MISSING",
		Message:  "Authentication token is required",
		Category: CategoryAuth,
	}
	AuthRefreshTokenExpired = ErrorDef{
		Code:     "AUTH_REFRESH_TOKEN_EXPIRED",
		Message:  "Your refresh token has expired. Please log in again",
		Category: CategoryAuth,
	}
	AuthRefreshTokenInvalid = ErrorDef{
		Code:     "AUTH_REFRESH_TOKEN_INVALID",
		Message:  "Invalid refresh token",
		Category: CategoryAuth,
	}

	// Authorization errors
	AuthPermissionDenied = ErrorDef{
		Code:     "AUTH_PERMISSION_DENIED",
		Message:  "You don't have permission to perform this action",
		Category: CategoryAuth,
	}
	AuthInsufficientRole = ErrorDef{
		Code:     "AUTH_INSUFFICIENT_ROLE",
		Message:  "Your role does not have access to this resource",
		Category: CategoryAuth,
	}
	AuthTenantAccessDenied = ErrorDef{
		Code:     "AUTH_TENANT_ACCESS_DENIED",
		Message:  "You don't have access to this organization",
		Category: CategoryAuth,
	}
	AuthSessionExpired = ErrorDef{
		Code:     "AUTH_SESSION_EXPIRED",
		Message:  "Your session has expired. Please log in again",
		Category: CategoryAuth,
	}
	AuthAccountLocked = ErrorDef{
		Code:     "AUTH_ACCOUNT_LOCKED",
		Message:  "Your account has been locked. Please contact support",
		Category: CategoryAuth,
	}
	AuthAccountDisabled = ErrorDef{
		Code:     "AUTH_ACCOUNT_DISABLED",
		Message:  "Your account has been disabled",
		Category: CategoryAuth,
	}
)

// ============================================================================
// VALIDATION ERRORS
// ============================================================================

var (
	ValidationTryToChangeRestrictedFields = ErrorDef{
		Code:     "VALIDATION_TRY_TO_CHANGE_RESTRICTED_FIELDS",
		Message:  "You are trying to change restricted fields",
		Category: CategoryValidation,
	}
	ValidationRequiredFields = ErrorDef{
		Code:     "VALIDATION_REQUIRED_FIELDS",
		Message:  "These fields are required",
		Category: CategoryValidation,
	}
	ValidationInvalidFormat = ErrorDef{
		Code:     "VALIDATION_INVALID_FORMAT",
		Message:  "Invalid format",
		Category: CategoryValidation,
	}
	ValidationInvalidEmail = ErrorDef{
		Code:     "VALIDATION_INVALID_EMAIL",
		Message:  "Invalid email address",
		Category: CategoryValidation,
	}
	ValidationInvalidPhone = ErrorDef{
		Code:     "VALIDATION_INVALID_PHONE",
		Message:  "Invalid phone number",
		Category: CategoryValidation,
	}
	ValidationOutOfRange = ErrorDef{
		Code:     "VALIDATION_OUT_OF_RANGE",
		Message:  "Value is out of allowed range",
		Category: CategoryValidation,
	}
	ValidationTooShort = ErrorDef{
		Code:     "VALIDATION_TOO_SHORT",
		Message:  "Value is too short",
		Category: CategoryValidation,
	}
	ValidationTooLong = ErrorDef{
		Code:     "VALIDATION_TOO_LONG",
		Message:  "Value is too long",
		Category: CategoryValidation,
	}
	ValidationInvalidType = ErrorDef{
		Code:     "VALIDATION_INVALID_TYPE",
		Message:  "Invalid value type",
		Category: CategoryValidation,
	}
	ValidationPasswordTooWeak = ErrorDef{
		Code:     "VALIDATION_PASSWORD_TOO_WEAK",
		Message:  "Password does not meet security requirements",
		Category: CategoryValidation,
	}
	ValidationInvalidDate = ErrorDef{
		Code:     "VALIDATION_INVALID_DATE",
		Message:  "Invalid date format",
		Category: CategoryValidation,
	}
	ValidationInvalidID = ErrorDef{
		Code:     "VALIDATION_INVALID_ID",
		Message:  "Invalid identifier format",
		Category: CategoryValidation,
	}
	ValidationInvalidValue = ErrorDef{
		Code:     "VALIDATION_INVALID_VALUE",
		Message:  "Invalid value",
		Category: CategoryValidation,
	}
)

// ============================================================================
// NOT FOUND ERRORS
// ============================================================================

var (
	NotFoundUser = ErrorDef{
		Code:     "NOT_FOUND_USER",
		Message:  "User not found",
		Category: CategoryNotFound,
	}
	NotFoundTenant = ErrorDef{
		Code:     "NOT_FOUND_TENANT",
		Message:  "Organization not found",
		Category: CategoryNotFound,
	}
	NotFoundRole = ErrorDef{
		Code:     "NOT_FOUND_ROLE",
		Message:  "Role not found",
		Category: CategoryNotFound,
	}
	NotFoundPermission = ErrorDef{
		Code:     "NOT_FOUND_PERMISSION",
		Message:  "Permission not found",
		Category: CategoryNotFound,
	}
	NotFoundProduct = ErrorDef{
		Code:     "NOT_FOUND_PRODUCT",
		Message:  "Product not found",
		Category: CategoryNotFound,
	}
	NotFoundOrder = ErrorDef{
		Code:     "NOT_FOUND_ORDER",
		Message:  "Order not found",
		Category: CategoryNotFound,
	}
	NotFoundVendor = ErrorDef{
		Code:     "NOT_FOUND_VENDOR",
		Message:  "Vendor not found",
		Category: CategoryNotFound,
	}
	NotFoundInventory = ErrorDef{
		Code:     "NOT_FOUND_INVENTORY",
		Message:  "Inventory item not found",
		Category: CategoryNotFound,
	}
	NotFoundConfig = ErrorDef{
		Code:     "NOT_FOUND_CONFIG",
		Message:  "Configuration not found",
		Category: CategoryNotFound,
	}
	NotFoundSession = ErrorDef{
		Code:     "NOT_FOUND_SESSION",
		Message:  "Session not found",
		Category: CategoryNotFound,
	}
	NotFoundResource = ErrorDef{
		Code:     "NOT_FOUND_RESOURCE",
		Message:  "Resource not found",
		Category: CategoryNotFound,
	}
)

// ============================================================================
// CONFLICT ERRORS
// ============================================================================

var (
	ConflictDuplicateResource = ErrorDef{
		Code:     "CONFLICT_DUPLICATE_RESOURCE",
		Message:  "A resource with this identifier already exists",
		Category: CategoryConflict,
	}
	ConflictDuplicateEmail = ErrorDef{
		Code:     "CONFLICT_DUPLICATE_EMAIL",
		Message:  "An account with this email already exists",
		Category: CategoryConflict,
	}
	ConflictDuplicateUsername = ErrorDef{
		Code:     "CONFLICT_DUPLICATE_USERNAME",
		Message:  "This username is already taken",
		Category: CategoryConflict,
	}
	ConflictOrderExists = ErrorDef{
		Code:     "CONFLICT_ORDER_EXISTS",
		Message:  "An order with this reference already exists",
		Category: CategoryConflict,
	}
	ConflictProductExists = ErrorDef{
		Code:     "CONFLICT_PRODUCT_EXISTS",
		Message:  "A product with this identifier already exists",
		Category: CategoryConflict,
	}
	ConflictVendorExists = ErrorDef{
		Code:     "CONFLICT_VENDOR_EXISTS",
		Message:  "A vendor with this name already exists",
		Category: CategoryConflict,
	}
	ConflictTenantExists = ErrorDef{
		Code:     "CONFLICT_TENANT_EXISTS",
		Message:  "An organization with this name already exists",
		Category: CategoryConflict,
	}
	ConflictResourceModified = ErrorDef{
		Code:     "CONFLICT_RESOURCE_MODIFIED",
		Message:  "The resource was modified by another user. Please refresh and try again",
		Category: CategoryConflict,
	}
)

// ============================================================================
// BUSINESS ERRORS
// ============================================================================

var (
	BusinessInsufficientStock = ErrorDef{
		Code:     "BUSINESS_INSUFFICIENT_STOCK",
		Message:  "Insufficient stock available",
		Category: CategoryBusiness,
	}
	BusinessOrderCancelled = ErrorDef{
		Code:     "BUSINESS_ORDER_CANCELLED",
		Message:  "This order has been cancelled",
		Category: CategoryBusiness,
	}
	BusinessOrderCompleted = ErrorDef{
		Code:     "BUSINESS_ORDER_COMPLETED",
		Message:  "This order has already been completed",
		Category: CategoryBusiness,
	}
	BusinessOrderCannotCancel = ErrorDef{
		Code:     "BUSINESS_ORDER_CANNOT_CANCEL",
		Message:  "This order cannot be cancelled in its current state",
		Category: CategoryBusiness,
	}
	BusinessInvalidOrderStatus = ErrorDef{
		Code:     "BUSINESS_INVALID_ORDER_STATUS",
		Message:  "Invalid order status transition",
		Category: CategoryBusiness,
	}
	BusinessVendorInactive = ErrorDef{
		Code:     "BUSINESS_VENDOR_INACTIVE",
		Message:  "This vendor is currently inactive",
		Category: CategoryBusiness,
	}
	BusinessProductInactive = ErrorDef{
		Code:     "BUSINESS_PRODUCT_INACTIVE",
		Message:  "This product is currently inactive",
		Category: CategoryBusiness,
	}
	BusinessLimitExceeded = ErrorDef{
		Code:     "BUSINESS_LIMIT_EXCEEDED",
		Message:  "Operation limit exceeded",
		Category: CategoryBusiness,
	}
	BusinessFeatureDisabled = ErrorDef{
		Code:     "BUSINESS_FEATURE_DISABLED",
		Message:  "This feature is currently disabled",
		Category: CategoryBusiness,
	}
	BusinessInvalidOperation = ErrorDef{
		Code:     "BUSINESS_INVALID_OPERATION",
		Message:  "This operation is not allowed",
		Category: CategoryBusiness,
	}
)

// ============================================================================
// INTERNAL ERRORS
// ============================================================================

var (
	InternalDatabaseError = ErrorDef{
		Code:     "INTERNAL_DATABASE_ERROR",
		Message:  "A database error occurred. Please try again later",
		Category: CategoryInternal,
	}
	InternalInvalidArgument = ErrorDef{
		Code:     "INTERNAL_INVALID_ARGUMENT",
		Message:  "An invalid argument occurred. Please check the arguments and try again",
		Category: CategoryInternal,
	}
	InternalServiceUnavailable = ErrorDef{
		Code:     "INTERNAL_SERVICE_UNAVAILABLE",
		Message:  "Service is temporarily unavailable. Please try again later",
		Category: CategoryInternal,
	}
	InternalGRPCError = ErrorDef{
		Code:     "INTERNAL_GRPC_ERROR",
		Message:  "A gRPC error occurred. Please try again later",
		Category: CategoryInternal,
	}
	InternalUnexpectedError = ErrorDef{
		Code:     "INTERNAL_UNEXPECTED_ERROR",
		Message:  "An unexpected error occurred. Please try again later",
		Category: CategoryInternal,
	}
	InternalCacheError = ErrorDef{
		Code:     "INTERNAL_CACHE_ERROR",
		Message:  "A cache error occurred. Please try again later",
		Category: CategoryInternal,
	}
	InternalConfigError = ErrorDef{
		Code:     "INTERNAL_CONFIG_ERROR",
		Message:  "A configuration error occurred",
		Category: CategoryInternal,
	}
	InternalExternalServiceError = ErrorDef{
		Code:     "INTERNAL_EXTERNAL_SERVICE_ERROR",
		Message:  "An external service error occurred. Please try again later",
		Category: CategoryInternal,
	}
	InternalTimeout = ErrorDef{
		Code:     "INTERNAL_TIMEOUT",
		Message:  "The operation timed out. Please try again",
		Category: CategoryInternal,
	}
)
