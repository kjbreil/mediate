package shows

import (
	"time"
)

// DeletedMedia represents media that was previously in Plex but has been deleted
type DeletedMedia struct {
	ID               uint      `gorm:"primaryKey"`
	
	// Original Plex identifiers
	RatingKey        string    `gorm:"index;not null"`        // Original Plex rating key
	HistoryKey       string    `gorm:"index"`                 // Plex history key if available
	LibrarySectionID int       `gorm:"index"`                 // Original library section
	
	// Media information (preserved from last known state)
	Title            string    `gorm:"not null"`
	MediaType        string    `gorm:"not null"` // movie, show, episode, etc.
	Year             int
	Duration         time.Duration
	Summary          string    `gorm:"type:text"`
	
	// For episodes/shows
	ShowTitle        string    // Parent show title for episodes
	SeasonNumber     int       // Season number for episodes
	EpisodeNumber    int       // Episode number
	TvdbID           int       `gorm:"index"` // TVDB ID if available
	
	// Deletion tracking
	LastSeenAt       time.Time `gorm:"not null"` // When it was last accessible
	DeletedAt        time.Time `gorm:"not null"` // When deletion was detected
	DetectionMethod  string    `gorm:"not null"` // How deletion was detected
	
	// Preserved viewing statistics
	TotalViews       int       `gorm:"default:0"`
	TotalWatchTime   time.Duration
	LastViewedAt     *time.Time
	FirstViewedAt    *time.Time
	
	CreatedAt        time.Time
	UpdatedAt        time.Time
	
	// Related viewing sessions
	ViewingSessions  []DeletedMediaSession `gorm:"foreignKey:DeletedMediaID"`
}

// DeletedMediaSession represents viewing sessions for deleted media
type DeletedMediaSession struct {
	ID               uint      `gorm:"primaryKey"`
	DeletedMediaID   uint      `gorm:"index;not null"`
	DeletedMedia     *DeletedMedia `gorm:"foreignKey:DeletedMediaID"`
	
	// Original session data
	HistoryKey       string    `gorm:"index"`
	AccountID        int       `gorm:"index"`
	PlexUsername     string    `gorm:"index"`
	
	// Session details
	ViewedAt         time.Time `gorm:"not null"`
	Duration         time.Duration
	ProgressPercent  float64   `gorm:"default:0"`
	Completed        bool      `gorm:"default:false"`
	
	// Device information
	DeviceType       string
	DeviceName       string
	Platform         string
	Location         string    // Remote/LAN
	
	// Playback details
	Paused           bool      `gorm:"default:false"`
	Stopped          bool      `gorm:"default:false"`
	
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// DeletedMediaSummary provides aggregated statistics for deleted media
type DeletedMediaSummary struct {
	TotalItems       int                    `json:"total_items"`
	TotalViews       int                    `json:"total_views"`
	TotalWatchTime   time.Duration          `json:"total_watch_time"`
	DeletedRecently  int                    `json:"deleted_recently"` // Last 30 days
	ByMediaType      map[string]int         `json:"by_media_type"`
	ByLibrary        map[int]int            `json:"by_library"`
	TopDeletedShows  []DeletedMediaStats    `json:"top_deleted_shows"`
	RecentDeletions  []*DeletedMedia        `json:"recent_deletions"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

// DeletedMediaStats provides statistics for individual deleted media
type DeletedMediaStats struct {
	Title           string        `json:"title"`
	MediaType       string        `json:"media_type"`
	TotalViews      int           `json:"total_views"`
	TotalWatchTime  time.Duration `json:"total_watch_time"`
	DeletedAt       time.Time     `json:"deleted_at"`
	DaysActive      int           `json:"days_active"` // Days between first and last view
}

// PlexHistoryResponse represents the XML response from Plex viewing history API
type PlexHistoryResponse struct {
	MediaContainer PlexHistoryContainer `xml:"MediaContainer"`
}

type PlexHistoryContainer struct {
	Size               int                `xml:"size,attr"`
	AllowSync          int                `xml:"allowSync,attr"`
	Identifier         string             `xml:"identifier,attr"`
	MediaTagPrefix     string             `xml:"mediaTagPrefix,attr"`
	MediaTagVersion    int                `xml:"mediaTagVersion,attr"`
	Videos             []PlexHistoryVideo `xml:"Video"`
}

type PlexHistoryVideo struct {
	HistoryKey         string `xml:"historyKey,attr"`
	RatingKey          string `xml:"ratingKey,attr"`
	Key                string `xml:"key,attr"`
	ParentRatingKey    string `xml:"parentRatingKey,attr"`
	GrandparentRatingKey string `xml:"grandparentRatingKey,attr"`
	Title              string `xml:"title,attr"`
	GrandparentTitle   string `xml:"grandparentTitle,attr"`
	Type               string `xml:"type,attr"`
	Thumb              string `xml:"thumb,attr"`
	ParentThumb        string `xml:"parentThumb,attr"`
	GrandparentThumb   string `xml:"grandparentThumb,attr"`
	ViewedAt           int64  `xml:"viewedAt,attr"`
	AccountID          int    `xml:"accountID,attr"`
	DeviceID           int    `xml:"deviceID,attr"`
	LibrarySectionID   int    `xml:"librarySectionID,attr"`
	
	// Episode specific
	ParentIndex        int    `xml:"parentIndex,attr"`  // Season number
	Index              int    `xml:"index,attr"`        // Episode number
	Year               int    `xml:"year,attr"`
	Duration           int    `xml:"duration,attr"`     // In milliseconds
}

// IsOrphaned checks if a media item's rating key returns 404 from Plex
func (dm *DeletedMedia) IsOrphaned() bool {
	// This would be implemented to check if the rating key is still accessible
	// For now, we assume if it's in our deleted media table, it's orphaned
	return true
}

// GetViewingStats calculates viewing statistics for this deleted media
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