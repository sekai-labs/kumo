package pages_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sekai-labs/kumo/internal/core/domain/pages"
)

func TestNewPagesProject(t *testing.T) {
	now := time.Now()
	t.Run("valid pages project", func(t *testing.T) {
		proj, err := pages.NewPagesProject(
			"proj-1",
			"docs-site",
			"docs-site.pages.dev",
			"production",
			[]string{"docs.example.com"},
			now,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proj.ID.String() != "proj-1" {
			t.Errorf("expected ID proj-1, got %s", proj.ID)
		}
		if proj.Name != "docs-site" {
			t.Errorf("expected Name docs-site, got %s", proj.Name)
		}
		if proj.ProductionBranch != "production" {
			t.Errorf("expected branch production, got %s", proj.ProductionBranch)
		}
		if len(proj.Domains) != 1 || proj.Domains[0] != "docs.example.com" {
			t.Errorf("unexpected domains: %v", proj.Domains)
		}
	})

	t.Run("default production branch", func(t *testing.T) {
		proj, err := pages.NewPagesProject("proj-2", "blog", "blog.pages.dev", "", nil, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if proj.ProductionBranch != "main" {
			t.Errorf("expected default main, got %s", proj.ProductionBranch)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := pages.NewPagesProject("", "blog", "blog.pages.dev", "main", nil, now)
		if !errors.Is(err, pages.ErrEmptyProjectID) {
			t.Errorf("expected ErrEmptyProjectID, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := pages.NewPagesProject("proj-1", "   ", "blog.pages.dev", "main", nil, now)
		if !errors.Is(err, pages.ErrEmptyProjectName) {
			t.Errorf("expected ErrEmptyProjectName, got %v", err)
		}
	})
}
