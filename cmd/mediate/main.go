package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kjbreil/mediate/pkg/analysis"
	"github.com/kjbreil/mediate/pkg/cli"
	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/jobs"
	"github.com/kjbreil/mediate/pkg/mcp"
	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/kjbreil/mediate/pkg/service"
)

//nolint:funlen // main function handles CLI parsing, configuration loading, and service initialization
func main() {
	// Parse command-line flags
	cliConfig := cli.ParseFlags()

	// Configure logging
	var logLevel slog.Level
	switch cliConfig.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// In MCP mode, log to stderr to avoid interfering with JSON-RPC on stdout
	logOutput := os.Stdout
	if cliConfig.Mode == "mcp" {
		logOutput = os.Stderr
	}

	logHandler := slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(logHandler)

	// Load configuration
	var c *config.Config
	var err error

	//nolint:nestif // configuration loading requires conditional logic for multiple scenarios
	if cliConfig.ConfigFile != "" {
		// Load configuration from specified file
		logger.Info("Loading configuration from file", "file", cliConfig.ConfigFile)
		c, err = config.LoadConfig(cliConfig.ConfigFile)
		if err != nil {
			logger.Error("Failed to load config file", "error", err)

			// If --create-config flag is set, create default config
			if cliConfig.CreateConfig {
				logger.Info("Creating default configuration file", "file", cliConfig.ConfigFile)
				err = config.CreateDefaultConfig(cliConfig.ConfigFile)
				if err != nil {
					logger.Error("Failed to create config file", "error", err)
					log.Fatalf("Failed to create config file: %v", err)
				}
				c, err = config.LoadConfig(cliConfig.ConfigFile)
				if err != nil {
					log.Fatalf("Failed to load newly created config file: %v", err)
				}
				logger.Info("Created and loaded default configuration")
			} else {
				logger.Info("Use --create-config flag to create a default configuration file")
				log.Fatalf("Failed to load config file: %v", err)
			}
		}
	} else {
		// Try to load from default location
		var home string
		var statErr error
		home, err = os.UserHomeDir()
		if err == nil {
			defaultPath := filepath.Join(home, ".config", "mediate", "config.yaml")
			_, statErr = os.Stat(defaultPath)
			if statErr == nil {
				logger.Info("Loading configuration from default location", "file", defaultPath)
				c, err = config.LoadConfig(defaultPath)
				if err != nil {
					logger.Error("Failed to load config from default location", "error", err)
				}
			}
		}

		// If no config was loaded, use hardcoded defaults
		if c == nil {
			logger.Warn("Using hardcoded configuration. This is not recommended for production use.")
			logger.Info("Create a config file with --config=/path/to/config.yaml --create-config")

			c = &config.Config{
				Plex: config.Plex{
					URL:   "http://plex.example.com:32400",
					Token: "your-plex-token",
					Ignored: []string{
						"Kids TV Shows",
						"Kids Movies",
					},
				},
				Sonarr: config.Sonarr{
					APIKey: "your-sonarr-api-key",
					URL:    "http://sonarr.example.com:8989",
				},
				Radarr: config.Radarr{
					APIKey: "your-radarr-api-key",
					URL:    "http://radarr.example.com:7878",
				},
				Database: config.Database{
					Path: "mediate.sqlite",
				},
				Automation: config.DefaultAutomation(),
			}
		}
	}

	// Check for analysis mode first
	if cliConfig.Analyze || cliConfig.ScanDeleted {
		// Initialize mediate for analysis
		var m *mediate.Mediate
		m, err = mediate.New(
			*c, // Dereference pointer to get the actual Config value
			mediate.WithLogger(logger),
		)
		if err != nil {
			log.Fatal(err)
		}

		// Load data for analysis
		err = m.LoadDataSync()
		if err != nil {
			logger.Error("Failed to load data", "error", err)
			m.Close()
			os.Exit(1)
		}

		// Run analysis
		runAnalysisMode(m, logger, cliConfig)
		m.Close()
		return
	}

	// Check operating mode
	if cliConfig.Mode == "mcp" {
		// Initialize mediate with fast loading for MCP mode
		var m *mediate.Mediate
		m, err = mediate.NewForMCP(
			*c, // Dereference pointer to get the actual Config value
			mediate.WithLogger(logger),
		)
		if err != nil {
			logger.Error("Failed to initialize mediate", "error", err)
			os.Exit(1)
		}

		// Run MCP server mode
		runMCPServer(m, logger, cliConfig)
		m.Close()
		return
	}

	// Initialize mediate for traditional job mode
	m, err := mediate.New(
		*c, // Dereference pointer to get the actual Config value
		mediate.WithLogger(logger),
	)
	if err != nil {
		logger.Error("Failed to initialize mediate", "error", err)
		os.Exit(1)
	}

	// Load data synchronously for job mode
	err = m.LoadDataSync()
	if err != nil {
		logger.Error("Failed to load data", "error", err)
		m.Close()
		os.Exit(1)
	}

	// Traditional job mode
	runJobMode(m, logger, cliConfig)
	m.Close()
}

// runMCPServer runs the application in MCP server mode.
func runMCPServer(m *mediate.Mediate, logger *slog.Logger, cliConfig *cli.Config) {
	logger.Info("Starting Mediate in MCP server mode",
		"transport", cliConfig.Transport,
		"port", cliConfig.Port)

	// Create MCP server
	mcpServer := mcp.NewMediateServer(m, logger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in a goroutine
	go func() {
		if err := mcpServer.Start(ctx); err != nil {
			logger.Error("MCP server error", "error", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	logger.Info("MCP server started, press Ctrl+C to stop")
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctrlC:
		logger.Info("Received interrupt signal")
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Shutdown
	logger.Info("Shutting down MCP server")
	cancel()
	if err := mcpServer.Close(); err != nil {
		logger.Error("Error closing MCP server", "error", err)
	}
	logger.Info("MCP server stopped")
}

// runJobMode runs the application in traditional job mode.
func runJobMode(m *mediate.Mediate, logger *slog.Logger, cliConfig *cli.Config) {
	logger.Info("Starting Mediate in job mode")

	// Create jobs
	j := jobs.New(m, logger)

	// Create service
	svc := service.NewService(logger)

	// Register scheduled jobs based on command-line arguments
	for _, jobName := range cliConfig.Jobs {
		interval := cliConfig.GetJobInterval(jobName)

		switch jobName {
		case "monitor":
			svc.AddJob("monitor", interval, j.MonitorJob)
			logger.Info("Registered monitor job", "interval", interval)

		case "download":
			svc.AddJob("download", interval, j.DownloadJob)
			logger.Info("Registered download job", "interval", interval)

		case "delete":
			svc.AddJob("delete", interval, j.DeleteJob)
			logger.Info("Registered delete job", "interval", interval)

		case "refresh":
			svc.AddJob("refresh", interval, j.RefreshJob)
			logger.Info("Registered refresh job", "interval", interval)

		default:
			logger.Warn("Unknown job", "name", jobName)
		}
	}

	// Register watcher jobs based on command-line flags
	if cliConfig.WatchPlex {
		svc.AddWatcherJob("plex-watch", j.PlexWatchJob)
		logger.Info("Registered Plex watcher job")
	}

	// Start the service
	logger.Info("Starting service")
	svc.Start()

	// Wait for interrupt signal
	logger.Info("Service started, press Ctrl+C to stop")
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ctrlC

	// Stop the service
	logger.Info("Stopping service")
	svc.Stop()
	logger.Info("Service stopped")
}

// runAnalysisMode runs the application in analysis mode.
func runAnalysisMode(m *mediate.Mediate, logger *slog.Logger, cliConfig *cli.Config) {
	logger.Info("Starting Mediate in analysis mode")

	err := analysis.RunAnalysis(cliConfig, m, logger)
	if err != nil {
		logger.Error("Analysis failed", "error", err)
		os.Exit(1)
	}
}
