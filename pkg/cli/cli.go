package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config represents the command-line configuration
type Config struct {
	// Job-related flags
	Jobs            []string
	Intervals       map[string]time.Duration
	DefaultInterval time.Duration
	WatchPlex       bool

	// General flags
	ConfigFile string
	LogLevel   string
	Help       bool
}

// ParseFlags parses command-line flags and returns a Config
func ParseFlags() *Config {
	cfg := &Config{
		Intervals: make(map[string]time.Duration),
	}

	// Define flags
	flag.StringVar(&cfg.ConfigFile, "config", "", "Path to configuration file")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	
	// Job selection and intervals
	jobsFlag := flag.String("jobs", "", "Comma-separated list of jobs to run (all, monitor, download, delete, refresh)")
	defaultIntervalFlag := flag.String("interval", "30m", "Default interval for all jobs")
	
	// Individual job intervals
	monitorIntervalFlag := flag.String("monitor-interval", "", "Interval for monitor job")
	downloadIntervalFlag := flag.String("download-interval", "", "Interval for download job")
	deleteIntervalFlag := flag.String("delete-interval", "", "Interval for delete job")
	refreshIntervalFlag := flag.String("refresh-interval", "", "Interval for refresh job")
	
	// Watcher flags
	flag.BoolVar(&cfg.WatchPlex, "watch-plex", false, "Enable Plex watching to trigger actions when media is played")
	
	// Help flag
	flag.BoolVar(&cfg.Help, "help", false, "Show help")
	flag.BoolVar(&cfg.Help, "h", false, "Show help (shorthand)")

	// Parse flags
	flag.Parse()

	// Show help if requested
	if cfg.Help {
		printHelp()
		os.Exit(0)
	}

	// Parse default interval
	var err error
	cfg.DefaultInterval, err = parseDuration(*defaultIntervalFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid default interval: %s\n", err)
		os.Exit(1)
	}

	// Parse job-specific intervals
	parseJobInterval(cfg, "monitor", *monitorIntervalFlag)
	parseJobInterval(cfg, "download", *downloadIntervalFlag)
	parseJobInterval(cfg, "delete", *deleteIntervalFlag)
	parseJobInterval(cfg, "refresh", *refreshIntervalFlag)

	// Parse jobs
	if *jobsFlag == "" {
		// Default to all jobs if none specified
		cfg.Jobs = []string{"monitor", "download", "delete", "refresh"}
	} else if *jobsFlag == "all" {
		cfg.Jobs = []string{"monitor", "download", "delete", "refresh"}
	} else {
		cfg.Jobs = strings.Split(*jobsFlag, ",")
		// Trim whitespace
		for i, job := range cfg.Jobs {
			cfg.Jobs[i] = strings.TrimSpace(job)
		}
	}

	return cfg
}

// parseJobInterval parses a job-specific interval
func parseJobInterval(cfg *Config, jobName, intervalStr string) {
	if intervalStr == "" {
		return
	}

	interval, err := parseDuration(intervalStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid interval for %s job: %s\n", jobName, err)
		os.Exit(1)
	}

	cfg.Intervals[jobName] = interval
}

// parseDuration parses a duration string with support for days
func parseDuration(s string) (time.Duration, error) {
	// Check for days format (e.g., "3d")
	if strings.HasSuffix(s, "d") {
		daysPart := strings.TrimSuffix(s, "d")
		days, err := time.ParseDuration(daysPart + "h")
		if err != nil {
			return 0, err
		}
		return days * 24, nil
	}

	// Standard duration parsing
	return time.ParseDuration(s)
}

// GetJobInterval returns the interval for a specific job
func (c *Config) GetJobInterval(jobName string) time.Duration {
	if interval, ok := c.Intervals[jobName]; ok {
		return interval
	}
	return c.DefaultInterval
}

// printHelp prints help information
func printHelp() {
	fmt.Println("Mediate - A tool to manage Plex, Sonarr, and Radarr")
	fmt.Println("\nUsage:")
	fmt.Println("  mediate [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nScheduled Jobs:")
	fmt.Println("  monitor      - Monitor episodes and set monitoring status")
	fmt.Println("  download     - Download episodes")
	fmt.Println("  delete       - Delete episodes")
	fmt.Println("  refresh      - Refresh shows and episodes")
	fmt.Println("\nWatchers:")
	fmt.Println("  --watch-plex - Watch for Plex playback events and trigger actions")
	fmt.Println("\nExamples:")
	fmt.Println("  mediate --jobs=monitor,download --interval=1h")
	fmt.Println("  mediate --jobs=all --delete-interval=1d --monitor-interval=30m")
	fmt.Println("  mediate --jobs=delete --watch-plex")
	fmt.Println("  mediate --watch-plex")
}
