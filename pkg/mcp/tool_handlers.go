package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
	"github.com/mark3labs/mcp-go/mcp"
)

// handleAnalyzeViewingHabits handles the analyze_viewing_habits tool.
func (s *MediateServer) handleAnalyzeViewingHabits(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling analyze_viewing_habits request")

	// Parse arguments
	timeframe := "month"    // default
	analysisType := "shows" // default

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if tf, exists := args["timeframe"]; exists {
			if tfStr, ok := tf.(string); ok {
				timeframe = tfStr
			}
		}
		if at, exists := args["analysis_type"]; exists {
			if atStr, ok := at.(string); ok {
				analysisType = atStr
			}
		}
	}

	// Get shows data
	shows := s.mediate.GetShows()
	if shows == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: Unable to retrieve shows data"),
			},
			IsError: true,
		}, nil
	}

	// Perform analysis
	analysis := &ViewingAnalysis{
		Timeframe:    timeframe,
		AnalysisType: analysisType,
		GeneratedAt:  time.Now(),
		Data:         make(map[string]interface{}),
	}

	switch analysisType {
	case "genres":
		analysis.Summary, analysis.Data = s.analyzeGenres(shows, timeframe)
	case "shows":
		analysis.Summary, analysis.Data = s.analyzeShows(shows, timeframe)
	case "patterns":
		analysis.Summary, analysis.Data = s.analyzePatterns(shows, timeframe)
	case "completion_rate":
		analysis.Summary, analysis.Data = s.analyzeCompletionRate(shows, timeframe)
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error: Unknown analysis type: %s", analysisType)),
			},
			IsError: true,
		}, nil
	}

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling analysis: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleGetRecommendations handles the get_recommendations tool.
func (s *MediateServer) handleGetRecommendations(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get_recommendations request")

	// Parse arguments
	mediaType := "shows"       // default
	basis := "viewing_history" // default

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if mt, exists := args["type"]; exists {
			if mtStr, ok := mt.(string); ok {
				mediaType = mtStr
			}
		}
		if b, exists := args["basis"]; exists {
			if bStr, ok := b.(string); ok {
				basis = bStr
			}
		}
	}

	// Generate recommendations
	recommendations := s.generateRecommendations(mediaType, basis, 0)

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(recommendations, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling recommendations: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleSearchMedia handles the search_media tool.
func (s *MediateServer) handleSearchMedia(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling search_media request")

	// Parse arguments
	var query string
	mediaType := "both" // default
	source := "all"     // default

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if q, exists := args["query"]; exists {
			if qStr, ok := q.(string); ok {
				query = qStr
			}
		}
		if mt, exists := args["type"]; exists {
			if mtStr, ok := mt.(string); ok {
				mediaType = mtStr
			}
		}
		if src, exists := args["source"]; exists {
			if srcStr, ok := src.(string); ok {
				source = srcStr
			}
		}
	}

	if query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: Query parameter is required"),
			},
			IsError: true,
		}, nil
	}

	// Perform search
	results := s.searchMedia(query, mediaType, source)

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling search results: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleAddToDownloads handles the add_to_downloads tool.
func (s *MediateServer) handleAddToDownloads(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling add_to_downloads request")

	// Parse arguments
	var items []*DownloadItem
	qualityProfile := ""

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if itemsRaw, exists := args["items"]; exists {
			if itemsJSON, err := json.Marshal(itemsRaw); err == nil {
				json.Unmarshal(itemsJSON, &items)
			}
		}
		if qp, exists := args["quality_profile"]; exists {
			if qpStr, ok := qp.(string); ok {
				qualityProfile = qpStr
			}
		}
	}

	if len(items) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: No items provided for download"),
			},
			IsError: true,
		}, nil
	}

	// Add to downloads
	response := s.addToDownloads(items, qualityProfile)

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling download response: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleGetSystemStatus handles the get_system_status tool.
func (s *MediateServer) handleGetSystemStatus(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get_system_status request")

	// Parse arguments
	detailed := false
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if d, exists := args["detailed"]; exists {
			if dBool, ok := d.(bool); ok {
				detailed = dBool
			}
		}
	}

	// Get system status
	status := s.getSystemStatus(detailed)

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling system status: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleAnalyzeShow handles the analyze_show tool.
func (s *MediateServer) handleAnalyzeShow(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling analyze_show request")

	// Parse arguments
	var showTitle string
	var tvdbID int
	timeframe := "all"
	var user string

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if st, exists := args["show_title"]; exists {
			if stStr, ok := st.(string); ok {
				showTitle = stStr
			}
		}
		if tid, exists := args["tvdb_id"]; exists {
			if tidFloat, ok := tid.(float64); ok {
				tvdbID = int(tidFloat)
			}
		}
		if tf, exists := args["timeframe"]; exists {
			if tfStr, ok := tf.(string); ok {
				timeframe = tfStr
			}
		}
		if u, exists := args["user"]; exists {
			if uStr, ok := u.(string); ok {
				user = uStr
			}
		}
	}

	// Find the show
	var targetShow *shows.Show
	if tvdbID > 0 {
		targetShow = s.mediate.DB.GetShow(tvdbID)
	} else if showTitle != "" {
		// Search for show by title
		allShows := s.mediate.GetShows()
		if allShows != nil {
			for _, show := range *allShows {
				if strings.EqualFold(show.Title, showTitle) {
					targetShow = show
					break
				}
			}
		}
	}

	if targetShow == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: Show not found"),
			},
			IsError: true,
		}, nil
	}

	// Analyze the show
	analysis := s.analyzeIndividualShow(targetShow, timeframe, user)

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling show analysis: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleAnalyzeEpisodes handles the analyze_episodes tool.
func (s *MediateServer) handleAnalyzeEpisodes(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling analyze_episodes request")

	// Parse arguments
	var showTitle string
	var tvdbID int
	var season int
	var user string
	sortBy := "episode_number"

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if st, exists := args["show_title"]; exists {
			if stStr, ok := st.(string); ok {
				showTitle = stStr
			}
		}
		if tid, exists := args["tvdb_id"]; exists {
			if tidFloat, ok := tid.(float64); ok {
				tvdbID = int(tidFloat)
			}
		}
		if s, exists := args["season"]; exists {
			if sFloat, ok := s.(float64); ok {
				season = int(sFloat)
			}
		}
		if u, exists := args["user"]; exists {
			if uStr, ok := u.(string); ok {
				user = uStr
			}
		}
		if sb, exists := args["sort_by"]; exists {
			if sbStr, ok := sb.(string); ok {
				sortBy = sbStr
			}
		}
	}

	// Find the show
	var targetShow *shows.Show
	if tvdbID > 0 {
		targetShow = s.mediate.DB.GetShow(tvdbID)
	} else if showTitle != "" {
		// Search for show by title
		allShows := s.mediate.GetShows()
		if allShows != nil {
			for _, show := range *allShows {
				if strings.EqualFold(show.Title, showTitle) {
					targetShow = show
					break
				}
			}
		}
	}

	if targetShow == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: Show not found"),
			},
			IsError: true,
		}, nil
	}

	// Analyze the episodes
	analysis := s.analyzeEpisodes(targetShow, season, user, sortBy, 0)

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling episode analysis: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleAnalyzeDeletedMedia handles the analyze_deleted_media tool.
func (s *MediateServer) handleAnalyzeDeletedMedia(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling analyze_deleted_media request")

	// Parse arguments
	action := "summary"
	var ratingKey string
	mediaType := "all"
	var libraryID int

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if a, exists := args["action"]; exists {
			if aStr, ok := a.(string); ok {
				action = aStr
			}
		}
		if rk, exists := args["rating_key"]; exists {
			if rkStr, ok := rk.(string); ok {
				ratingKey = rkStr
			}
		}
		if mt, exists := args["media_type"]; exists {
			if mtStr, ok := mt.(string); ok {
				mediaType = mtStr
			}
		}
		if lid, exists := args["library_id"]; exists {
			if lidFloat, ok := lid.(float64); ok {
				libraryID = int(lidFloat)
			}
		}
	}

	var result interface{}
	var err error

	switch action {
	case "summary":
		result, err = s.mediate.DB.GetDeletedMediaSummary()
	case "list":
		var deletedMedia []*shows.DeletedMedia
		deletedMedia, err = s.mediate.DB.GetDeletedMedia(0, 0) // Get all
		if err == nil {
			// Filter by media type and library if specified
			var filtered []*shows.DeletedMedia
			for _, media := range deletedMedia {
				if mediaType != "all" && media.MediaType != mediaType {
					continue
				}
				if libraryID > 0 && media.LibrarySectionID != libraryID {
					continue
				}
				filtered = append(filtered, media)
			}
			result = filtered
		}
	case "details":
		if ratingKey == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.NewTextContent("Error: rating_key required for details action"),
				},
				IsError: true,
			}, nil
		}
		result, err = s.mediate.DB.GetDeletedMediaByRatingKey(ratingKey)
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error: Unknown action: %s", action)),
			},
			IsError: true,
		}, nil
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error retrieving deleted media data: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling deleted media data: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}

// handleScanDeletedMedia handles the scan_deleted_media tool.
func (s *MediateServer) handleScanDeletedMedia(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling scan_deleted_media request")

	// Parse arguments
	forceRescan := false
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if fr, exists := args["force_rescan"]; exists {
			if frBool, ok := fr.(bool); ok {
				forceRescan = frBool
			}
		}
	}

	// Create Plex history client - placeholder implementation
	// Note: You would need to implement actual Plex configuration access
	s.logger.Info("Starting orphaned record detection scan", "force_rescan", forceRescan)

	// For now, return a placeholder result since we need Plex config integration
	result := map[string]interface{}{
		"scan_completed_at": time.Now(),
		"status":            "scan_not_implemented",
		"message":           "Plex configuration integration needed for actual scanning",
		"force_rescan":      forceRescan,
	}

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling scan results: %v", err)),
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(resultJSON)),
		},
	}, nil
}
