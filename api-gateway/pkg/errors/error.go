package errors

import (
	"fmt"
	"net/http"
	"runtime"
)

// StackTrace represents a stack trace
type StackTrace []uintptr

// AppError represents an application error
type AppError struct {
	Code       int        `json:"code"`
	Message    string     `json:"message"`
	Details    string     `json:"details,omitempty"`
	Internal   error      `json:"-"`
	StackTrace StackTrace `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap implements the unwrap interface for error wrapping
func (e *AppError) Unwrap() error {
	return e.Internal
}

// WithStack captures the stack trace
func (e *AppError) WithStack() *AppError {
	if e.StackTrace == nil {
		const depth = 32
		var pcs [depth]uintptr
		n := runtime.Callers(2, pcs[:])
		e.StackTrace = pcs[0:n]
	}

	return e
}

// Stack returns the stack trace as a string
func (e *AppError) Stack() string {
	if e.StackTrace == nil {
		return ""
	}

	frames := runtime.CallersFrames(e.StackTrace)
	var stack string
	for {
		frame, more := frames.Next()
		stack += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return stack
}

// NewAppError creates a new application error
func NewAppError(code int, message string, details string) *AppError {
	appError := &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}

	return appError.WithStack()
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *AppError {
	appError := &AppError{
		Code:     http.StatusInternalServerError,
		Message:  message,
		Internal: err,
	}

	return appError.WithStack()
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	appError := &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
	}

	return appError.WithStack()
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	appError := &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}

	return appError.WithStack()
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	appError := &AppError{
		Code:    http.StatusNotFound,
		Message: message,
	}

	return appError.WithStack()
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details string) *AppError {
	appError := &AppError{
		Code:    http.StatusUnprocessableEntity,
		Message: message,
		Details: details,
	}

	return appError.WithStack()
}

// ErrorResponse generates a standardized error response
func ErrorResponse(err error) (int, interface{}) {
	if appErr, ok := err.(*AppError); ok {
		response := map[string]interface{}{
			"error": appErr.Message,
			"code":  appErr.Code,
		}

		if appErr.Details != "" {
			response["details"] = appErr.Details
		}

		return appErr.Code, response
	}

	// Generic internal server error for unexpected errors
	return http.StatusInternalServerError, map[string]interface{}{
		"error": "Internal server error",
		"code":  http.StatusInternalServerError,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError returns the AppError if the error is an AppError
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
