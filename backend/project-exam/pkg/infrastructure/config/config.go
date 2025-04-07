package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Ethereum EthereumConfig
	Log      LogConfig
}

// ServerConfig holds configuration related to the HTTP server
type ServerConfig struct {
	Port         string
	Mode         string // debug or release
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	RateLimit    RateLimitConfig
	Auth         AuthConfig
}

// RateLimitConfig configures the rate limiter
type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled bool
	APIKeys map[string]string // map[apiKey]userID
}

// EthereumConfig holds configuration related to Ethereum client
type EthereumConfig struct {
	RPCURL          string
	RequestTimeout  time.Duration
	DefaultGasLimit uint64
	RetryAttempts   int
	RetryDelay      time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json or text
	OutputPath string // stdout, stderr, or filepath
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Parse API keys from environment
	apiKeys := make(map[string]string)
	apiKeysStr := getEnv("API_KEYS", "")
	if apiKeysStr != "" {
		keyPairs := strings.Split(apiKeysStr, ",")
		for _, pair := range keyPairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				apiKeys[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			RateLimit: RateLimitConfig{
				Limit:  getIntEnv("RATE_LIMIT", 100),
				Window: getDurationEnv("RATE_LIMIT_WINDOW", 15*time.Minute),
			},
			Auth: AuthConfig{
				Enabled: getBoolEnv("AUTH_ENABLED", false),
				APIKeys: apiKeys,
			},
		},
		Ethereum: EthereumConfig{
			RPCURL:          getEnv("ETHEREUM_RPC_URL", "https://mainnet.infura.io/v3/YOUR_INFURA_KEY"),
			RequestTimeout:  getDurationEnv("ETHEREUM_REQUEST_TIMEOUT", 10*time.Second),
			DefaultGasLimit: getUint64Env("ETHEREUM_DEFAULT_GAS_LIMIT", 21000),
			RetryAttempts:   getIntEnv("ETHEREUM_RETRY_ATTEMPTS", 3),
			RetryDelay:      getDurationEnv("ETHEREUM_RETRY_DELAY", 1*time.Second),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT", "stdout"),
		},
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getUint64Env(key string, defaultValue uint64) uint64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value := strings.ToLower(valueStr)
	return value == "true" || value == "1" || value == "yes" || value == "y"
}
