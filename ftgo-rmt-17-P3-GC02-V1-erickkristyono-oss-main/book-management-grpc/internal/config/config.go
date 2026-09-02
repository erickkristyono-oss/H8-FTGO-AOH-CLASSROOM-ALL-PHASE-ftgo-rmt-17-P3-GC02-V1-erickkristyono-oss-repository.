// Package config memuat konfigurasi aplikasi dari environment variable.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config menampung seluruh konfigurasi runtime.
type Config struct {
	GRPCPort         string        // port server gRPC
	HTTPPort         string        // port REST gateway + Swagger
	JWTSecret        string        // secret penandatangan JWT
	JWTTTL           time.Duration // masa berlaku token
	DBDriver         string        // "memory" atau "postgres"
	DatabaseURL      string        // DSN Postgres (bila DBDriver=postgres)
	SchedInterval    time.Duration // interval job scheduler
	OverdueBorrowTTL time.Duration // default durasi peminjaman
}

// Load membaca konfigurasi dari environment.
func Load() Config {
	return Config{
		GRPCPort:         getEnv("GRPC_PORT", "50051"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		JWTSecret:        getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTTTL:           getDuration("JWT_TTL", 24*time.Hour),
		DBDriver:         getEnv("DB_DRIVER", "memory"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		SchedInterval:    getDuration("SCHED_INTERVAL", time.Hour),
		OverdueBorrowTTL: getDuration("BORROW_DURATION", 7*24*time.Hour),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	// Dukung format durasi Go ("1h", "30m")
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return def
}
