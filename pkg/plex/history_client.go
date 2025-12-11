package plex

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
)

// HistoryClient handles Plex viewing history operations.
type HistoryClient struct {
	baseURL string
	token   string
	logger  *slog.Logger
	client  *http.Client
}

// NewHistoryClient creates a new Plex history client.
func NewHistoryClient(baseURL, token string, logger *slog.Logger) *HistoryClient {
	return &HistoryClient{
		baseURL: baseURL,
		token:   token,
		logger:  logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetViewingHistory retrieves all viewing history from Plex.
func (hc *HistoryClient) GetViewingHistory(ctx context.Context, start, size int) (*shows.PlexHistoryResponse, error) {
	url := fmt.Sprintf("%s/status/sessions/history/all", hc.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("X-Plex-Token", hc.token)
	if start > 0 {
		q.Add("X-Plex-Container-Start", strconv.Itoa(start))
	}
	if size > 0 {
		q.Add("X-Plex-Container-Size", strconv.Itoa(size))
	}
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("X-Plex-Token", hc.token)

	hc.logger.DebugContext(ctx, "Fetching Plex viewing history", "url", req.URL.String())

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch viewing history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("plex API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var historyResp shows.PlexHistoryResponse
	err = xml.Unmarshal(body, &historyResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	hc.logger.InfoContext(ctx, "Retrieved viewing history",
		"total_records", historyResp.MediaContainer.Size,
		"returned_records", len(historyResp.MediaContainer.Videos))

	return &historyResp, nil
}

// GetAllViewingHistory retrieves all viewing history with pagination.
func (hc *HistoryClient) GetAllViewingHistory(ctx context.Context) (*shows.PlexHistoryResponse, error) {
	const batchSize = 100
	var allVideos []shows.PlexHistoryVideo
	start := 0

	for {
		resp, err := hc.GetViewingHistory(ctx, start, batchSize)
		if err != nil {
			return nil, err
		}

		if len(resp.MediaContainer.Videos) == 0 {
			break
		}

		allVideos = append(allVideos, resp.MediaContainer.Videos...)

		// If we got fewer records than requested, we've reached the end
		if len(resp.MediaContainer.Videos) < batchSize {
			break
		}

		start += batchSize

		// Add a small delay to be respectful to the Plex server
		time.Sleep(100 * time.Millisecond)
	}

	// Return consolidated response
	return &shows.PlexHistoryResponse{
		MediaContainer: shows.PlexHistoryContainer{
			Size:   len(allVideos),
			Videos: allVideos,
		},
	}, nil
}

// CheckMediaExists verifies if media with given rating key still exists in Plex.
func (hc *HistoryClient) CheckMediaExists(ctx context.Context, ratingKey string) (bool, error) {
	url := fmt.Sprintf("%s/library/metadata/%s", hc.baseURL, ratingKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	// Add token
	q := req.URL.Query()
	q.Add("X-Plex-Token", hc.token)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("X-Plex-Token", hc.token)

	resp, err := hc.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check media existence: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
}

// DetectOrphanedRecords identifies viewing history records for deleted media.
func (hc *HistoryClient) DetectOrphanedRecords(ctx context.Context) ([]shows.PlexHistoryVideo, error) {
	hc.logger.InfoContext(ctx, "Starting orphaned record detection")

	history, err := hc.GetAllViewingHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get viewing history: %w", err)
	}

	var orphanedRecords []shows.PlexHistoryVideo
	processedRatingKeys := make(map[string]bool)

	for _, video := range history.MediaContainer.Videos {
		// Skip if we've already checked this rating key
		if processedRatingKeys[video.RatingKey] {
			continue
		}
		processedRatingKeys[video.RatingKey] = true

		var exists bool
		var checkErr error
		exists, checkErr = hc.CheckMediaExists(ctx, video.RatingKey)
		if checkErr != nil {
			hc.logger.WarnContext(ctx, "Failed to check media existence",
				"rating_key", video.RatingKey,
				"error", checkErr)
			continue
		}

		if !exists {
			hc.logger.DebugContext(ctx, "Found orphaned record",
				"rating_key", video.RatingKey,
				"title", video.Title)

			// Find all history records for this orphaned media
			for _, v := range history.MediaContainer.Videos {
				if v.RatingKey == video.RatingKey {
					orphanedRecords = append(orphanedRecords, v)
				}
			}
		}

		// Add a small delay to avoid overwhelming the server
		time.Sleep(50 * time.Millisecond)
	}

	hc.logger.InfoContext(ctx, "Orphaned record detection complete",
		"total_checked", len(processedRatingKeys),
		"orphaned_records", len(orphanedRecords))

	return orphanedRecords, nil
}

// ConvertToDeletedMedia converts Plex history records to DeletedMedia structs.
func (hc *HistoryClient) ConvertToDeletedMedia(orphanedRecords []shows.PlexHistoryVideo) []*shows.DeletedMedia {
	mediaMap := make(map[string]*shows.DeletedMedia)

	for _, record := range orphanedRecords {
		if existing, exists := mediaMap[record.RatingKey]; exists {
			// Add session to existing media
			session := &shows.DeletedMediaSession{
				HistoryKey: record.HistoryKey,
				AccountID:  record.AccountID,
				ViewedAt:   time.Unix(record.ViewedAt, 0),
				Duration:   time.Duration(record.Duration) * time.Millisecond,
				DeviceType: hc.getDeviceTypeFromID(record.DeviceID),
			}
			existing.ViewingSessions = append(existing.ViewingSessions, *session)
			existing.TotalViews++

			// Update timing info
			viewedAt := time.Unix(record.ViewedAt, 0)
			if existing.FirstViewedAt == nil || viewedAt.Before(*existing.FirstViewedAt) {
				existing.FirstViewedAt = &viewedAt
			}
			if existing.LastViewedAt == nil || viewedAt.After(*existing.LastViewedAt) {
				existing.LastViewedAt = &viewedAt
			}
		} else {
			// Create new deleted media record
			viewedAt := time.Unix(record.ViewedAt, 0)
			media := &shows.DeletedMedia{
				RatingKey:        record.RatingKey,
				HistoryKey:       record.HistoryKey,
				LibrarySectionID: record.LibrarySectionID,
				Title:            record.Title,
				MediaType:        record.Type,
				Year:             record.Year,
				Duration:         time.Duration(record.Duration) * time.Millisecond,
				ShowTitle:        record.GrandparentTitle,
				SeasonNumber:     record.ParentIndex,
				EpisodeNumber:    record.Index,
				LastSeenAt:       time.Now().AddDate(0, 0, -1), // Estimate
				DeletedAt:        time.Now(),
				DetectionMethod:  "api_404_check",
				TotalViews:       1,
				FirstViewedAt:    &viewedAt,
				LastViewedAt:     &viewedAt,
				ViewingSessions: []shows.DeletedMediaSession{
					{
						HistoryKey: record.HistoryKey,
						AccountID:  record.AccountID,
						ViewedAt:   viewedAt,
						Duration:   time.Duration(record.Duration) * time.Millisecond,
						DeviceType: hc.getDeviceTypeFromID(record.DeviceID),
					},
				},
			}
			mediaMap[record.RatingKey] = media
		}
	}

	// Convert map to slice
	var result []*shows.DeletedMedia
	for _, media := range mediaMap {
		// Calculate total watch time
		var totalWatchTime time.Duration
		for _, session := range media.ViewingSessions {
			totalWatchTime += session.Duration
		}
		media.TotalWatchTime = totalWatchTime

		result = append(result, media)
	}

	return result
}

// getDeviceTypeFromID maps device ID to device type (simplified).
func (hc *HistoryClient) getDeviceTypeFromID(deviceID int) string {
	// This is a simplified mapping - in practice you'd want to query
	// the devices endpoint to get actual device information
	switch deviceID {
	case 0:
		return "Unknown"
	default:
		return fmt.Sprintf("Device-%d", deviceID)
	}
}
