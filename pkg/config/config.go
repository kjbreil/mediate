package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Plex     Plex     `yaml:"plex"`
	Sonarr   Sonarr   `yaml:"sonarr"`
	Radarr   Radarr   `yaml:"radarr"`
	Database Database `yaml:"database"`
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
