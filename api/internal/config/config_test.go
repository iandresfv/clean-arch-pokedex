package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error for a clean environment: %v", err)
	}

	if got, want := cfg.Server.Addr(), "localhost:8080"; got != want {
		t.Errorf("Server.Addr() = %q, want %q", got, want)
	}
	if got, want := cfg.Log.Level, slog.LevelInfo; got != want {
		t.Errorf("Log.Level = %v, want %v", got, want)
	}
	if got, want := cfg.Env, EnvDevelopment; got != want {
		t.Errorf("Env = %q, want %q", got, want)
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true for default environment")
	}
}

func TestLoadInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "non-numeric port",
			env:     map[string]string{"SERVER_PORT": "abc"},
			wantErr: "SERVER_PORT must be an integer",
		},
		{
			name:    "port out of range",
			env:     map[string]string{"SERVER_PORT": "70000"},
			wantErr: "SERVER_PORT must be between 1 and 65535",
		},
		{
			name:    "unknown log level",
			env:     map[string]string{"LOG_LEVEL": "verbose"},
			wantErr: "LOG_LEVEL must be one of debug, info, warn, error",
		},
		{
			name:    "malformed duration",
			env:     map[string]string{"SERVER_READ_TIMEOUT": "15"},
			wantErr: "SERVER_READ_TIMEOUT must be a duration",
		},
		{
			name:    "negative duration",
			env:     map[string]string{"SERVER_READ_TIMEOUT": "-5s"},
			wantErr: "SERVER_READ_TIMEOUT must be positive",
		},
		{
			name:    "non-boolean flag",
			env:     map[string]string{"TLS_ENABLED": "yes please"},
			wantErr: "TLS_ENABLED must be a boolean",
		},
		{
			name:    "invalid sslmode",
			env:     map[string]string{"DB_SSLMODE": "maybe"},
			wantErr: "is not a valid libpq sslmode",
		},
		{
			name:    "min conns above max conns",
			env:     map[string]string{"DB_MAX_CONNS": "5", "DB_MIN_CONNS": "10"},
			wantErr: "DB_MIN_CONNS must be between 0 and DB_MAX_CONNS",
		},
		{
			name:    "tls enabled without certificate paths",
			env:     map[string]string{"TLS_ENABLED": "true"},
			wantErr: "TLS_CERT_PATH and TLS_KEY_PATH are required",
		},
		{
			name: "wildcard origin with credentials",
			env: map[string]string{
				"CORS_ALLOWED_ORIGINS":   "*",
				"CORS_ALLOW_CREDENTIALS": "true",
			},
			wantErr: `cannot contain "*"`,
		},
		{
			name:    "invalid trusted proxy cidr",
			env:     map[string]string{"TRUSTED_PROXIES": "10.0.0.1"},
			wantErr: "is not valid CIDR notation",
		},
		{
			name:    "production without database password",
			env:     map[string]string{"APP_ENV": "production", "CORS_ALLOWED_ORIGINS": "https://pokedex.app"},
			wantErr: "DB_PASSWORD is required",
		},
		{
			name: "production with loopback origin",
			env: map[string]string{
				"APP_ENV":              "production",
				"DB_PASSWORD":          "s3cret",
				"CORS_ALLOWED_ORIGINS": "http://localhost:5173",
			},
			wantErr: "loopback origin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			_, err := Load()
			if err == nil {
				t.Fatal("Load() succeeded, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// TestLoadReportsAllErrors verifies that configuration problems are accumulated
// rather than returned one per restart.
func TestLoadReportsAllErrors(t *testing.T) {
	clearEnv(t)
	t.Setenv("SERVER_PORT", "abc")
	t.Setenv("LOG_LEVEL", "verbose")
	t.Setenv("DB_SSLMODE", "maybe")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() succeeded, want error")
	}
	for _, want := range []string{"SERVER_PORT", "LOG_LEVEL", "DB_SSLMODE"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s; errors should accumulate", err, want)
		}
	}
}

func TestLoadValidOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("SERVER_HOST", "0.0.0.0")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example , https://b.example")
	t.Setenv("DB_MAX_CONN_LIFETIME", "45m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if got, want := cfg.Server.Addr(), "0.0.0.0:9000"; got != want {
		t.Errorf("Server.Addr() = %q, want %q", got, want)
	}
	if got, want := cfg.Log.Level, slog.LevelDebug; got != want {
		t.Errorf("Log.Level = %v, want %v", got, want)
	}
	if got, want := cfg.Database.MaxConnLifetime, 45*time.Minute; got != want {
		t.Errorf("MaxConnLifetime = %v, want %v", got, want)
	}
	want := []string{"https://a.example", "https://b.example"}
	if got := cfg.CORS.AllowedOrigins; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("AllowedOrigins = %v, want %v (entries must be trimmed)", got, want)
	}
}

func TestDatabaseDSNEscapesPassword(t *testing.T) {
	db := Database{
		Host: "db", Port: 5432, User: "pokedex",
		Password: "p@ss:w/rd?", Name: "pokedex", SSLMode: "disable",
	}

	dsn := db.DSN()
	if strings.Contains(dsn, "p@ss:w/rd?") {
		t.Errorf("DSN() left special characters unescaped: %s", dsn)
	}
	for _, want := range []string{"postgres://", "db:5432", "sslmode=disable"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("DSN() = %q, want it to contain %q", dsn, want)
		}
	}
}

// TestServerAddrBracketsIPv6 guards the use of net.JoinHostPort: a bare
// concatenation produces an unusable address for IPv6 hosts.
func TestServerAddrBracketsIPv6(t *testing.T) {
	s := Server{Host: "::1", Port: 8080}
	if got, want := s.Addr(), "[::1]:8080"; got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

// clearEnv removes every variable Load reads so a test starts from a known
// state regardless of the developer's shell or CI environment.
func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"APP_ENV",
		"SERVER_HOST", "SERVER_PORT", "SERVER_READ_TIMEOUT", "SERVER_READ_HEADER_TIMEOUT",
		"SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT", "SERVER_SHUTDOWN_TIMEOUT",
		"SERVER_HANDLER_TIMEOUT", "SERVER_MAX_HEADER_BYTES",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_CONNS", "DB_MIN_CONNS", "DB_MAX_CONN_LIFETIME", "DB_MAX_CONN_IDLE_TIME",
		"DB_CONNECT_TIMEOUT", "DB_STATEMENT_TIMEOUT",
		"LOG_LEVEL", "LOG_FORMAT",
		"CORS_ALLOWED_ORIGINS", "CORS_ALLOW_CREDENTIALS", "CORS_MAX_AGE",
		"RATE_LIMIT_ENABLED", "RATE_LIMIT_RPM", "RATE_LIMIT_BURST", "TRUSTED_PROXIES",
		"TLS_ENABLED", "TLS_CERT_PATH", "TLS_KEY_PATH",
		"REDIS_ENABLED", "REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB",
		"CACHE_ENABLED", "CACHE_TTL", "CACHE_MAX_ENTRIES",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
}
