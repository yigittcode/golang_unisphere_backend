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
		Port string `yaml:"port" env:"SERVER_PORT"`
		Mode string `yaml:"mode" env:"SERVER_MODE"`
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
	// Load default config with sane defaults
	config := &Config{}
	setDefaults(config)

	// Try to read config file if it exists
	if _, err := os.Stat(configPath); err == nil {
	// Read file
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into Config structure
		err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	}

	// Override with environment variables
	err := loadFromEnv(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load from environment: %w", err)
	}

	// Validate config
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for the configuration
func setDefaults(config *Config) {
	// Server defaults
	config.Server.Port = "8080"
	config.Server.Mode = "development"

	// Database defaults
	config.Database.Driver = "postgres"
	config.Database.Host = "localhost"
	config.Database.Port = "5432"
	config.Database.User = "postgres"
	config.Database.Password = "postgres"
	config.Database.DBName = "unisphere"
	config.Database.SSLMode = "disable"
	config.Database.MaxIdleConns = 5
	config.Database.MaxOpenConns = 20
	config.Database.ConnMaxLifetime = "1h"

	// JWT defaults
	config.JWT.AccessTokenExpiration = "1h"
	config.JWT.RefreshTokenExpiration = "720h"
	config.JWT.Issuer = "unisphere.app"

	// Logging defaults
	config.Logging.Level = "info"
	config.Logging.Format = "json"
}

// loadFromEnv overrides configuration with environment variables
func loadFromEnv(config *Config) error {
	// Recursively process the config structure and look for env tags
	err := processStructFields(config)
	if err != nil {
		return err
	}

	return nil
}

// validateConfig ensures that the configuration is valid
func validateConfig(config *Config) error {
	// Ensure required fields are set
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	// Validate JWT expiration formats
	if _, err := time.ParseDuration(config.JWT.AccessTokenExpiration); err != nil {
		return fmt.Errorf("invalid JWT access token expiration format: %w", err)
	}

	if _, err := time.ParseDuration(config.JWT.RefreshTokenExpiration); err != nil {
		return fmt.Errorf("invalid JWT refresh token expiration format: %w", err)
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
