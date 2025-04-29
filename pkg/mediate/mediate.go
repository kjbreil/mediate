package mediate

import (
	"context"
	"github.com/kjbreil/go-plex/library"
	"github.com/kjbreil/go-plex/plex"
	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/movies"
	"github.com/kjbreil/mediate/pkg/shows"
	"github.com/kjbreil/mediate/pkg/store"
	"golift.io/starr"
	"golift.io/starr/radarr"
	"golift.io/starr/sonarr"
	"log/slog"
	"time"
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

func New(c config.Config, options ...MediateOptions) (*Mediate, error) {
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

	m.DB, err = store.InitDB()
	if err != nil {
		return nil, err
	}

	sonarrConfig := starr.New(c.Sonarr.ApiKey, c.Sonarr.URL, 0)
	m.sonarr = sonarr.New(sonarrConfig)

	radarrConfig := starr.New(c.Radarr.ApiKey, c.Radarr.URL, 0)
	m.radarr = radarr.New(radarrConfig)

	err = m.plex.InitLibraries()
	if err != nil {
		return nil, err
	}

	for _, lib := range m.plex.Libraries {
		m.DB.AddLibrary(lib)
	}

	// go func() {
	_ = m.loadShows()
	// if err != nil {
	// 	return
	// }
	start := time.Now()
	m.plex.PopulateLibraries()()
	m.logger.Info("plex library refreshed", "duration", time.Since(start))
	_ = m.loadPlex()
	// }()

	err = m.loadShows()
	if err != nil {
		return nil, err
	}

	err = m.loadMovies()
	if err != nil {
		return nil, err
	}

	return &m, nil
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
