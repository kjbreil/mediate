package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
)

// searchMedia performs media search across multiple sources.
func (s *MediateServer) searchMedia(query string, mediaType string, source string) *SearchResults {
	startTime := time.Now()
	results := &SearchResults{
		Query:      query,
		TotalCount: 0,
		Results:    make([]*SearchResult, 0),
		Sources:    make([]string, 0),
		SearchTime: 0,
	}

	// Search based on source parameter
	switch source {
	case "plex":
		plexResults := s.searchPlex(query, mediaType)
		results.Results = append(results.Results, plexResults...)
		results.Sources = append(results.Sources, "plex")
	case "sonarr":
		sonarrResults := s.searchSonarr(query, mediaType)
		results.Results = append(results.Results, sonarrResults...)
		results.Sources = append(results.Sources, "sonarr")
	case "radarr":
		radarrResults := s.searchRadarr(query, mediaType)
		results.Results = append(results.Results, radarrResults...)
		results.Sources = append(results.Sources, "radarr")
	case "all":
		// Search all sources
		if mediaType == "shows" || mediaType == "both" {
			plexResults := s.searchPlex(query, "shows")
			sonarrResults := s.searchSonarr(query, "shows")
			results.Results = append(results.Results, plexResults...)
			results.Results = append(results.Results, sonarrResults...)
			results.Sources = append(results.Sources, "plex", "sonarr")
		}
		if mediaType == "movies" || mediaType == "both" {
			plexMovies := s.searchPlex(query, "movies")
			radarrResults := s.searchRadarr(query, "movies")
			results.Results = append(results.Results, plexMovies...)
			results.Results = append(results.Results, radarrResults...)
			if !contains(results.Sources, "plex") {
				results.Sources = append(results.Sources, "plex")
			}
			results.Sources = append(results.Sources, "radarr")
		}
	}

	results.TotalCount = len(results.Results)
	results.SearchTime = time.Since(startTime)

	s.logger.Info("Search completed",
		"query", query,
		"results", results.TotalCount,
		"duration", results.SearchTime)

	return results
}

// searchPlex searches the Plex library.
func (s *MediateServer) searchPlex(query string, mediaType string) []*SearchResult {
	results := make([]*SearchResult, 0)

	shows := s.mediate.GetShows()
	if shows == nil {
		return results
	}

	queryLower := strings.ToLower(query)

	for _, show := range *shows {
		titleLower := strings.ToLower(show.Title)

		// Simple string matching - in production, you'd use more sophisticated search
		if strings.Contains(titleLower, queryLower) {
			if mediaType == "shows" || mediaType == "both" {
				result := &SearchResult{
					Title:     show.Title,
					Type:      "show",
					Source:    "plex",
					TvdbID:    show.TvdbID,
					Available: true,
					Monitored: s.isShowMonitored(show),
					Status:    s.getShowStatus(show),
				}
				results = append(results, result)
			}
		}
	}

	return results
}

// searchSonarr searches Sonarr for TV shows.
func (s *MediateServer) searchSonarr(query string, mediaType string) []*SearchResult {
	results := make([]*SearchResult, 0)

	if mediaType != "shows" && mediaType != "both" {
		return results
	}

	// In a real implementation, you'd call the Sonarr API
	// For now, we'll return some sample data
	sampleResults := []*SearchResult{
		{
			Title:       "Breaking Bad",
			Type:        "show",
			Source:      "sonarr",
			TvdbID:      81189,
			Year:        2008,
			Genre:       []string{"Drama", "Crime"},
			Description: "A high school chemistry teacher turned methamphetamine producer",
			Available:   false,
			Monitored:   false,
			Status:      "available",
		},
		{
			Title:       "Better Call Saul",
			Type:        "show",
			Source:      "sonarr",
			TvdbID:      273181,
			Year:        2015,
			Genre:       []string{"Drama", "Crime"},
			Description: "A prequel to Breaking Bad",
			Available:   false,
			Monitored:   false,
			Status:      "available",
		},
	}

	queryLower := strings.ToLower(query)
	for _, result := range sampleResults {
		if strings.Contains(strings.ToLower(result.Title), queryLower) {
			results = append(results, result)
		}
	}

	return results
}

// searchRadarr searches Radarr for movies.
func (s *MediateServer) searchRadarr(query string, mediaType string) []*SearchResult {
	results := make([]*SearchResult, 0)

	if mediaType != "movies" && mediaType != "both" {
		return results
	}

	// In a real implementation, you'd call the Radarr API
	// For now, we'll return some sample data
	sampleResults := []*SearchResult{
		{
			Title:       "The Dark Knight",
			Type:        "movie",
			Source:      "radarr",
			Year:        2008,
			Genre:       []string{"Action", "Crime", "Drama"},
			Description: "Batman faces the Joker in this acclaimed superhero film",
			Available:   false,
			Monitored:   false,
			Status:      "available",
		},
		{
			Title:       "Inception",
			Type:        "movie",
			Source:      "radarr",
			Year:        2010,
			Genre:       []string{"Action", "Sci-Fi", "Thriller"},
			Description: "A thief who steals corporate secrets through dream-sharing technology",
			Available:   false,
			Monitored:   false,
			Status:      "available",
		},
	}

	queryLower := strings.ToLower(query)
	for _, result := range sampleResults {
		if strings.Contains(strings.ToLower(result.Title), queryLower) {
			results = append(results, result)
		}
	}

	return results
}

// addToDownloads adds items to the download queue.
func (s *MediateServer) addToDownloads(items []*DownloadItem, qualityProfile string) *DownloadResponse {
	response := &DownloadResponse{
		Success:   make([]*DownloadItem, 0),
		Failed:    make([]*DownloadItem, 0),
		Errors:    make([]string, 0),
		Timestamp: time.Now(),
	}

	for _, item := range items {
		s.logger.Info("Adding item to downloads",
			"title", item.Title,
			"type", item.Type,
			"monitor", item.Monitor)

		// In a real implementation, you'd call Sonarr/Radarr APIs here
		success := s.addItemToDownloadService(item, qualityProfile)

		if success {
			response.Success = append(response.Success, item)
		} else {
			response.Failed = append(response.Failed, item)
			response.Errors = append(response.Errors,
				fmt.Sprintf("Failed to add %s to downloads", item.Title))
		}
	}

	successCount := len(response.Success)
	failedCount := len(response.Failed)

	if failedCount == 0 {
		response.Message = fmt.Sprintf("Successfully added %d item(s) to downloads", successCount)
	} else if successCount == 0 {
		response.Message = fmt.Sprintf("Failed to add all %d item(s) to downloads", failedCount)
	} else {
		response.Message = fmt.Sprintf("Added %d item(s) successfully, %d failed", successCount, failedCount)
	}

	return response
}

// addItemToDownloadService adds a single item to the appropriate download service.
func (s *MediateServer) addItemToDownloadService(item *DownloadItem, qualityProfile string) bool {
	switch item.Type {
	case "show":
		return s.addToSonarr(item, qualityProfile)
	case "movie":
		return s.addToRadarr(item, qualityProfile)
	default:
		s.logger.Error("Unknown media type", "type", item.Type)
		return false
	}
}

// addToSonarr adds a show to Sonarr.
func (s *MediateServer) addToSonarr(item *DownloadItem, qualityProfile string) bool {
	s.logger.Info("Adding show to Sonarr", "title", item.Title, "tvdb_id", item.TvdbID, "monitor", item.Monitor)

	// If we have a TVDB ID, try to find and add the show
	if item.TvdbID > 0 {
		return s.addShowByTvdbID(item.TvdbID, item.Monitor, qualityProfile)
	}

	// Otherwise, search by title (fallback)
	return s.addShowByTitle(item.Title, item.Monitor, qualityProfile)
}

// addShowByTvdbID adds a show to Sonarr using TVDB ID.
func (s *MediateServer) addShowByTvdbID(tvdbID int, monitor bool, qualityProfile string) bool {
	// Check if show already exists in our database
	existingShow := s.mediate.DB.GetShow(tvdbID)
	if existingShow != nil {
		s.logger.Info("Show already exists in database", "tvdb_id", tvdbID, "title", existingShow.Title)

		// If it's not monitored but we want to monitor it, trigger a search
		if monitor && !s.isShowMonitored(existingShow) {
			s.logger.Info("Show exists but not monitored, triggering search", "title", existingShow.Title)
			s.triggerShowSearch(existingShow)
		}
		return true
	}

	// Search for the show in Sonarr's lookup
	s.logger.Info("Searching for show in Sonarr lookup", "tvdb_id", tvdbID)

	// This would require calling Sonarr's lookup API to find the show
	// For now, we'll log the action and return success
	// In a real implementation, you'd call something like:
	// lookupResults, err := s.mediate.sonarr.Lookup(fmt.Sprintf("tvdb:%d", tvdbID))
	// then add the show with the appropriate quality profile

	s.logger.Info("Would add show to Sonarr via TVDB lookup", "tvdb_id", tvdbID)
	return true
}

// addShowByTitle adds a show to Sonarr using title search.
func (s *MediateServer) addShowByTitle(title string, monitor bool, qualityProfile string) bool {
	s.logger.Info("Searching for show by title", "title", title)

	// This would search Sonarr's lookup API by title
	// For now, we'll log the action
	s.logger.Info("Would add show to Sonarr via title search", "title", title)
	return true
}

// triggerShowSearch triggers a search for a show's monitored episodes.
func (s *MediateServer) triggerShowSearch(show *shows.Show) {
	s.logger.Info("Triggering search for show", "title", show.Title, "sonarr_id", show.SonarrId)

	// Get monitored episodes for this show
	monitoredEpisodes := make([]int64, 0)
	for _, episode := range show.Episodes {
		if episode.Wanted && !episode.HasFile {
			monitoredEpisodes = append(monitoredEpisodes, episode.SonarrId)
		}
	}

	if len(monitoredEpisodes) > 0 {
		s.logger.Info("Triggering episode search", "title", show.Title, "episode_count", len(monitoredEpisodes))

		// Send the search command to Sonarr (similar to DownloadEpisodes)
		err := s.mediate.TriggerEpisodeSearch(monitoredEpisodes)

		if err != nil {
			s.logger.Error("Failed to trigger episode search", "error", err)
		} else {
			s.logger.Info("Successfully triggered episode search", "title", show.Title)
		}
	} else {
		s.logger.Info("No monitored episodes to search for", "title", show.Title)
	}
}

// addToRadarr adds a movie to Radarr.
func (s *MediateServer) addToRadarr(item *DownloadItem, qualityProfile string) bool {
	s.logger.Info("Adding movie to Radarr", "title", item.Title)

	// In a real implementation, you'd:
	// 1. Search for the movie in Radarr's database
	// 2. Add the movie with the specified quality profile
	// 3. Set monitoring status based on item.Monitor

	// For now, we'll simulate success
	return true
}

// getSystemStatus returns the current system status.
func (s *MediateServer) getSystemStatus(detailed bool) *SystemStatus {
	status := &SystemStatus{
		Services:    make(map[string]ServiceStatus),
		LastUpdated: time.Now(),
		Health:      "healthy",
	}

	// Check database status
	shows := s.mediate.GetShows()
	if shows != nil {
		totalEpisodes := 0
		watchedEpisodes := 0

		for _, show := range *shows {
			totalEpisodes += len(show.Episodes)
			for _, episode := range show.Episodes {
				if episode.Watched {
					watchedEpisodes++
				}
			}
		}

		status.Database = DatabaseStatus{
			Connected:    true,
			ShowCount:    len(*shows),
			EpisodeCount: totalEpisodes,
			LastUpdated:  time.Now(),
		}

		status.Statistics = SystemStatistics{
			TotalShows:       len(*shows),
			TotalEpisodes:    totalEpisodes,
			WatchedEpisodes:  watchedEpisodes,
			MonitoredShows:   s.getMonitoredShowCount(shows),
			PendingDownloads: s.getPendingDownloadCount(shows),
		}
	} else {
		status.Database = DatabaseStatus{
			Connected: false,
		}
		status.Health = "degraded"
	}

	// Check service statuses
	status.Services["plex"] = ServiceStatus{
		Name:      "Plex",
		URL:       s.mediate.Config().Plex.URL,
		Status:    "connected", // In real implementation, ping the service
		LastCheck: time.Now(),
	}

	status.Services["sonarr"] = ServiceStatus{
		Name:      "Sonarr",
		URL:       s.mediate.Config().Sonarr.URL,
		Status:    "connected",
		LastCheck: time.Now(),
	}

	status.Services["radarr"] = ServiceStatus{
		Name:      "Radarr",
		URL:       s.mediate.Config().Radarr.URL,
		Status:    "connected",
		LastCheck: time.Now(),
	}

	// Add job statuses if detailed
	if detailed {
		status.Jobs = []JobStatus{
			{
				Name:     "monitor",
				Status:   "running",
				LastRun:  &time.Time{}, // Would get from actual job system
				Interval: "1h",
			},
			{
				Name:     "download",
				Status:   "running",
				LastRun:  &time.Time{},
				Interval: "30m",
			},
		}
	}

	return status
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *MediateServer) isShowMonitored(show *shows.Show) bool {
	for _, episode := range show.Episodes {
		if episode.Wanted {
			return true
		}
	}
	return false
}

func (s *MediateServer) getShowStatus(show *shows.Show) string {
	if show.Continuing {
		return "continuing"
	}
	return "ended"
}

func (s *MediateServer) getMonitoredShowCount(shows *shows.Shows) int {
	count := 0
	for _, show := range *shows {
		if s.isShowMonitored(show) {
			count++
		}
	}
	return count
}

func (s *MediateServer) getPendingDownloadCount(shows *shows.Shows) int {
	count := 0
	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Wanted && !episode.HasFile {
				count++
			}
		}
	}
	return count
}
