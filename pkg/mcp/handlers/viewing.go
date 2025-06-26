package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/kjbreil/mediate/pkg/mcp"
	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/mark3labs/mcp-go/mcp"
)

// ViewingHandler handles viewing-related MCP operations
type ViewingHandler struct {
	mediate *mediate.Mediate
	logger  *slog.Logger
}

// NewViewingHandler creates a new viewing handler
func NewViewingHandler(m *mediate.Mediate, logger *slog.Logger) *ViewingHandler {
	return &ViewingHandler{
		mediate: m,
		logger:  logger,
	}
}

// AnalyzeViewingHabits analyzes user viewing patterns
func (h *ViewingHandler) AnalyzeViewingHabits(ctx context.Context, timeframe string, analysisType string) (*mcp.ViewingAnalysis, error) {
	h.logger.Info("Analyzing viewing habits", "timeframe", timeframe, "analysis_type", analysisType)

	analysis := &mcp.ViewingAnalysis{
		Timeframe:    timeframe,
		AnalysisType: analysisType,
		GeneratedAt:  time.Now(),
		Data:         make(map[string]interface{}),
	}

	shows := h.mediate.GetShows()
	if shows == nil {
		return nil, fmt.Errorf("failed to get shows from database")
	}

	switch analysisType {
	case "genres":
		analysis.Summary, analysis.Data = h.analyzeGenres(shows, timeframe)
	case "shows":
		analysis.Summary, analysis.Data = h.analyzeShows(shows, timeframe)
	case "patterns":
		analysis.Summary, analysis.Data = h.analyzePatterns(shows, timeframe)
	case "completion_rate":
		analysis.Summary, analysis.Data = h.analyzeCompletionRate(shows, timeframe)
	default:
		return nil, fmt.Errorf("unknown analysis type: %s", analysisType)
	}

	return analysis, nil
}

// analyzeGenres analyzes viewing patterns by genre
func (h *ViewingHandler) analyzeGenres(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	genreCount := make(map[string]int)
	totalEpisodes := 0

	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Watched && h.isInTimeframe(episode.LastViewedAt, timeframe) {
				// For now, we'll use basic genre categorization
				// In a real implementation, you'd get genre data from Plex/TVDB
				genres := h.getShowGenres(show.Title)
				for _, genre := range genres {
					genreCount[genre]++
				}
				totalEpisodes++
			}
		}
	}

	summary := fmt.Sprintf("Watched %d episodes across %d genres in the last %s", 
		totalEpisodes, len(genreCount), timeframe)

	data := map[string]interface{}{
		"total_episodes": totalEpisodes,
		"genre_breakdown": genreCount,
		"top_genre": h.getTopGenre(genreCount),
	}

	return summary, data
}

// analyzeShows analyzes viewing patterns by show
func (h *ViewingHandler) analyzeShows(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	showStats := make(map[string]map[string]interface{})
	totalWatched := 0

	for _, show := range *shows {
		watchedCount := 0
		totalCount := len(show.Episodes)
		
		for _, episode := range show.Episodes {
			if episode.Watched && h.isInTimeframe(episode.LastViewedAt, timeframe) {
				watchedCount++
				totalWatched++
			}
		}

		if watchedCount > 0 {
			completionRate := float64(watchedCount) / float64(totalCount) * 100
			showStats[show.Title] = map[string]interface{}{
				"watched_episodes": watchedCount,
				"total_episodes":   totalCount,
				"completion_rate":  completionRate,
				"status":          show.Status(),
			}
		}
	}

	summary := fmt.Sprintf("Watched %d episodes across %d shows in the last %s", 
		totalWatched, len(showStats), timeframe)

	data := map[string]interface{}{
		"total_watched": totalWatched,
		"shows":        showStats,
		"most_watched": h.getMostWatchedShow(showStats),
	}

	return summary, data
}

// analyzePatterns analyzes viewing time patterns
func (h *ViewingHandler) analyzePatterns(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	hourCounts := make(map[int]int)
	dayCounts := make(map[string]int)
	totalSessions := 0

	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Watched && episode.LastViewedAt != nil && 
			   h.isInTimeframe(episode.LastViewedAt, timeframe) {
				
				viewTime := *episode.LastViewedAt
				hour := viewTime.Hour()
				day := viewTime.Weekday().String()
				
				hourCounts[hour]++
				dayCounts[day]++
				totalSessions++
			}
		}
	}

	peakHour := h.getPeakHour(hourCounts)
	peakDay := h.getPeakDay(dayCounts)

	summary := fmt.Sprintf("Most active viewing time: %s at %d:00, with %d total sessions", 
		peakDay, peakHour, totalSessions)

	data := map[string]interface{}{
		"total_sessions":   totalSessions,
		"hourly_breakdown": hourCounts,
		"daily_breakdown":  dayCounts,
		"peak_hour":       peakHour,
		"peak_day":        peakDay,
	}

	return summary, data
}

// analyzeCompletionRate analyzes show completion rates
func (h *ViewingHandler) analyzeCompletionRate(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	completedShows := 0
	inProgressShows := 0
	notStartedShows := 0
	totalShows := len(*shows)

	completionRates := make([]float64, 0)

	for _, show := range *shows {
		watchedCount := 0
		totalCount := len(show.Episodes)

		for _, episode := range show.Episodes {
			if episode.Watched {
				watchedCount++
			}
		}

		completionRate := float64(watchedCount) / float64(totalCount) * 100
		completionRates = append(completionRates, completionRate)

		if completionRate == 100 {
			completedShows++
		} else if completionRate > 0 {
			inProgressShows++
		} else {
			notStartedShows++
		}
	}

	avgCompletion := h.calculateAverage(completionRates)

	summary := fmt.Sprintf("Average completion rate: %.1f%% (%d completed, %d in progress, %d not started)", 
		avgCompletion, completedShows, inProgressShows, notStartedShows)

	data := map[string]interface{}{
		"total_shows":        totalShows,
		"completed_shows":    completedShows,
		"in_progress_shows":  inProgressShows,
		"not_started_shows":  notStartedShows,
		"average_completion": avgCompletion,
	}

	return summary, data
}

// Helper functions

func (h *ViewingHandler) isInTimeframe(timestamp *time.Time, timeframe string) bool {
	if timestamp == nil {
		return false
	}

	now := time.Now()
	switch timeframe {
	case "week":
		return timestamp.After(now.AddDate(0, 0, -7))
	case "month":
		return timestamp.After(now.AddDate(0, -1, 0))
	case "quarter":
		return timestamp.After(now.AddDate(0, -3, 0))
	case "year":
		return timestamp.After(now.AddDate(-1, 0, 0))
	case "all":
		return true
	default:
		return false
	}
}

func (h *ViewingHandler) getShowGenres(title string) []string {
	// Simplified genre detection - in real implementation, 
	// you'd query Plex/TVDB API for actual genre data
	genreMap := map[string][]string{
		"The Office":       {"Comedy", "Workplace"},
		"Breaking Bad":     {"Drama", "Crime"},
		"Game of Thrones":  {"Fantasy", "Drama"},
		"Friends":          {"Comedy", "Romance"},
		"Stranger Things":  {"Sci-Fi", "Horror"},
	}

	if genres, exists := genreMap[title]; exists {
		return genres
	}
	return []string{"Unknown"}
}

func (h *ViewingHandler) getTopGenre(genreCount map[string]int) string {
	maxCount := 0
	topGenre := ""
	
	for genre, count := range genreCount {
		if count > maxCount {
			maxCount = count
			topGenre = genre
		}
	}
	
	return topGenre
}

func (h *ViewingHandler) getMostWatchedShow(showStats map[string]map[string]interface{}) string {
	maxWatched := 0
	mostWatched := ""
	
	for show, stats := range showStats {
		if watched, ok := stats["watched_episodes"].(int); ok && watched > maxWatched {
			maxWatched = watched
			mostWatched = show
		}
	}
	
	return mostWatched
}

func (h *ViewingHandler) getPeakHour(hourCounts map[int]int) int {
	maxCount := 0
	peakHour := 0
	
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}
	
	return peakHour
}

func (h *ViewingHandler) getPeakDay(dayCounts map[string]int) string {
	maxCount := 0
	peakDay := ""
	
	for day, count := range dayCounts {
		if count > maxCount {
			maxCount = count
			peakDay = day
		}
	}
	
	return peakDay
}

func (h *ViewingHandler) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	
	return sum / float64(len(values))
}