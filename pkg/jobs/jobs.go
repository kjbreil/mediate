package jobs

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/kjbreil/mediate/pkg/shows"
)

const rewatchResetThreshold = 0.5

// Jobs contains all the job functions.
type Jobs struct {
	mediate *mediate.Mediate
	logger  *slog.Logger
}

// New creates a new Jobs instance.
func New(m *mediate.Mediate, logger *slog.Logger) *Jobs {
	return &Jobs{
		mediate: m,
		logger:  logger,
	}
}

// MonitorJob handles monitoring episodes and setting monitoring status.
func (j *Jobs) MonitorJob() error {
	j.logger.Info("Running monitor job")

	// Monitor all pilot episodes that don't have files and aren't wanted
	err := j.mediate.MonitorEpisodes(
		j.mediate.GetShows().Find(shows.Finders[shows.AllPilots]).HasFile(false).Wanted(false),
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to monitor episodes: %w", err)
	}

	// Mark only pilot episodes as unwatched
	_, errs := j.mediate.MarkOnlyPilotUnwatched()
	if len(errs) > 0 {
		return fmt.Errorf("errors marking pilots as unwatched: %v", errs)
	}

	// Set monitored status
	// SetMonitored doesn't return an error, so we don't need to check it
	j.mediate.SetMonitored()

	return nil
}

// DownloadJob handles downloading episodes.
func (j *Jobs) DownloadJob() error {
	j.logger.Info("Running download job")

	cfg := j.mediate.Config().Automation

	// Get recently watched episodes
	episodes := j.mediate.RecentlyWatched()

	for _, e := range episodes {
		// Find the next N episodes that are aired but not downloaded
		nextEpisodes := j.mediate.DB.NextXEpisodes(cfg.EpisodesAhead, e).
			Downloading(false).
			Aired(true).
			HasFile(false)

		if len(nextEpisodes) > 0 {
			// DownloadEpisodes doesn't return an error
			j.mediate.DownloadEpisodes(nextEpisodes)

			// RefreshShowsEpisodes doesn't return an error
			j.mediate.RefreshShowsEpisodes(nextEpisodes)
		}
	}

	return nil
}

// DeleteJob handles deleting episodes.
func (j *Jobs) DeleteJob() error {
	j.logger.Info("Running delete job")

	cfg := j.mediate.Config().Automation

	// Save original values and restore after
	originalWindowDuration := shows.WindowDuration
	originalMaxShowRating := shows.MaxShowRating
	defer func() {
		shows.WindowDuration = originalWindowDuration //nolint:reassign // Restoring original value
		shows.MaxShowRating = originalMaxShowRating   //nolint:reassign // Restoring original value
	}()

	// Set rating threshold from config
	shows.MaxShowRating = cfg.DeleteMaxRating //nolint:reassign // Setting from config

	// Delete watched episodes that can be deleted (watched cleanup window)
	//nolint:reassign // Setting from config
	shows.WindowDuration = time.Hour * 24 * time.Duration(cfg.WatchedCleanupDays)
	watchedEpisodes := j.mediate.GetShows().Find(shows.Finders[shows.WatchedCanDelete])

	if len(watchedEpisodes) > 0 {
		j.logger.Info("Deleting watched episodes", "count", len(watchedEpisodes))
		err := j.mediate.DeleteEpisodes(watchedEpisodes)
		if err != nil {
			return fmt.Errorf("failed to delete watched episodes: %w", err)
		}
	}

	// Delete unwatched episodes that can be deleted (unwatched cleanup window)
	//nolint:reassign // Setting from config
	shows.WindowDuration = time.Hour * 24 * time.Duration(cfg.UnwatchedCleanupDays)
	unwatchedEpisodes := j.mediate.GetShows().Find(shows.Finders[shows.NotWatchedCanDelete])

	if len(unwatchedEpisodes) > 0 {
		j.logger.Info("Deleting unwatched episodes", "count", len(unwatchedEpisodes))
		err := j.mediate.DeleteEpisodes(unwatchedEpisodes)
		if err != nil {
			return fmt.Errorf("failed to delete unwatched episodes: %w", err)
		}
	}

	return nil
}

// RefreshJob handles refreshing shows and episodes.
func (j *Jobs) RefreshJob() error {
	j.logger.Info("Running refresh job")

	// Get recently watched episodes
	episodes := j.mediate.RecentlyWatched()

	// Refresh those shows and episodes
	// RefreshShowsEpisodes doesn't return an error
	j.mediate.RefreshShowsEpisodes(episodes)

	return nil
}

// PlexWatchJob sets up a Plex watcher.
func (j *Jobs) PlexWatchJob() error {
	j.logger.Info("Setting up Plex watch job")

	cfg := j.mediate.Config().Automation

	j.mediate.OnPlexPlaying(func(pp *mediate.PlexPlaying) {
		ep := pp.Episode()
		if ep == nil {
			return
		}

		if pp.Changed {
			j.logger.Info("Plex reporting show being watched",
				"title", ep.Title,
				"season", ep.Season,
				"episode", ep.Episode,
				"left", pp.TimeLeft().Minutes(),
			)

			// UpdateEpisode doesn't return an error
			j.mediate.UpdateEpisode(ep)

			// Download next N episodes that are aired but not downloaded
			nextEpisodes := j.mediate.DB.NextXEpisodes(cfg.EpisodesAhead, ep).
				HasFile(false).
				Aired(true).
				Downloading(false)

			if len(nextEpisodes) > 0 {
				// DownloadEpisodes doesn't return an error
				j.mediate.DownloadEpisodes(nextEpisodes)
			}
		}

		// Debug logging for rewatch check
		j.logger.Debug("Rewatch check",
			"show", ep.ShowTitle,
			"progress", fmt.Sprintf("%.1f%%", pp.Progress()*100),
			"duration", ep.Duration,
			"viewed", pp.Viewed(),
			"triggered", pp.RewatchResetTriggered())

		// Rewatch reset: mark subsequent episodes as unwatched after 50% progress
		if pp.Progress() >= rewatchResetThreshold && !pp.RewatchResetTriggered() {
			j.logger.Info("Rewatch detected, resetting subsequent episodes",
				"show", ep.ShowTitle, "season", ep.Season, "episode", ep.Episode,
				"progress", fmt.Sprintf("%.0f%%", pp.Progress()*100))

			count, errs := j.mediate.MarkSubsequentEpisodesUnwatched(ep)
			if count > 0 {
				j.logger.Info("Reset subsequent episodes", "count", count)
			}
			if len(errs) > 0 {
				j.logger.Warn("Some episodes failed to reset", "failed", len(errs))
			}

			pp.SetRewatchResetTriggered(true)
		}
	})

	// This job doesn't return as it sets up a callback
	// Return nil to indicate success
	return nil
}
