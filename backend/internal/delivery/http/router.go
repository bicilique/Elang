package http

import "github.com/gin-gonic/gin"

// RouteConfig holds all handlers and router configuration
// Simplified to only use ApplicationHandler and MonitoringHandler for maintainability
type RouteConfig struct {
	Router              *gin.Engine
	AppHandler          ApplicationHandler
	DependenciesHandler DependenciesHandler
}

func (c *RouteConfig) Setup() {
	// Apply global middleware
	c.Router.Use(gin.Logger())
	c.Router.Use(gin.Recovery())
	c.Router.Use(corsMiddleware()) // Add CORS support

	// Health check endpoint (no auth required)
	c.Router.GET("/health", healthCheck)

	// Main API group
	api := c.Router.Group("/api")
	{
		// Application Management APIs (CRUD and monitoring control)
		c.setupApplicationRoutes(api)

		// Dependencies related routes
		c.setupDependenciesRoute(api)
	}
}

func (c *RouteConfig) setupApplicationRoutes(api *gin.RouterGroup) {
	apps := api.Group("/applications")
	{
		// Application CRUD operations
		apps.POST("/add", c.AppHandler.AddApplication)
		apps.GET("/list", c.AppHandler.ListApplications)
		apps.GET("/:app_id/list", c.AppHandler.ListApplicationDependency)
		apps.PATCH("/:app_id/recover", c.AppHandler.RecoverApplication)
		apps.DELETE("/:app_id/remove", c.AppHandler.RemoveApplication)

		// Dependency management
		apps.POST("/add/dependencies", c.AppHandler.AddApplicationDependency) // Fixed typo
		apps.PATCH("/update/dependencies", c.AppHandler.UpdateApplicationDependency)
		apps.PATCH("/remove/dependencies", c.AppHandler.RemoveApplicationDependency) // Fixed typo

		// Status and monitoring
		// apps.GET("/status", c.AppHandler.GetAllApplicationsStatus)
		apps.GET("/:app_id/status", c.AppHandler.GetApplicationStatus)
		apps.GET("/:app_id/scan", c.AppHandler.ScanApplication) // Scan Application dependencies by Checking depedency version in OSV
	}
}

func (c *RouteConfig) setupDependenciesRoute(api *gin.RouterGroup) {
	scan := api.Group("/scan")
	{
		// Scan Application dependencies by Checking depedency version in OSV
		scan.POST("/dependencies", c.DependenciesHandler.ScanApplication)
		// Get SBOM by its ID
		scan.GET("/dependencies/:app_name/:sbom_id", c.DependenciesHandler.GetSBOM)

		// Start monitoring application dependencies for changes
		scan.POST("/:app_id/start", c.DependenciesHandler.MonitorApplicationDepedencies)

		// Stop monitoring application dependencies
		scan.POST("/:app_id/stop", c.DependenciesHandler.StopMonitoringApplication)

		// Get monitoring status of all applications
		scan.GET("/:app_id/status", c.DependenciesHandler.GetAllApplicationsStatus)
	}
}

// corsMiddleware provides CORS support for cross-origin requests
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// healthCheck provides a simple health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "elang-v1",
		"version": "1.0",
		"features": gin.H{
			"enhanced_security_detection": true,
			"progressive_monitoring":      true,
			"tag_monitoring":              true,
			"commit_monitoring":           true,
		},
	})
}
