package config

import (
	"context"
	delivery "elang-backend/internal/delivery/http"
	"elang-backend/internal/helper"
	"elang-backend/internal/model/dto"
	"elang-backend/internal/repository"
	"elang-backend/internal/services"
	"elang-backend/internal/usecase"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AppConfig struct {
	Log    *logrus.Logger
	Config *Configurations
	DB     *gorm.DB
}

func Bootstrap(Config *AppConfig) {
	// Initialize other components if needed
	repos := initializeRepositories(Config.DB)

	// Initialize services with repositories, logger, and configurations
	services := initializeServices(repos, Config.Log, Config.Config)

	// Initialize HTTP handlers
	server := setupHTTPServer(services)

	// Start HTTP server with graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startHTTPServer(ctx, server)

}

// startHTTPServer starts the HTTP server with graceful shutdown
func startHTTPServer(ctx context.Context, server *http.Server) {
	go func() {
		log.Printf("🌐 Starting HTTP server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("🛑 Shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("✅ HTTP server stopped gracefully")
	}
}

func setupHTTPServer(services *Services) *http.Server {
	router := gin.Default()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes with simplified handlers
	routeConfig := &delivery.RouteConfig{
		Router:              router,
		AppHandler:          *delivery.NewApplicationHandler(services.ApplicationService),
		DependenciesHandler: *delivery.NewDependenciesHandler(services.DepedenciesService),
	}
	routeConfig.Setup()

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

func initializeRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		App:              repository.NewAppRepository(db),
		Depedency:        repository.NewDependencyRepository(db),
		AppDepedency:     repository.NewAppDependencyRepository(db),
		DepedencyVersion: repository.NewDependencyVersionRepository(db),
		Runtime:          repository.NewRuntimeRepository(db),
		Framework:        repository.NewFrameworkRepository(db),
		AuditTrail:       repository.NewAuditTrailRepository(db),
	}
}

func initializeServices(repos *Repositories, log *logrus.Logger, cfg *Configurations) *Services {
	basicRepos := dto.BasicRepositories{
		AppRepository:              repos.App,
		DepedencyRepository:        repos.Depedency,
		AppToDepedencyRepository:   repos.AppDepedency,
		DepedencyVersionRepository: repos.DepedencyVersion,
		RunTimeRepository:          repos.Runtime,
		FrameWorkRepository:        repos.Framework,
	}
	dependencyParser := helper.NewDependencyParser()
	objectStorageService := usecase.NewMinioUsecase(cfg.MINIO_ENDPOINT, cfg.MINIO_ACCESS_KEY, cfg.MINIO_SECRET_KEY, cfg.MINIO_BUCKET_NAME, cfg.MINIO_USE_SSL)
	return &Services{
		ObjectStorageService: objectStorageService,
		ApplicationService:   services.NewApplicationService(basicRepos, *dependencyParser, objectStorageService),
		DepedenciesService:   services.NewDependenciesService(basicRepos, *dependencyParser, objectStorageService),
	}
}

type Services struct {
	// GithubApiService     usecase.GitHubAPIInterface     // GitHub API service
	// MessagingService     usecase.MessagingInterface     // Messaging service (e.g., Telegram)
	ObjectStorageService usecase.ObjectStorageInterface // Minio object storage service
	ApplicationService   services.ApplicationInterface  // Application management service
	DepedenciesService   services.DependenciesInterface // Scan service for dependency scanning
}

type Repositories struct {
	App              repository.ApplicationRepository       // Manages applications
	Depedency        repository.DependencyRepository        // Manages dependencies
	AppDepedency     repository.AppDependencyRepository     // App to Dependency mapping
	DepedencyVersion repository.DependencyVersionRepository // Versioning for dependencies
	Runtime          repository.RuntimeRepository           // Manages runtimes
	Framework        repository.FrameworkRepository         // Manages frameworks
	AuditTrail       repository.AuditTrailRepository        // Audit trail tracking
}
