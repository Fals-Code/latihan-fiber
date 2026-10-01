package config

import (
	"testing"
	"time"
)

func TestJWTConfigRequiresEnvironmentAndParsesSeconds(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-secret")
	t.Setenv("JWT_EXPIRES_IN", "3600")
	secret, ttl, err := JWTConfig()
	if err != nil {
		t.Fatal(err)
	}
	if secret != "test-only-secret" || ttl != time.Hour {
		t.Fatalf("unexpected config: %q %s", secret, ttl)
	}
}

func TestJWTConfigRejectsMissingSecretAndInvalidExpiration(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_EXPIRES_IN", "3600")
	if _, _, err := JWTConfig(); err == nil {
		t.Fatal("expected missing secret error")
	}
	t.Setenv("JWT_SECRET", "test-only-secret")
	for _, value := range []string{"", "0", "-1", "invalid"} {
		t.Setenv("JWT_EXPIRES_IN", value)
		if _, _, err := JWTConfig(); err == nil {
			t.Fatalf("expected invalid expiration error for %q", value)
		}
	}
}
