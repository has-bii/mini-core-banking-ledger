package identifier

import "testing"

func TestValidUUID(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"0195c1de-3b2a-7000-8000-000000000001", true},
		{"0195C1DE3B2A70008000000000000001", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"", false}, {"not-a-uuid", false},
		{"0195c1de-3b2a-7000-8000-00000000000z", false},
		{"0195c1de_3b2a-7000-8000-000000000001", false},
		{"-195c1de-3b2a-7000-8000-000000000001", false},
		{"0195c1de3b2a7000800000000000000", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			if got := ValidUUID(tc.value); got != tc.valid {
				t.Fatalf("ValidUUID(%q) = %v, want %v", tc.value, got, tc.valid)
			}
		})
	}
}
