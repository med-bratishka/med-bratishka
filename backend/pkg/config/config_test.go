package config

import "testing"

func TestValidateRequiresTwoFactorKey(t *testing.T) {
	cfg := defaultConfig()
	cfg.Auth.TwoFactorEncryptionKey = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing two factor key to fail validation")
	}
}

func TestDefaultConfigIsValid(t *testing.T) {
	if err := defaultConfig().Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
}
