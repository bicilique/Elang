package model

type ScanSummary struct {
	TotalDependencies    int `json:"total_dependencies"`
	TotalVulnerabilities int `json:"total_vulnerabilities"`
	Critical             int `json:"critical"`
	High                 int `json:"high"`
	Medium               int `json:"medium"`
	Low                  int `json:"low"`
	Ignored              int `json:"ignored"`
	None                 int `json:"none"`
}

type ScanPolicy struct {
	FailOn []string `json:"fail_on"`
	Status string   `json:"status"`
	Reason string   `json:"reason"`
}

type ScanArtifacts struct {
	VulnerabilityReport string `json:"vulnerability_report"`
	SBOM                string `json:"sbom"`
}

type ScanFinding struct {
	Dependency       string   `json:"dependency"`
	Version          string   `json:"version"`
	Severity         string   `json:"severity"`
	VulnerabilityIDs []string `json:"vulnerability_ids"`
	Recommendation   string   `json:"recommendation"`
}

type ScanApplicationResult struct {
	AppID      string        `json:"app_id"`
	AppName    string        `json:"app_name"`
	ScanStatus string        `json:"scan_status"`
	Summary    ScanSummary   `json:"summary"`
	Policies   ScanPolicy    `json:"policies"`
	Artifacts  ScanArtifacts `json:"artifacts"`
	Findings   []ScanFinding `json:"findings"`
}

type DependencyInfoRequest struct {
	Name          string `json:"name" binding:"required"`
	Owner         string `json:"owner"`
	Repo          string `json:"repo"`
	Version       string `json:"version" binding:"required"`
	RepositoryURL string `json:"repository_url"`
	IsGitHubRepo  bool   `json:"is_github_repo" default:"true"`
}

type UpdateApplicationDependencyRequest struct {
	AppID   string                 `json:"app_id"`
	Updates []UpdateDependencyItem `json:"dependencies"`
}

type UpdateDependencyItem struct {
	DependencyID  string `json:"dependency_id"`
	Owner         string `json:"owner,omitempty"`          // Optional
	Repo          string `json:"repo,omitempty"`           // Optional
	UsedVersion   string `json:"used_version"`             // Required
	RepositoryURL string `json:"repository_url,omitempty"` // Optional
}

type UpdateApplicationDependencyResponse struct {
	AppID   string   `json:"app_id"`
	Updated []string `json:"updated"`
	Failed  []string `json:"failed"`
	Message string   `json:"message"`
}
