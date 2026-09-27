package ports

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnauthorized = errors.New("unauthorized: invalid or expired Cloudflare API token")
	ErrForbidden    = errors.New("forbidden: insufficient permissions for this resource")
	ErrNotFound     = errors.New("resource not found")
)

type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limit exceeded: retry after %s (%s)", e.RetryAfter, e.Message)
	}
	return fmt.Sprintf("rate limit exceeded: %s", e.Message)
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited
}

var ErrRateLimited = errors.New("rate limit exceeded")

func NewRateLimitError(retryAfter time.Duration, msg string) error {
	if msg == "" {
		msg = "too many requests"
	}
	return &RateLimitError{
		RetryAfter: retryAfter,
		Message:    msg,
	}
}
