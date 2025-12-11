package mediate

import (
	"fmt"
	"sync"
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
	"golift.io/starr/sonarr"
)

func (m *Mediate) loadShows() error {
	allSeries, err := m.sonarr.GetAllSeries()
	if err != nil {
		return err
	}

	buf := make(chan struct{}, 10)
	defer close(buf)
	wg := &sync.WaitGroup{}

	start := time.Now()
	for _, ser := range allSeries {
		s, _ := m.DB.AddSonarrSeries(ser)

		wg.Add(1)
		buf <- struct{}{}
		go func(s *shows.Show) {
			defer func() {
				<-buf
				wg.Done()
			}()
			err = m.UpdateEpisodes(s)
			if err != nil {
				m.logger.Error("could not update episodes", "show", s.Title, "err", err.Error())
				return
			}
		}(s)
	}
	wg.Wait()
	m.UpdateDownloading()
	m.logger.Info("loading shows completed", "duration", time.Since(start))

	return nil
}

func (m *Mediate) UpdateDownloading() {
	_, _ = m.sonarr.SendCommand(&sonarr.CommandRequest{
		Name: "RefreshMonitoredDownloads",
	})

	time.Sleep(time.Millisecond * 100)

	queue, _ := m.sonarr.GetQueue(10000, 10)
	filesDownloading := make([]int, 0, len(queue.Records))
	for _, q := range queue.Records {
		filesDownloading = append(filesDownloading, int(q.EpisodeID))
	}

	m.DB.MarkDownloading(m.DB.SonarrIDs(filesDownloading...).TvdbIDs()...)
}

func (m *Mediate) UpdateEpisode(ep *shows.Episode) {
	if err := m.RefreshShow(m.DB.GetShow(ep.TvdbID)); err != nil {
		m.logger.Error("Failed to refresh show", "error", err)
	}
	m.UpdateDownloading()
}

func (m *Mediate) UpdateEpisodes(s *shows.Show) error {
	buf := make(chan struct{}, 10)
	defer close(buf)
	wg := &sync.WaitGroup{}

	episodes, err := m.sonarr.GetSeriesEpisodes(&sonarr.GetEpisode{SeriesID: s.SonarrID})
	if err != nil {
		return err
	}
	for _, ep := range episodes {
		e, _ := m.DB.AddSonarrEpisode(s, ep)

		if ep.HasFile {
			wg.Add(1)
			buf <- struct{}{}
			go func(ep *sonarr.Episode) {
				defer func() {
					<-buf
					wg.Done()
				}()

				files, _ := m.sonarr.GetEpisodeFiles(ep.EpisodeFileID)
				e.DownloadedAt = time.Time{}
				for _, file := range files {
					if file.DateAdded.After(e.DownloadedAt) {
						e.DownloadedAt = file.DateAdded
					}
				}
				m.DB.Save(e)
			}(ep)
		}
	}

	wg.Wait()

	return nil
}

func (m *Mediate) DownloadEpisodes(episodes shows.Episodes) {
	if len(episodes) == 0 {
		return
	}
	m.logger.Info(fmt.Sprintf("downloading %d episodes", len(episodes)))
	_ = m.MonitorEpisodes(episodes, true)

	_, _ = m.sonarr.SendCommand(&sonarr.CommandRequest{
		Name:       "EpisodeSearch",
		EpisodeIDs: episodes.SonarrIDs(),
	})
}
