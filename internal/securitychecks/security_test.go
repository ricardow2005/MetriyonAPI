package securitychecks

import (
	"strings"
	"testing"

	"forge-api-client/internal/security"
)

func TestProtectorRoundTrip(t *testing.T) {
	protector, err := security.NewProtector(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := protector.Encrypt("very-secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "very-secret-token" || !strings.HasPrefix(encrypted, "forge:enc:") {
		t.Fatal("value was not encrypted")
	}
	plain, err := protector.Decrypt(encrypted)
	if err != nil || plain != "very-secret-token" {
		t.Fatalf("round trip failed: %q %v", plain, err)
	}
}

func TestSanitize(t *testing.T) {
	got := security.SanitizeText("Authorization: Bearer abc123 password=secret")
	if strings.Contains(got, "abc123") || strings.Contains(got, "secret") {
		t.Fatalf("sensitive data leaked: %s", got)
	}
}
