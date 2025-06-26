package mcp

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
)

// analyzeGenres analyzes viewing patterns by genre
func (s *MediateServer) analyzeGenres(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	genreCount := make(map[string]int)
	totalEpisodes := 0

	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Watched && s.isInTimeframe(episode.LastViewedAt, timeframe) {
				genres := s.getShowGenres(show.Title)
				for _, genre := range genres {
					genreCount[genre]++
				}
				totalEpisodes++
			}
		}
	}

	summary := fmt.Sprintf("Watched %d episodes across %d genres in the last %s", 
		totalEpisodes, len(genreCount), timeframe)

	// Sort genres by count
	type genreStats struct {
		Genre string `json:"genre"`
		Count int    `json:"count"`
	}
	
	var sortedGenres []genreStats
	for genre, count := range genreCount {
		sortedGenres = append(sortedGenres, genreStats{Genre: genre, Count: count})
	}
	
	sort.Slice(sortedGenres, func(i, j int) bool {
		return sortedGenres[i].Count > sortedGenres[j].Count
	})

	data := map[string]interface{}{
		"total_episodes":   totalEpisodes,
		"total_genres":     len(genreCount),
		"genre_breakdown":  genreCount,
		"sorted_genres":    sortedGenres,
		"top_genre":        s.getTopGenre(genreCount),
	}

	return summary, data
}

// analyzeShows analyzes viewing patterns by show
func (s *MediateServer) analyzeShows(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	showStats := make(map[string]map[string]interface{})
	totalWatched := 0

	for _, show := range *shows {
		watchedCount := 0
		totalCount := len(show.Episodes)
		
		for _, episode := range show.Episodes {
			if episode.Watched && s.isInTimeframe(episode.LastViewedAt, timeframe) {
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
				"continuing":       show.Continuing,
				"rating":          show.Rating,
			}
		}
	}

	summary := fmt.Sprintf("Watched %d episodes across %d shows in the last %s", 
		totalWatched, len(showStats), timeframe)

	// Get top shows by episodes watched
	type showRank struct {
		Title   string `json:"title"`
		Watched int    `json:"watched"`
	}
	
	var topShows []showRank
	for title, stats := range showStats {
		if watched, ok := stats["watched_episodes"].(int); ok {
			topShows = append(topShows, showRank{Title: title, Watched: watched})
		}
	}
	
	sort.Slice(topShows, func(i, j int) bool {
		return topShows[i].Watched > topShows[j].Watched
	})

	data := map[string]interface{}{
		"total_watched":     totalWatched,
		"shows_watched":     len(showStats),
		"shows":            showStats,
		"top_shows":        topShows,
		"most_watched":     s.getMostWatchedShow(showStats),
	}

	return summary, data
}

// analyzePatterns analyzes viewing time patterns
func (s *MediateServer) analyzePatterns(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	hourCounts := make(map[int]int)
	dayCounts := make(map[string]int)
	totalSessions := 0

	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Watched && episode.LastViewedAt != nil && 
			   s.isInTimeframe(episode.LastViewedAt, timeframe) {
				
				viewTime := *episode.LastViewedAt
				hour := viewTime.Hour()
				day := viewTime.Weekday().String()
				
				hourCounts[hour]++
				dayCounts[day]++
				totalSessions++
			}
		}
	}

	peakHour := s.getPeakHour(hourCounts)
	peakDay := s.getPeakDay(dayCounts)

	summary := fmt.Sprintf("Most active viewing time: %s at %d:00, with %d total sessions", 
		peakDay, peakHour, totalSessions)

	data := map[string]interface{}{
		"total_sessions":   totalSessions,
		"hourly_breakdown": hourCounts,
		"daily_breakdown":  dayCounts,
		"peak_hour":       peakHour,
		"peak_day":        peakDay,
		"viewing_insights": s.getViewingInsights(hourCounts, dayCounts),
	}

	return summary, data
}

// analyzeCompletionRate analyzes show completion rates
func (s *MediateServer) analyzeCompletionRate(shows *shows.Shows, timeframe string) (string, map[string]interface{}) {
	completedShows := 0
	inProgressShows := 0
	notStartedShows := 0
	totalShows := len(*shows)

	completionRates := make([]float64, 0)
	showCompletions := make(map[string]float64)

	for _, show := range *shows {
		watchedCount := 0
		totalCount := len(show.Episodes)

		for _, episode := range show.Episodes {
			if episode.Watched {
				watchedCount++
			}
		}

		completionRate := 0.0
		if totalCount > 0 {
			completionRate = float64(watchedCount) / float64(totalCount) * 100
		}
		
		completionRates = append(completionRates, completionRate)
		showCompletions[show.Title] = completionRate

		if completionRate == 100 {
			completedShows++
		} else if completionRate > 0 {
			inProgressShows++
		} else {
			notStartedShows++
		}
	}

	avgCompletion := s.calculateAverage(completionRates)

	summary := fmt.Sprintf("Average completion rate: %.1f%% (%d completed, %d in progress, %d not started)", 
		avgCompletion, completedShows, inProgressShows, notStartedShows)

	data := map[string]interface{}{
		"total_shows":        totalShows,
		"completed_shows":    completedShows,
		"in_progress_shows":  inProgressShows,
		"not_started_shows":  notStartedShows,
		"average_completion": avgCompletion,
		"show_completions":   showCompletions,
	}

	return summary, data
}

// generateRecommendations generates personalized recommendations
func (s *MediateServer) generateRecommendations(mediaType string, basis string, limit int) []*Recommendation {
	s.logger.Info("Generating recommendations", "type", mediaType, "basis", basis, "limit", limit)

	recommendations := make([]*Recommendation, 0)
	
	// Get user's viewing data for analysis
	shows := s.mediate.GetShows()
	if shows == nil {
		return recommendations
	}

	switch basis {
	case "viewing_history":
		recommendations = s.generateFromHistory(shows, mediaType, limit)
	case "similar_shows":
		recommendations = s.generateSimilar(shows, mediaType, limit)
	case "popular":
		recommendations = s.generatePopular(mediaType, limit)
	case "new_releases":
		recommendations = s.generateNewReleases(mediaType, limit)
	default:
		recommendations = s.generateFromHistory(shows, mediaType, limit)
	}

	return recommendations
}

// generateFromHistory generates recommendations based on viewing history
func (s *MediateServer) generateFromHistory(shows *shows.Shows, mediaType string, limit int) []*Recommendation {
	// Analyze user's preferred genres
	genreScores := make(map[string]int)
	
	for _, show := range *shows {
		watchedCount := 0
		for _, episode := range show.Episodes {
			if episode.Watched {
				watchedCount++
			}
		}
		
		if watchedCount > 0 {
			genres := s.getShowGenres(show.Title)
			for _, genre := range genres {
				genreScores[genre] += watchedCount
			}
		}
	}

	// Generate sample recommendations based on preferred genres
	recommendations := []*Recommendation{
		{
			Title:       "Better Call Saul",
			Type:        "show",
			Score:       0.95,
			Reason:      "Based on your viewing of crime dramas",
			Genre:       []string{"Drama", "Crime"},
			Year:        2015,
			Description: "A prequel to Breaking Bad that follows the story of small-time lawyer Jimmy McGill",
			Available:   false,
			CreatedAt:   time.Now(),
		},
		{
			Title:       "The Wire",
			Type:        "show",
			Score:       0.92,
			Reason:      "Highly rated crime drama similar to your preferences",
			Genre:       []string{"Drama", "Crime"},
			Year:        2002,
			Description: "A realistic look at the drug trade in Baltimore through the eyes of both law enforcement and drug dealers",
			Available:   false,
			CreatedAt:   time.Now(),
		},
		{
			Title:       "Succession",
			Type:        "show",
			Score:       0.89,
			Reason:      "Popular drama series with complex characters",
			Genre:       []string{"Drama", "Comedy"},
			Year:        2018,
			Description: "A media empire patriarch considers his succession plan",
			Available:   false,
			CreatedAt:   time.Now(),
		},
	}

	if limit < len(recommendations) {
		recommendations = recommendations[:limit]
	}

	return recommendations
}

// generateSimilar generates recommendations based on similar shows
func (s *MediateServer) generateSimilar(shows *shows.Shows, mediaType string, limit int) []*Recommendation {
	// Simplified similar show recommendations
	return []*Recommendation{
		{
			Title:       "Dark",
			Type:        "show",
			Score:       0.88,
			Reason:      "Complex sci-fi drama with mystery elements",
			Genre:       []string{"Sci-Fi", "Mystery"},
			Year:        2017,
			Description: "A family saga with a supernatural twist set in a German town",
			Available:   false,
			CreatedAt:   time.Now(),
		},
	}
}

// generatePopular generates popular recommendations
func (s *MediateServer) generatePopular(mediaType string, limit int) []*Recommendation {
	return []*Recommendation{
		{
			Title:       "The Bear",
			Type:        "show",
			Score:       0.94,
			Reason:      "Currently trending comedy-drama",
			Genre:       []string{"Comedy", "Drama"},
			Year:        2022,
			Description: "A chef returns to Chicago to run his deceased brother's Italian beef restaurant",
			Available:   false,
			CreatedAt:   time.Now(),
		},
	}
}

// generateNewReleases generates new release recommendations
func (s *MediateServer) generateNewReleases(mediaType string, limit int) []*Recommendation {
	return []*Recommendation{
		{
			Title:       "House of the Dragon",
			Type:        "show",
			Score:       0.85,
			Reason:      "New fantasy epic series",
			Genre:       []string{"Fantasy", "Drama"},
			Year:        2022,
			Description: "A Game of Thrones prequel set 200 years before the original series",
			Available:   false,
			CreatedAt:   time.Now(),
		},
	}
}

// Helper functions

func (s *MediateServer) isInTimeframe(timestamp *time.Time, timeframe string) bool {
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

func (s *MediateServer) getShowGenres(title string) []string {
	// Simplified genre mapping - in production, you'd use TVDB/TMDB APIs
	genreMap := map[string][]string{
		"The Office":       {"Comedy", "Workplace"},
		"Breaking Bad":     {"Drama", "Crime"},
		"Game of Thrones":  {"Fantasy", "Drama"},
		"Friends":          {"Comedy", "Romance"},
		"Stranger Things":  {"Sci-Fi", "Horror"},
		"Better Call Saul": {"Drama", "Crime"},
		"The Wire":         {"Drama", "Crime"},
		"Succession":       {"Drama", "Comedy"},
		"Dark":             {"Sci-Fi", "Mystery"},
		"The Bear":         {"Comedy", "Drama"},
	}

	// Try exact match first
	if genres, exists := genreMap[title]; exists {
		return genres
	}

	// Try partial match
	titleLower := strings.ToLower(title)
	for showTitle, genres := range genreMap {
		if strings.Contains(strings.ToLower(showTitle), titleLower) ||
		   strings.Contains(titleLower, strings.ToLower(showTitle)) {
			return genres
		}
	}

	return []string{"Drama"} // Default genre
}

func (s *MediateServer) getTopGenre(genreCount map[string]int) string {
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

func (s *MediateServer) getMostWatchedShow(showStats map[string]map[string]interface{}) string {
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

func (s *MediateServer) getPeakHour(hourCounts map[int]int) int {
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

func (s *MediateServer) getPeakDay(dayCounts map[string]int) string {
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

func (s *MediateServer) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	
	return sum / float64(len(values))
}

func (s *MediateServer) getViewingInsights(hourCounts map[int]int, dayCounts map[string]int) []string {
	insights := make([]string, 0)
	
	// Peak viewing time insights
	peakHour := s.getPeakHour(hourCounts)
	if peakHour >= 20 && peakHour <= 23 {
		insights = append(insights, "You're a night owl - most viewing happens in the evening")
	} else if peakHour >= 12 && peakHour <= 17 {
		insights = append(insights, "Afternoon viewing is your preference")
	}
	
	// Weekend vs weekday insights
	weekendCount := dayCounts["Saturday"] + dayCounts["Sunday"]
	weekdayCount := dayCounts["Monday"] + dayCounts["Tuesday"] + dayCounts["Wednesday"] + 
	                dayCounts["Thursday"] + dayCounts["Friday"]
	
	if weekendCount > weekdayCount {
		insights = append(insights, "You watch more on weekends than weekdays")
	} else {
		insights = append(insights, "You're consistent with weekday viewing")
	}
	
	return insights
}