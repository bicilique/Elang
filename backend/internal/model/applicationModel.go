package model

type AddApplicationRequest struct {
	AppName     string `form:"app_name" binding:"required"`
	RuntimeType string `form:"runtime_type" binding:"required"`
	Framework   string `form:"framework" binding:"required"`
	Description string `form:"description"`
	// File will be handled as multipart.FileHeader in handler, not here
}

type ListApplicationDependencyResponse struct {
	AppID        string                        `json:"app_id"`
	AppName      string                        `json:"app_name"`
	Dependencies []ApplicationDependencyDetail `json:"dependencies"`
	Message      string                        `json:"message"`
}

type ApplicationDependencyDetail struct {
	DependencyID  string  `json:"dependency_id"`
	Name          string  `json:"name"`
	Owner         string  `json:"owner"`
	Repo          string  `json:"repo"`
	UsedVersion   string  `json:"used_version"`
	IsMonitored   bool    `json:"is_monitored"`
	RepositoryURL string  `json:"repository_url"`
	LastTag       *string `json:"latest_tag,omitempty"`
	DefaultBranch *string `json:"default_branch,omitempty"`
}

type AddApplicationResponse struct {
	AppID           string      `json:"app_id"`
	AppName         string      `json:"app_name"`
	RuntimeType     string      `json:"runtime_type"`
	Framework       string      `json:"framework"`
	Description     string      `json:"description"`
	Status          string      `json:"status"`
	DependencyParse interface{} `json:"dependency_parse"`
	Message         string      `json:"message"`
}

// ListApplicationsResponse is a top-level response
type ListApplicationsResponse struct {
	Applications []ApplicationSummary `json:"applications"`
	Message      string               `json:"message"`
}

type ApplicationSummary struct {
	AppID       string `json:"app_id"`
	AppName     string `json:"app_name"`
	RuntimeType string `json:"runtime_type"`
	Framework   string `json:"framework"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type ApplicationStatus struct {
	AppID           string `json:"app_id"`
	AppName         string `json:"app_name"`
	Status          string `json:"status"`
	DependencyCount int    `json:"dependency_count"`
	LastUpdated     string `json:"last_updated,omitempty"`
}
