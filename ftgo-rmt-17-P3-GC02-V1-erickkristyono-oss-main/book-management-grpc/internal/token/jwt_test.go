package token

import (
	"testing"
	"time"
)

func TestGenerateAndParse(t *testing.T) {
	m := NewManager("super-secret", time.Hour)
	tok, err := m.Generate("user-123", "alice")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	claims, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if claims.Sub != "user-123" {
		t.Errorf("Sub = %q, want user-123", claims.Sub)
	}
	if claims.Username != "alice" {
		t.Errorf("Username = %q, want alice", claims.Username)
	}
	if claims.Exp <= claims.Iat {
		t.Errorf("Exp (%d) harus lebih besar dari Iat (%d)", claims.Exp, claims.Iat)
	}
}

func TestParseExpiredToken(t *testing.T) {
	// TTL negatif => token langsung kedaluwarsa.
	m := NewManager("secret", -time.Minute)
	tok, err := m.Generate("u1", "bob")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if _, err := m.Parse(tok); err != ErrExpiredToken {
		t.Fatalf("Parse error = %v, want ErrExpiredToken", err)
	}
}

func TestParseTamperedToken(t *testing.T) {
	m := NewManager("secret", time.Hour)
	tok, _ := m.Generate("u1", "bob")
	// Ubah 1 karakter pada payload -> signature tidak cocok lagi.
	tampered := tok[:len(tok)-3] + "abc"
	if _, err := m.Parse(tampered); err != ErrInvalidToken {
		t.Fatalf("Parse error = %v, want ErrInvalidToken", err)
	}
}

func TestParseWrongSecret(t *testing.T) {
	m1 := NewManager("secret-A", time.Hour)
	m2 := NewManager("secret-B", time.Hour)
	tok, _ := m1.Generate("u1", "bob")
	if _, err := m2.Parse(tok); err != ErrInvalidToken {
		t.Fatalf("Parse dengan secret berbeda err = %v, want ErrInvalidToken", err)
	}
}
