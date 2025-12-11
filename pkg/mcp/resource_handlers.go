package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleViewingHistoryResource handles the viewing history resource.
func (s *MediateServer) handleViewingHistoryResource(
	ctx context.Context,
	request mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	s.logger.InfoContext(ctx, "Handling viewing history resource request")

	shows := s.mediate.GetShows()
	if shows == nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     `{"error": "Unable to retrieve viewing history"}`,
			},
		}, nil
	}

	// Build viewing history
	history := make([]map[string]interface{}, 0)

	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Watched && episode.LastViewedAt != nil {
				history = append(history, map[string]interface{}{
					"show_title":    show.Title,
					"episode_title": episode.Title,
					"season":        episode.Season,
					"episode":       episode.Episode,
					"watched_at":    episode.LastViewedAt,
					"duration":      episode.Duration,
					"rating":        show.Rating,
				})
			}
		}
	}

	// Sort by watch date (most recent first)
	// In a real implementation, you'd sort this properly

	result := map[string]interface{}{
		"total_watched": len(history),
		"history":       history,
		"generated_at":  time.Now(),
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     `{"error": "Failed to marshal viewing history"}`,
			},
		}, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(resultJSON),
		},
	}, nil
}

// handleLibraryStatsResource handles the library statistics resource.
func (s *MediateServer) handleLibraryStatsResource(
	ctx context.Context,
	request mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	s.logger.InfoContext(ctx, "Handling library stats resource request")

	shows := s.mediate.GetShows()
	if shows == nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     `{"error": "Unable to retrieve library statistics"}`,
			},
		}, nil
	}

	// Calculate statistics
	totalShows := len(*shows)
	totalEpisodes := 0
	watchedEpisodes := 0
	monitoredShows := 0
	continuingShows := 0

	for _, show := range *shows {
		totalEpisodes += len(show.Episodes)

		if show.Continuing {
			continuingShows++
		}

		// Check if show has monitored episodes
		hasMonitored := false
		for _, episode := range show.Episodes {
			if episode.Watched {
				watchedEpisodes++
			}
			if episode.Wanted {
				hasMonitored = true
			}
		}

		if hasMonitored {
			monitoredShows++
		}
	}

	stats := map[string]interface{}{
		"total_shows":      totalShows,
		"total_episodes":   totalEpisodes,
		"watched_episodes": watchedEpisodes,
		"monitored_shows":  monitoredShows,
		"continuing_shows": continuingShows,
		"completion_rate":  float64(watchedEpisodes) / float64(totalEpisodes) * 100,
		"last_updated":     time.Now(),
	}

	resultJSON, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     `{"error": "Failed to marshal library statistics"}`,
			},
		}, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(resultJSON),
		},
	}, nil
}

// handleDownloadQueueResource handles the download queue resource.
func (s *MediateServer) handleDownloadQueueResource(
	ctx context.Context,
	request mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	s.logger.InfoContext(ctx, "Handling download queue resource request")

	shows := s.mediate.GetShows()
	if shows == nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     `{"error": "Unable to retrieve download queue"}`,
			},
		}, nil
	}

	// Build download queue
	queue := make([]map[string]interface{}, 0)

	for _, show := range *shows {
		for _, episode := range show.Episodes {
			if episode.Downloading {
				queue = append(queue, map[string]interface{}{
					"show_title":    show.Title,
					"episode_title": episode.Title,
					"season":        episode.Season,
					"episode":       episode.Episode,
					"status":        "downloading",
					"has_file":      episode.HasFile,
					"air_date":      episode.AirDate,
				})
			} else if episode.Wanted && !episode.HasFile {
				queue = append(queue, map[string]interface{}{
					"show_title":    show.Title,
					"episode_title": episode.Title,
					"season":        episode.Season,
					"episode":       episode.Episode,
					"status":        "wanted",
					"has_file":      episode.HasFile,
					"air_date":      episode.AirDate,
				})
			}
		}
	}

	result := map[string]interface{}{
		"queue_size":   len(queue),
		"queue":        queue,
		"last_updated": time.Now(),
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     `{"error": "Failed to marshal download queue"}`,
			},
		}, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(resultJSON),
		},
	}, nil
}
