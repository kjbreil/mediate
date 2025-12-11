package mcp

import (
	"context"
	"log/slog"

	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/kjbreil/mediate/pkg/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MediateServer implements the MCP server for Mediate.
type MediateServer struct {
	mediate *mediate.Mediate
	logger  *slog.Logger
	server  *server.MCPServer
}

// NewMediateServer creates a new MCP server instance.
func NewMediateServer(m *mediate.Mediate, logger *slog.Logger) *MediateServer {
	srv := &MediateServer{
		mediate: m,
		logger:  logger,
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"mediate",
		"1.0.0",
		server.WithLogging(),
	)

	srv.server = mcpServer
	srv.registerTools()
	srv.registerResources()

	return srv
}

// registerTools registers all MCP tools.
func (s *MediateServer) registerTools() {
	// Viewing habits analysis tool
	s.server.AddTool(mcp.Tool{
		Name:        "analyze_viewing_habits",
		Description: "Analyze user viewing patterns and habits",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"timeframe": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"week", "month", "quarter", "year", "all"},
					"description": "Time period to analyze",
				},
				"analysis_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"genres", "shows", "patterns", "completion_rate"},
					"description": "Type of analysis to perform",
				},
			},
		},
	}, s.handleAnalyzeViewingHabits)

	// Recommendations tool
	s.server.AddTool(mcp.Tool{
		Name:        "get_recommendations",
		Description: "Get personalized show/movie recommendations",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"shows", "movies", "both"},
					"description": "Type of media to recommend",
				},
				"basis": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"viewing_history", "similar_shows", "popular", "new_releases"},
					"description": "Basis for recommendations",
				},
			},
		},
	}, s.handleGetRecommendations)

	// Search media tool
	s.server.AddTool(mcp.Tool{
		Name:        "search_media",
		Description: "Search for shows or movies across Plex, Sonarr, and Radarr",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query (title, actor, director, etc.)",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"shows", "movies", "both"},
					"description": "Type of media to search",
				},
				"source": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"plex", "sonarr", "radarr", "all"},
					"description": "Which service to search",
				},
			},
			Required: []string{"query"},
		},
	}, s.handleSearchMedia)

	// Add to downloads tool
	s.server.AddTool(mcp.Tool{
		Name:        "add_to_downloads",
		Description: "Add shows or movies to download queue",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"items": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title": map[string]interface{}{
								"type": "string",
							},
							"type": map[string]interface{}{
								"type": "string",
								"enum": []string{"show", "movie"},
							},
							"tvdb_id": map[string]interface{}{
								"type": "integer",
							},
							"monitor": map[string]interface{}{
								"type":    "boolean",
								"default": true,
							},
						},
						"required": []string{"title", "type"},
					},
				},
				"quality_profile": map[string]interface{}{
					"type":        "string",
					"description": "Quality profile to use",
				},
			},
			Required: []string{"items"},
		},
	}, s.handleAddToDownloads)

	// System status tool
	s.server.AddTool(mcp.Tool{
		Name:        "get_system_status",
		Description: "Get status of Mediate system and connected services",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"detailed": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "Include detailed service information",
				},
			},
		},
	}, s.handleGetSystemStatus)

	// Individual show analysis tool
	s.server.AddTool(mcp.Tool{
		Name:        "analyze_show",
		Description: "Analyze viewing habits for a specific show with per-user breakdown",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"show_title": map[string]interface{}{
					"type":        "string",
					"description": "Title of the show to analyze",
				},
				"tvdb_id": map[string]interface{}{
					"type":        "integer",
					"description": "TVDB ID of the show (optional, alternative to title)",
				},
				"timeframe": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"week", "month", "quarter", "year", "all"},
					"description": "Time period to analyze",
					"default":     "all",
				},
				"user": map[string]interface{}{
					"type":        "string",
					"description": "Specific Plex username to analyze (optional)",
				},
			},
		},
	}, s.handleAnalyzeShow)

	// Episode analysis tool
	s.server.AddTool(mcp.Tool{
		Name:        "analyze_episodes",
		Description: "Analyze viewing data for individual episodes with detailed metrics",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"show_title": map[string]interface{}{
					"type":        "string",
					"description": "Title of the show to analyze episodes for",
				},
				"tvdb_id": map[string]interface{}{
					"type":        "integer",
					"description": "TVDB ID of the show (optional, alternative to title)",
				},
				"season": map[string]interface{}{
					"type":        "integer",
					"description": "Specific season to analyze (optional)",
				},
				"user": map[string]interface{}{
					"type":        "string",
					"description": "Specific Plex username to analyze (optional)",
				},
				"sort_by": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"view_count", "completion_rate", "air_date", "episode_number"},
					"description": "How to sort the episode results",
					"default":     "episode_number",
				},
			},
		},
	}, s.handleAnalyzeEpisodes)

	// Deleted media analysis tool
	s.server.AddTool(mcp.Tool{
		Name:        "analyze_deleted_media",
		Description: "Analyze viewing data for deleted/orphaned Plex media",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"scan", "summary", "list", "details"},
					"description": "Action to perform: scan for new deletions, get summary, list deleted media, or get details",
					"default":     "summary",
				},
				"rating_key": map[string]interface{}{
					"type":        "string",
					"description": "Specific Plex rating key for details action",
				},
				"media_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"movie", "show", "episode", "all"},
					"description": "Filter by media type",
					"default":     "all",
				},
				"library_id": map[string]interface{}{
					"type":        "integer",
					"description": "Filter by library section ID",
				},
			},
		},
	}, s.handleAnalyzeDeletedMedia)

	// Scan for deleted media tool
	s.server.AddTool(mcp.Tool{
		Name:        "scan_deleted_media",
		Description: "Scan Plex for orphaned viewing history records",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"force_rescan": map[string]interface{}{
					"type":        "boolean",
					"description": "Force a complete rescan even if recently scanned",
					"default":     false,
				},
			},
		},
	}, s.handleScanDeletedMedia)
}

// registerResources registers all MCP resources.
func (s *MediateServer) registerResources() {
	// Viewing history resource
	s.server.AddResource(mcp.Resource{
		URI:         "viewing://history",
		Name:        "Viewing History",
		Description: "Real-time access to user viewing history",
		MIMEType:    "application/json",
	}, s.handleViewingHistoryResource)

	// Library stats resource
	s.server.AddResource(mcp.Resource{
		URI:         "library://stats",
		Name:        "Library Statistics",
		Description: "Current library statistics and information",
		MIMEType:    "application/json",
	}, s.handleLibraryStatsResource)

	// Download queue resource
	s.server.AddResource(mcp.Resource{
		URI:         "downloads://queue",
		Name:        "Download Queue",
		Description: "Current download queue status and progress",
		MIMEType:    "application/json",
	}, s.handleDownloadQueueResource)
}

// Start starts the MCP server with stdio transport.
func (s *MediateServer) Start(ctx context.Context) error {
	s.logger.Info("Starting Mediate MCP server")

	// Start with stdio transport (for Claude Desktop integration)
	return server.ServeStdio(s.server)
}

// Close closes the MCP server and cleans up resources.
func (s *MediateServer) Close() error {
	s.logger.Info("Shutting down Mediate MCP server")
	return nil
}

// GetMediate returns the underlying Mediate instance.
func (s *MediateServer) GetMediate() *mediate.Mediate {
	return s.mediate
}

// GetDB returns the database instance.
func (s *MediateServer) GetDB() *store.Store {
	return s.mediate.DB
}
