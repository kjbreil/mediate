package analysis

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kjbreil/mediate/pkg/cli"
	"github.com/kjbreil/mediate/pkg/mcp"
	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/kjbreil/mediate/pkg/shows"
)

// RunAnalysis executes the analysis based on CLI configuration.
func RunAnalysis(cfg *cli.Config, m *mediate.Mediate, logger *slog.Logger) error {
	if cfg.ScanDeleted {
		return runScanDeleted(m, logger)
	}

	if cfg.Analyze {
		return runAnalyze(cfg, m, logger)
	}

	return errors.New("no analysis action specified")
}

// runScanDeleted scans for deleted media.
func runScanDeleted(m *mediate.Mediate, _ *slog.Logger) error {
	fmt.Println("🔍 Scanning for deleted/orphaned media...")

	// For now, use hardcoded config values since GetConfig() doesn't exist
	// This should be replaced with actual config access
	if m == nil {
		return errors.New("plex configuration not found. Please check config file")
	}

	// TODO: Add proper config access to get Plex URL and token
	// For now, return a message that this feature needs configuration
	fmt.Println("⚠️  Deleted media scanning requires Plex configuration integration")
	fmt.Println("   This feature needs to be connected to your actual Plex server config")
	return nil
}

// runAnalyze runs the specified analysis.
func runAnalyze(cfg *cli.Config, m *mediate.Mediate, logger *slog.Logger) error {
	// Create MCP server for analysis functions
	mcpServer := mcp.NewMediateServer(m, logger)

	switch cfg.AnalyzeType {
	case "habits":
		return analyzeViewingHabits(mcpServer)
	case "show":
		if cfg.AnalyzeShow == "" {
			return errors.New("--show parameter required for show analysis")
		}
		return analyzeShow(mcpServer, cfg.AnalyzeShow)
	case "episodes":
		if cfg.AnalyzeShow == "" {
			return errors.New("--show parameter required for episode analysis")
		}
		return analyzeEpisodes(mcpServer, cfg.AnalyzeShow)
	case "deleted":
		return analyzeDeletedMedia(mcpServer)
	default:
		return fmt.Errorf("unknown analysis type: %s", cfg.AnalyzeType)
	}
}

// analyzeViewingHabits analyzes overall viewing habits.
func analyzeViewingHabits(server *mcp.MediateServer) error {
	fmt.Println("📊 Analyzing viewing habits...")

	// Use a simple approach for now - create JSON output and parse key stats
	showsData := server.GetMediate().GetShows()
	if showsData == nil {
		return errors.New("unable to retrieve shows data")
	}

	var totalEpisodes, watchedEpisodes int
	showStats := make(map[string]int)
	genreStats := make(map[string]int)

	for _, show := range *showsData {
		var showWatched int
		for _, episode := range show.Episodes {
			totalEpisodes++
			if episode.Watched {
				watchedEpisodes++
				showWatched++
			}
		}
		if showWatched > 0 {
			showStats[show.Title] = showWatched
		}
		// Simple genre mapping
		genres := getSimpleGenres(show.Title)
		for _, genre := range genres {
			genreStats[genre] += showWatched
		}
	}

	fmt.Printf("\n📊 Overall Statistics:\n")
	fmt.Printf("  • Total Episodes: %d\n", totalEpisodes)
	fmt.Printf("  • Watched Episodes: %d\n", watchedEpisodes)
	if totalEpisodes > 0 {
		completionRate := float64(watchedEpisodes) / float64(totalEpisodes) * 100
		fmt.Printf("  • Overall Completion Rate: %.1f%%\n", completionRate)
	}

	fmt.Printf("\n🎭 Top Genres:\n")
	printGenreStats(genreStats)

	fmt.Printf("\n📺 Most Watched Shows:\n")
	printShowStats(showStats)

	return nil
}

// analyzeShow analyzes a specific show.
func analyzeShow(server *mcp.MediateServer, showTitle string) error {
	fmt.Printf("🎬 Analyzing show: %s\n", showTitle)

	// Find the show
	showsData := server.GetMediate().GetShows()
	if showsData == nil {
		return errors.New("unable to retrieve shows data")
	}

	var targetShow *shows.Show
	for _, show := range *showsData {
		if strings.EqualFold(show.Title, showTitle) {
			targetShow = show
			break
		}
	}

	if targetShow == nil {
		return fmt.Errorf("show not found: %s", showTitle)
	}

	// Calculate basic statistics
	var totalEpisodes, watchedEpisodes int
	var seasons = make(map[int]int) // season -> episode count

	for _, episode := range targetShow.Episodes {
		totalEpisodes++
		seasons[episode.Season]++
		if episode.Watched {
			watchedEpisodes++
		}
	}

	completionRate := 0.0
	if totalEpisodes > 0 {
		completionRate = float64(watchedEpisodes) / float64(totalEpisodes) * 100
	}

	// Print results
	fmt.Printf("📚 Show: %s (%d seasons, %d episodes)\n",
		targetShow.Title, len(seasons), totalEpisodes)
	fmt.Printf("📊 Overall Stats:\n")
	fmt.Printf("  • Total Episodes: %d\n", totalEpisodes)
	fmt.Printf("  • Watched Episodes: %d\n", watchedEpisodes)
	fmt.Printf("  • Completion Rate: %.1f%%\n", completionRate)
	fmt.Printf("  • Status: %s\n", getShowStatus(targetShow))

	if len(seasons) > 0 {
		fmt.Printf("\n📅 Season Breakdown:\n")
		for season, count := range seasons {
			var seasonWatched int
			for _, episode := range targetShow.Episodes {
				if episode.Season == season && episode.Watched {
					seasonWatched++
				}
			}
			seasonCompletion := 0.0
			if count > 0 {
				seasonCompletion = float64(seasonWatched) / float64(count) * 100
			}
			fmt.Printf("  • Season %d: %d/%d episodes (%.1f%%)\n",
				season, seasonWatched, count, seasonCompletion)
		}
	}

	return nil
}

// analyzeEpisodes analyzes episodes for a specific show.
func analyzeEpisodes(server *mcp.MediateServer, showTitle string) error {
	fmt.Printf("📺 Analyzing episodes for: %s\n", showTitle)

	// Find the show
	showsData := server.GetMediate().GetShows()
	if showsData == nil {
		return errors.New("unable to retrieve shows data")
	}

	var targetShow *shows.Show
	for _, show := range *showsData {
		if strings.EqualFold(show.Title, showTitle) {
			targetShow = show
			break
		}
	}

	if targetShow == nil {
		return fmt.Errorf("show not found: %s", showTitle)
	}

	var totalEpisodes, watchedEpisodes int
	var totalWatchTime time.Duration

	// Sort episodes by view count (simulated since we don't have real view counts yet)
	type episodeInfo struct {
		episode *shows.Episode
		watched bool
	}

	var episodeList []episodeInfo
	for _, episode := range targetShow.Episodes {
		totalEpisodes++
		if episode.Watched {
			watchedEpisodes++
			if episode.LastViewedAt != nil {
				totalWatchTime += episode.Duration
			}
		}
		episodeList = append(episodeList, episodeInfo{
			episode: episode,
			watched: episode.Watched,
		})
	}

	fmt.Printf("📊 Episode Analysis Summary:\n")
	fmt.Printf("  • Total Episodes: %d\n", totalEpisodes)
	fmt.Printf("  • Watched Episodes: %d\n", watchedEpisodes)
	if totalEpisodes > 0 {
		avgViews := float64(watchedEpisodes) / float64(totalEpisodes)
		fmt.Printf("  • Average Completion Rate: %.1f%%\n", avgViews*100)
	}
	fmt.Printf("  • Total Watch Time: %v\n", totalWatchTime)

	fmt.Printf("\n📋 Episode List (most recent first):\n")
	count := 20
	if len(episodeList) < count {
		count = len(episodeList)
	}

	// Sort by season/episode descending (most recent episodes first)
	for i := range len(episodeList) - 1 {
		for j := i + 1; j < len(episodeList); j++ {
			ep1 := episodeList[i].episode
			ep2 := episodeList[j].episode
			if ep2.Season > ep1.Season || (ep2.Season == ep1.Season && ep2.Episode > ep1.Episode) {
				episodeList[i], episodeList[j] = episodeList[j], episodeList[i]
			}
		}
	}

	for i := range count {
		ep := episodeList[i].episode
		status := "⭕ Not Watched"
		if episodeList[i].watched {
			status = "✅ Watched"
		}

		lastViewed := "Never"
		if ep.LastViewedAt != nil {
			lastViewed = ep.LastViewedAt.Format("2006-01-02")
		}

		fmt.Printf("  S%02dE%02d: %s %s (Last: %s)\n",
			ep.Season, ep.Episode, ep.Title, status, lastViewed)
	}

	return nil
}

// analyzeDeletedMedia analyzes deleted media.
func analyzeDeletedMedia(server *mcp.MediateServer) error {
	fmt.Println("🗑️  Analyzing deleted media...")

	summary, err := server.GetDB().GetDeletedMediaSummary()
	if err != nil {
		return fmt.Errorf("error getting deleted media summary: %w", err)
	}

	fmt.Printf("📊 Deleted Media Summary:\n")
	fmt.Printf("  • Total Deleted Items: %d\n", summary.TotalItems)
	fmt.Printf("  • Total Views Lost: %d\n", summary.TotalViews)
	fmt.Printf("  • Recently Deleted (30 days): %d\n", summary.DeletedRecently)

	if len(summary.ByMediaType) > 0 {
		fmt.Printf("\n📚 By Media Type:\n")
		for mediaType, count := range summary.ByMediaType {
			fmt.Printf("  • %s: %d items\n", mediaType, count)
		}
	}

	if len(summary.TopDeletedShows) > 0 {
		fmt.Printf("\n🏆 Most Watched Deleted Content:\n")
		count := 5
		if len(summary.TopDeletedShows) < count {
			count = len(summary.TopDeletedShows)
		}
		for i := range count {
			stats := summary.TopDeletedShows[i]
			fmt.Printf("  %d. %s (%s): %d views, %v watch time\n",
				i+1, stats.Title, stats.MediaType, stats.TotalViews, stats.TotalWatchTime)
		}
	}

	return nil
}

// Helper functions for printing formatted data

// getSimpleGenres returns simplified genre mapping.
func getSimpleGenres(title string) []string {
	genreMap := map[string][]string{
		"What We Do in the Shadows": {"Comedy", "Horror"},
		"The Office":                {"Comedy", "Workplace"},
		"Breaking Bad":              {"Drama", "Crime"},
		"Game of Thrones":           {"Fantasy", "Drama"},
		"Friends":                   {"Comedy", "Romance"},
		"Stranger Things":           {"Sci-Fi", "Horror"},
		"Better Call Saul":          {"Drama", "Crime"},
		"The Wire":                  {"Drama", "Crime"},
		"Succession":                {"Drama", "Comedy"},
		"Dark":                      {"Sci-Fi", "Mystery"},
		"The Bear":                  {"Comedy", "Drama"},
	}

	// Try exact match first
	if genres, exists := genreMap[title]; exists {
		return genres
	}

	// Try partial match (case insensitive)
	titleLower := strings.ToLower(title)
	for showTitle, genres := range genreMap {
		if strings.Contains(strings.ToLower(showTitle), titleLower) ||
			strings.Contains(titleLower, strings.ToLower(showTitle)) {
			return genres
		}
	}

	return []string{"Drama"} // Default genre
}

// printGenreStats prints genre statistics in a formatted way.
func printGenreStats(genreStats map[string]int) {
	type genreStat struct {
		genre string
		count int
	}

	var genres []genreStat
	for genre, count := range genreStats {
		genres = append(genres, genreStat{genre, count})
	}

	// Sort by count descending
	for i := range len(genres) - 1 {
		for j := i + 1; j < len(genres); j++ {
			if genres[j].count > genres[i].count {
				genres[i], genres[j] = genres[j], genres[i]
			}
		}
	}

	count := 5
	if len(genres) < count {
		count = len(genres)
	}
	for i := range count {
		fmt.Printf("  • %s: %d episodes\n", genres[i].genre, genres[i].count)
	}
}

// printShowStats prints show statistics in a formatted way.
func printShowStats(showStats map[string]int) {
	type showStat struct {
		title string
		count int
	}

	var shows []showStat
	for title, count := range showStats {
		shows = append(shows, showStat{title, count})
	}

	// Sort by count descending
	for i := range len(shows) - 1 {
		for j := i + 1; j < len(shows); j++ {
			if shows[j].count > shows[i].count {
				shows[i], shows[j] = shows[j], shows[i]
			}
		}
	}

	count := 10
	if len(shows) < count {
		count = len(shows)
	}
	for i := range count {
		fmt.Printf("  • %s: %d episodes\n", shows[i].title, shows[i].count)
	}
}

// getShowStatus returns a simplified status for a show.
func getShowStatus(show *shows.Show) string {
	if show.Continuing {
		return "Continuing"
	}
	return "Ended"
}
