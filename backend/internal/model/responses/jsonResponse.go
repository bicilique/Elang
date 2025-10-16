package responses

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ListRunningMonitoringsResponse represents the response for listing running monitoring jobs
type ListRunningMonitoringsResponse struct {
	RunningJobs  []RunningMonitoringJob `json:"running_jobs"`
	TotalJobs    int                    `json:"total_jobs"`
	SystemStatus SystemStatus           `json:"system_status"`
}

// RunningMonitoringJob represents a currently running monitoring job
type RunningMonitoringJob struct {
	JobID                 string    `json:"job_id"`
	AppID                 string    `json:"app_id"`
	AppName               string    `json:"app_name"`
	Status                string    `json:"status"`
	StartTime             time.Time `json:"start_time"`
	LastActivity          time.Time `json:"last_activity"`
	DependenciesTotal     int       `json:"dependencies_total"`
	DependenciesProcessed int       `json:"dependencies_processed"`
	SecurityDetections    int       `json:"security_detections"`
	Progress              float64   `json:"progress"`
}

// SystemStatus represents the overall system status
type SystemStatus struct {
	MonitoringEnabled   bool `json:"monitoring_enabled"`
	ScheduledMonitoring bool `json:"scheduled_monitoring"`
	EnhancedDetection   bool `json:"enhanced_detection"`
	AuditTrailEnabled   bool `json:"audit_trail_enabled"`
	AIIntegrationReady  bool `json:"ai_integration_ready"`
}

func JSONSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// JSONSuccessResponseWithCount sends a response with a count field which is the number of items in the response.
// The data can be any type of data, but it must be serializable to JSON.
func JSONSuccessResponseWithCount(c *gin.Context, statusCode int, message string, count int64, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"count":   count,
		"data":    data,
	})
}

// JSONErrorResponse sends a JSON response with a success status set to false.
// The response will have three fields: "success", "message" and "error".
// The "success" field will always be false.
// The "message" field will contain the message passed to the function.
// The "error" field will contain the data passed to the function, it can be any type of data.
func JSONErrorResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
		"error":   data,
	})
	c.Abort()
}
