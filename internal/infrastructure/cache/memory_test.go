package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_BasicOperations(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	var result string
	found, err := c.Get(ctx, "key1", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Errorf("expected found to be false, got true")
	}

	err = c.Set(ctx, "key1", "hello world", time.Minute)
	if err != nil {
		t.Fatalf("set error: %v", err)
	}

	found, err = c.Get(ctx, "key1", &result)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if !found {
		t.Fatalf("expected found to be true")
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}

	err = c.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}

	found, err = c.Get(ctx, "key1", &result)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if found {
		t.Errorf("expected key1 to be deleted")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	err := c.Set(ctx, "quick", 42, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("set error: %v", err)
	}

	var val int
	found, err := c.Get(ctx, "quick", &val)
	if err != nil || !found || val != 42 {
		t.Fatalf("expected immediate get to succeed, got found=%v, val=%d, err=%v", found, val, err)
	}

	time.Sleep(35 * time.Millisecond)

	found, err = c.Get(ctx, "quick", &val)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if found {
		t.Errorf("expected item to be expired and not found")
	}
}

func TestMemoryCache_PurgeExpired(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "exp1", "val1", 10*time.Millisecond)
	_ = c.Set(ctx, "exp2", "val2", 10*time.Millisecond)
	_ = c.Set(ctx, "keep", "val3", 10*time.Minute)

	time.Sleep(20 * time.Millisecond)

	purged := c.PurgeExpired()
	if purged != 2 {
		t.Errorf("expected 2 purged items, got %d", purged)
	}

	var keepVal string
	found, err := c.Get(ctx, "keep", &keepVal)
	if err != nil || !found || keepVal != "val3" {
		t.Errorf("expected keepVal to remain, found=%v", found)
	}
}

func TestMemoryCache_Flush(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCache()

	_ = c.Set(ctx, "k1", 1, 0)
	_ = c.Set(ctx, "k2", 2, 0)

	if err := c.Flush(ctx); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	var val int
	found, _ := c.Get(ctx, "k1", &val)
	if found {
		t.Errorf("expected k1 flushed")
	}
	found, _ = c.Get(ctx, "k2", &val)
	if found {
		t.Errorf("expected k2 flushed")
	}
}

func TestMemoryCache_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewMemoryCache()
	var val string
	_, err := c.Get(ctx, "key", &val)
	if err == nil {
		t.Errorf("expected context canceled error")
	}

	err = c.Set(ctx, "key", "v", 0)
	if err == nil {
		t.Errorf("expected context canceled error")
	}
}
