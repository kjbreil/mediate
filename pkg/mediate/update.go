package mediate

import (
	"fmt"
	"github.com/kjbreil/mediate/pkg/movies"
	"github.com/kjbreil/mediate/pkg/shows"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
	"path/filepath"
	"time"
)

func (m *Mediate) MarkOnlyPilotUnwatched() (rtn []*shows.Episode, errors []error) {
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

	return
}

func (m *Mediate) RecentlyWatched() shows.Episodes {
	var episodes shows.Episodes
	m.DB.Where("last_viewed_at >?", time.Now().Add(-time.Hour*24*7)).Find(&episodes)
	return episodes
}

// SetMonitored sets the unwatched episodes to monitored and if the show is continuing, sets the most recent season to monitored
func (m *Mediate) SetMonitored() {

	for _, show := range *m.DB.GetShows() {

		if show.Ignore {
			continue
		}
		if show.Title != "Workaholics" {
			continue
		}

		err := m.RefreshShow(show)
		if err != nil {
			continue
		}

		// if the show is continuing and rating is above 5 then set future episodes to be downloaded
		if show.Continuing {
			if show.Rating >= 5 && show.Rating <= 8 {
				var ser *sonarr.Series
				ser, err = m.sonarr.GetSeriesByID(show.SonarrId)
				if err != nil {
					panic(err)
				}

				for _, sea := range ser.Seasons {
					sea.Monitored = false
				}

				_, err = m.sonarr.UpdateSeries(ser, false)
				if err != nil {
					return
				}
				episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
					if s.SonarrId != show.SonarrId || !e.Watched || e.IsPilot() || !e.Wanted {
						return false
					}
					return true
				})
				err = m.MonitorEpisodes(episodes, false)
				if err != nil {
					return
				}

				episodes = m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
					if s.SonarrId != show.SonarrId {
						return false
					}
					if e.IsPilot() && !e.Wanted {
						return true
					}
					if e.Wanted {
						return false
					}
					if e.HasNotAired() {
						return true
					}
					if e.PlexRatingKey == "" {
						return false
					}
					if !e.Watched {
						return true
					}
					return false
				})
				err = m.MonitorEpisodes(episodes, true)
				if err != nil {
					continue
				}

				err = m.RefreshShow(show)
				if err != nil {
					continue
				}

				episodes = show.GetEpisodes().Wanted(true).HasFile(false).Aired(true)
				m.DownloadEpisodes(episodes)

			}
			if show.Rating < 5 {
				var ser *sonarr.Series

				ser, err = m.sonarr.GetSeriesByID(show.SonarrId)
				if err != nil {
					panic(err)
				}

				for _, sea := range ser.Seasons {
					sea.Monitored = false
				}
				episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
					if s.SonarrId != show.SonarrId || !e.Watched || e.IsPilot() || !e.Wanted {
						return false
					}
					return true
				})
				err = m.MonitorEpisodes(episodes, false)
				if err != nil {
					continue
				}
			}
		} else {
			if show.Rating < 9 {
				err = m.UnMonitorAll(show)
				if err != nil {
					continue
				}
				err = m.MonitorPilot(show)
				if err != nil {
					continue
				}
				episodes := show.GetEpisodes().Wanted(true).HasFile(false).Aired(true)
				m.DownloadEpisodes(episodes)
			}

			if show.Rating >= 9 {
				err = m.MonitorAll(show)
				if err != nil {
					continue
				}
			}

		}

		episodes := show.GetEpisodes().Wanted(true).HasFile(false).Aired(true)
		m.DownloadEpisodes(episodes)
	}

}

func (m *Mediate) DeleteEpisodes(episodes []*shows.Episode) error {
	m.logger.Info(fmt.Sprintf("Deleting %d episodes", len(episodes)))
	var toDelete []int64

	for _, ep := range episodes {
		if ep.Wanted {
			toDelete = append(toDelete, ep.SonarrId)
			m.logger.Info("Marking episode as unmonitored", "sonarrId", ep.SonarrId, "season", ep.Season, "episode", ep.Episode)
		}
	}

	if len(toDelete) > 0 {
		m.logger.Info(fmt.Sprintf("Unmonitoring %d episodes in Sonarr", len(toDelete)))
		_, err := m.sonarr.MonitorEpisode(toDelete, false)
		if err != nil {
			m.logger.Error("Failed to unmonitor episodes", "err", err.Error())
			return err
		}
	}

	for _, ep := range episodes {
		if ep.HasFile {
			m.logger.Info("Deleting episode file", "sonarrFileId", ep.SonarrFileId, "season", ep.Season, "episode", ep.Episode)
			err := m.sonarr.DeleteEpisodeFile(ep.SonarrFileId)
			if err != nil {
				m.logger.Error("Failed to delete episode file", "sonarrFileId", ep.SonarrFileId, "err", err.Error())
				return err
			}
		}
	}
	m.logger.Info("Successfully completed episode deletion")
	return nil
}

func (m *Mediate) MonitorEpisodes(episodes shows.Episodes, toMonitor bool) error {
	_, err := m.sonarr.MonitorEpisode(episodes.SonarrIds(), toMonitor)
	return err
}

func (m *Mediate) UnMonitorAll(show *shows.Show) error {
	episodes := show.GetEpisodes()
	err := m.MonitorEpisodes(episodes, false)
	if err != nil {
		return err
	}
	ser, err := m.sonarr.GetSeriesByID(show.SonarrId)
	ser.Monitored = true
	if err != nil {
		panic(err)
	}

	for _, sea := range ser.Seasons {
		sea.Monitored = false
	}

	_, err = m.sonarr.UpdateSeries(ser, false)
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

	ser, err := m.sonarr.GetSeriesByID(show.SonarrId)
	if err != nil {
		return err
	}

	for _, sea := range ser.Seasons {
		sea.Monitored = sea.SeasonNumber != 0
	}

	_, err = m.sonarr.UpdateSeries(ser, false)
	if err != nil {
		return err
	}

	return m.RefreshShow(show)
}

func (m *Mediate) MonitorPilot(show *shows.Show) error {
	episodes := m.GetShows().Find(func(s *shows.Show, e *shows.Episode) bool {
		if s.SonarrId != show.SonarrId {
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
		err := m.radarr.DeleteMovie(mov.RadarrID, true, false)
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

func (m *Mediate) refreshPlexLibraries() {
	// m.plex
}
