package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Configurations struct {
	// Server configuration
	PORT string
	MODE string

	// Database configuration
	DB_HOST     string
	DB_PORT     string
	DB_USER     string
	DB_PASSWORD string
	DB_NAME     string
	DB_SSLMODE  string

	// Object Storage (Minio) configuration
	MINIO_ENDPOINT    string
	MINIO_ACCESS_KEY  string
	MINIO_SECRET_KEY  string
	MINIO_BUCKET_NAME string
	MINIO_USE_SSL     bool

	// Messaging service configuration
	MESSAGING_SERVICE_URL string
}

func LoadConfigurations() *Configurations {

	if os.Getenv("DEVELOPER_HOST") == "true" {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env file")
		}

	}
	return &Configurations{
		// Server configuration
		PORT: os.Getenv("PORT"),
		MODE: os.Getenv("MODE"),

		// Database configuration
		DB_HOST:     getEnvWithDefault("DB_HOST", "localhost"),
		DB_PORT:     getEnvWithDefault("DB_PORT", "5432"),
		DB_USER:     getEnvWithDefault("DB_USER", "postgres"),
		DB_PASSWORD: getEnvWithDefault("DB_PASSWORD", ""),
		DB_NAME:     getEnvWithDefault("DB_NAME", "go_messaging"),
		DB_SSLMODE:  getEnvWithDefault("DB_SSLMODE", "disable"),

		// Object Storage (Minio) configuration
		MINIO_ENDPOINT:    getEnvWithDefault("STORAGE_ENDPOINT", "localhost:9000"),
		MINIO_ACCESS_KEY:  getEnvWithDefault("STORAGE_ACCESS_KEY", "minioadmin"),
		MINIO_SECRET_KEY:  getEnvWithDefault("STORAGE_SECRET_KEY", "minioadmin"),
		MINIO_BUCKET_NAME: getEnvWithDefault("BUCKET_NAME", "silent-patch-detector"),
		MINIO_USE_SSL:     getEnvWithDefault("STRORAGE_SSL", "false") == "true",

		// Messaging service configuration
		MESSAGING_SERVICE_URL: getEnvWithDefault("MESSAGING_SERVICE_URL", ""),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
