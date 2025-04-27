package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config structure represents the application configuration
type Config struct {
	Server struct {
		Port        string `yaml:"port" env:"SERVER_PORT"`
		Mode        string `yaml:"mode" env:"SERVER_MODE"`
		StoragePath string `yaml:"storage_path" env:"STORAGE_PATH"`
	} `yaml:"server"`

	Database struct {
		Driver          string `yaml:"driver" env:"DB_DRIVER"`
		Host            string `yaml:"host" env:"DB_HOST"`
		Port            string `yaml:"port" env:"DB_PORT"`
		User            string `yaml:"user" env:"DB_USER"`
		Password        string `yaml:"password" env:"DB_PASSWORD"`
		DBName          string `yaml:"dbname" env:"DB_NAME"`
		SSLMode         string `yaml:"sslmode" env:"DB_SSLMODE"`
		MaxIdleConns    int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
		MaxOpenConns    int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
		ConnMaxLifetime string `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"`
	} `yaml:"database"`

	JWT struct {
		Secret                 string `yaml:"secret" env:"JWT_SECRET"`
		AccessTokenExpiration  string `yaml:"access_token_expiration" env:"JWT_ACCESS_TOKEN_EXPIRATION"`
		RefreshTokenExpiration string `yaml:"refresh_token_expiration" env:"JWT_REFRESH_TOKEN_EXPIRATION"`
		Issuer                 string `yaml:"issuer" env:"JWT_ISSUER"`
	} `yaml:"jwt"`

	Logging struct {
		Level  string `yaml:"level" env:"LOG_LEVEL"`
		Format string `yaml:"format" env:"LOG_FORMAT"`
	} `yaml:"logging"`
}

// LoadConfig loads configuration from a file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}
	setDefaults(config)

	if _, err := os.Stat(configPath); err == nil {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
		err = yaml.Unmarshal(file, config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	err := loadFromEnv(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load from environment: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for the configuration
func setDefaults(config *Config) {
	config.Server.Port = "8080"
	config.Server.Mode = "development"
	config.Server.StoragePath = "./uploads"

	config.Database.Driver = "postgres"
	config.Database.Host = "localhost"
	config.Database.Port = "5432"
	config.Database.User = "postgres"
	config.Database.DBName = "unisphere"
	config.Database.SSLMode = "disable"
	config.Database.MaxIdleConns = 5
	config.Database.MaxOpenConns = 20
	config.Database.ConnMaxLifetime = "1h"

	config.JWT.AccessTokenExpiration = "1h"
	config.JWT.RefreshTokenExpiration = "720h"
	config.JWT.Issuer = "unisphere.app"

	config.Logging.Level = "info"
	config.Logging.Format = "text"
}

// loadFromEnv overrides configuration with environment variables
func loadFromEnv(config *Config) error {
	err := processStructFields(config)
	if err != nil {
		return err
	}
	return nil
}

// validateConfig ensures that the configuration is valid
func validateConfig(config *Config) error {
	// Ensure required non-secret fields are set
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver (DB_DRIVER) is required")
	}
	if config.Database.Host == "" {
		return fmt.Errorf("database host (DB_HOST) is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user (DB_USER) is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name (DB_NAME) is required")
	}
	if config.Server.Port == "" {
		return fmt.Errorf("server port (SERVER_PORT) is required")
	}

	// Ensure secrets are provided (likely via ENV vars)
	if config.Database.Password == "" {
		return fmt.Errorf("database password (DB_PASSWORD) is required and should be set via environment variable")
	}
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret (JWT_SECRET) is required and should be set via environment variable")
	}

	// Validate time duration formats
	if _, err := time.ParseDuration(config.JWT.AccessTokenExpiration); err != nil {
		return fmt.Errorf("invalid JWT access token expiration format (JWT_ACCESS_TOKEN_EXPIRATION): %w", err)
	}
	if _, err := time.ParseDuration(config.JWT.RefreshTokenExpiration); err != nil {
		return fmt.Errorf("invalid JWT refresh token expiration format (JWT_REFRESH_TOKEN_EXPIRATION): %w", err)
	}
	if _, err := time.ParseDuration(config.Database.ConnMaxLifetime); err != nil {
		return fmt.Errorf("invalid database connection max lifetime format (DB_CONN_MAX_LIFETIME): %w", err)
	}

	// Validate server mode
	mode := strings.ToLower(config.Server.Mode)
	if mode != "development" && mode != "production" {
		return fmt.Errorf("invalid server mode '%s' (SERVER_MODE): must be 'development' or 'production'", config.Server.Mode)
	}

	// Validate log level
	level := strings.ToLower(config.Logging.Level)
	if level != "debug" && level != "info" && level != "warn" && level != "error" {
		return fmt.Errorf("invalid log level '%s' (LOG_LEVEL): must be 'debug', 'info', 'warn', or 'error'", config.Logging.Level)
	}

	// Validate log format
	format := strings.ToLower(config.Logging.Format)
	if format != "text" && format != "json" {
		return fmt.Errorf("invalid log format '%s' (LOG_FORMAT): must be 'text' or 'json'", config.Logging.Format)
	}

	return nil
}

// GetPostgresConnectionString returns postgres connection string
func (c *Config) GetPostgresConnectionString() string {
	sslMode := c.Database.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		sslMode,
	)
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvAsInt gets an environment variable as an integer or returns a default value
func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// GetEnvAsBool gets an environment variable as a boolean or returns a default value
func GetEnvAsBool(key string, defaultValue bool) bool {
	valueStr := GetEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	// Convert string to lowercase for checking
	valueLower := strings.ToLower(valueStr)
	if valueLower == "true" || valueLower == "1" || valueLower == "yes" {
		return true
	}
	if valueLower == "false" || valueLower == "0" || valueLower == "no" {
		return false
	}

	return defaultValue
}

// GetEnvAsDuration gets an environment variable as a duration or returns a default value
func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := GetEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
