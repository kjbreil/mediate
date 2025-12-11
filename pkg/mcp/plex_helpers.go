package mcp

import (
	"log/slog"

	"github.com/kjbreil/mediate/pkg/plex"
)

// NewPlexHistoryClient creates a new Plex history client.
func NewPlexHistoryClient(baseURL, token string, logger *slog.Logger) *plex.HistoryClient {
	return plex.NewHistoryClient(baseURL, token, logger)
}
