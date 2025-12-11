package shows

import (
	"time"
)

// DeletedMedia represents media that was previously in Plex but has been deleted.
type DeletedMedia struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Original Plex identifiers
	RatingKey        string `gorm:"index;not null" json:"rating_key"`         // Original Plex rating key
	HistoryKey       string `gorm:"index"          json:"history_key"`        // Plex history key if available
	LibrarySectionID int    `gorm:"index"          json:"library_section_id"` // Original library section

	// Media information (preserved from last known state)
	Title     string        `gorm:"not null"  json:"title"`
	MediaType string        `gorm:"not null"  json:"media_type"` // movie, show, episode, etc.
	Year      int           `                 json:"year"`
	Duration  time.Duration `                 json:"duration"`
	Summary   string        `gorm:"type:text" json:"summary"`

	// For episodes/shows
	ShowTitle     string `json:"show_title"`                  // Parent show title for episodes
	SeasonNumber  int    `json:"season_number"`               // Season number for episodes
	EpisodeNumber int    `json:"episode_number"`              // Episode number
	TvdbID        int    `json:"tvdb_id"        gorm:"index"` // TVDB ID if available

	// Deletion tracking
	LastSeenAt      time.Time `gorm:"not null" json:"last_seen_at"`     // When it was last accessible
	DeletedAt       time.Time `gorm:"not null" json:"deleted_at"`       // When deletion was detected
	DetectionMethod string    `gorm:"not null" json:"detection_method"` // How deletion was detected

	// Preserved viewing statistics
	TotalViews     int           `gorm:"default:0" json:"total_views"`
	TotalWatchTime time.Duration `                 json:"total_watch_time"`
	LastViewedAt   *time.Time    `                 json:"last_viewed_at"`
	FirstViewedAt  *time.Time    `                 json:"first_viewed_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Related viewing sessions
	ViewingSessions []DeletedMediaSession `gorm:"foreignKey:DeletedMediaID" json:"viewing_sessions,omitempty"`
}

// DeletedMediaSession represents viewing sessions for deleted media.
type DeletedMediaSession struct {
	ID             uint          `gorm:"primaryKey"                json:"id"`
	DeletedMediaID uint          `gorm:"index;not null"            json:"deleted_media_id"`
	DeletedMedia   *DeletedMedia `gorm:"foreignKey:DeletedMediaID" json:"deleted_media,omitempty"`

	// Original session data
	HistoryKey   string `gorm:"index" json:"history_key"`
	AccountID    int    `gorm:"index" json:"account_id"`
	PlexUsername string `gorm:"index" json:"plex_username"`

	// Session details
	ViewedAt        time.Time     `gorm:"not null"      json:"viewed_at"`
	Duration        time.Duration `                     json:"duration"`
	ProgressPercent float64       `gorm:"default:0"     json:"progress_percent"`
	Completed       bool          `gorm:"default:false" json:"completed"`

	// Device information
	DeviceType string `json:"device_type"`
	DeviceName string `json:"device_name"`
	Platform   string `json:"platform"`
	Location   string `json:"location"` // Remote/LAN

	// Playback details
	Paused  bool `gorm:"default:false" json:"paused"`
	Stopped bool `gorm:"default:false" json:"stopped"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeletedMediaSummary provides aggregated statistics for deleted media.
type DeletedMediaSummary struct {
	TotalItems      int                 `json:"total_items"`
	TotalViews      int                 `json:"total_views"`
	TotalWatchTime  time.Duration       `json:"total_watch_time"`
	DeletedRecently int                 `json:"deleted_recently"` // Last 30 days
	ByMediaType     map[string]int      `json:"by_media_type"`
	ByLibrary       map[int]int         `json:"by_library"`
	TopDeletedShows []DeletedMediaStats `json:"top_deleted_shows"`
	RecentDeletions []*DeletedMedia     `json:"recent_deletions"`
	GeneratedAt     time.Time           `json:"generated_at"`
}

// DeletedMediaStats provides statistics for individual deleted media.
type DeletedMediaStats struct {
	Title          string        `json:"title"`
	MediaType      string        `json:"media_type"`
	TotalViews     int           `json:"total_views"`
	TotalWatchTime time.Duration `json:"total_watch_time"`
	DeletedAt      time.Time     `json:"deleted_at"`
	DaysActive     int           `json:"days_active"` // Days between first and last view
}

// PlexHistoryResponse represents the XML response from Plex viewing history API.
type PlexHistoryResponse struct {
	MediaContainer PlexHistoryContainer `xml:"MediaContainer"`
}

type PlexHistoryContainer struct {
	Size            int                `xml:"size,attr"`
	AllowSync       int                `xml:"allowSync,attr"`
	Identifier      string             `xml:"identifier,attr"`
	MediaTagPrefix  string             `xml:"mediaTagPrefix,attr"`
	MediaTagVersion int                `xml:"mediaTagVersion,attr"`
	Videos          []PlexHistoryVideo `xml:"Video"`
}

type PlexHistoryVideo struct {
	HistoryKey           string `xml:"historyKey,attr"`
	RatingKey            string `xml:"ratingKey,attr"`
	Key                  string `xml:"key,attr"`
	ParentRatingKey      string `xml:"parentRatingKey,attr"`
	GrandparentRatingKey string `xml:"grandparentRatingKey,attr"`
	Title                string `xml:"title,attr"`
	GrandparentTitle     string `xml:"grandparentTitle,attr"`
	Type                 string `xml:"type,attr"`
	Thumb                string `xml:"thumb,attr"`
	ParentThumb          string `xml:"parentThumb,attr"`
	GrandparentThumb     string `xml:"grandparentThumb,attr"`
	ViewedAt             int64  `xml:"viewedAt,attr"`
	AccountID            int    `xml:"accountID,attr"`
	DeviceID             int    `xml:"deviceID,attr"`
	LibrarySectionID     int    `xml:"librarySectionID,attr"`

	// Episode specific
	ParentIndex int `xml:"parentIndex,attr"` // Season number
	Index       int `xml:"index,attr"`       // Episode number
	Year        int `xml:"year,attr"`
	Duration    int `xml:"duration,attr"` // In milliseconds
}

// IsOrphaned checks if a media item's rating key returns 404 from Plex.
func (dm *DeletedMedia) IsOrphaned() bool {
	// This would be implemented to check if the rating key is still accessible
	// For now, we assume if it's in our deleted media table, it's orphaned
	return true
}

// GetViewingStats calculates viewing statistics for this deleted media.
func (dm *DeletedMedia) GetViewingStats() *DeletedMediaStats {
	totalWatchTime := time.Duration(0)
	for _, session := range dm.ViewingSessions {
		totalWatchTime += session.Duration
	}

	var daysActive int
	if dm.FirstViewedAt != nil && dm.LastViewedAt != nil {
		daysActive = int(dm.LastViewedAt.Sub(*dm.FirstViewedAt).Hours() / 24)
	}

	return &DeletedMediaStats{
		Title:          dm.Title,
		MediaType:      dm.MediaType,
		TotalViews:     dm.TotalViews,
		TotalWatchTime: totalWatchTime,
		DeletedAt:      dm.DeletedAt,
		DaysActive:     daysActive,
	}
}
