package session_test

import (
	"backend/internal/modules/session"
	"errors"
	"strings"
	"testing"
)

func TestService_RejectsMalformedTokens(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, token string
	}{
		{"empty", ""},
		{"short", strings.Repeat("A", 42)},
		{"padded", strings.Repeat("A", 43) + "="},
		{"invalid alphabet", strings.Repeat("!", 43)},
		{"noncanonical trailing bits", strings.Repeat("A", 42) + "B"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// A nil store proves malformed input is rejected before persistence.
			svc := session.NewService(nil)
			got, err := svc.FindByToken(t.Context(), tt.token)
			if !errors.Is(err, session.ErrNotFound) || got != (session.Session{}) {
				t.Fatalf("FindByToken() = %+v, %v; want not found", got, err)
			}
			if err := svc.Revoke(t.Context(), tt.token); !errors.Is(err, session.ErrNotFound) {
				t.Fatalf("Revoke() = %v, want not found", err)
			}
		})
	}
}
