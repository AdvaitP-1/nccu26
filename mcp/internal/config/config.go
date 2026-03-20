// Package config centralises runtime configuration for the MCP server.
//
// Values are read from environment variables with sensible local-dev defaults.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all tunables for the MCP server process.
type Config struct {
	// ServerAddr is the host:port the MCP server listens on.
	ServerAddr string

	// BackendBaseURL is the root URL of the structural analysis backend
	// (e.g. "http://localhost:8000").
	BackendBaseURL string

	// BackendTimeout governs individual HTTP calls to the backend.
	BackendTimeout time.Duration

	// RiskThreshold is the maximum per-file risk score (0-100) that is
	// still considered safe for a micro-commit.  Anything above is blocked.
	RiskThreshold int

	// BlockOnCritical, when true, rejects any micro-commit whose overlap
	// set contains at least one "critical" severity entry.
	BlockOnCritical bool
}

// Load reads configuration from the environment.
func Load() Config {
	return Config{
		ServerAddr:      envStr("MCP_SERVER_ADDR", ":9090"),
		BackendBaseURL:  envStr("MCP_BACKEND_URL", "http://localhost:8000"),
		BackendTimeout:  envDuration("MCP_BACKEND_TIMEOUT", 10*time.Second),
		RiskThreshold:   envInt("MCP_RISK_THRESHOLD", 70),
		BlockOnCritical: envBool("MCP_BLOCK_ON_CRITICAL", true),
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
