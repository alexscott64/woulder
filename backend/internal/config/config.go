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
	Auth     AuthConfig
	Upload   UploadConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port                   string
	GinMode                string
	CORS                   CORSConfig
	DisableBackgroundSyncs bool
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
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// WeatherConfig holds weather API configuration
type WeatherConfig struct {
	OpenWeatherMapAPIKey  string
	MountainProjectAPIKey string
	PreferOpenMeteo       bool
	// OfflineMode, when true, instructs the weather service to skip all
	// Open-Meteo / OpenWeatherMap API calls on the per-request hot path
	// and serve weather data exclusively from the local database. This is
	// intended for development to avoid hitting Open-Meteo rate limits
	// while iterating on the UI. Refresh the DB manually with
	// `cmd/sync_weather`. Loaded from WEATHER_OFFLINE_MODE (default false).
	OfflineMode bool
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	DurationMinutes int
}

// AuthConfig holds general app auth configuration.
type AuthConfig struct {
	JWTSecret          string
	AccessTokenMinutes int
	RefreshTokenDays   int
	AdminEmail         string
	AdminPassword      string
	AdminDisplayName   string
}

// UploadConfig holds local upload storage configuration.
type UploadConfig struct {
	StorageDriver string
	Dir           string
	MaxBytes      int64
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:                   getEnv("PORT", "8080"),
			GinMode:                getEnv("GIN_MODE", "release"),
			DisableBackgroundSyncs: getEnvAsBool("DISABLE_BACKGROUND_SYNCS", false),
			CORS: CORSConfig{
				AllowOrigins:     []string{"*"}, // TODO: Configure per environment
				AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
				MaxAge:           12 * time.Hour,
			},
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", ""),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", ""),
			SSLMode:         getEnv("DB_SSLMODE", "require"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
			ConnMaxIdleTime: time.Duration(getEnvAsInt("DB_CONN_MAX_IDLE_TIME_MINUTES", 5)) * time.Minute,
		},
		Weather: WeatherConfig{
			OpenWeatherMapAPIKey:  getEnv("OPENWEATHERMAP_API_KEY", ""),
			MountainProjectAPIKey: getEnv("MOUNTAIN_PROJECT_API_KEY", ""),
			PreferOpenMeteo:       true, // Open-Meteo is primary, OpenWeatherMap is fallback
			OfflineMode:           getEnvAsBool("WEATHER_OFFLINE_MODE", false),
		},
		Cache: CacheConfig{
			DurationMinutes: getEnvAsInt("CACHE_DURATION", 10),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnv("APP_JWT_SECRET", "development-only-change-me"),
			AccessTokenMinutes: getEnvAsInt("APP_ACCESS_TOKEN_MINUTES", 15),
			RefreshTokenDays:   getEnvAsInt("APP_REFRESH_TOKEN_DAYS", 30),
			AdminEmail:         getEnv("APP_ADMIN_EMAIL", getEnv("MONEY_USERNAME", "")),
			AdminPassword:      getEnv("APP_ADMIN_PASSWORD", getEnv("MONEY_PASSWORD", "")),
			AdminDisplayName:   getEnv("APP_ADMIN_DISPLAY_NAME", "Money Creek Admin"),
		},
		Upload: UploadConfig{
			StorageDriver: getEnv("UPLOAD_STORAGE_DRIVER", "local"),
			Dir:           getEnv("UPLOAD_DIR", "./uploads"),
			MaxBytes:      int64(getEnvAsInt("UPLOAD_MAX_BYTES", 10*1024*1024)),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
