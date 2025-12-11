package shows

import (
	"fmt"
	"time"
)

// ViewingSession represents a single viewing session for an episode by a specific user.
type ViewingSession struct {
	ID            uint     `gorm:"primaryKey"`
	EpisodeTvdbID int      `gorm:"index"`
	Episode       *Episode `gorm:"foreignKey:EpisodeTvdbID;references:TvdbID"`

	// User information
	PlexUserID   int    `gorm:"index"` // Plex user ID
	PlexUsername string `gorm:"index"` // Plex username for easier querying

	// Session details
	StartedAt       time.Time
	EndedAt         *time.Time
	Duration        time.Duration // How long they actually watched
	ProgressPercent float64       // How far through the episode (0-100)
	Completed       bool          // Did they finish the episode

	// Context
	DeviceType string // TV, Mobile, Web, etc.
	DeviceName string // Specific device name
	Location   string // Home, Remote, etc.

	// Behavior
	Paused     bool // Did they pause during viewing
	PauseCount int  // How many times they paused
	Skipped    bool // Did they skip/fast forward
	SkipCount  int  // How many times they skipped

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ViewingSessions is a slice of ViewingSession.
type ViewingSessions []*ViewingSession

// ByUser filters viewing sessions by username.
func (vs ViewingSessions) ByUser(username string) ViewingSessions {
	var result ViewingSessions
	for _, session := range vs {
		if session.PlexUsername == username {
			result = append(result, session)
		}
	}
	return result
}

// Completed filters to only completed viewing sessions.
func (vs ViewingSessions) Completed() ViewingSessions {
	var result ViewingSessions
	for _, session := range vs {
		if session.Completed {
			result = append(result, session)
		}
	}
	return result
}

// InTimeframe filters sessions within a time period.
func (vs ViewingSessions) InTimeframe(since time.Time) ViewingSessions {
	var result ViewingSessions
	for _, session := range vs {
		if session.StartedAt.After(since) {
			result = append(result, session)
		}
	}
	return result
}

// TotalWatchTime calculates total watch time across all sessions.
func (vs ViewingSessions) TotalWatchTime() time.Duration {
	var total time.Duration
	for _, session := range vs {
		total += session.Duration
	}
	return total
}

// UserStats calculates viewing statistics per user.
type UserStats struct {
	Username          string
	TotalSessions     int
	TotalWatchTime    time.Duration
	CompletedEpisodes int
	AverageCompletion float64
	FavoriteDevice    string
	ViewingPeak       string // Most active viewing hour
}

// GetUserStats calculates statistics for each user.
func (vs ViewingSessions) GetUserStats() map[string]*UserStats {
	userMap := make(map[string]*UserStats)
	deviceCounts := make(map[string]map[string]int) // user -> device -> count
	hourCounts := make(map[string]map[int]int)      // user -> hour -> count

	for _, session := range vs {
		username := session.PlexUsername

		// Initialize user if not exists
		if _, exists := userMap[username]; !exists {
			userMap[username] = &UserStats{
				Username: username,
			}
			deviceCounts[username] = make(map[string]int)
			hourCounts[username] = make(map[int]int)
		}

		stats := userMap[username]
		stats.TotalSessions++
		stats.TotalWatchTime += session.Duration

		if session.Completed {
			stats.CompletedEpisodes++
		}

		// Track device usage
		deviceCounts[username][session.DeviceType]++

		// Track viewing hours
		hour := session.StartedAt.Hour()
		hourCounts[username][hour]++
	}

	// Calculate averages and favorites
	for username, stats := range userMap {
		if stats.TotalSessions > 0 {
			stats.AverageCompletion = float64(stats.CompletedEpisodes) / float64(stats.TotalSessions) * 100
		}

		// Find favorite device
		maxCount := 0
		for device, count := range deviceCounts[username] {
			if count > maxCount {
				maxCount = count
				stats.FavoriteDevice = device
			}
		}

		// Find peak viewing hour
		maxHourCount := 0
		peakHour := 0
		for hour, count := range hourCounts[username] {
			if count > maxHourCount {
				maxHourCount = count
				peakHour = hour
			}
		}
		stats.ViewingPeak = formatHour(peakHour)
	}

	return userMap
}

func formatHour(hour int) string {
	if hour == 0 {
		return "12:00 AM"
	}
	if hour < 12 {
		return fmt.Sprintf("%d:00 AM", hour)
	}
	if hour == 12 {
		return "12:00 PM"
	}
	return fmt.Sprintf("%d:00 PM", hour-12)
}
