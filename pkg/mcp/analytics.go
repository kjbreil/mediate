package mcp

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
)

// analyzeGenres analyzes viewing patterns by genre.
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
		"total_episodes":  totalEpisodes,
		"total_genres":    len(genreCount),
		"genre_breakdown": genreCount,
		"sorted_genres":   sortedGenres,
		"top_genre":       s.getTopGenre(genreCount),
	}

	return summary, data
}

// analyzeShows analyzes viewing patterns by show.
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
				"rating":           show.Rating,
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
		"total_watched": totalWatched,
		"shows_watched": len(showStats),
		"shows":         showStats,
		"top_shows":     topShows,
		"most_watched":  s.getMostWatchedShow(showStats),
	}

	return summary, data
}

// analyzePatterns analyzes viewing time patterns.
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
		"peak_hour":        peakHour,
		"peak_day":         peakDay,
		"viewing_insights": s.getViewingInsights(hourCounts, dayCounts),
	}

	return summary, data
}

// analyzeCompletionRate analyzes show completion rates.
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

// generateRecommendations generates personalized recommendations.
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

// generateFromHistory generates recommendations based on viewing history.
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

	// Return all recommendations without limit

	return recommendations
}

// generateSimilar generates recommendations based on similar shows.
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

// generatePopular generates popular recommendations.
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

// generateNewReleases generates new release recommendations.
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

// ShowAnalysis represents detailed analysis for a specific show.
type ShowAnalysis struct {
	Show          ShowInfo                  `json:"show"`
	Timeframe     string                    `json:"timeframe"`
	UserFilter    string                    `json:"user_filter,omitempty"`
	OverallStats  ShowStats                 `json:"overall_stats"`
	UserBreakdown map[string]*UserShowStats `json:"user_breakdown"`
	SeasonStats   map[int]*SeasonStats      `json:"season_stats"`
	ViewingTrends *ViewingTrends            `json:"viewing_trends"`
	GeneratedAt   time.Time                 `json:"generated_at"`
}

type ShowInfo struct {
	Title      string  `json:"title"`
	TvdbID     int     `json:"tvdb_id"`
	Continuing bool    `json:"continuing"`
	Rating     float64 `json:"rating"`
	Seasons    int     `json:"seasons"`
	Episodes   int     `json:"total_episodes"`
}

type ShowStats struct {
	TotalViews          int           `json:"total_views"`
	UniqueViewers       int           `json:"unique_viewers"`
	TotalWatchTime      time.Duration `json:"total_watch_time"`
	AverageCompletion   float64       `json:"average_completion"`
	MostWatchedEpisode  string        `json:"most_watched_episode"`
	LeastWatchedEpisode string        `json:"least_watched_episode"`
	BingeScore          float64       `json:"binge_score"` // How often episodes are watched consecutively
}

type UserShowStats struct {
	Username         string        `json:"username"`
	EpisodesWatched  int           `json:"episodes_watched"`
	TotalWatchTime   time.Duration `json:"total_watch_time"`
	CompletionRate   float64       `json:"completion_rate"`
	FavoriteSeason   int           `json:"favorite_season"`
	ViewingFrequency string        `json:"viewing_frequency"`
	LastWatched      *time.Time    `json:"last_watched,omitempty"`
}

type SeasonStats struct {
	Season         int           `json:"season"`
	EpisodeCount   int           `json:"episode_count"`
	TotalViews     int           `json:"total_views"`
	AverageViews   float64       `json:"average_views"`
	CompletionRate float64       `json:"completion_rate"`
	TotalWatchTime time.Duration `json:"total_watch_time"`
	MostPopular    string        `json:"most_popular_episode"`
}

type ViewingTrends struct {
	DailyViewing   map[string]int `json:"daily_viewing"`
	HourlyViewing  map[int]int    `json:"hourly_viewing"`
	WeeklyTrend    []int          `json:"weekly_trend"`
	PopularDevices []string       `json:"popular_devices"`
}

// analyzeIndividualShow provides detailed analysis for a specific show.
func (s *MediateServer) analyzeIndividualShow(show *shows.Show, timeframe string, userFilter string) *ShowAnalysis {
	analysis := &ShowAnalysis{
		Show: ShowInfo{
			Title:      show.Title,
			TvdbID:     show.TvdbID,
			Continuing: show.Continuing,
			Rating:     show.Rating,
			Episodes:   len(show.Episodes),
		},
		Timeframe:     timeframe,
		UserFilter:    userFilter,
		UserBreakdown: make(map[string]*UserShowStats),
		SeasonStats:   make(map[int]*SeasonStats),
		GeneratedAt:   time.Now(),
	}

	// Get viewing sessions for this show (simulated for now since we don't have real session data yet)
	sessions := s.getViewingSessionsForShow(show, timeframe)

	// Filter by user if specified
	if userFilter != "" {
		sessions = sessions.ByUser(userFilter)
	}

	// Calculate overall stats
	analysis.OverallStats = s.calculateShowStats(show, sessions)

	// Calculate per-user breakdown
	userStats := sessions.GetUserStats()
	for username, stats := range userStats {
		analysis.UserBreakdown[username] = &UserShowStats{
			Username:         stats.Username,
			EpisodesWatched:  stats.CompletedEpisodes,
			TotalWatchTime:   stats.TotalWatchTime,
			CompletionRate:   stats.AverageCompletion,
			ViewingFrequency: s.calculateViewingFrequency(stats.TotalSessions, timeframe),
		}
	}

	// Calculate season stats
	analysis.SeasonStats = s.calculateSeasonStats(show, sessions)

	// Calculate viewing trends
	analysis.ViewingTrends = s.calculateViewingTrends(sessions)

	// Calculate season count
	maxSeason := 0
	for _, episode := range show.Episodes {
		if episode.Season > maxSeason {
			maxSeason = episode.Season
		}
	}
	analysis.Show.Seasons = maxSeason

	return analysis
}

// EpisodeAnalysis represents detailed analysis for episodes.
type EpisodeAnalysis struct {
	Show        ShowInfo        `json:"show"`
	Season      int             `json:"season,omitempty"`
	UserFilter  string          `json:"user_filter,omitempty"`
	Episodes    []*EpisodeStats `json:"episodes"`
	Summary     EpisodeSummary  `json:"summary"`
	SortBy      string          `json:"sort_by"`
	GeneratedAt time.Time       `json:"generated_at"`
}

type EpisodeStats struct {
	Title             string                       `json:"title"`
	Season            int                          `json:"season"`
	Episode           int                          `json:"episode"`
	TvdbID            int                          `json:"tvdb_id"`
	AirDate           *time.Time                   `json:"air_date,omitempty"`
	Duration          time.Duration                `json:"duration"`
	ViewCount         int                          `json:"view_count"`
	UniqueViewers     int                          `json:"unique_viewers"`
	AverageCompletion float64                      `json:"average_completion"`
	TotalWatchTime    time.Duration                `json:"total_watch_time"`
	LastWatched       *time.Time                   `json:"last_watched,omitempty"`
	UserStats         map[string]*UserEpisodeStats `json:"user_stats"`
	PopularityRank    int                          `json:"popularity_rank"`
	SkipRate          float64                      `json:"skip_rate"`
	ReplayRate        float64                      `json:"replay_rate"`
}

type UserEpisodeStats struct {
	Username       string        `json:"username"`
	ViewCount      int           `json:"view_count"`
	CompletionRate float64       `json:"completion_rate"`
	LastWatched    *time.Time    `json:"last_watched,omitempty"`
	WatchTime      time.Duration `json:"watch_time"`
}

type EpisodeSummary struct {
	TotalEpisodes  int           `json:"total_episodes"`
	TotalViews     int           `json:"total_views"`
	AverageViews   float64       `json:"average_views"`
	TotalWatchTime time.Duration `json:"total_watch_time"`
	MostPopular    string        `json:"most_popular"`
	LeastPopular   string        `json:"least_popular"`
	HighestRated   string        `json:"highest_rated"`
}

// analyzeEpisodes provides detailed analysis for episodes.
func (s *MediateServer) analyzeEpisodes(
	show *shows.Show,
	season int,
	userFilter string,
	sortBy string,
	limit int,
) *EpisodeAnalysis {
	analysis := &EpisodeAnalysis{
		Show: ShowInfo{
			Title:      show.Title,
			TvdbID:     show.TvdbID,
			Continuing: show.Continuing,
			Episodes:   len(show.Episodes),
		},
		Season:      season,
		UserFilter:  userFilter,
		SortBy:      sortBy,
		Episodes:    make([]*EpisodeStats, 0),
		GeneratedAt: time.Now(),
	}

	// Filter episodes by season if specified
	episodes := show.Episodes
	if season > 0 {
		episodes = make([]*shows.Episode, 0)
		for _, ep := range show.Episodes {
			if ep.Season == season {
				episodes = append(episodes, ep)
			}
		}
	}

	totalViews := 0
	totalWatchTime := time.Duration(0)

	// Analyze each episode
	for _, episode := range episodes {
		sessions := s.getViewingSessionsForEpisode(episode)

		if userFilter != "" {
			sessions = sessions.ByUser(userFilter)
		}

		episodeStats := &EpisodeStats{
			Title:       episode.Title,
			Season:      episode.Season,
			Episode:     episode.Episode,
			TvdbID:      episode.TvdbID,
			AirDate:     episode.AirDate,
			Duration:    episode.Duration,
			ViewCount:   episode.ViewCount,
			LastWatched: episode.LastViewedAt,
			UserStats:   make(map[string]*UserEpisodeStats),
		}

		// Calculate episode metrics from sessions
		episodeStats.UniqueViewers = len(sessions.GetUserStats())
		episodeStats.TotalWatchTime = sessions.TotalWatchTime()

		// Calculate completion rate
		if len(sessions) > 0 {
			completed := sessions.Completed()
			episodeStats.AverageCompletion = float64(len(completed)) / float64(len(sessions)) * 100
		}

		// Per-user stats for this episode
		userStats := sessions.GetUserStats()
		for username, stats := range userStats {
			episodeStats.UserStats[username] = &UserEpisodeStats{
				Username:       stats.Username,
				ViewCount:      stats.TotalSessions,
				CompletionRate: stats.AverageCompletion,
				WatchTime:      stats.TotalWatchTime,
			}
		}

		analysis.Episodes = append(analysis.Episodes, episodeStats)
		totalViews += episodeStats.ViewCount
		totalWatchTime += episodeStats.TotalWatchTime
	}

	// Sort episodes based on sortBy parameter
	s.sortEpisodes(analysis.Episodes, sortBy)

	// No limit applied - return all episodes

	// Calculate summary
	analysis.Summary = EpisodeSummary{
		TotalEpisodes:  len(episodes),
		TotalViews:     totalViews,
		TotalWatchTime: totalWatchTime,
	}

	if len(episodes) > 0 {
		analysis.Summary.AverageViews = float64(totalViews) / float64(len(episodes))
	}

	// Find most/least popular
	if len(analysis.Episodes) > 0 {
		analysis.Summary.MostPopular = analysis.Episodes[0].Title
		analysis.Summary.LeastPopular = analysis.Episodes[len(analysis.Episodes)-1].Title
	}

	return analysis
}

// Helper functions for the new analysis

func (s *MediateServer) getViewingSessionsForShow(show *shows.Show, timeframe string) shows.ViewingSessions {
	// TODO: Query actual viewing sessions from database
	// For now, return simulated data based on existing episode data
	sessions := make(shows.ViewingSessions, 0)

	for _, episode := range show.Episodes {
		if episode.Watched && episode.LastViewedAt != nil {
			session := &shows.ViewingSession{
				EpisodeTvdbID:   episode.TvdbID,
				PlexUsername:    "primary_user", // Simulated
				StartedAt:       *episode.LastViewedAt,
				Duration:        episode.Duration,
				ProgressPercent: 100,
				Completed:       true,
			}
			sessions = append(sessions, session)
		}
	}

	return sessions.InTimeframe(s.getTimeframeCutoff(timeframe))
}

func (s *MediateServer) getViewingSessionsForEpisode(episode *shows.Episode) shows.ViewingSessions {
	// TODO: Query actual viewing sessions from database
	sessions := make(shows.ViewingSessions, 0)

	if episode.Watched && episode.LastViewedAt != nil {
		session := &shows.ViewingSession{
			EpisodeTvdbID:   episode.TvdbID,
			PlexUsername:    "primary_user", // Simulated
			StartedAt:       *episode.LastViewedAt,
			Duration:        episode.Duration,
			ProgressPercent: 100,
			Completed:       true,
		}
		sessions = append(sessions, session)
	}

	return sessions
}

func (s *MediateServer) calculateShowStats(show *shows.Show, sessions shows.ViewingSessions) ShowStats {
	stats := ShowStats{}

	stats.TotalViews = len(sessions)
	stats.TotalWatchTime = sessions.TotalWatchTime()

	userStats := sessions.GetUserStats()
	stats.UniqueViewers = len(userStats)

	if len(sessions) > 0 {
		completed := sessions.Completed()
		stats.AverageCompletion = float64(len(completed)) / float64(len(sessions)) * 100
	}

	return stats
}

func (s *MediateServer) calculateSeasonStats(show *shows.Show, sessions shows.ViewingSessions) map[int]*SeasonStats {
	seasonMap := make(map[int]*SeasonStats)

	for _, episode := range show.Episodes {
		season := episode.Season
		if _, exists := seasonMap[season]; !exists {
			seasonMap[season] = &SeasonStats{
				Season: season,
			}
		}

		seasonMap[season].EpisodeCount++

		if episode.Watched {
			seasonMap[season].TotalViews++
		}
	}

	// Calculate averages
	for _, stats := range seasonMap {
		if stats.EpisodeCount > 0 {
			stats.AverageViews = float64(stats.TotalViews) / float64(stats.EpisodeCount)
		}
	}

	return seasonMap
}

func (s *MediateServer) calculateViewingTrends(sessions shows.ViewingSessions) *ViewingTrends {
	trends := &ViewingTrends{
		DailyViewing:  make(map[string]int),
		HourlyViewing: make(map[int]int),
		WeeklyTrend:   make([]int, 7),
	}

	for _, session := range sessions {
		day := session.StartedAt.Weekday().String()
		hour := session.StartedAt.Hour()
		weekday := int(session.StartedAt.Weekday())

		trends.DailyViewing[day]++
		trends.HourlyViewing[hour]++
		trends.WeeklyTrend[weekday]++
	}

	return trends
}

func (s *MediateServer) calculateViewingFrequency(totalSessions int, timeframe string) string {
	days := s.getTimeframeDays(timeframe)
	if days == 0 {
		return "unknown"
	}

	avgPerDay := float64(totalSessions) / float64(days)

	if avgPerDay >= 2 {
		return "heavy"
	} else if avgPerDay >= 0.5 {
		return "regular"
	} else if avgPerDay >= 0.1 {
		return "light"
	}
	return "rare"
}

func (s *MediateServer) sortEpisodes(episodes []*EpisodeStats, sortBy string) {
	switch sortBy {
	case "view_count":
		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].ViewCount > episodes[j].ViewCount
		})
	case "completion_rate":
		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].AverageCompletion > episodes[j].AverageCompletion
		})
	case "air_date":
		sort.Slice(episodes, func(i, j int) bool {
			if episodes[i].AirDate == nil && episodes[j].AirDate == nil {
				return false
			}
			if episodes[i].AirDate == nil {
				return false
			}
			if episodes[j].AirDate == nil {
				return true
			}
			return episodes[i].AirDate.After(*episodes[j].AirDate)
		})
	case "episode_number":
		sort.Slice(episodes, func(i, j int) bool {
			if episodes[i].Season != episodes[j].Season {
				return episodes[i].Season < episodes[j].Season
			}
			return episodes[i].Episode < episodes[j].Episode
		})
	default:
		// Default to episode number sorting
		sort.Slice(episodes, func(i, j int) bool {
			if episodes[i].Season != episodes[j].Season {
				return episodes[i].Season < episodes[j].Season
			}
			return episodes[i].Episode < episodes[j].Episode
		})
	}
}

func (s *MediateServer) getTimeframeCutoff(timeframe string) time.Time {
	now := time.Now()
	switch timeframe {
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "quarter":
		return now.AddDate(0, -3, 0)
	case "year":
		return now.AddDate(-1, 0, 0)
	default:
		return time.Time{} // All time
	}
}

func (s *MediateServer) getTimeframeDays(timeframe string) int {
	switch timeframe {
	case "week":
		return 7
	case "month":
		return 30
	case "quarter":
		return 90
	case "year":
		return 365
	default:
		return 0
	}
}

// analyzeDeletedMediaTrends analyzes viewing trends for deleted media.
func (s *MediateServer) analyzeDeletedMediaTrends(timeframe string) (*DeletedMediaAnalysis, error) {
	deletedMedia, err := s.mediate.DB.GetDeletedMedia(0, 0)
	if err != nil {
		return nil, err
	}

	analysis := &DeletedMediaAnalysis{
		Timeframe:   timeframe,
		GeneratedAt: time.Now(),
		TotalItems:  len(deletedMedia),
		MediaStats:  make(map[string]*DeletedMediaTypeStats),
	}

	// Calculate cutoff time for timeframe filtering
	cutoff := s.getTimeframeCutoff(timeframe)

	var totalViews int
	var totalWatchTime time.Duration
	viewsByUser := make(map[string]int)
	deletionsByMonth := make(map[string]int)

	for _, media := range deletedMedia {
		// Skip if outside timeframe (based on deletion date)
		if !cutoff.IsZero() && media.DeletedAt.Before(cutoff) {
			continue
		}

		analysis.ActiveItems++
		totalViews += media.TotalViews
		totalWatchTime += media.TotalWatchTime

		// Track by media type
		if _, exists := analysis.MediaStats[media.MediaType]; !exists {
			analysis.MediaStats[media.MediaType] = &DeletedMediaTypeStats{
				MediaType: media.MediaType,
			}
		}
		stats := analysis.MediaStats[media.MediaType]
		stats.Count++
		stats.TotalViews += media.TotalViews
		stats.TotalWatchTime += media.TotalWatchTime

		// Track views by user from sessions
		for _, session := range media.ViewingSessions {
			viewsByUser[session.PlexUsername]++
		}

		// Track deletions by month
		monthKey := media.DeletedAt.Format("2006-01")
		deletionsByMonth[monthKey]++
	}

	analysis.TotalViews = totalViews
	analysis.TotalWatchTime = totalWatchTime
	analysis.ViewsByUser = viewsByUser
	analysis.DeletionsByMonth = deletionsByMonth

	// Calculate most watched deleted content
	if len(deletedMedia) > 0 {
		// Sort by view count
		mostWatched := deletedMedia[0]
		for _, media := range deletedMedia[1:] {
			if media.TotalViews > mostWatched.TotalViews {
				mostWatched = media
			}
		}
		analysis.MostWatchedDeleted = &DeletedMediaHighlight{
			Title:          mostWatched.Title,
			MediaType:      mostWatched.MediaType,
			TotalViews:     mostWatched.TotalViews,
			TotalWatchTime: mostWatched.TotalWatchTime,
			DeletedAt:      mostWatched.DeletedAt,
		}
	}

	return analysis, nil
}

// DeletedMediaAnalysis represents comprehensive analysis of deleted media.
type DeletedMediaAnalysis struct {
	Timeframe          string                            `json:"timeframe"`
	GeneratedAt        time.Time                         `json:"generated_at"`
	TotalItems         int                               `json:"total_items"`
	ActiveItems        int                               `json:"active_items"`
	TotalViews         int                               `json:"total_views"`
	TotalWatchTime     time.Duration                     `json:"total_watch_time"`
	ViewsByUser        map[string]int                    `json:"views_by_user"`
	DeletionsByMonth   map[string]int                    `json:"deletions_by_month"`
	MediaStats         map[string]*DeletedMediaTypeStats `json:"media_stats"`
	MostWatchedDeleted *DeletedMediaHighlight            `json:"most_watched_deleted"`
}

// DeletedMediaTypeStats provides statistics for a specific media type.
type DeletedMediaTypeStats struct {
	MediaType      string        `json:"media_type"`
	Count          int           `json:"count"`
	TotalViews     int           `json:"total_views"`
	TotalWatchTime time.Duration `json:"total_watch_time"`
	AverageViews   float64       `json:"average_views"`
}

// DeletedMediaHighlight highlights notable deleted media.
type DeletedMediaHighlight struct {
	Title          string        `json:"title"`
	MediaType      string        `json:"media_type"`
	TotalViews     int           `json:"total_views"`
	TotalWatchTime time.Duration `json:"total_watch_time"`
	DeletedAt      time.Time     `json:"deleted_at"`
}

// CalculateDeletedMediaImpact analyzes the viewing impact of deleted media.
func (s *MediateServer) CalculateDeletedMediaImpact() (*DeletedMediaImpact, error) {
	deletedSummary, err := s.mediate.DB.GetDeletedMediaSummary()
	if err != nil {
		return nil, err
	}

	// Get current active media stats for comparison
	shows := s.mediate.GetShows()
	if shows == nil {
		return nil, errors.New("unable to get current shows data")
	}

	var activeViews int
	var activeWatchTime time.Duration
	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Watched {
				activeViews++
				if episode.LastViewedAt != nil {
					// Estimate watch time (would be better with actual data)
					activeWatchTime += episode.Duration
				}
			}
		}
	}

	impact := &DeletedMediaImpact{
		DeletedViews:     deletedSummary.TotalViews,
		DeletedWatchTime: deletedSummary.TotalWatchTime,
		ActiveViews:      activeViews,
		ActiveWatchTime:  activeWatchTime,
		GeneratedAt:      time.Now(),
	}

	// Calculate impact percentages
	totalViews := activeViews + deletedSummary.TotalViews
	if totalViews > 0 {
		impact.DeletedViewsPercent = float64(deletedSummary.TotalViews) / float64(totalViews) * 100
	}

	totalWatchTime := activeWatchTime + deletedSummary.TotalWatchTime
	if totalWatchTime > 0 {
		impact.DeletedWatchTimePercent = float64(deletedSummary.TotalWatchTime) / float64(totalWatchTime) * 100
	}

	return impact, nil
}

// DeletedMediaImpact represents the impact analysis of deleted media.
type DeletedMediaImpact struct {
	DeletedViews            int           `json:"deleted_views"`
	DeletedWatchTime        time.Duration `json:"deleted_watch_time"`
	DeletedViewsPercent     float64       `json:"deleted_views_percent"`
	DeletedWatchTimePercent float64       `json:"deleted_watch_time_percent"`
	ActiveViews             int           `json:"active_views"`
	ActiveWatchTime         time.Duration `json:"active_watch_time"`
	GeneratedAt             time.Time     `json:"generated_at"`
}
