// Package security menyediakan hashing password.
//
// Implementasi ini memakai KDF berbasis standard library (salt acak + iterasi
// SHA-256) sehingga tidak butuh dependency eksternal dan mudah diuji. Format
// encoded: "sha256$<iter>$<salt_b64>$<hash_b64>".
//
// Ini memakai bcrypt/argon2 (golang.org/x/crypto).
// Lihat komentar VarianBcrypt di bawah untuk contoh penggantinya.
package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const defaultIterations = 120_000

// Hasher mengimplementasikan kontrak hashing yang dibutuhkan usecase.
type Hasher struct {
	Iterations int
}

// NewHasher membuat Hasher dengan jumlah iterasi default.
func NewHasher() *Hasher { return &Hasher{Iterations: defaultIterations} }

// derive menurunkan kunci dari password + salt dengan iterasi SHA-256.
func derive(password string, salt []byte, iter int) []byte {
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(password))
	sum := h.Sum(nil)
	for i := 1; i < iter; i++ {
		h.Reset()
		h.Write(sum)
		h.Write(salt)
		sum = h.Sum(nil)
	}
	return sum
}

// Hash menghasilkan encoded hash dari sebuah password plaintext.
func (h *Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}
	iter := h.Iterations
	if iter <= 0 {
		iter = defaultIterations
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	dk := derive(password, salt, iter)
	return fmt.Sprintf("sha256$%d$%s$%s",
		iter,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(dk),
	), nil
}

// Verify memeriksa apakah password cocok dengan encoded hash.
// Perbandingan memakai constant-time comparison agar aman dari timing attack.
func (h *Hasher) Verify(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got := derive(password, salt, iter)
	return subtle.ConstantTimeCompare(got, want) == 1
}

// --- Varian Bcrypt (produksi) ------------------------------------------------
// Untuk memakai bcrypt, tambahkan "golang.org/x/crypto/bcrypt" ke go.mod lalu:
//
//   func (h *Hasher) Hash(pw string) (string, error) {
//       b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
//       return string(b), err
//   }
//   func (h *Hasher) Verify(pw, encoded string) bool {
//       return bcrypt.CompareHashAndPassword([]byte(encoded), []byte(pw)) == nil
//   }
