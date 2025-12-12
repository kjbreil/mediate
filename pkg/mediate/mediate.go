package mediate

import (
	"context"
	"log/slog"
	"time"

	"github.com/kjbreil/go-plex/pkg/library"
	"github.com/kjbreil/go-plex/pkg/plex"
	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/movies"
	"github.com/kjbreil/mediate/pkg/shows"
	"github.com/kjbreil/mediate/pkg/store"
	"golift.io/starr"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
)

type Mediate struct {
	plex   *plex.Plex
	sonarr *sonarr.Sonarr
	radarr *radarr.Radarr
	logger *slog.Logger

	Movies movies.Movies
	config config.Config

	ctx    context.Context
	cancel context.CancelFunc
	DB     *store.Store
}

func New(c config.Config, _ ...Options) (*Mediate, error) {
	var err error
	m := Mediate{
		plex:   nil,
		sonarr: nil,
		radarr: nil,
		logger: slog.Default(),
		config: c,
		Movies: make(movies.Movies),
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.plex, err = plex.New(
		c.Plex.URL,
		c.Plex.Token,
		plex.WithCacheLibrary("plex-library-cache.json"),
		plex.WithLogger(m.logger),
	)
	if err != nil {
		return nil, err
	}

	m.DB, err = store.InitDBWithPath(c.Database.Path)
	if err != nil {
		return nil, err
	}

	sonarrConfig := starr.New(c.Sonarr.APIKey, c.Sonarr.URL, 0)
	m.sonarr = sonarr.New(sonarrConfig)

	radarrConfig := starr.New(c.Radarr.APIKey, c.Radarr.URL, 0)
	m.radarr = radarr.New(radarrConfig)

	err = m.plex.InitLibraries()
	if err != nil {
		return nil, err
	}

	for _, lib := range m.plex.Libraries {
		m.DB.AddLibrary(lib)
	}

	return &m, nil
}

// NewForMCP creates a new Mediate instance with fast initialization for MCP mode
// Heavy data loading is deferred to background goroutines.
func NewForMCP(c config.Config, options ...Options) (*Mediate, error) {
	m, err := New(c, options...)
	if err != nil {
		return nil, err
	}

	// Start heavy data loading in background to avoid blocking MCP initialization
	go func() {
		m.logger.Info("Starting background data loading for MCP mode")

		// Load shows data
		var loadErr error
		start := time.Now()
		loadErr = m.loadShows()
		if loadErr != nil {
			m.logger.Error("Failed to load shows", "error", loadErr)
		} else {
			m.logger.Info("Background shows loading completed", "duration", time.Since(start))
		}

		// Populate Plex libraries
		start = time.Now()
		m.plex.PopulateLibraries()()
		m.logger.Info("Background plex library refresh completed", "duration", time.Since(start))

		// Load Plex data
		loadErr = m.loadPlex()
		if loadErr != nil {
			m.logger.Error("Failed to load plex data", "error", loadErr)
		}

		// Load movies
		loadErr = m.loadMovies()
		if loadErr != nil {
			m.logger.Error("Failed to load movies", "error", loadErr)
		}

		m.logger.Info("Background data loading completed")
	}()

	return m, nil
}

// LoadDataSync loads all data synchronously (for traditional job mode).
func (m *Mediate) LoadDataSync() error {
	m.logger.Info("Loading shows from Sonarr...")
	start := time.Now()
	err := m.loadShows()
	if err != nil {
		return err
	}
	m.logger.Info("Shows loaded from Sonarr", "duration", time.Since(start))

	m.logger.Info("Refreshing Plex libraries...")
	start = time.Now()
	m.plex.PopulateLibraries()()
	m.logger.Info("Plex libraries refreshed", "duration", time.Since(start))

	m.logger.Info("Loading Plex show data...")
	start = time.Now()
	err = m.loadPlex()
	if err != nil {
		return err
	}
	m.logger.Info("Plex show data loaded", "duration", time.Since(start))

	m.logger.Info("Loading movies...")
	start = time.Now()
	err = m.loadMovies()
	if err != nil {
		return err
	}
	m.logger.Info("Movies loaded", "duration", time.Since(start))

	m.logger.Info("Data loading complete")
	return nil
}

func (m *Mediate) loadPlex() error {
	var err error
	for _, lib := range m.plex.Libraries {
		if lib.Type == library.TypeShow {
			err = m.loadPlexShows(lib)
			if err != nil {
				return err
			}
		}

		// if lib.Type == library.TypeMovie {
		// 	err = m.loadPlexMovies(lib)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}
	return nil
}

func (m *Mediate) Close() {
	m.cancel()
	m.plex.Close()
}

func (m *Mediate) GetShows() *shows.Shows {
	return m.DB.GetShows()
}

func (m *Mediate) Config() config.Config {
	return m.config
}

// TriggerEpisodeSearch triggers a search for specific episodes in Sonarr.
func (m *Mediate) TriggerEpisodeSearch(episodeIDs []int64) error {
	if len(episodeIDs) == 0 {
		return nil
	}

	_, err := m.sonarr.SendCommand(&sonarr.CommandRequest{
		Name:       "EpisodeSearch",
		EpisodeIDs: episodeIDs,
	})

	return err
}

// MarkSubsequentEpisodesUnwatched marks all episodes after the given episode as unwatched in Plex.
func (m *Mediate) MarkSubsequentEpisodesUnwatched(ep *shows.Episode) (int, []error) {
	episodes := m.DB.AllEpisodesAfter(ep)

	var errors []error
	count := 0

	for _, e := range episodes {
		if e.PlexRatingKey == "" {
			continue
		}

		if err := m.plex.UnScrobble(e.PlexRatingKey); err != nil {
			m.logger.Error("Failed to unscrobble episode",
				"show", e.ShowTitle, "s", e.Season, "e", e.Episode, "error", err)
			errors = append(errors, err)
			continue
		}

		e.Watched = false
		e.LastViewedAt = nil
		m.DB.Save(e)
		count++
	}

	return count, errors
}
