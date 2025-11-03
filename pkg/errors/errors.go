package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with additional context
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Internal   error                  `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Internal
}

// WithInternal adds an internal error
func (e *AppError) WithInternal(err error) *AppError {
	e.Internal = err
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// New creates a new AppError
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// Common error codes
const (
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeRateLimit          = "RATE_LIMIT_EXCEEDED"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       = "TOKEN_INVALID"
	ErrCodeUserExists         = "USER_ALREADY_EXISTS"
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeNodeNotFound       = "NODE_NOT_FOUND"
	ErrCodeNodeUnavailable    = "NODE_UNAVAILABLE"
	ErrCodeSessionNotFound    = "SESSION_NOT_FOUND"
	ErrCodeQuotaExceeded      = "QUOTA_EXCEEDED"
	ErrCodeConfigInvalid      = "INVALID_CONFIGURATION"
)

// Predefined errors
var (
	// General errors
	ErrInternal = New(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest = New(ErrCodeBadRequest, "Bad request", http.StatusBadRequest)
	ErrUnauthorized = New(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden = New(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrNotFound = New(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrConflict = New(ErrCodeConflict, "Resource conflict", http.StatusConflict)
	ErrRateLimit = New(ErrCodeRateLimit, "Rate limit exceeded", http.StatusTooManyRequests)

	// Validation errors
	ErrValidation = New(ErrCodeValidation, "Validation failed", http.StatusBadRequest)
	ErrInvalidInput = New(ErrCodeBadRequest, "Invalid input", http.StatusBadRequest)

	// Database errors
	ErrDatabase = New(ErrCodeDatabaseError, "Database error", http.StatusInternalServerError)

	// Auth errors
	ErrInvalidCredentials = New(ErrCodeInvalidCredentials, "Invalid credentials", http.StatusUnauthorized)
	ErrTokenExpired = New(ErrCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
	ErrTokenInvalid = New(ErrCodeTokenInvalid, "Invalid token", http.StatusUnauthorized)

	// User errors
	ErrUserExists = New(ErrCodeUserExists, "User already exists", http.StatusConflict)
	ErrUserNotFound = New(ErrCodeUserNotFound, "User not found", http.StatusNotFound)
	ErrUserInactive = New(ErrCodeUnauthorized, "User account is inactive", http.StatusUnauthorized)

	// Node errors
	ErrNodeNotFound = New(ErrCodeNodeNotFound, "VPN node not found", http.StatusNotFound)
	ErrNodeUnavailable = New(ErrCodeNodeUnavailable, "No available VPN nodes", http.StatusServiceUnavailable)
	ErrNodeAtCapacity = New(ErrCodeNodeUnavailable, "Node at maximum capacity", http.StatusServiceUnavailable)

	// Session errors
	ErrSessionNotFound = New(ErrCodeSessionNotFound, "Session not found", http.StatusNotFound)
	ErrSessionExpired = New(ErrCodeUnauthorized, "Session has expired", http.StatusUnauthorized)

	// Quota errors
	ErrQuotaExceeded = New(ErrCodeQuotaExceeded, "Quota exceeded", http.StatusForbidden)

	// Config errors
	ErrConfigInvalid = New(ErrCodeConfigInvalid, "Invalid configuration", http.StatusBadRequest)
)

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Is checks if an error is a specific type
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetAppError extracts an AppError from an error chain
func GetAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Return internal error as default
	return ErrInternal.WithInternal(err)
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationError creates a validation error with field details
func NewValidationError(errors []ValidationError) *AppError {
	details := make(map[string]interface{})
	details["fields"] = errors
	return ErrValidation.WithDetails(details)
}
