package cloudflare

import (
	"errors"
	"net/http"
	"testing"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/shared"
	"github.com/sekai-labs/kumo/internal/core/ports"
)

func TestMapError(t *testing.T) {
	if MapError(nil) != nil {
		t.Errorf("expected nil error mapping to nil")
	}

	genericErr := errors.New("custom non-cf error")
	if MapError(genericErr) != genericErr {
		t.Errorf("expected non-cf error to pass through unchanged")
	}

	tests := []struct {
		name       string
		statusCode int
		header     http.Header
		expected   error
	}{
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			expected:   ports.ErrUnauthorized,
		},
		{
			name:       "403 Forbidden",
			statusCode: 403,
			expected:   ports.ErrForbidden,
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			expected:   ports.ErrNotFound,
		},
		{
			name:       "429 Rate Limited",
			statusCode: 429,
			header:     http.Header{"Retry-After": []string{"30"}},
			expected:   ports.ErrRateLimited,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &cf.Error{
				StatusCode: tt.statusCode,
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Header:     tt.header,
				},
				Errors: []shared.ErrorData{
					{Message: "API error message"},
				},
			}

			mapped := MapError(apiErr)
			if !errors.Is(mapped, tt.expected) {
				t.Fatalf("expected error %v, got %v", tt.expected, mapped)
			}

			if tt.statusCode == 429 {
				var rle *ports.RateLimitError
				if errors.As(mapped, &rle) {
					if rle.RetryAfter != 30*time.Second {
						t.Errorf("expected RetryAfter 30s, got %v", rle.RetryAfter)
					}
				} else {
					t.Errorf("expected mapped error to be RateLimitError")
				}
			}
		})
	}
}
