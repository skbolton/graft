package hookhelpers

import "testing"

func TestOnCreate(t *testing.T) {
	cases := []struct {
		name   string
		oldSha string
		want   bool
		wantEr bool
	}{
		{"all zeros marks creation", "0000000000000000000000000000000000000000", true, false},
		{"real sha is a branch switch", "787cd8dda5e49dbe4f4676d6e1fbae77ca71ba96", false, false},
		{"short sha is not creation", "0000", false, false},
		{"garbage is not creation", "not-a-sha", false, false},
		{"empty is a usage error", "", false, true},
		{"whitespace is a usage error", "  ", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := OnCreate(tc.oldSha)
			if got != tc.want {
				t.Errorf("OnCreate(%q) = %v, want %v", tc.oldSha, got, tc.want)
			}
			if (err != nil) != tc.wantEr {
				t.Errorf("OnCreate(%q) error = %v, wantErr %v", tc.oldSha, err, tc.wantEr)
			}
		})
	}
}
