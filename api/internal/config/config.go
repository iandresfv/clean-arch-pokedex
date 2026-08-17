// Package config provides environment-based configuration loading and validation.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Config aggregates every runtime setting the server needs. It is populated once
// at startup by Load and treated as immutable afterwards.
type Config struct {
	Env       string
	Server    Server
	Database  Database
	Log       Log
	CORS      CORS
	RateLimit RateLimit
	TLS       TLS
	Redis     Redis
	Cache     Cache
}

// IsProduction reports whether the process is running with production defaults,
// which tighten several validation rules in validate.
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

type Server struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	HandlerTimeout    time.Duration
	MaxHeaderBytes    int
}

type Database struct {
	Host             string
	Port             int
	User             string
	Password         string
	Name             string
	SSLMode          string
	MaxConns         int
	MinConns         int
	MaxConnLifetime  time.Duration
	MaxConnIdleTime  time.Duration
	ConnectTimeout   time.Duration
	StatementTimeout time.Duration
}

type Log struct {
	Level  slog.Level
	Format string
}

type CORS struct {
	AllowedOrigins   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type RateLimit struct {
	Enabled        bool
	RequestsPerMin int
	Burst          int
	TrustedProxies []string
}

type TLS struct {
	Enabled  bool
	CertPath string
	KeyPath  string
}

type Redis struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
}

type Cache struct {
	Enabled    bool
	TTL        time.Duration
	MaxEntries int
}

// Addr returns the host:port the HTTP server binds to.
func (s Server) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

// DSN returns a PostgreSQL connection string in URL form. The password is
// escaped so that special characters do not corrupt the URL.
func (d Database) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   net.JoinHostPort(d.Host, strconv.Itoa(d.Port)),
		Path:   d.Name,
	}
	q := u.Query()
	q.Set("sslmode", d.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

// Load reads configuration from the environment, applying defaults and
// validating every value. All problems are collected and returned together so a
// misconfigured deployment reports every error at once instead of one per
// restart.
func Load() (*Config, error) {
	l := &loader{}
	env := l.enum("APP_ENV", EnvDevelopment, EnvDevelopment, EnvProduction)

	// Development ships with working defaults so the project runs on a fresh
	// clone. Production has no safe default for a credential, so the variable
	// becomes mandatory and its absence fails startup.
	dbPassword := l.str("DB_PASSWORD", "postgres")
	if env == EnvProduction {
		dbPassword = l.mustStr("DB_PASSWORD")
	}

	cfg := &Config{
		Env: env,
		Server: Server{
			Host:              l.str("SERVER_HOST", "localhost"),
			Port:              l.int("SERVER_PORT", 8080),
			ReadTimeout:       l.duration("SERVER_READ_TIMEOUT", 15*time.Second),
			ReadHeaderTimeout: l.duration("SERVER_READ_HEADER_TIMEOUT", 5*time.Second),
			WriteTimeout:      l.duration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:       l.duration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:   l.duration("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
			HandlerTimeout:    l.duration("SERVER_HANDLER_TIMEOUT", 10*time.Second),
			MaxHeaderBytes:    l.int("SERVER_MAX_HEADER_BYTES", 1<<20),
		},
		Database: Database{
			Host:             l.str("DB_HOST", "localhost"),
			Port:             l.int("DB_PORT", 5432),
			User:             l.str("DB_USER", "postgres"),
			Password:         dbPassword,
			Name:             l.str("DB_NAME", "pokedex"),
			SSLMode:          l.str("DB_SSLMODE", "disable"),
			MaxConns:         l.int("DB_MAX_CONNS", 25),
			MinConns:         l.int("DB_MIN_CONNS", 5),
			MaxConnLifetime:  l.duration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
			MaxConnIdleTime:  l.duration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
			ConnectTimeout:   l.duration("DB_CONNECT_TIMEOUT", 30*time.Second),
			StatementTimeout: l.duration("DB_STATEMENT_TIMEOUT", 10*time.Second),
		},
		Log: Log{
			Level:  l.logLevel("LOG_LEVEL", slog.LevelInfo),
			Format: l.enum("LOG_FORMAT", "json", "json", "text"),
		},
		CORS: CORS{
			AllowedOrigins:   l.strSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
			AllowCredentials: l.bool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           l.duration("CORS_MAX_AGE", 12*time.Hour),
		},
		RateLimit: RateLimit{
			Enabled:        l.bool("RATE_LIMIT_ENABLED", true),
			RequestsPerMin: l.int("RATE_LIMIT_RPM", 120),
			Burst:          l.int("RATE_LIMIT_BURST", 30),
			TrustedProxies: l.strSlice("TRUSTED_PROXIES", nil),
		},
		TLS: TLS{
			Enabled:  l.bool("TLS_ENABLED", false),
			CertPath: l.str("TLS_CERT_PATH", ""),
			KeyPath:  l.str("TLS_KEY_PATH", ""),
		},
		Redis: Redis{
			Enabled:  l.bool("REDIS_ENABLED", false),
			Addr:     l.str("REDIS_ADDR", "localhost:6379"),
			Password: l.str("REDIS_PASSWORD", ""),
			DB:       l.int("REDIS_DB", 0),
		},
		Cache: Cache{
			Enabled:    l.bool("CACHE_ENABLED", true),
			TTL:        l.duration("CACHE_TTL", 5*time.Minute),
			MaxEntries: l.int("CACHE_MAX_ENTRIES", 1024),
		},
	}

	if err := errors.Join(append(l.errs, cfg.validate()...)...); err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}
	return cfg, nil
}

// validate enforces invariants that cannot be expressed by parsing alone.
func (c *Config) validate() []error {
	var errs []error

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", c.Server.Port))
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		errs = append(errs, fmt.Errorf("DB_PORT must be between 1 and 65535, got %d", c.Database.Port))
	}
	if c.Database.Name == "" {
		errs = append(errs, errors.New("DB_NAME must not be empty"))
	}
	if c.Database.User == "" {
		errs = append(errs, errors.New("DB_USER must not be empty"))
	}
	if c.Database.MaxConns < 1 {
		errs = append(errs, fmt.Errorf("DB_MAX_CONNS must be at least 1, got %d", c.Database.MaxConns))
	}
	if c.Database.MinConns < 0 || c.Database.MinConns > c.Database.MaxConns {
		errs = append(errs, fmt.Errorf(
			"DB_MIN_CONNS must be between 0 and DB_MAX_CONNS (%d), got %d",
			c.Database.MaxConns, c.Database.MinConns))
	}
	switch c.Database.SSLMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		errs = append(errs, fmt.Errorf("DB_SSLMODE %q is not a valid libpq sslmode", c.Database.SSLMode))
	}
	if c.RateLimit.Enabled && c.RateLimit.RequestsPerMin < 1 {
		errs = append(errs, fmt.Errorf("RATE_LIMIT_RPM must be at least 1, got %d", c.RateLimit.RequestsPerMin))
	}
	if c.RateLimit.Enabled && c.RateLimit.Burst < 1 {
		errs = append(errs, fmt.Errorf("RATE_LIMIT_BURST must be at least 1, got %d", c.RateLimit.Burst))
	}
	for _, cidr := range c.RateLimit.TrustedProxies {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXIES entry %q is not valid CIDR notation", cidr))
		}
	}
	if c.TLS.Enabled {
		if c.TLS.CertPath == "" || c.TLS.KeyPath == "" {
			errs = append(errs, errors.New("TLS_CERT_PATH and TLS_KEY_PATH are required when TLS_ENABLED=true"))
		}
	}
	// A wildcard origin combined with credentials is rejected outright by every
	// browser, so accepting it here would produce a server that looks configured
	// but cannot serve the client.
	if c.CORS.AllowCredentials {
		for _, o := range c.CORS.AllowedOrigins {
			if o == "*" {
				errs = append(errs, errors.New(
					`CORS_ALLOWED_ORIGINS cannot contain "*" while CORS_ALLOW_CREDENTIALS=true`))
			}
		}
	}
	// Localhost origins in production almost always mean the variable was never
	// set for the environment, which would break the deployed client silently.
	if c.IsProduction() {
		for _, o := range c.CORS.AllowedOrigins {
			if strings.Contains(o, "localhost") || strings.Contains(o, "127.0.0.1") {
				errs = append(errs, fmt.Errorf(
					"CORS_ALLOWED_ORIGINS contains loopback origin %q while APP_ENV=production", o))
			}
		}
	}
	if c.Cache.Enabled && c.Cache.MaxEntries < 1 {
		errs = append(errs, fmt.Errorf("CACHE_MAX_ENTRIES must be at least 1, got %d", c.Cache.MaxEntries))
	}
	if c.Redis.Enabled && c.Redis.Addr == "" {
		errs = append(errs, errors.New("REDIS_ADDR is required when REDIS_ENABLED=true"))
	}

	return errs
}

// loader reads environment variables, accumulating parse failures instead of
// returning on the first one.
type loader struct {
	errs []error
}

func (l *loader) str(key, fallback string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	return v
}

// mustStr records an error when the variable is absent or empty. It exists for
// values that have no safe default, such as a production database password.
func (l *loader) mustStr(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		l.errs = append(l.errs, fmt.Errorf("%s is required but not set", key))
		return ""
	}
	return v
}

func (l *loader) int(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s must be an integer, got %q", key, raw))
		return fallback
	}
	return v
}

func (l *loader) bool(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s must be a boolean, got %q", key, raw))
		return fallback
	}
	return v
}

func (l *loader) duration(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	v, err := time.ParseDuration(raw)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s must be a duration such as 30s or 5m, got %q", key, raw))
		return fallback
	}
	if v <= 0 {
		l.errs = append(l.errs, fmt.Errorf("%s must be positive, got %q", key, raw))
		return fallback
	}
	return v
}

func (l *loader) strSlice(key string, fallback []string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (l *loader) enum(key, fallback string, allowed ...string) string {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	raw = strings.ToLower(raw)
	if slices.Contains(allowed, raw) {
		return raw
	}
	l.errs = append(l.errs, fmt.Errorf("%s must be one of %s, got %q", key, strings.Join(allowed, ", "), raw))
	return fallback
}

func (l *loader) logLevel(key string, fallback slog.Level) slog.Level {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(raw)); err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s must be one of debug, info, warn, error, got %q", key, raw))
		return fallback
	}
	return lvl
}
