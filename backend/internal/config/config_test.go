package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"DATABASE_URL": "postgres://localhost/ledger_test", "APP_ENV": "dev",
		"HTTP_ADDR": ":8080", "LOG_LEVEL": "info", "HTTP_READ_TIMEOUT": "", "SHUTDOWN_TIMEOUT": "",
	} {
		t.Setenv(key, value)
	}
}

func TestLoadEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	setValidEnv(t)
	// Restore the original environment after testing an initially absent key.
	t.Setenv("SEED_PASSWORD", "")
	if err := os.Unsetenv("SEED_PASSWORD"); err != nil {
		t.Fatal(err)
	}
	content := "# local config\nexport SEED_PASSWORD='local # password' # comment\nAPP_ENV=prod\nDATABASE_URL=ignored\n"
	if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "dev" || cfg.DatabaseURL != "postgres://localhost/ledger_test" || os.Getenv("SEED_PASSWORD") != "local # password" {
		t.Fatal("file values were not loaded or existing environment was overwritten")
	}
}

func TestLoadRejectsMalformedEnvFile(t *testing.T) {
	for _, input := range []string{"INVALID@KEY=value", "KEY='unclosed"} {
		t.Run(input, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.WriteFile(".env", []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "load .env:") {
				t.Fatalf("expected env file error, got %v", err)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	setValidEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "dev" || cfg.Addr != ":8080" || cfg.LogLevel != "info" || cfg.ReadTimeout != 5*time.Second || cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadOverrides(t *testing.T) {
	setValidEnv(t)
	t.Setenv("APP_ENV", "prod")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("HTTP_READ_TIMEOUT", "2s")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")
	cfg, err := Load()
	if err != nil || cfg.Env != "prod" || cfg.LogLevel != "debug" || cfg.ReadTimeout != 2*time.Second || cfg.ShutdownTimeout != 3*time.Second {
		t.Fatalf("Load() = %+v, %v", cfg, err)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"DATABASE_URL", ""}, {"APP_ENV", "unknown"}, {"LOG_LEVEL", "verbose"},
		{"HTTP_READ_TIMEOUT", "bad"}, {"HTTP_READ_TIMEOUT", "0s"},
		{"SHUTDOWN_TIMEOUT", "-1s"},
	} {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv(tc.key, tc.value)
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("want error naming %s, got %v", tc.key, err)
			}
		})
	}
}

func TestLoadAggregatesErrors(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("APP_ENV", "unknown")
	t.Setenv("HTTP_READ_TIMEOUT", "bad")
	_, err := Load()
	if err == nil {
		t.Fatal("expected configuration errors")
	}
	for _, key := range []string{"DATABASE_URL", "APP_ENV", "HTTP_READ_TIMEOUT"} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("missing %s in %v", key, err)
		}
	}
}
