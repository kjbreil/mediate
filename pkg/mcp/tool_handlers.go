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
	s.logger.InfoContext(ctx, "Handling analyze_viewing_habits request")

	// Parse arguments
	timeframe := timeframeMonth    // default
	analysisType := mediaTypeShows // default

	//nolint:nestif // argument parsing requires type assertions for JSON interface{} values
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if tf, exists := args["timeframe"]; exists {
			if tfStr, tfOk := tf.(string); tfOk {
				timeframe = tfStr
			}
		}
		if at, exists := args["analysis_type"]; exists {
			if atStr, atOk := at.(string); atOk {
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
	s.logger.InfoContext(ctx, "Handling get_recommendations request")

	// Parse arguments
	mediaType := "shows"       // default
	basis := "viewing_history" // default

	//nolint:nestif // argument parsing requires type assertions for JSON interface{} values
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if mt, exists := args["type"]; exists {
			if mtStr, mtOk := mt.(string); mtOk {
				mediaType = mtStr
			}
		}
		if b, exists := args["basis"]; exists {
			if bStr, bOk := b.(string); bOk {
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
	s.logger.InfoContext(ctx, "Handling search_media request")

	// Parse arguments
	var query string
	mediaType := mediaTypeBoth // default
	source := sourceAll        // default

	//nolint:nestif // argument parsing requires type assertions for JSON interface{} values
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if q, exists := args["query"]; exists {
			if qStr, qOk := q.(string); qOk {
				query = qStr
			}
		}
		if mt, exists := args["type"]; exists {
			if mtStr, mtOk := mt.(string); mtOk {
				mediaType = mtStr
			}
		}
		if src, exists := args["source"]; exists {
			if srcStr, srcOk := src.(string); srcOk {
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
	s.logger.InfoContext(ctx, "Handling add_to_downloads request")

	// Parse arguments
	var items []*DownloadItem
	qualityProfile := ""

	//nolint:nestif // argument parsing requires type assertions and JSON marshaling for complex structures
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if itemsRaw, exists := args["items"]; exists {
			var itemsJSON []byte
			var unmarshalErr error
			itemsJSON, err := json.Marshal(itemsRaw)
			if err == nil {
				unmarshalErr = json.Unmarshal(itemsJSON, &items)
				if unmarshalErr != nil {
					s.logger.ErrorContext(ctx, "Failed to unmarshal items", "error", unmarshalErr)
				}
			}
		}
		if qp, exists := args["quality_profile"]; exists {
			if qpStr, qpOk := qp.(string); qpOk {
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
	s.logger.InfoContext(ctx, "Handling get_system_status request")

	// Parse arguments
	detailed := false
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if d, exists := args["detailed"]; exists {
			if dBool, dOk := d.(bool); dOk {
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
	s.logger.InfoContext(ctx, "Handling analyze_show request")

	// Parse arguments
	showTitle, tvdbID, timeframe, user := s.parseAnalyzeShowArgs(request)

	// Find the show
	targetShow := s.findShowByTitleOrID(showTitle, tvdbID)
	if targetShow == nil {
		return s.showNotFoundError(), nil
	}

	// Analyze the show
	analysis := s.analyzeIndividualShow(targetShow, timeframe, user)

	// Convert to JSON and return result
	return s.marshalAnalysisResult(analysis, "show analysis")
}

// handleAnalyzeEpisodes handles the analyze_episodes tool.
func (s *MediateServer) handleAnalyzeEpisodes(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.InfoContext(ctx, "Handling analyze_episodes request")

	// Parse arguments
	var showTitle string
	var tvdbID int
	var season int
	var user string
	sortBy := "episode_number"

	//nolint:nestif // argument parsing requires type assertions for JSON interface{} values
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if st, exists := args["show_title"]; exists {
			if stStr, stOk := st.(string); stOk {
				showTitle = stStr
			}
		}
		if tid, exists := args["tvdb_id"]; exists {
			if tidFloat, tidOk := tid.(float64); tidOk {
				tvdbID = int(tidFloat)
			}
		}
		if s, exists := args["season"]; exists {
			if sFloat, sOk := s.(float64); sOk {
				season = int(sFloat)
			}
		}
		if u, exists := args["user"]; exists {
			if uStr, uOk := u.(string); uOk {
				user = uStr
			}
		}
		if sb, exists := args["sort_by"]; exists {
			if sbStr, sbOk := sb.(string); sbOk {
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
	_ mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	s.logger.InfoContext(ctx, "Handling analyze_deleted_media request")

	// Get deleted media summary from the database
	summary, err := s.mediate.DB.GetDeletedMediaSummary()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error getting deleted media summary: %v", err)),
			},
			IsError: true,
		}, nil
	}

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(summary, "", "  ")
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
	s.logger.InfoContext(ctx, "Handling scan_deleted_media request")

	// Parse arguments
	forceRescan := false
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if fr, exists := args["force_rescan"]; exists {
			if frBool, frOk := fr.(bool); frOk {
				forceRescan = frBool
			}
		}
	}

	// Create Plex history client - placeholder implementation
	// Note: You would need to implement actual Plex configuration access
	s.logger.InfoContext(ctx, "Starting orphaned record detection scan", "force_rescan", forceRescan)

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

// Helper functions for handleAnalyzeShow

// parseAnalyzeShowArgs extracts arguments from the analyze_show request.
func (s *MediateServer) parseAnalyzeShowArgs(request mcp.CallToolRequest) (string, int, string, string) {
	var showTitle string
	var tvdbID int
	timeframe := "all"
	var user string

	//nolint:nestif // argument parsing requires type assertions for JSON interface{} values
	if args, argsOk := request.Params.Arguments.(map[string]interface{}); argsOk {
		if st, exists := args["show_title"]; exists {
			if stStr, stOk := st.(string); stOk {
				showTitle = stStr
			}
		}
		if tid, exists := args["tvdb_id"]; exists {
			if tidFloat, tidOk := tid.(float64); tidOk {
				tvdbID = int(tidFloat)
			}
		}
		if tf, exists := args["timeframe"]; exists {
			if tfStr, tfOk := tf.(string); tfOk {
				timeframe = tfStr
			}
		}
		if u, exists := args["user"]; exists {
			if uStr, uOk := u.(string); uOk {
				user = uStr
			}
		}
	}

	return showTitle, tvdbID, timeframe, user
}

// findShowByTitleOrID finds a show by TVDB ID or title.
func (s *MediateServer) findShowByTitleOrID(showTitle string, tvdbID int) *shows.Show {
	if tvdbID > 0 {
		return s.mediate.DB.GetShow(tvdbID)
	}

	if showTitle != "" {
		allShows := s.mediate.GetShows()
		if allShows != nil {
			for _, show := range *allShows {
				if strings.EqualFold(show.Title, showTitle) {
					return show
				}
			}
		}
	}

	return nil
}

// showNotFoundError returns an error result for show not found.
func (s *MediateServer) showNotFoundError() *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent("Error: Show not found"),
		},
		IsError: true,
	}
}

// marshalAnalysisResult marshals analysis data to JSON and returns a result.
func (s *MediateServer) marshalAnalysisResult(analysis interface{}, dataType string) (*mcp.CallToolResult, error) {
	resultJSON, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Error marshaling %s: %v", dataType, err)),
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
