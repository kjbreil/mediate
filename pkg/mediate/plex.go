package mediate

import (
	"fmt"
	"sync"
	"time"

	"github.com/kjbreil/go-plex/library"
	"github.com/kjbreil/go-plex/plex"
	"github.com/kjbreil/mediate/pkg/shows"
)

type plexActivity struct {
	playing map[int]*PlexPlaying
	m       sync.Mutex
}

type PlexPlaying struct {
	state   string
	changed time.Time
	episode *shows.Episode
	viewed  time.Duration
	m       sync.Mutex
	Changed bool
}

func (pp *PlexPlaying) Episode() *shows.Episode {
	return pp.episode
}

func (pp *PlexPlaying) TimeLeft() time.Duration {
	if pp.episode != nil {
		return pp.episode.Duration - pp.viewed
	}
	return 0
}

func (m *Mediate) RefreshShowsEpisodes(episodes shows.Episodes) {
	shows := make(map[int]*shows.Show)

	for _, e := range episodes {
		if _, ok := shows[e.ShowTvdbID]; !ok {
			shows[e.ShowTvdbID] = m.DB.GetShow(e.TvdbID)
		}
	}
	for _, s := range shows {
		if err := m.RefreshShow(s); err != nil {
			m.logger.Error("Failed to refresh show", "title", s.Title, "error", err)
		}
	}
}
func (m *Mediate) RefreshShow(s *shows.Show) error {
	m.logger.Info("Refreshing show", "title", s.Title)
	err := m.UpdateEpisodes(s)
	if err != nil {
		return err
	}
	m.UpdateDownloading()
	plexShow := m.plex.Libraries.FindShow(s.PlexRatingKey)
	if plexShow == nil {
		return fmt.Errorf("could not find show in plex rating key %s", s.PlexRatingKey)
	}
	err = m.plex.GetShowEpisodes(plexShow)
	if err != nil {
		return err
	}
	err = m.loadPlexShows(s.Library)
	if err != nil {
		return err
	}
	episodes := m.DB.GetEpisodes(s.TvdbID).HasFile(true).InPlex(false)
	if len(episodes) > 0 {
		var scanErr = m.plex.ScanLibrary(s.Library)
		if scanErr != nil {
			m.logger.Error("Failed to scan library", "library", s.Library, "error", scanErr)
		}
	}
	m.logger.Info("Finished refreshing show", "title", s.Title)
	return nil
}

func (m *Mediate) OnPlexPlaying(f func(pp *PlexPlaying)) {
	playing := &plexActivity{
		playing: make(map[int]*PlexPlaying),
		m:       sync.Mutex{},
	}

	m.plex.Websocket.OnPlaying(func(n plex.NotificationContainer) {
		playing.m.Lock()
		defer playing.m.Unlock()

		for _, ps := range n.PlaySessionStateNotification {
			ep := m.DB.GetEpisodeFromRatingKey(ps.RatingKey)
			// if ep == nil {
			// 	m.RefreshShow(m.DB.GetShow(ep.TvdbID))
			// 	ep = m.DB.GetEpisodeFromRatingKey(ps.RatingKey)
			if ep == nil {
				return
			}
			// }
			pp, ok := playing.playing[ep.TvdbID]
			if !ok {
				pp = &PlexPlaying{
					episode: ep,
					m:       sync.Mutex{},
				}
				playing.playing[ep.TvdbID] = pp
			}
			pp.m.Lock()
			pp.viewed = time.Millisecond * time.Duration(ps.ViewOffset)
			pp.Changed = false
			if pp.state != ps.State {
				pp.state = ps.State
				pp.Changed = true
				pp.changed = time.Now()
			}
			f(pp)
			pp.m.Unlock()
		}
	})
	m.plex.SubscribeToNotifications()

	// m.plex.Webhook = plex.NewWebhook(8080, net.ParseIP("10.0.2.2"))
	//
	// m.plex.Webhook.OnPlay(func(w plex.WebhookEvent) {
	// 	playing.m.Lock()
	// 	defer playing.m.Unlock()
	//
	// 	ep := m.DB.GetEpisodeFromRatingKey(w.Metadata.RatingKey)
	// 	if ep != nil {
	// 		pp, ok := playing.playing[ep.TvdbID]
	// 		if !ok {
	// 			pp = &PlexPlaying{
	// 				episode: ep,
	// 				m:       sync.Mutex{},
	// 			}
	// 			playing.playing[ep.TvdbID] = pp
	// 		}
	// 		pp.m.Lock()
	// 		f(pp)
	// 		pp.m.Unlock()
	// 	}
	//
	// })

	// m.plex.ServeWebhook()
}

func (m *Mediate) loadPlexShows(lib *library.Library) error {
	for _, show := range *m.DB.GetShows() {
		m.loadPlexShow(lib, show)
	}

	return nil
}

func (m *Mediate) loadPlexShow(lib *library.Library, show *shows.Show) {
	plexShow, _, _ := lib.Shows.FindTvdbID(show.TvdbID)
	if plexShow == nil {
		return
	}

	show.PlexRatingKey = plexShow.RatingKey
	show.Rating = plexShow.UserRating
	show.Library = lib
	show.Ignore = m.config.Plex.Ignore(show.LibraryTitle())
	show.LibraryUUID = lib.UUID

	var ss = *show
	ss.Episodes = nil
	m.DB.Save(ss)

	m.loadPlexEpisodes(lib, show)
}

func (m *Mediate) loadPlexEpisodes(lib *library.Library, show *shows.Show) {
	for _, episode := range show.Episodes {
		_, _, plexEpisode := lib.Shows.FindTvdbID(episode.TvdbID)
		if plexEpisode == nil {
			continue
		}

		episode.PlexRatingKey = plexEpisode.RatingKey
		episode.Watched = plexEpisode.Watched
		episode.LastViewedAt = plexEpisode.LastViewedAt
		episode.UpdatedAt = plexEpisode.UpdatedAt
		episode.Duration = time.Millisecond * time.Duration(plexEpisode.Duration)
		m.DB.Save(episode)
	}
}
