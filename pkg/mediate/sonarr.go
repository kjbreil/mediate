package mediate

import (
	"fmt"
	"github.com/kjbreil/mediate/pkg/shows"
	"golift.io/starr/sonarr"
	"sync"
	"time"
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

		s := m.Shows.NewShow(int(ser.TvdbID))

		s.Title = ser.Title
		s.SonarrId = ser.ID
		s.Continuing = ser.Status == "continuing"

		wg.Add(1)
		buf <- struct{}{}
		go func(s *shows.Show) {
			defer func() {
				<-buf
				wg.Done()
			}()
			err = m.UpdateEpisodes(s)
			if err != nil {
				m.logger.Error(fmt.Sprintf("could not update episodes for %s", s.Title), "err", err.Error())
				return
			}
		}(s)

	}
	wg.Wait()
	m.UpdateDownloading()
	m.logger.Info(fmt.Sprintln("loading shows took", time.Since(start)))

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

	m.Shows.MarkDownloading(m.Shows.SonarrIdToTvdbId(filesDownloading...)...)
}

func (m *Mediate) UpdateEpisode(ep *shows.Episode) {
	m.UpdateEpisodes(m.Shows.GetShow(ep.TvdbID))
	m.UpdateDownloading()
}

func (m *Mediate) UpdateEpisodes(s *shows.Show) error {

	buf := make(chan struct{}, 10)
	defer close(buf)
	wg := &sync.WaitGroup{}

	episodes, err := m.sonarr.GetSeriesEpisodes(s.SonarrId)
	if err != nil {
		return err
	}
	for _, ep := range episodes {

		e := s.AddSonarrEpisode(ep)

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
			}(ep)
		}
	}

	wg.Wait()

	return nil
}

func (m *Mediate) DownloadEpisodes(episodes shows.EpisodesArr) {
	if len(episodes) == 0 {
		return
	}
	m.logger.Info(fmt.Sprintf("downloading %d episodes", len(episodes)))
	_ = m.MonitorEpisodes(episodes, true)

	_, _ = m.sonarr.SendCommand(&sonarr.CommandRequest{
		Name:       "EpisodeSearch",
		EpisodeIDs: episodes.SonarrIds(),
	})
}
