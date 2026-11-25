//go:build integration

package testutil

import (
	platformredis "backend/internal/platform/redis"
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis returns a client scoped to a private key namespace. It fails the test
// on setup errors and removes only keys in that namespace at cleanup.
func Redis(t *testing.T) *redis.Client {
	t.Helper()
	_, values := testConfig(t)
	url := strings.TrimSpace(values["REDIS_URL"])
	if url == "" {
		t.Fatal("backend/.env.test must contain REDIS_URL for a test Redis instance")
	}
	admin, err := platformredis.Open(t.Context(), url)
	if err != nil {
		t.Fatalf("open test Redis: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close() })

	prefix := "test_" + strings.ToLower(rand.Text()) + ":"
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var keys []string
		iter := admin.Scan(ctx, 0, prefix+"*", 100).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			t.Errorf("scan test Redis namespace: %v", err)
			return
		}
		if len(keys) > 0 {
			if err := admin.Del(ctx, keys...).Err(); err != nil {
				t.Errorf("delete test Redis namespace: %v", err)
			}
		}
	})

	options := *admin.Options()
	client := redis.NewClient(&options)
	client.AddHook(redisNamespace{prefix: prefix})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// Namespace commands without replacing Redis behavior. Reject unsupported
// commands rather than accidentally bypassing isolation.
type redisNamespace struct{ prefix string }

// DialHook leaves dialing unchanged because namespace isolation applies to commands.
func (h redisNamespace) DialHook(next redis.DialHook) redis.DialHook { return next }

// ProcessHook wraps next to scope each command to the private namespace.
func (h redisNamespace) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if err := h.namespace(cmd); err != nil {
			return err
		}
		return next(ctx, cmd)
	}
}

// namespace rewrites key arguments in place and rejects commands that cannot
// be safely scoped to the private namespace.
func (h redisNamespace) namespace(cmd redis.Cmder) error {
	args := cmd.Args()
	switch cmd.Name() {
	case "get", "set", "pttl", "persist", "expireat", "exists":
		args[1] = h.prefix + args[1].(string)
	case "del":
		for i := 1; i < len(args); i++ {
			args[i] = h.prefix + args[i].(string)
		}
	case "scan":
		matched := false
		for i := 2; i+1 < len(args); i++ {
			if args[i] == "match" {
				args[i+1] = h.prefix + args[i+1].(string)
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("test Redis scan requires a namespace pattern")
		}
	case "ping", "hello", "auth", "select":
		// Connection-local initialization has no keyspace effects.
	case "client":
		if len(args) < 2 || (args[1] != "setinfo" && args[1] != "setname") {
			return fmt.Errorf("test Redis client command is not namespace-safe")
		}
	default:
		return fmt.Errorf("test Redis command %q is not namespace-safe", cmd.Name())
	}
	return nil
}

// ProcessPipelineHook scopes every command and skips the entire pipeline if
// any command cannot be safely scoped.
func (h redisNamespace) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			if err := h.namespace(cmd); err != nil {
				return err
			}
		}
		return next(ctx, cmds)
	}
}
