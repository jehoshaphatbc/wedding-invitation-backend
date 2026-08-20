package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment  string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	DBTimezone   string

	JWTSecret          string
	JWTAccessExpiry    int
	JWTRefreshExpiry   int
	JWTIssuer          string

	SuperAdminName     string
	SuperAdminEmail    string
	SuperAdminPassword string

	ServerPort string
	GinMode    string

	FRONTEND_URL string
}

func Load() (*Config, error) {
	godotenv.Load()

	accessExpiry, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINUTES", "15"))
	refreshExpiry, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "30"))

	cfg := &Config{
		Environment:  getEnv("ENVIRONMENT", "development"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "wedding_invitation"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		DBTimezone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),

		JWTSecret:          getEnv("JWT_SECRET", "change-this-in-production"),
		JWTAccessExpiry:    accessExpiry,
		JWTRefreshExpiry:   refreshExpiry,
		JWTIssuer:          "wedding-invitation-backend",

		SuperAdminName:     getEnv("SUPER_ADMIN_NAME", "Super Admin"),
		SuperAdminEmail:    getEnv("SUPER_ADMIN_EMAIL", "admin@wedding.com"),
		SuperAdminPassword: getEnv("SUPER_ADMIN_PASSWORD", "SuperAdmin123!"),

		ServerPort: getEnv("SERVER_PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		FRONTEND_URL: getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode, c.DBTimezone,
	)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
