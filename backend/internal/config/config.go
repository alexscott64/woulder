package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Weather  WeatherConfig
	Cache    CacheConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string
	GinMode string
	CORS    CORSConfig
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// WeatherConfig holds weather API configuration
type WeatherConfig struct {
	OpenWeatherMapAPIKey  string
	MountainProjectAPIKey string
	PreferOpenMeteo       bool
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	DurationMinutes int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "release"),
			CORS: CORSConfig{
				AllowOrigins:     []string{"*"}, // TODO: Configure per environment
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
				MaxAge:           12 * time.Hour,
			},
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "require"),
		},
		Weather: WeatherConfig{
			OpenWeatherMapAPIKey:  getEnv("OPENWEATHERMAP_API_KEY", ""),
			MountainProjectAPIKey: getEnv("MOUNTAIN_PROJECT_API_KEY", ""),
			PreferOpenMeteo:       true, // Open-Meteo is primary, OpenWeatherMap is fallback
		},
		Cache: CacheConfig{
			DurationMinutes: getEnvAsInt("CACHE_DURATION", 10),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	return nil
}

// ConnectionString returns a PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
