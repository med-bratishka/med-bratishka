package totp

import (
	"testing"
	"time"
)

func TestValidateAcceptsCurrentAndSkew(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Unix(60, 0)
	code := hotp(secret, uint64(now.Unix()/defaultPeriod))

	if !Validate(secret, code, now, 0) {
		t.Fatal("expected current TOTP code to be valid")
	}
	if !Validate(secret, code, now.Add(defaultPeriod*time.Second), 1) {
		t.Fatal("expected previous-window TOTP code to be valid with skew")
	}
}

func TestValidateRejectsInvalidCode(t *testing.T) {
	if Validate("JBSWY3DPEHPK3PXP", "000000", time.Unix(60, 0), 1) {
		t.Fatal("expected invalid TOTP code to be rejected")
	}
}
