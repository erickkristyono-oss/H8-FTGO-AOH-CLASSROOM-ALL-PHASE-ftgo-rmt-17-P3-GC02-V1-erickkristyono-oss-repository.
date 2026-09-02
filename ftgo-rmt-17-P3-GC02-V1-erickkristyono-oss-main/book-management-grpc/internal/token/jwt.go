// Package token menyediakan pembuatan & verifikasi JSON Web Token (JWT)
// dengan algoritma HS256, diimplementasikan memakai standard library saja
// (crypto/hmac, crypto/sha256, encoding/base64, encoding/json).
//
// Token yang dihasilkan kompatibel dengan JWT standar (RFC 7519) sehingga bisa
// diverifikasi oleh tools lain seperti jwt.io.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	// ErrInvalidToken dikembalikan bila format/signature token tidak valid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken dikembalikan bila token sudah kedaluwarsa.
	ErrExpiredToken = errors.New("token has expired")
)

// Claims adalah payload JWT yang dipakai aplikasi.
type Claims struct {
	Sub      string `json:"sub"`      // user id
	Username string `json:"username"` // username
	Iat      int64  `json:"iat"`      // issued at (unix)
	Exp      int64  `json:"exp"`      // expiry (unix)
}

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Manager membuat & memverifikasi token dengan secret dan durasi tertentu.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager membuat Manager baru. ttl adalah masa berlaku token.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func (m *Manager) sign(signingInput string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(signingInput))
	return b64(mac.Sum(nil))
}

// Generate membuat token JWT baru untuk userID & username.
func (m *Manager) Generate(userID, username string) (string, error) {
	now := time.Now()
	h := header{Alg: "HS256", Typ: "JWT"}
	c := Claims{
		Sub:      userID,
		Username: username,
		Iat:      now.Unix(),
		Exp:      now.Add(m.ttl).Unix(),
	}
	hb, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	cb, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	signingInput := b64(hb) + "." + b64(cb)
	return signingInput + "." + m.sign(signingInput), nil
}

// Parse memverifikasi signature & expiry lalu mengembalikan Claims.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	expectedSig := m.sign(signingInput)
	// Bandingkan signature secara constant-time.
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(parts[2])) != 1 {
		return nil, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, ErrInvalidToken
	}
	if c.Exp > 0 && time.Now().Unix() >= c.Exp {
		return nil, ErrExpiredToken
	}
	return &c, nil
}
