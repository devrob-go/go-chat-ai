package configs

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	PRODUCTION_ENV  = "production"
	STAGING_ENV     = "staging"
	DEVELOPMENT_ENV = "development"
)

// Config holds application configuration
type Config struct {
	Environment        string
	ChatServicePort    string
	RestGatewayPort    string
	LogLevel           string
	LogJSONFormat      bool
	HealthCheckTimeout int // in seconds
	ServerReadTimeout  int // in seconds
	ServerWriteTimeout int // in seconds

	// Security Configuration
	TLSEnabled    bool
	TLSCertFile   string
	TLSKeyFile    string
	MinTLSVersion uint16
	MaxTLSVersion uint16

	// Auth Service Configuration
	AuthServiceHost     string
	AuthServicePort     string
	AuthServiceTLS      bool
	AuthServiceCertFile string
	AuthServiceKeyFile  string
	AuthServiceCAFile   string

	// OpenAI Configuration
	OpenAIAPIKey      string
	OpenAIModel       string
	OpenAIMaxTokens   int
	OpenAITemperature float64
	OpenAITimeout     int // in seconds

	// Database Configuration (if needed for chat history)
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
	PostgresHost         string
	PostgresPort         string
	DBSSLMode            string
	DBMaxConnections     int
	DBMaxIdleConnections int
	DBConnectionTimeout  int // in seconds
	MigrationsDir        string

	// Rate Limiting
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   int // in seconds

	// Security Headers
	SecurityHeadersEnabled bool
	HSTSMaxAge             int // in seconds
	ContentSecurityPolicy  string

	// Logging Security
	LogSensitiveData  bool
	LogRequestHeaders bool
	LogResponseBody   bool
}

// LoadConfig loads and validates configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file only if it exists, without overwriting existing env vars
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	// Parse TLS version strings
	minTLSVersion := parseTLSVersion(getEnv("MIN_TLS_VERSION", "1.2"))
	maxTLSVersion := parseTLSVersion(getEnv("MAX_TLS_VERSION", "1.3"))

	// Parse OpenAI temperature
	openAITemp, err := strconv.ParseFloat(getEnv("OPENAI_TEMPERATURE", "0.7"), 64)
	if err != nil {
		openAITemp = 0.7
	}

	// Parse OpenAI max tokens
	openAIMaxTokens, err := strconv.Atoi(getEnv("OPENAI_MAX_TOKENS", "1000"))
	if err != nil {
		openAIMaxTokens = 1000
	}

	// Parse OpenAI timeout
	openAITimeout, err := strconv.Atoi(getEnv("OPENAI_TIMEOUT", "30"))
	if err != nil {
		openAITimeout = 30
	}

	cfg := &Config{
		Environment:        getEnv("APP_ENV", "development"),
		ChatServicePort:    getEnv("APP_PORT", "8082"),
		RestGatewayPort:    getEnv("REST_PORT", "8083"),
		LogLevel:           getEnv("LOG_LEVEL", "debug"),
		LogJSONFormat:      getEnvAsBool("LOG_JSON_FORMAT", false),
		HealthCheckTimeout: getEnvAsInt("HEALTH_CHECK_TIMEOUT", 30),
		ServerReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
		ServerWriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),

		// Security Configuration
		TLSEnabled:    getEnvAsBool("TLS_ENABLED", false),
		TLSCertFile:   getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:    getEnv("TLS_KEY_FILE", ""),
		MinTLSVersion: minTLSVersion,
		MaxTLSVersion: maxTLSVersion,

		// Auth Service Configuration
		AuthServiceHost:     getEnv("AUTH_SERVICE_HOST", "localhost"),
		AuthServicePort:     getEnv("AUTH_SERVICE_PORT", "8081"),
		AuthServiceTLS:      getEnvAsBool("AUTH_SERVICE_TLS", true),
		AuthServiceCertFile: getEnv("AUTH_SERVICE_CERT_FILE", ""),
		AuthServiceKeyFile:  getEnv("AUTH_SERVICE_KEY_FILE", ""),
		AuthServiceCAFile:   getEnv("AUTH_SERVICE_CA_FILE", ""),

		// OpenAI Configuration
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:       getEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
		OpenAIMaxTokens:   openAIMaxTokens,
		OpenAITemperature: openAITemp,
		OpenAITimeout:     openAITimeout,

		// Database Configuration
		PostgresUser:         getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:     getEnv("POSTGRES_PASSWORD", "password"),
		PostgresDB:           getEnv("POSTGRES_DB", "chat_db"),
		PostgresHost:         getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:         getEnv("POSTGRES_PORT", "5432"),
		DBSSLMode:            getEnv("DB_SSL_MODE", "disable"),
		DBMaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 10),
		DBMaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
		DBConnectionTimeout:  getEnvAsInt("DB_CONNECTION_TIMEOUT", 30),
		MigrationsDir:        getEnv("MIGRATIONS_DIR", "./storage/migrations"),

		// Rate Limiting
		RateLimitEnabled:  getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsInt("RATE_LIMIT_WINDOW", 60),

		// Security Headers
		SecurityHeadersEnabled: getEnvAsBool("SECURITY_HEADERS_ENABLED", true),
		HSTSMaxAge:             getEnvAsInt("HSTS_MAX_AGE", 31536000),
		ContentSecurityPolicy:  getEnv("CONTENT_SECURITY_POLICY", "default-src 'self'"),

		// Logging Security
		LogSensitiveData:  getEnvAsBool("LOG_SENSITIVE_DATA", false),
		LogRequestHeaders: getEnvAsBool("LOG_REQUEST_HEADERS", false),
		LogResponseBody:   getEnvAsBool("LOG_RESPONSE_BODY", false),
	}

	// Validate required configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}

	if c.AuthServiceHost == "" {
		return fmt.Errorf("AUTH_SERVICE_HOST is required")
	}

	// Only validate TLS certificates if TLS is actually enabled
	if c.AuthServiceTLS && c.TLSEnabled {
		if c.AuthServiceCertFile == "" {
			return fmt.Errorf("AUTH_SERVICE_CERT_FILE is required when TLS is enabled")
		}
		if c.AuthServiceKeyFile == "" {
			return fmt.Errorf("AUTH_SERVICE_KEY_FILE is required when TLS is enabled")
		}
	}

	return nil
}

// GetAuthServiceEndpoint returns the full auth service endpoint
func (c *Config) GetAuthServiceEndpoint() string {
	protocol := "http"
	if c.AuthServiceTLS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%s", protocol, c.AuthServiceHost, c.AuthServicePort)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func parseTLSVersion(version string) uint16 {
	switch strings.ToLower(version) {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}
