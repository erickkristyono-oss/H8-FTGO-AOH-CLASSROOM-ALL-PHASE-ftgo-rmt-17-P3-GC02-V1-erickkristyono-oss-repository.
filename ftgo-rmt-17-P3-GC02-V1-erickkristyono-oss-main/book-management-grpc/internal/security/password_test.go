package security

import "testing"

func TestHashAndVerify(t *testing.T) {
	h := NewHasher()
	encoded, err := h.Hash("rahasia123")
	if err != nil {
		t.Fatalf("Hash error: %v", err)
	}
	if encoded == "rahasia123" {
		t.Fatal("hash tidak boleh sama dengan plaintext")
	}
	if !h.Verify("rahasia123", encoded) {
		t.Error("Verify seharusnya true untuk password yang benar")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	h := NewHasher()
	encoded, _ := h.Hash("benar")
	if h.Verify("salah", encoded) {
		t.Error("Verify seharusnya false untuk password yang salah")
	}
}

func TestHashProducesDifferentSalts(t *testing.T) {
	h := NewHasher()
	a, _ := h.Hash("samepass")
	b, _ := h.Hash("samepass")
	if a == b {
		t.Error("dua hash untuk password sama harus berbeda (salt acak)")
	}
}

func TestHashEmptyPassword(t *testing.T) {
	h := NewHasher()
	if _, err := h.Hash(""); err == nil {
		t.Error("Hash password kosong harus error")
	}
}
