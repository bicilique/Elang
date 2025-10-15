package model

// CommitDetail represents detailed information about a Git commit.
type CommitDetail struct {
	SHA       string       `json:"sha"`
	Author    CommitPerson `json:"author"`
	Committer CommitPerson `json:"committer"`
	Message   string       `json:"message"`
	Stats     CommitStats  `json:"stats"`
	Files     []CommitFile `json:"files"`
}

// CommitPerson represents the author or committer of a commit.
type CommitPerson struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

// CommitStats represents the statistics of a commit.
type CommitStats struct {
	Total     int `json:"total"`
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}

// CommitFile represents a file changed in a commit.
type CommitFile struct {
	Filename    string `json:"filename"`
	Status      string `json:"status"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Changes     int    `json:"changes"`
	BlobURL     string `json:"blob_url"`
	RawURL      string `json:"raw_url"`
	ContentsURL string `json:"contents_url"`
	Patch       string `json:"patch"`
}

// CompareCommitResult represents the result of a GitHub compare commits API call.
type CompareCommitResult struct {
	URL             string              `json:"url"`
	HTMLURL         string              `json:"html_url"`
	PermalinkURL    string              `json:"permalink_url"`
	DiffURL         string              `json:"diff_url"`
	PatchURL        string              `json:"patch_url"`
	BaseCommit      CommitSummary       `json:"base_commit"`
	MergeBaseCommit CommitSummary       `json:"merge_base_commit"`
	Status          string              `json:"status"`
	AheadBy         int                 `json:"ahead_by"`
	BehindBy        int                 `json:"behind_by"`
	TotalCommits    int                 `json:"total_commits"`
	Commits         []CommitSummary     `json:"commits"`
	Files           []CompareFileChange `json:"files"`
}

// CommitSummary represents a summary of a commit in the compare result.
type CommitSummary struct {
	SHA         string `json:"sha"`
	NodeID      string `json:"node_id"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CommentsURL string `json:"comments_url"`
}

// CompareFileChange represents a file changed in the compare result.
type CompareFileChange struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
	Patch     string `json:"patch"`
}
