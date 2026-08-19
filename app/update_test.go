package app

import "testing"

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"0.6.3", "0.6.2", true},
		{"v1.0.0", "0.9.9", true},
		{"0.6.2", "0.6.2", false},
		{"0.6.1", "0.6.2", false},
		{"1.2", "1.2.0", false},
		{"1.10.0", "1.9.9", true},
	}
	for _, test := range tests {
		if got := isNewerVersion(test.latest, test.current); got != test.want {
			t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", test.latest, test.current, got, test.want)
		}
	}
}
