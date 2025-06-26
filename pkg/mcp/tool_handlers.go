package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleAnalyzeViewingHabits handles the analyze_viewing_habits tool
func (s *MediateServer) handleAnalyzeViewingHabits(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling analyze_viewing_habits request")

	// Parse arguments
	timeframe := "month" // default
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

// handleGetRecommendations handles the get_recommendations tool
func (s *MediateServer) handleGetRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling get_recommendations request")

	// Parse arguments
	mediaType := "shows" // default
	basis := "viewing_history" // default
	limit := 10 // default

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
		if l, exists := args["limit"]; exists {
			if lFloat, ok := l.(float64); ok {
				limit = int(lFloat)
			}
		}
	}

	// Generate recommendations
	recommendations := s.generateRecommendations(mediaType, basis, limit)

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

// handleSearchMedia handles the search_media tool
func (s *MediateServer) handleSearchMedia(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.logger.Info("Handling search_media request")

	// Parse arguments
	var query string
	mediaType := "both" // default
	source := "all" // default

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

// handleAddToDownloads handles the add_to_downloads tool
func (s *MediateServer) handleAddToDownloads(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

// handleGetSystemStatus handles the get_system_status tool
func (s *MediateServer) handleGetSystemStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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