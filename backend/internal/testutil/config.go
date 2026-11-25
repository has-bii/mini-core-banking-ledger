//go:build integration

package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

// testConfig returns the backend root and values read directly from .env.test.
// It fails the test if it cannot find the root or read the file.
func testConfig(t *testing.T) (string, map[string]string) {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("cannot locate backend/go.mod")
		}
		root = parent
	}
	values, err := godotenv.Read(filepath.Join(root, ".env.test"))
	if err != nil {
		t.Fatalf("read backend/.env.test: %v", err)
	}
	return root, values
}
