package config

import (
	"elang-backend/internal/entity"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	Connection *gorm.DB
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabase creates a new database connection
func NewDatabase(config Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	_ = gormLogger

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: gormLogger,
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	return &Database{Connection: db}, nil
}

// AutoMigrate runs database migrations
func (d *Database) AutoMigrate() error {
	log.Println("🔄 Starting database migration...")

	// Core entity migration
	err := d.Connection.AutoMigrate(
		&entity.App{},
		&entity.Dependency{},
		&entity.DependencyVersion{},
		&entity.Framework{},
		&entity.Runtime{},
		&entity.AppDependency{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate core entity: %w", err)
	}
	log.Println("✅ Core entity migrated successfully")

	// Enhanced entity migration for Security Detector V2
	err = d.Connection.AutoMigrate(
		&entity.MonitoringJob{},
		&entity.AuditTrail{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate enhanced entity: %w", err)
	}
	log.Println("✅ Enhanced entity migrated successfully")

	log.Println("✅ Database migration completed successfully")
	return nil
}

// Seed seeds initial data into the database
func (d *Database) Seed() {
	runTimeTypes := []entity.Runtime{
		{Name: "Node.js"},
		{Name: "Python"},
		{Name: "Java"},
		{Name: "Go"},
		{Name: "Ruby"},
		{Name: "PHP"},
		{Name: "DotNet"},
		{Name: "Gradle"},
	}

	// Seed Runtime Types and build a map of name to ID
	runtimeIDMap := make(map[string]int)
	for _, rt := range runTimeTypes {
		var existing entity.Runtime
		result := d.Connection.Where("name = ?", rt.Name).First(&existing)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				d.Connection.Create(&rt)
				log.Printf("Seeded runtime: %s\n", rt.Name)
				d.Connection.Where("name = ?", rt.Name).First(&existing)
			}
			// else log error
		} else {
			log.Printf("Runtime %s already exists, skipping seeding.\n", rt.Name)
		}
		runtimeIDMap[rt.Name] = existing.ID
	}

	frameworks := []struct {
		Name    string
		Runtime string // runtime name, not ID
	}{
		{Name: "Express", Runtime: "Node.js"},
		{Name: "Django", Runtime: "Python"},
		{Name: "Spring", Runtime: "Java"},
		{Name: "Gin", Runtime: "Go"},
		{Name: "Rails", Runtime: "Ruby"},
		{Name: "Laravel", Runtime: "PHP"},
		{Name: "ASP.NET", Runtime: "DotNet"},
		{Name: "Flask", Runtime: "Python"},
		{Name: "React", Runtime: "Node.js"},
		{Name: "Vue.js", Runtime: "Node.js"},
		{Name: "Angular", Runtime: "Node.js"},
		{Name: "Spring Boot", Runtime: "Java"},
		{Name: "Echo", Runtime: "Go"},
		{Name: "Symfony", Runtime: "PHP"},
		{Name: "Ruby on Rails", Runtime: "Ruby"},
		{Name: "CodeIgniter", Runtime: "PHP"},
		{Name: "Native", Runtime: "Gradle"},
	}

	// Seed Frameworks (no runtime association, case-insensitive check)
	for _, fw := range frameworks {
		name := strings.TrimSpace(fw.Name)
		var existing entity.Framework
		result := d.Connection.Where("LOWER(name) = ?", strings.ToLower(name)).First(&existing)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				framework := entity.Framework{Name: name}
				d.Connection.Create(&framework)
				log.Printf("Seeded framework: %s\n", name)
			} else {
				log.Printf("Error checking framework %s: %v\n", name, result.Error)
			}
		} else {
			log.Printf("Framework %s already exists, skipping seeding.\n", name)
		}
	}
	log.Println("✅ Database seeding completed successfully.")
}

// Ping tests the database connection
func (d *Database) Ping() error {
	sqlDB, err := d.Connection.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.Connection.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
