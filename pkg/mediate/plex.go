package mediate

import (
	"fmt"
	"github.com/kjbreil/go-plex/library"
	"github.com/kjbreil/go-plex/plex"
	"github.com/kjbreil/mediate/pkg/movies"
	"github.com/kjbreil/mediate/pkg/shows"
	"sync"
	"time"
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
	movie   *movies.Movie
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
	episodes := m.Shows.GetShowEpisodes(s.TvdbID).HasFile(true).InPlex(false)
	if len(episodes) > 0 {
		m.plex.ScanLibrary(s.Library)
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
			ep := m.Shows.GetEpisode(ps.RatingKey)
			if ep != nil {
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

		}

	})

	m.plex.SubscribeToNotifications()

}

func (m *Mediate) loadPlexShows(lib *library.Library) error {

	for _, show := range m.Shows.Slice() {

		if show.Title == "Workaholics" {
			fmt.Println("here")
		}
		plexShow, _, _ := lib.Shows.FindTvdbID(show.TvdbID)
		if plexShow == nil {
			continue
		}
		show.PlexRatingKey = plexShow.RatingKey
		show.Rating = plexShow.UserRating
		show.Library = lib
		show.Ignore = m.config.Plex.Ignore(show.LibraryTitle())

		for _, season := range show.Episodes {
			for _, episode := range season {
				_, _, plexEpisode := lib.Shows.FindTvdbID(episode.TvdbID)
				if plexEpisode != nil {
					episode.PlexRatingKey = plexEpisode.RatingKey
					episode.Watched = plexEpisode.Watched
					episode.LastViewedAt = plexEpisode.LastViewedAt
					episode.UpdatedAt = plexEpisode.UpdatedAt
					episode.Duration = time.Millisecond * time.Duration(plexEpisode.Duration)
				}
			}
		}
	}

	return nil
}

func (m *Mediate) loadPlexMovies(lib *library.Library) error {

	for _, movie := range m.Movies {
		plexMovie := lib.Movies.FindTMDB(movie.TmdbID)
		if plexMovie == nil {
			continue
		}
		movie.PlexRatingKey = plexMovie.RatingKey
		movie.Rating = plexMovie.UserRating
		movie.Library = lib.Title
		movie.Ignore = m.config.Plex.Ignore(movie.Library)
		movie.Watched = plexMovie.Watched
		movie.LastViewedAt = plexMovie.LastViewedAt
		movie.UpdatedAt = plexMovie.UpdatedAt
	}

	return nil
}
