package crypto

import "testing"

func TestSecretBoxEncryptDecryptRoundTrip(t *testing.T) {
	box, err := NewSecretBox("test-key")
	if err != nil {
		t.Fatalf("NewSecretBox returned error: %v", err)
	}

	ciphertext, err := box.EncryptString("TOTPSECRET", []byte("totp:1:patient"))
	if err != nil {
		t.Fatalf("EncryptString returned error: %v", err)
	}
	if ciphertext == "" || ciphertext == "TOTPSECRET" {
		t.Fatalf("unexpected ciphertext: %q", ciphertext)
	}

	plain, err := box.DecryptString(ciphertext, []byte("totp:1:patient"))
	if err != nil {
		t.Fatalf("DecryptString returned error: %v", err)
	}
	if plain != "TOTPSECRET" {
		t.Fatalf("expected plaintext TOTPSECRET, got %q", plain)
	}
}

func TestSecretBoxRejectsWrongAAD(t *testing.T) {
	box, err := NewSecretBox("test-key")
	if err != nil {
		t.Fatalf("NewSecretBox returned error: %v", err)
	}

	ciphertext, err := box.EncryptString("TOTPSECRET", []byte("totp:1:patient"))
	if err != nil {
		t.Fatalf("EncryptString returned error: %v", err)
	}

	if _, err := box.DecryptString(ciphertext, []byte("totp:2:patient")); err == nil {
		t.Fatal("expected decrypt with wrong AAD to fail")
	}
}

func TestSecretBoxRejectsEmptyKey(t *testing.T) {
	if _, err := NewSecretBox(""); err == nil {
		t.Fatal("expected empty key to fail")
	}
}
