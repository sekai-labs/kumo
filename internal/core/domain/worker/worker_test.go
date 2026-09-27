package worker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/worker"
)

func TestNewWorkerScript(t *testing.T) {
	now := time.Now()
	t.Run("valid script", func(t *testing.T) {
		ws, err := worker.NewWorkerScript("my-api", now, now, "bundled", true, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ID.String() != "my-api" {
			t.Errorf("expected ID my-api, got %s", ws.ID)
		}
		if ws.UsageModel != "bundled" {
			t.Errorf("expected usage model bundled, got %s", ws.UsageModel)
		}
		if !ws.Logpush {
			t.Errorf("expected logpush true")
		}
	})

	t.Run("default usage model", func(t *testing.T) {
		ws, err := worker.NewWorkerScript("my-api", now, now, "", false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.UsageModel != "standard" {
			t.Errorf("expected default standard, got %s", ws.UsageModel)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := worker.NewWorkerScript("   ", now, now, "bundled", false, false)
		if !errors.Is(err, worker.ErrEmptyScriptID) {
			t.Errorf("expected ErrEmptyScriptID, got %v", err)
		}
	})
}
