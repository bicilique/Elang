package http

import (
	"elang-backend/internal/model/responses"
	"elang-backend/internal/services"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// DependenciesHandler handles HTTP requests related to dependencies
type DependenciesHandler struct {
	dependencyService services.DependenciesInterface
}

func NewDependenciesHandler(depService services.DependenciesInterface) *DependenciesHandler {
	return &DependenciesHandler{
		dependencyService: depService,
	}
}

// Add methods to handle scan-related requests
// For example, a method to initiate a scan
func (h *DependenciesHandler) ScanApplication(c *gin.Context) {

	var req struct {
		AppName     string `form:"name" binding:"required"`
		Runtime     string `form:"runtime" binding:"required"`
		Version     string `form:"version"`
		Description string `form:"description,omitempty"`
	}

	if err := c.ShouldBind(&req); err != nil {
		slog.Error("Failed to bind request", "error", err)
		responses.JSONErrorResponse(c, 400, err.Error(), nil)
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		responses.JSONErrorResponse(c, 400, "failed to get file: "+err.Error(), nil)
		return
	}
	defer file.Close()

	fileBytes := make([]byte, fileHeader.Size)
	_, err = file.Read(fileBytes)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to read file: "+err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	result, err := h.dependencyService.ScanDependencies(
		ctx,
		req.AppName,
		req.Runtime,
		req.Version,
		req.Description,
		fileHeader.Filename,
		string(fileBytes),
	)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to scan application: "+err.Error(), nil)
		return
	}

	responses.JSONSuccessResponse(c, 200, "application scanned successfully", result)
}

// GetSBOM retrieves the SBOM for a given application and SBOM ID
func (h *DependenciesHandler) GetSBOM(c *gin.Context) {
	sbomId := c.Param("sbom_id")
	appName := c.Param("app_name")
	if sbomId == "" || appName == "" {
		responses.JSONErrorResponse(c, 400, "sbom_id and app_name are required", nil)
		return
	}

	ctx := c.Request.Context()
	sbomData, err := h.dependencyService.GetSBOMById(ctx, appName, sbomId)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to get SBOM: "+err.Error(), nil)
		return
	}

	responses.JSONSuccessResponse(c, 200, "SBOM retrieved successfully", sbomData)
}

// MonitorApplicationDepedencies monitors application dependencies for changes
func (h *DependenciesHandler) MonitorApplicationDepedencies(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "app_id is required", nil)
		return
	}

	ctx := c.Request.Context()
	err := h.dependencyService.StartMonitoringApplication(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to monitor application dependencies: "+err.Error(), nil)
		return
	}

	message := fmt.Sprintf("start monitoring application %s successfully", appUID)
	responses.JSONSuccessResponse(c, 200, message, nil)
}

// StopMonitoringApplication stops monitoring an application's dependencies
func (h *DependenciesHandler) StopMonitoringApplication(c *gin.Context) {

	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "app_uid is required", nil)
		return
	}

	ctx := c.Request.Context()
	err := h.dependencyService.StopMonitoringApplication(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to stop monitoring application: "+err.Error(), nil)
		return
	}

	message := fmt.Sprintf("stopped monitoring application %s successfully", appUID)
	responses.JSONSuccessResponse(c, 200, message, nil)
}

// GetAllApplicationsStatus retrieves the monitoring status of all applications
func (h *DependenciesHandler) GetAllApplicationsStatus(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "app_uid is required", nil)
		return
	}
	ctx := c.Request.Context()
	result, err := h.dependencyService.GetMonitoringStatus(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to get applications status: "+err.Error(), nil)
		return
	}

	responses.JSONSuccessResponse(c, 200, "applications status retrieved successfully", result)
}
