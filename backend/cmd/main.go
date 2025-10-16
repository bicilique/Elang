package main

import "elang-backend/internal/config"

func main() {

	// Load configurations
	configs := config.LoadConfigurations()

	// Initialize database connection
	dbConfig := config.Config{
		Host:     configs.DB_HOST,
		Port:     configs.DB_PORT,
		User:     configs.DB_USER,
		Password: configs.DB_PASSWORD,
		DBName:   configs.DB_NAME,
		SSLMode:  configs.DB_SSLMODE,
	}
	database, err := config.NewDatabase(dbConfig)
	if err != nil {
		panic(err)
	}

	// Run database migrations
	if err := database.AutoMigrate(); err != nil {
		panic(err)
	}
	database.Seed()

	// Initialize logger
	logger := config.NewLogger()

	// Create AppConfig
	appConfig := &config.AppConfig{
		Log:    logger,
		Config: configs,
		DB:     database.Connection,
	}

	// Bootstrap the application
	config.Bootstrap(appConfig)
}
