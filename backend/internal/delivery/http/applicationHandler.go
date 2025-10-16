package http

import (
	"elang-backend/internal/model"
	"elang-backend/internal/model/responses"
	"elang-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	applicationService services.ApplicationInterface
}

func NewApplicationHandler(appService services.ApplicationInterface) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: appService,
	}
}

// Add methods to handle application-related requests
func (h *ApplicationHandler) AddApplication(c *gin.Context) {
	var req model.AddApplicationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
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
	result, err := h.applicationService.AddApplication(
		ctx,
		req.AppName,
		req.RuntimeType,
		req.Framework,
		req.Description,
		fileHeader.Filename,
		string(fileBytes),
	)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to add application: "+err.Error(), nil)
		return
	}

	responses.JSONSuccessResponse(c, 200, "application added successfully", result)
}

// AddApplicationDependency handles adding new dependencies to an existing application (batch supported)
func (h *ApplicationHandler) AddApplicationDependency(c *gin.Context) {
	var req struct {
		AppID        string                        `json:"app_id" binding:"required"`
		Dependencies []model.DependencyInfoRequest `json:"dependencies" binding:"required,dive,required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSONErrorResponse(c, 400, "invalid request: "+err.Error(), nil)
		return
	}
	ctx := c.Request.Context()
	resp, err := h.applicationService.AddApplicationDependency(ctx, req.AppID, req.Dependencies)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to add dependencies: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "dependencies processed", resp)
}

// UpdateApplicationDependency handles batch updates to application dependencies (version, status, GitHub URL)
func (h *ApplicationHandler) UpdateApplicationDependency(c *gin.Context) {
	var req model.UpdateApplicationDependencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSONErrorResponse(c, 400, "invalid request: "+err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.applicationService.UpdateApplicationDependency(ctx, req.AppID, &req)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to update dependencies: "+err.Error(), nil)
		return
	}

	responses.JSONSuccessResponse(c, 200, "dependencies updated", resp)
}

// RemoveApplicationDependency handles batch removal of dependencies from an application
func (h *ApplicationHandler) RemoveApplicationDependency(c *gin.Context) {
	var req struct {
		AppID         string   `json:"app_id" binding:"required"`
		DependencyIDs []string `json:"dependencies" binding:"required,dive,required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.JSONErrorResponse(c, 400, "invalid request: "+err.Error(), nil)
		return
	}
	ctx := c.Request.Context()
	result, err := h.applicationService.RemoveApplicationDependency(ctx, req.AppID, req.DependencyIDs)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to remove dependencies: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 201, "dependencies removed", result)
}

// ListApplicationDependency handles listing dependencies for a given application
func (h *ApplicationHandler) ListApplicationDependency(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "missing app_id parameter", nil)
		return
	}
	ctx := c.Request.Context()
	resp, err := h.applicationService.ListApplicationDependency(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to list dependencies: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "dependencies fetched", resp)
}

// RemoveApplication handles soft-deleting (inactivating) an application
func (h *ApplicationHandler) RemoveApplication(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "missing app_id parameter", nil)
		return
	}
	ctx := c.Request.Context()
	err := h.applicationService.RemoveApplication(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to remove application: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "application removed (inactivated)", nil)
}

// RecoverApplication handles re-activating an application
func (h *ApplicationHandler) RecoverApplication(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "missing app_id parameter", nil)
		return
	}
	ctx := c.Request.Context()
	err := h.applicationService.RecoverApplication(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to recover application: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "application recovered (activated)", nil)
}

// ListApplications handles listing all applications
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := h.applicationService.ListApplications(ctx)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to list applications: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "applications fetched", resp)
}

// GetApplicationStatus handles fetching the status of a single application
func (h *ApplicationHandler) GetApplicationStatus(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "missing app_id parameter", nil)
		return
	}
	ctx := c.Request.Context()
	resp, err := h.applicationService.GetApplicationStatus(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to get application status: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "application status fetched", resp)
}

// ScanApplication handles scanning an application's dependencies against OSV
func (h *ApplicationHandler) ScanApplication(c *gin.Context) {
	appUID := c.Param("app_id")
	if appUID == "" {
		responses.JSONErrorResponse(c, 400, "missing app_id parameter", nil)
		return
	}
	ctx := c.Request.Context()
	resp, err := h.applicationService.ScanApplicationDependencies(ctx, appUID)
	if err != nil {
		responses.JSONErrorResponse(c, 500, "failed to scan application: "+err.Error(), nil)
		return
	}
	responses.JSONSuccessResponse(c, 200, "application scan initiated", resp)
}
