// Package security provides secure error handling that prevents information leakage
package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Error codes for client responses (no sensitive info)
const (
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeRateLimited        = "RATE_LIMITED"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "TIMEOUT"
)

// User-facing error messages (safe to expose)
var userFacingMessages = map[string]string{
	ErrCodeBadRequest:         "The request was malformed or invalid.",
	ErrCodeUnauthorized:       "Authentication is required.",
	ErrCodeForbidden:          "You don't have permission to access this resource.",
	ErrCodeNotFound:           "The requested resource was not found.",
	ErrCodeConflict:           "The request conflicts with the current state.",
	ErrCodeRateLimited:        "Too many requests. Please try again later.",
	ErrCodeValidation:         "The provided data is invalid.",
	ErrCodeInternal:           "An unexpected error occurred. Please try again.",
	ErrCodeServiceUnavailable: "The service is temporarily unavailable.",
	ErrCodeTimeout:            "The request timed out. Please try again.",
}

// SecureError wraps errors with safe public messages and internal details
type SecureError struct {
	// Public (safe to expose to clients)
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`

	// Internal (never expose to clients)
	internalError error
	stackTrace    string
	timestamp     time.Time
	component     string
	userID        string
	action        string
}

// Error implements the error interface
func (e *SecureError) Error() string {
	return e.Message
}

// Unwrap returns the internal error for error chain handling
func (e *SecureError) Unwrap() error {
	return e.internalError
}

// InternalError returns the internal error (for logging only)
func (e *SecureError) InternalError() error {
	return e.internalError
}

// StackTrace returns the stack trace (for logging only)
func (e *SecureError) StackTrace() string {
	return e.stackTrace
}

// LogDetails returns details safe for logging (no client exposure)
func (e *SecureError) LogDetails() map[string]interface{} {
	details := map[string]interface{}{
		"code":       e.Code,
		"request_id": e.RequestID,
		"timestamp":  e.timestamp,
		"component":  e.component,
	}

	if e.internalError != nil {
		details["internal_error"] = e.internalError.Error()
	}

	if e.userID != "" {
		details["user_id"] = e.userID
	}

	if e.action != "" {
		details["action"] = e.action
	}

	return details
}

// ClientResponse returns a safe response for clients
func (e *SecureError) ClientResponse() map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":       e.Code,
			"message":    e.Message,
			"request_id": e.RequestID,
		},
	}
}

// HTTPStatus returns the appropriate HTTP status code
func (e *SecureError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeBadRequest, ErrCodeValidation:
		return 400
	case ErrCodeUnauthorized:
		return 401
	case ErrCodeForbidden:
		return 403
	case ErrCodeNotFound:
		return 404
	case ErrCodeConflict:
		return 409
	case ErrCodeRateLimited:
		return 429
	case ErrCodeServiceUnavailable:
		return 503
	case ErrCodeTimeout:
		return 504
	default:
		return 500
	}
}

// SecureErrorBuilder builds secure errors with a fluent interface
type SecureErrorBuilder struct {
	error *SecureError
}

// NewSecureError creates a new secure error builder
func NewSecureError(code string) *SecureErrorBuilder {
	requestID := generateRequestID()

	message, ok := userFacingMessages[code]
	if !ok {
		message = userFacingMessages[ErrCodeInternal]
	}

	return &SecureErrorBuilder{
		error: &SecureError{
			Code:      code,
			Message:   message,
			RequestID: requestID,
			timestamp: time.Now().UTC(),
		},
	}
}

// WithInternalError adds the internal error
func (b *SecureErrorBuilder) WithInternalError(err error) *SecureErrorBuilder {
	b.error.internalError = err
	return b
}

// WithStackTrace captures the current stack trace
func (b *SecureErrorBuilder) WithStackTrace() *SecureErrorBuilder {
	b.error.stackTrace = captureStackTrace(3) // Skip builder frames
	return b
}

// WithComponent adds the component name
func (b *SecureErrorBuilder) WithComponent(component string) *SecureErrorBuilder {
	b.error.component = component
	return b
}

// WithUserID adds the user ID (for logging only)
func (b *SecureErrorBuilder) WithUserID(userID string) *SecureErrorBuilder {
	b.error.userID = userID
	return b
}

// WithAction adds the action being performed
func (b *SecureErrorBuilder) WithAction(action string) *SecureErrorBuilder {
	b.error.action = action
	return b
}

// WithMessage overrides the default message (must be safe for clients)
func (b *SecureErrorBuilder) WithMessage(message string) *SecureErrorBuilder {
	// Sanitize the message
	b.error.Message = sanitizeErrorMessage(message)
	return b
}

// WithRequestID sets a specific request ID
func (b *SecureErrorBuilder) WithRequestID(requestID string) *SecureErrorBuilder {
	b.error.RequestID = requestID
	return b
}

// Build returns the constructed SecureError
func (b *SecureErrorBuilder) Build() *SecureError {
	return b.error
}

// Helper functions

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func captureStackTrace(skip int) string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()

		// Skip runtime and standard library
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}

		builder.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))

		if !more {
			break
		}
	}

	return builder.String()
}

func sanitizeErrorMessage(message string) string {
	// Remove potential sensitive patterns
	patterns := []string{
		// File paths
		`(?:/[a-zA-Z0-9._-]+)+`,
		// SQL errors
		`(?i)sql|query|database|table|column`,
		// Stack traces
		`(?i)panic|goroutine|runtime`,
		// Credentials
		`(?i)password|secret|key|token|credential`,
		// IP addresses
		`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`,
	}

	result := message
	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(result), strings.ToLower(pattern)) {
			// If sensitive pattern detected, use generic message
			return userFacingMessages[ErrCodeInternal]
		}
	}

	// Truncate long messages
	if len(result) > 200 {
		result = result[:200] + "..."
	}

	return result
}

// Common error constructors

// BadRequest creates a bad request error
func BadRequest(internalErr error) *SecureError {
	return NewSecureError(ErrCodeBadRequest).
		WithInternalError(internalErr).
		WithStackTrace().
		Build()
}

// Unauthorized creates an unauthorized error
func Unauthorized(internalErr error) *SecureError {
	return NewSecureError(ErrCodeUnauthorized).
		WithInternalError(internalErr).
		Build()
}

// Forbidden creates a forbidden error
func Forbidden(internalErr error) *SecureError {
	return NewSecureError(ErrCodeForbidden).
		WithInternalError(internalErr).
		Build()
}

// NotFound creates a not found error
func NotFound(internalErr error) *SecureError {
	return NewSecureError(ErrCodeNotFound).
		WithInternalError(internalErr).
		Build()
}

// Conflict creates a conflict error
func Conflict(internalErr error) *SecureError {
	return NewSecureError(ErrCodeConflict).
		WithInternalError(internalErr).
		WithStackTrace().
		Build()
}

// RateLimited creates a rate limited error
func RateLimited(retryAfter time.Duration) *SecureError {
	err := NewSecureError(ErrCodeRateLimited).Build()
	err.Message = fmt.Sprintf("Too many requests. Please try again in %s.", retryAfter.Round(time.Second))
	return err
}

// ValidationError creates a validation error with safe field info
func ValidationError(field, issue string) *SecureError {
	err := NewSecureError(ErrCodeValidation).Build()
	// Only include field name if it's a known safe field
	safeFields := map[string]bool{
		"email": true, "username": true, "password": true,
		"country": true, "protocol": true, "node_id": true,
	}
	if safeFields[field] {
		err.Message = fmt.Sprintf("Invalid %s: %s", field, issue)
	}
	return err
}

// InternalError creates an internal server error
func InternalError(internalErr error) *SecureError {
	return NewSecureError(ErrCodeInternal).
		WithInternalError(internalErr).
		WithStackTrace().
		Build()
}

// ServiceUnavailable creates a service unavailable error
func ServiceUnavailable(service string) *SecureError {
	return NewSecureError(ErrCodeServiceUnavailable).
		WithInternalError(fmt.Errorf("service unavailable: %s", service)).
		Build()
}

// Timeout creates a timeout error
func Timeout(operation string) *SecureError {
	return NewSecureError(ErrCodeTimeout).
		WithInternalError(fmt.Errorf("timeout: %s", operation)).
		Build()
}

// ErrorRegistry tracks errors for monitoring and alerting
type ErrorRegistry struct {
	errors       map[string]int64
	recentErrors []*SecureError
	maxRecent    int
	mu           sync.RWMutex
}

// NewErrorRegistry creates a new error registry
func NewErrorRegistry(maxRecent int) *ErrorRegistry {
	if maxRecent <= 0 {
		maxRecent = 100
	}
	return &ErrorRegistry{
		errors:       make(map[string]int64),
		recentErrors: make([]*SecureError, 0, maxRecent),
		maxRecent:    maxRecent,
	}
}

// Record records an error
func (r *ErrorRegistry) Record(err *SecureError) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors[err.Code]++

	r.recentErrors = append(r.recentErrors, err)
	if len(r.recentErrors) > r.maxRecent {
		r.recentErrors = r.recentErrors[1:]
	}
}

// Stats returns error statistics
func (r *ErrorRegistry) Stats() map[string]int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]int64, len(r.errors))
	for k, v := range r.errors {
		result[k] = v
	}
	return result
}

// RecentErrors returns recent errors (for debugging)
func (r *ErrorRegistry) RecentErrors() []*SecureError {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*SecureError, len(r.recentErrors))
	copy(result, r.recentErrors)
	return result
}

// PanicRecovery provides safe panic recovery with stack trace sanitization
type PanicRecovery struct {
	component string
	registry  *ErrorRegistry
}

// NewPanicRecovery creates a new panic recovery handler
func NewPanicRecovery(component string, registry *ErrorRegistry) *PanicRecovery {
	return &PanicRecovery{
		component: component,
		registry:  registry,
	}
}

// Recover handles a panic and returns a safe error
func (pr *PanicRecovery) Recover() *SecureError {
	if r := recover(); r != nil {
		var internalErr error
		switch v := r.(type) {
		case error:
			internalErr = v
		case string:
			internalErr = errors.New(v)
		default:
			internalErr = fmt.Errorf("panic: %v", v)
		}

		secErr := NewSecureError(ErrCodeInternal).
			WithInternalError(internalErr).
			WithStackTrace().
			WithComponent(pr.component).
			Build()

		if pr.registry != nil {
			pr.registry.Record(secErr)
		}

		return secErr
	}
	return nil
}

// WrapWithRecovery wraps a function with panic recovery
func WrapWithRecovery(fn func() error, component string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var internalErr error
			switch v := r.(type) {
			case error:
				internalErr = v
			case string:
				internalErr = errors.New(v)
			default:
				internalErr = fmt.Errorf("panic: %v", v)
			}

			err = NewSecureError(ErrCodeInternal).
				WithInternalError(internalErr).
				WithStackTrace().
				WithComponent(component).
				Build()
		}
	}()

	return fn()
}
