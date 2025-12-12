package mediate

import (
	"path/filepath"
	"time"

	"github.com/kjbreil/mediate/pkg/movies"
	"github.com/kjbreil/mediate/pkg/shows"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

func (m *Mediate) MarkOnlyPilotUnwatched() ([]*shows.Episode, []error) {
	var rtn []*shows.Episode
	var errors []error
	for _, show := range *m.DB.GetShows() {
		if show.Ignore {
			continue
		}

		if e := show.Episodes.OnlyPilot(); e != nil {
			err := m.plex.UnScrobble(e.PlexRatingKey)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return rtn, errors
}

func (m *Mediate) RecentlyWatched() shows.Episodes {
	var episodes shows.Episodes
	m.DB.Where("last_viewed_at >?", time.Now().Add(-time.Hour*24*7)).Find(&episodes)
	return episodes
}

// SetMonitored sets the unwatched episodes to monitored and if the show is continuing, sets the most recent season to monitored.
func (m *Mediate) SetMonitored() {
	for _, show := range *m.DB.GetShows() {
		if show.Ignore {
			continue
		}

		if err := m.RefreshShow(show); err != nil {
			continue
		}

		if show.Continuing {
			m.handleContinuingShow(show)
		} else {
			m.handleEndedShow(show)
		}

		episodes := show.GetEpisodes().Wanted(true).HasFile(false).Aired(true)
		m.DownloadEpisodes(episodes)
	}
}

func (m *Mediate) handleContinuingShow(show *shows.Show) {
	cfg := m.config.Automation
	// Medium rated shows (between keep threshold and keep-all threshold)
	if show.Rating >= cfg.KeepMinRating && show.Rating < cfg.KeepAllMinRating {
		m.setupMediumRatedContinuingShow(show)
		return
	}

	if show.Rating < cfg.KeepMinRating {
		m.unmonitorLowRatedShow(show)
	}
}

func (m *Mediate) handleEndedShow(show *shows.Show) {
	cfg := m.config.Automation
	// For ended shows below keep-all threshold, unmonitor and keep only pilot
	if show.Rating < cfg.KeepAllMinRating {
		if err := m.UnMonitorAll(show); err != nil {
			return
		}
		if err := m.MonitorPilot(show); err != nil {
			return
		}
		episodes := show.GetEpisodes().Wanted(true).HasFile(false).Aired(true)
		m.DownloadEpisodes(episodes)
		return
	}

	// For highly rated ended shows, monitor all episodes
	if show.Rating >= cfg.KeepAllMinRating {
		_ = m.MonitorAll(show)
	}
}

func (m *Mediate) setupMediumRatedContinuingShow(show *shows.Show) {
	ser, err := m.sonarr.GetSeriesByID(show.SonarrID)
	if err != nil {
		panic(err)
	}

	for _, sea := range ser.Seasons {
		sea.Monitored = false
	}

	updateInput := &sonarr.AddSeriesInput{
		ID:                ser.ID,
		Title:             ser.Title,
		TitleSlug:         ser.TitleSlug,
		TvdbID:            ser.TvdbID,
		ImdbID:            ser.ImdbID,
		TvMazeID:          ser.TvMazeID,
		TvRageID:          ser.TvRageID,
		Path:              ser.Path,
		QualityProfileID:  ser.QualityProfileID,
		LanguageProfileID: ser.LanguageProfileID,
		SeriesType:        ser.SeriesType,
		Monitored:         ser.Monitored,
		SeasonFolder:      ser.SeasonFolder,
		UseSceneNumbering: ser.UseSceneNumbering,
		Tags:              ser.Tags,
		Seasons:           ser.Seasons,
		Images:            ser.Images,
	}

	if _, err = m.sonarr.UpdateSeries(updateInput, false); err != nil {
		return
	}

	m.unmonitorWatchedEpisodes(show)
	m.monitorUnwatchedEpisodes(show)

	if err = m.RefreshShow(show); err != nil {
		return
	}

	episodes := show.GetEpisodes().Wanted(true).HasFile(false).Aired(true)
	m.DownloadEpisodes(episodes)
}

func (m *Mediate) unmonitorLowRatedShow(show *shows.Show) {
	ser, err := m.sonarr.GetSeriesByID(show.SonarrID)
	if err != nil {
		panic(err)
	}

	for _, sea := range ser.Seasons {
		sea.Monitored = false
	}

	episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
		return s.SonarrID == show.SonarrID && e.Watched && !e.IsPilot() && e.Wanted
	})

	_ = m.MonitorEpisodes(episodes, false)
}

func (m *Mediate) unmonitorWatchedEpisodes(show *shows.Show) {
	episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
		return s.SonarrID == show.SonarrID && e.Watched && !e.IsPilot() && e.Wanted
	})
	_ = m.MonitorEpisodes(episodes, false)
}

func (m *Mediate) monitorUnwatchedEpisodes(show *shows.Show) {
	episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
		if s.SonarrID != show.SonarrID {
			return false
		}
		if e.IsPilot() && !e.Wanted {
			return true
		}
		if e.Wanted || e.PlexRatingKey == "" {
			return false
		}
		return e.HasNotAired() || !e.Watched
	})
	_ = m.MonitorEpisodes(episodes, true)
}

func (m *Mediate) DeleteEpisodes(episodes []*shows.Episode) error {
	m.logger.Info("Deleting episodes", "count", len(episodes))
	var toDelete []int64

	for _, ep := range episodes {
		if ep.Wanted {
			toDelete = append(toDelete, ep.SonarrID)
			m.logger.Info(
				"Marking episode as unmonitored",
				"sonarrId",
				ep.SonarrID,
				"season",
				ep.Season,
				"episode",
				ep.Episode,
			)
		}
	}

	if len(toDelete) > 0 {
		m.logger.Info("Unmonitoring episodes in Sonarr", "count", len(toDelete))
		var err error
		_, err = m.sonarr.MonitorEpisode(toDelete, false)
		if err != nil {
			m.logger.Error("Failed to unmonitor episodes", "err", err.Error())
			return err
		}
	}

	for _, ep := range episodes {
		if ep.HasFile {
			m.logger.Info(
				"Deleting episode file",
				"sonarrFileId",
				ep.SonarrFileID,
				"season",
				ep.Season,
				"episode",
				ep.Episode,
			)
			var err = m.sonarr.DeleteEpisodeFile(ep.SonarrFileID)
			if err != nil {
				m.logger.Error("Failed to delete episode file", "sonarrFileId", ep.SonarrFileID, "err", err.Error())
				return err
			}
		}
	}
	m.logger.Info("Successfully completed episode deletion")
	return nil
}

func (m *Mediate) MonitorEpisodes(episodes shows.Episodes, toMonitor bool) error {
	_, err := m.sonarr.MonitorEpisode(episodes.SonarrIDs(), toMonitor)
	return err
}

func (m *Mediate) UnMonitorAll(show *shows.Show) error {
	episodes := show.GetEpisodes()
	err := m.MonitorEpisodes(episodes, false)
	if err != nil {
		return err
	}
	var ser *sonarr.Series
	ser, err = m.sonarr.GetSeriesByID(show.SonarrID)
	ser.Monitored = true
	if err != nil {
		panic(err)
	}

	for _, sea := range ser.Seasons {
		sea.Monitored = false
	}

	updateInput := &sonarr.AddSeriesInput{
		ID:                ser.ID,
		Title:             ser.Title,
		TitleSlug:         ser.TitleSlug,
		TvdbID:            ser.TvdbID,
		ImdbID:            ser.ImdbID,
		TvMazeID:          ser.TvMazeID,
		TvRageID:          ser.TvRageID,
		Path:              ser.Path,
		QualityProfileID:  ser.QualityProfileID,
		LanguageProfileID: ser.LanguageProfileID,
		SeriesType:        ser.SeriesType,
		Monitored:         ser.Monitored,
		SeasonFolder:      ser.SeasonFolder,
		UseSceneNumbering: ser.UseSceneNumbering,
		Tags:              ser.Tags,
		Seasons:           ser.Seasons,
		Images:            ser.Images,
	}
	_, err = m.sonarr.UpdateSeries(updateInput, false)
	if err != nil {
		return err
	}

	return m.RefreshShow(show)
}

func (m *Mediate) MonitorAll(show *shows.Show) error {
	episodes := show.GetEpisodes()
	err := m.MonitorEpisodes(episodes, true)
	if err != nil {
		return err
	}

	var ser *sonarr.Series
	ser, err = m.sonarr.GetSeriesByID(show.SonarrID)
	if err != nil {
		return err
	}

	for _, sea := range ser.Seasons {
		sea.Monitored = sea.SeasonNumber != 0
	}

	updateInput := &sonarr.AddSeriesInput{
		ID:                ser.ID,
		Title:             ser.Title,
		TitleSlug:         ser.TitleSlug,
		TvdbID:            ser.TvdbID,
		ImdbID:            ser.ImdbID,
		TvMazeID:          ser.TvMazeID,
		TvRageID:          ser.TvRageID,
		Path:              ser.Path,
		QualityProfileID:  ser.QualityProfileID,
		LanguageProfileID: ser.LanguageProfileID,
		SeriesType:        ser.SeriesType,
		Monitored:         ser.Monitored,
		SeasonFolder:      ser.SeasonFolder,
		UseSceneNumbering: ser.UseSceneNumbering,
		Tags:              ser.Tags,
		Seasons:           ser.Seasons,
		Images:            ser.Images,
	}
	_, err = m.sonarr.UpdateSeries(updateInput, false)
	if err != nil {
		return err
	}

	return m.RefreshShow(show)
}

func (m *Mediate) MonitorPilot(show *shows.Show) error {
	episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
		if s.SonarrID != show.SonarrID {
			return false
		}
		if e.IsPilot() && !e.Wanted {
			return true
		}

		return false
	})
	err := m.MonitorEpisodes(episodes, true)
	if err != nil {
		return err
	}
	return m.RefreshShow(show)
}

func (m *Mediate) DeleteMovies(moviesList []*movies.Movie) error {
	for _, mov := range moviesList {
		var err error
		err = m.radarr.DeleteMovie(mov.RadarrID, true, false)
		if err != nil {
			return err
		}
		_, err = m.radarr.AddMovie(&radarr.AddMovieInput{
			RootFolderPath:   filepath.Dir(mov.Path),
			TmdbID:           int64(mov.TmdbID),
			QualityProfileID: mov.QualityProfileID,
			Monitored:        false,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
