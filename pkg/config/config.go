package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Plex       Plex       `yaml:"plex"`
	Sonarr     Sonarr     `yaml:"sonarr"`
	Radarr     Radarr     `yaml:"radarr"`
	Database   Database   `yaml:"database"`
	Automation Automation `yaml:"automation"`
}

// Automation configuration for media management behavior.
type Automation struct {
	DeleteMaxRating      float64 `yaml:"delete_max_rating"`      // Delete episodes if show rating <= this (default: 3.0)
	KeepMinRating        float64 `yaml:"keep_min_rating"`        // Keep shows rated >= this selectively (default: 5.0)
	KeepAllMinRating     float64 `yaml:"keep_all_min_rating"`    // Keep ALL episodes for shows rated >= this (default: 9.0)
	WatchedCleanupDays   int     `yaml:"watched_cleanup_days"`   // Days after watching before deletion (default: 5)
	UnwatchedCleanupDays int     `yaml:"unwatched_cleanup_days"` // Days for unwatched episode cleanup (default: 30)
	EpisodesAhead        int     `yaml:"episodes_ahead"`         // Episodes to download ahead of current (default: 3)
}

// Default values for Automation config.
const (
	DefaultDeleteMaxRating      = 3.0
	DefaultKeepMinRating        = 5.0
	DefaultKeepAllMinRating     = 9.0
	DefaultWatchedCleanupDays   = 5
	DefaultUnwatchedCleanupDays = 30
	DefaultEpisodesAhead        = 3
)

// DefaultAutomation returns an Automation config with default values.
func DefaultAutomation() Automation {
	return Automation{
		DeleteMaxRating:      DefaultDeleteMaxRating,
		KeepMinRating:        DefaultKeepMinRating,
		KeepAllMinRating:     DefaultKeepAllMinRating,
		WatchedCleanupDays:   DefaultWatchedCleanupDays,
		UnwatchedCleanupDays: DefaultUnwatchedCleanupDays,
		EpisodesAhead:        DefaultEpisodesAhead,
	}
}

// ApplyDefaults fills in zero values with defaults.
func (a *Automation) ApplyDefaults() {
	defaults := DefaultAutomation()
	if a.DeleteMaxRating == 0 {
		a.DeleteMaxRating = defaults.DeleteMaxRating
	}
	if a.KeepMinRating == 0 {
		a.KeepMinRating = defaults.KeepMinRating
	}
	if a.KeepAllMinRating == 0 {
		a.KeepAllMinRating = defaults.KeepAllMinRating
	}
	if a.WatchedCleanupDays == 0 {
		a.WatchedCleanupDays = defaults.WatchedCleanupDays
	}
	if a.UnwatchedCleanupDays == 0 {
		a.UnwatchedCleanupDays = defaults.UnwatchedCleanupDays
	}
	if a.EpisodesAhead == 0 {
		a.EpisodesAhead = defaults.EpisodesAhead
	}
}

// Plex configuration.
type Plex struct {
	URL     string   `yaml:"url"`
	Token   string   `yaml:"token"`
	Ignored []string `yaml:"ignored"`
}

func (p *Plex) Ignore(toIgnore string) bool {
	if toIgnore == "" {
		return true
	}
	for _, ig := range p.Ignored {
		if ig == toIgnore {
			return true
		}
	}
	return false
}

// Sonarr configuration.
type Sonarr struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// Radarr configuration.
type Radarr struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// Database configuration.
type Database struct {
	Path string `yaml:"path"`
}

// LoadConfig loads the configuration from a file.
func LoadConfig(path string) (*Config, error) {
	// Expand home directory if needed
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, ".config", "mediate", "config.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// CreateDefaultConfig creates a default configuration file.
func CreateDefaultConfig(path string) error {
	// Create default config
	config := Config{
		Plex: Plex{
			URL:   "http://plex.example.com:32400",
			Token: "your-plex-token",
			Ignored: []string{
				"Kids TV Shows",
				"Kids Movies",
			},
		},
		Sonarr: Sonarr{
			URL:    "http://sonarr.example.com:8989",
			APIKey: "your-sonarr-api-key",
		},
		Radarr: Radarr{
			URL:    "http://radarr.example.com:7878",
			APIKey: "your-radarr-api-key",
		},
		Database: Database{
			Path: "mediate.sqlite",
		},
		Automation: DefaultAutomation(),
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write to file
	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
