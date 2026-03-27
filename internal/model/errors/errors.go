// Package errors defines structured error types for common failure cases.
package errors

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for common failure cases
var (
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrTokenExpired     = errors.New("token expired")
)

// RateLimitedError is returned when the API rate limit is exceeded.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
}

// NotFoundError provides context about what resource was not found.
type NotFoundError struct {
	Resource string // e.g. "PR", "review", "thread", "comment"
	ID       string // identifier that was looked up
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s' not found", e.Resource, e.ID)
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}

// IsNotFound checks if an error is a not-found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsRateLimited checks if an error is a rate limit error
func IsRateLimited(err error) bool {
	var rateLimitErr *RateLimitedError
	return errors.As(err, &rateLimitErr)
}

// IsPermissionDenied checks if an error is a permission denied error
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

// IsTokenExpired checks if an error is a token expired error
func IsTokenExpired(err error) bool {
	return errors.Is(err, ErrTokenExpired)
}
