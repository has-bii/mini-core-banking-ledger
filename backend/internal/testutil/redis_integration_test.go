//go:build integration

package testutil_test

import (
	"backend/internal/testutil"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedis_NamespaceIsolation(t *testing.T) {
	t.Parallel()
	first := testutil.Redis(t)
	second := testutil.Redis(t)
	const key = "session:shared-test-key"
	if err := first.Set(t.Context(), key, "first", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := second.Set(t.Context(), key, "second", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if got, err := first.Get(t.Context(), key).Result(); err != nil || got != "first" {
		t.Fatalf("first namespace = %q, %v; want first", got, err)
	}
	if err := second.Del(t.Context(), key).Err(); err != nil {
		t.Fatal(err)
	}
	if got, err := first.Get(t.Context(), key).Result(); err != nil || got != "first" {
		t.Fatalf("deleting second namespace changed first: %q, %v", got, err)
	}
	if _, err := second.Get(t.Context(), key).Result(); !errors.Is(err, redis.Nil) {
		t.Fatalf("second namespace after delete = %v, want redis.Nil", err)
	}
	if err := first.FlushDB(t.Context()).Err(); err == nil {
		t.Fatal("shared database flush must be rejected")
	}
}
