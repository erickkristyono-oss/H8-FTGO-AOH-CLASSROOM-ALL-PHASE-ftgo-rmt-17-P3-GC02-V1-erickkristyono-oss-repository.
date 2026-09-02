// Package idgen menghasilkan UUID versi 4 (RFC 4122) memakai standard library.
// Dibuat agar layer inti tidak bergantung pada dependency eksternal.
package idgen

import (
	"crypto/rand"
	"fmt"
)

// NewUUID mengembalikan string UUID v4 acak, contoh:
// "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d".
func NewUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("idgen: gagal membaca random: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // versi 4
	b[8] = (b[8] & 0x3f) | 0x80 // varian 10xx
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
