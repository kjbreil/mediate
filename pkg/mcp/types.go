package mcp

import "time"

// ViewingAnalysis represents the result of viewing habits analysis.
type ViewingAnalysis struct {
	Timeframe    string                 `json:"timeframe"`
	AnalysisType string                 `json:"analysis_type"`
	Summary      string                 `json:"summary"`
	Data         map[string]interface{} `json:"data"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

// Recommendation represents a media recommendation.
type Recommendation struct {
	Title       string    `json:"title"`
	Type        string    `json:"type"` // "show" or "movie"
	TvdbID      int       `json:"tvdb_id,omitempty"`
	Score       float64   `json:"score"`
	Reason      string    `json:"reason"`
	Genre       []string  `json:"genre,omitempty"`
	Year        int       `json:"year,omitempty"`
	Description string    `json:"description,omitempty"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
}

// SearchResult represents a search result from various sources.
type SearchResult struct {
	Title       string   `json:"title"`
	Type        string   `json:"type"`   // "show" or "movie"
	Source      string   `json:"source"` // "plex", "sonarr", "radarr"
	TvdbID      int      `json:"tvdb_id,omitempty"`
	Year        int      `json:"year,omitempty"`
	Genre       []string `json:"genre,omitempty"`
	Description string   `json:"description,omitempty"`
	Available   bool     `json:"available"`
	Monitored   bool     `json:"monitored,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// SearchResults contains results from multiple sources.
type SearchResults struct {
	Query      string          `json:"query"`
	TotalCount int             `json:"total_count"`
	Results    []*SearchResult `json:"results"`
	Sources    []string        `json:"sources"`
	SearchTime time.Duration   `json:"search_time"`
}

// DownloadItem represents an item to be added to downloads.
type DownloadItem struct {
	Title          string `json:"title"`
	Type           string `json:"type"` // "show" or "movie"
	TvdbID         int    `json:"tvdb_id,omitempty"`
	Monitor        bool   `json:"monitor"`
	QualityProfile string `json:"quality_profile,omitempty"`
}

// DownloadRequest represents a request to add items to download.
type DownloadRequest struct {
	Items          []*DownloadItem `json:"items"`
	QualityProfile string          `json:"quality_profile,omitempty"`
}

// DownloadResponse represents the response after adding downloads.
type DownloadResponse struct {
	Success   []*DownloadItem `json:"success"`
	Failed    []*DownloadItem `json:"failed"`
	Errors    []string        `json:"errors,omitempty"`
	Message   string          `json:"message"`
	Timestamp time.Time       `json:"timestamp"`
}

// SystemStatus represents the current system status.
type SystemStatus struct {
	Services    map[string]ServiceStatus `json:"services"`
	Database    DatabaseStatus           `json:"database"`
	Jobs        []JobStatus              `json:"jobs"`
	Statistics  SystemStatistics         `json:"statistics"`
	Health      string                   `json:"health"` // "healthy", "degraded", "unhealthy"
	LastUpdated time.Time                `json:"last_updated"`
}

// ServiceStatus represents the status of an external service.
type ServiceStatus struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Status    string    `json:"status"` // "connected", "disconnected", "error"
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
	Version   string    `json:"version,omitempty"`
}

// DatabaseStatus represents database health.
type DatabaseStatus struct {
	Connected    bool      `json:"connected"`
	ShowCount    int       `json:"show_count"`
	EpisodeCount int       `json:"episode_count"`
	LastUpdated  time.Time `json:"last_updated"`
	Size         string    `json:"size,omitempty"`
}

// JobStatus represents the status of background jobs.
type JobStatus struct {
	Name     string     `json:"name"`
	Status   string     `json:"status"` // "running", "stopped", "error"
	LastRun  *time.Time `json:"last_run,omitempty"`
	NextRun  *time.Time `json:"next_run,omitempty"`
	Interval string     `json:"interval,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// SystemStatistics represents system usage statistics.
type SystemStatistics struct {
	TotalShows       int `json:"total_shows"`
	TotalEpisodes    int `json:"total_episodes"`
	MonitoredShows   int `json:"monitored_shows"`
	WatchedEpisodes  int `json:"watched_episodes"`
	PendingDownloads int `json:"pending_downloads"`
}
