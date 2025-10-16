package http

import "github.com/gin-gonic/gin"

type RouteConfig struct {
	Router              *gin.Engine
	AppHandler          ApplicationHandler
	DependenciesHandler DependenciesHandler
}

// Setup initializes all routes and applies global middleware.
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

// setupApplicationRoutes registers application management and monitoring endpoints under /api/applications.
func (c *RouteConfig) setupApplicationRoutes(api *gin.RouterGroup) {
	apps := api.Group("/applications")
	{
		// Application CRUD operations
		apps.POST("/add", c.AppHandler.AddApplication)                    // Add new application
		apps.GET("/list", c.AppHandler.ListApplications)                  // List all applications
		apps.GET("/:app_id/list", c.AppHandler.ListApplicationDependency) // List dependencies for an application
		apps.PATCH("/:app_id/recover", c.AppHandler.RecoverApplication)   // Recover a deleted application
		apps.DELETE("/:app_id/remove", c.AppHandler.RemoveApplication)    // Remove an application

		// Dependency management for applications
		apps.POST("/add/dependencies", c.AppHandler.AddApplicationDependency)        // Add dependencies to an application
		apps.PATCH("/update/dependencies", c.AppHandler.UpdateApplicationDependency) // Update application dependencies
		apps.PATCH("/remove/dependencies", c.AppHandler.RemoveApplicationDependency) // Remove dependencies from an application

		// Monitoring control
		apps.GET("/:app_id/status", c.AppHandler.GetApplicationStatus) // Get application status
		apps.GET("/:app_id/scan", c.AppHandler.ScanApplication)        // Scan application dependencies (OSV)
	}
}

// setupDependenciesRoute registers dependency scanning and monitoring endpoints under /api/scan.
func (c *RouteConfig) setupDependenciesRoute(api *gin.RouterGroup) {
	scan := api.Group("/scan")
	{
		// Scan application dependencies (OSV)
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

// corsMiddleware provides CORS support for cross-origin requests.
// Allows all origins and common HTTP methods/headers.
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

// healthCheck provides a simple health check endpoint.
// Returns service status and enabled features.
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
