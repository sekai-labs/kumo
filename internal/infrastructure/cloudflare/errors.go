package cloudflare

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v4"
	"github.com/sekai-labs/kumo/internal/core/ports"
)

func MapError(err error) error {
	if err == nil {
		return nil
	}

	var cfErr *cf.Error
	if errors.As(err, &cfErr) {
		switch cfErr.StatusCode {
		case http.StatusUnauthorized:
			return ports.ErrUnauthorized
		case http.StatusForbidden:
			return ports.ErrForbidden
		case http.StatusNotFound:
			return ports.ErrNotFound
		case http.StatusTooManyRequests:
			var retryAfter time.Duration
			if cfErr.Response != nil {
				retryHeader := cfErr.Response.Header.Get("Retry-After")
				if retryHeader != "" {
					if seconds, err := strconv.Atoi(retryHeader); err == nil && seconds > 0 {
						retryAfter = time.Duration(seconds) * time.Second
					}
				}
			}
			msg := "too many requests"
			if len(cfErr.Errors) > 0 && cfErr.Errors[0].Message != "" {
				msg = cfErr.Errors[0].Message
			}
			return ports.NewRateLimitError(retryAfter, msg)
		}
	}

	return err
}
