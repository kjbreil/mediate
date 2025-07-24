package mediate

import (
	"github.com/kjbreil/mediate/pkg/movies"
	"golift.io/starr/radarr"
)

func (m *Mediate) loadMovies() error {
	allMovies, err := m.radarr.GetMovie(&radarr.GetMovie{TMDBID: 0})
	if err != nil {
		return err
	}
	for _, radarrMovie := range allMovies {

		movie := &movies.Movie{
			Title:            radarrMovie.Title,
			PlexRatingKey:    "",
			Year:             radarrMovie.Year,
			TmdbID:           int(radarrMovie.TmdbID),
			RadarrID:         radarrMovie.ID,
			Wanted:           radarrMovie.Monitored,
			HasFile:          radarrMovie.HasFile,
			Path:             radarrMovie.Path,
			QualityProfileID: radarrMovie.QualityProfileID,
		}
		if radarrMovie.MovieFile != nil {
			movie.DownloadedAt = radarrMovie.MovieFile.DateAdded
		}
		m.Movies[int(radarrMovie.TmdbID)] = movie
	}

	return nil
}
