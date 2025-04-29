package main

import (
	"github.com/kjbreil/mediate/pkg/cli"
	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/jobs"
	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/kjbreil/mediate/pkg/service"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

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

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(logHandler)

	// Load configuration
	var c config.Config
	if cliConfig.ConfigFile != "" {
		// TODO: Implement config file loading
		logger.Info("Loading configuration from file", "file", cliConfig.ConfigFile)
	} else {
		// Use default configuration
		c = config.Config{
			Plex: config.Plex{
				URL:   "http://plex1.kaygel.io:32400",
				Token: "-HacSX44mXL1WHVACUZ5",
				Ignored: []string{
					"Kids TV Shows",
					"Kids Movies",
				},
			},
			Sonarr: config.Sonarr{
				ApiKey: "67bd04cc551149188947a0024a7f5c1e",
				URL:    "http://10.0.1.22:8989/show/",
			},
			Radarr: config.Radarr{
				ApiKey: "e2eab479a088404387c7b1b48eab5287",
				URL:    "http://10.0.1.22:7878/film/",
			},
		}
	}

	// Initialize mediate
	m, err := mediate.New(
		c,
		mediate.WithLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	// Create jobs
	j := jobs.New(m, logger)

	// Create service
	svc := service.NewService(logger)

	// Register jobs based on command-line arguments
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
			
		case "plex-watch":
			svc.AddJob("plex-watch", interval, j.PlexWatchJob)
			logger.Info("Registered plex-watch job", "interval", interval)
			
		default:
			logger.Warn("Unknown job", "name", jobName)
		}
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
