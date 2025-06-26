package mcp

import (
	"context"
	"log/slog"

	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MediateServer implements the MCP server for Mediate
type MediateServer struct {
	mediate *mediate.Mediate
	logger  *slog.Logger
	server  *server.MCPServer
}

// NewMediateServer creates a new MCP server instance
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

// registerTools registers all MCP tools
func (s *MediateServer) registerTools() {
	// Viewing habits analysis tool
	s.server.AddTool(mcp.Tool{
		Name:        "analyze_viewing_habits",
		Description: "Analyze user viewing patterns and habits",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"timeframe": map[string]interface{}{
					"type": "string",
					"enum": []string{"week", "month", "quarter", "year", "all"},
					"description": "Time period to analyze",
				},
				"analysis_type": map[string]interface{}{
					"type": "string",
					"enum": []string{"genres", "shows", "patterns", "completion_rate"},
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
					"type": "string",
					"enum": []string{"shows", "movies", "both"},
					"description": "Type of media to recommend",
				},
				"basis": map[string]interface{}{
					"type": "string",
					"enum": []string{"viewing_history", "similar_shows", "popular", "new_releases"},
					"description": "Basis for recommendations",
				},
				"limit": map[string]interface{}{
					"type": "integer",
					"default": 10,
					"description": "Number of recommendations",
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
					"type": "string",
					"description": "Search query (title, actor, director, etc.)",
				},
				"type": map[string]interface{}{
					"type": "string",
					"enum": []string{"shows", "movies", "both"},
					"description": "Type of media to search",
				},
				"source": map[string]interface{}{
					"type": "string",
					"enum": []string{"plex", "sonarr", "radarr", "all"},
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
								"type": "boolean",
								"default": true,
							},
						},
						"required": []string{"title", "type"},
					},
				},
				"quality_profile": map[string]interface{}{
					"type": "string",
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
					"type": "boolean",
					"default": false,
					"description": "Include detailed service information",
				},
			},
		},
	}, s.handleGetSystemStatus)
}

// registerResources registers all MCP resources
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

// Start starts the MCP server with stdio transport
func (s *MediateServer) Start(ctx context.Context) error {
	s.logger.Info("Starting Mediate MCP server")
	
	// Start with stdio transport (for Claude Desktop integration)
	return server.ServeStdio(s.server)
}

// Close closes the MCP server and cleans up resources
func (s *MediateServer) Close() error {
	s.logger.Info("Shutting down Mediate MCP server")
	return nil
}