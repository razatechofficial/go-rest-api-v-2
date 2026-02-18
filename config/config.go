// Package config handles all application configuration.
// It uses cleanenv to load from YAML files and environment variables.
// Environment variables override YAML values (12-factor app principle).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config is the root configuration struct.
// Tags tell cleanenv how to map values:
// - yaml:"field" = map to YAML field
// - env:"VAR_NAME" = map to environment variable
// - env-default:"value" = default if not set
// - env-required:"true" = must be provided
type Config struct {
	App      AppConfig      `yaml:"app" env-prefix:"APP_"`
	Server   ServerConfig   `yaml:"server" env-prefix:"SERVER_"`
	Database DatabaseConfig `yaml:"database" env-prefix:"DB_"`
	Log      LogConfig      `yaml:"log" env-prefix:"LOG_"`
}

// AppConfig holds general application settings.
// These are used for identification and behavior tuning.
type AppConfig struct {
	// Name is the application name used in logs and headers
	Name string `yaml:"name" env:"NAME" env-default:"Enterprise API"`

	// Version is the current release version
	Version string `yaml:"version" env:"VERSION" env-default:"1.0.0"`

	// Environment determines behavior: development, staging, production
	// In production, we enable stricter security and disable debug features
	Environment string `yaml:"environment" env:"ENV" env-default:"development"`
}

// ServerConfig holds HTTP server settings.
// Timeouts prevent slowloris attacks and resource exhaustion.
type ServerConfig struct {
	// Host binds to all interfaces by default (0.0.0.0)
	// Use localhost (127.0.0.1) for local development only
	Host string `yaml:"host" env:"HOST" env-default:"0.0.0.0"`

	// Port is the HTTP listening port
	Port int `yaml:"port" env:"PORT" env-default:"8080"`

	// ReadTimeout is max time to read request body
	// Prevents clients from keeping connections open forever
	ReadTimeout time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"10s"`

	// WriteTimeout is max time to write response
	// Prevents slow clients from blocking server
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`

	// IdleTimeout is max time between requests on persistent connection
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"60s"`

	// Mode is Gin framework mode: debug, release, test
	// Release mode disables debug prints and enables caching
	Mode string `yaml:"mode" env:"MODE" env-default:"debug"`
}

// DatabaseConfig holds PostgreSQL connection settings.
// These are used to create the connection pool.
type DatabaseConfig struct {
	// Host is the database server address
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`

	// Port is the PostgreSQL port (default 5432)
	Port int `yaml:"port" env:"PORT" env-default:"5432"`

	// User is the database username
	User string `yaml:"user" env:"USER" env-default:"postgres"`

	// Password is REQUIRED - no default for security
	// Must be provided via environment variable
	Password string `yaml:"password" env:"PASSWORD" env-required:"true"`

	// Database is the database name
	Database string `yaml:"database" env:"DATABASE" env-default:"enterprise_db"`

	// SSLMode controls connection encryption
	// disable = no SSL (dev only), require = SSL required (production)
	SSLMode string `yaml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`

	// MaxOpenConns is the maximum number of open connections to the database
	// Rule of thumb: (CPU cores * 2) + effective disk spindles
	MaxOpenConns int32 `yaml:"max_open_conns" env:"MAX_OPEN_CONNS" env-default:"25"`

	// MinIdleConns is the minimum number of idle connections to keep open
	// Helps handle sudden traffic spikes without creating new connections
	MinIdleConns int32 `yaml:"min_idle_conns" env:"MIN_IDLE_CONNS" env-default:"10"`

	// MaxConnLifetime is how long a connection may be reused
	// Helps prevent issues with stale connections or database restarts
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env:"MAX_CONN_LIFETIME" env-default:"1h"`

	// MaxConnIdleTime is how long an idle connection stays open
	// Closes unused connections to save resources
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env:"MAX_CONN_IDLE_TIME" env-default:"30m"`
}

// LogConfig holds structured logging settings.
// We use Zap for high-performance JSON logging.
type LogConfig struct {
	// Level controls verbosity: debug, info, warn, error, fatal
	Level string `yaml:"level" env:"LEVEL" env-default:"info"`

	// Format is output format: json (production) or console (development)
	Format string `yaml:"format" env:"FORMAT" env-default:"console"`
}

// ConnectionString builds PostgreSQL connection URL.
// This format is used by pgx (PostgreSQL driver).
func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

// Address returns server bind address in host:port format.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// IsProduction returns true if running in production environment.
// Used to enable security features and disable debug tools.
func (a AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

// IsDevelopment returns true if running in development.
func (a AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}

// Load reads configuration from multiple sources:
// 1. Load .env from root (secrets)
// 2. Detect environment
// 3. Load environments/{env}.yaml
// 4. Override with env vars
func Load() (*Config, error) {
	var cfg Config

	// Step 1: Load .env from root (if exists)
	// This populates OS environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		// .env file is optional - continue with existing env vars
		fmt.Println("No .env file found, using existing environment variables")
	}

	// Step 2: Detect environment from ENV variable
	// Falls back to "local" if not set
	env := getEnvironment()

	// Step 3: Load from environments/{env}.yaml
	configFile := filepath.Join("environments", env+".yaml")

	if fileExists(configFile) {
		fmt.Printf("Loading config from: %s\n", configFile)
		if err := cleanenv.ReadConfig(configFile, &cfg); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", configFile, err)
		}
	} else {
		fmt.Printf("Warning: %s not found, using defaults\n", configFile)
	}

	// Step 4: Override with environment variables
	// This includes vars from .env file and system env
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment: %w", err)
	}

	// Set detected environment if not already set
	if cfg.App.Environment == "" {
		cfg.App.Environment = env
	}

	// Validate
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &cfg, nil
}

// getEnvironment detects current environment
func getEnvironment() string {
	// Check common environment variable names
	for _, key := range []string{"ENV", "APP_ENV", "ENVIRONMENT"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return "local" // Default to local
}

// fileExists checks if file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// validate checks configuration for errors and security issues.
func validate(cfg *Config) error {
	// In production, require SSL for database connections
	if cfg.App.IsProduction() && cfg.Database.SSLMode == "disable" {
		return fmt.Errorf("database SSL cannot be disabled in production")
	}

	// Validate timeout values are reasonable
	if cfg.Server.ReadTimeout < 1*time.Second {
		return fmt.Errorf("read timeout too short")
	}

	return nil
}
