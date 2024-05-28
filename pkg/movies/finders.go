package movies

import "time"

type FindFunc func(movie *Movie) bool

var (
	WindowDuration time.Duration = time.Hour * 24 * 30
	MaxMovieRating float64       = 3
)

type FinderKey int

const (
	WatchedCanDelete FinderKey = iota
	DownloadedAfterWatched
	MissingFromPlex
	DownloadedNotWanted
)

var Finders map[FinderKey]FindFunc

func init() {
	Finders = map[FinderKey]FindFunc{
		WatchedCanDelete: func(movie *Movie) bool {
			return movie.Watched &&
				movie.WatchedOutside(WindowDuration) &&
				movie.Rating <= MaxMovieRating &&
				!movie.AddedWithin(WindowDuration)
		},
		DownloadedAfterWatched: func(movie *Movie) bool {
			return movie.HasFile && movie.Watched && movie.WatchedBeforeDownload() && movie.Rating <= MaxMovieRating
		},
		MissingFromPlex: func(movie *Movie) bool {
			return movie.PlexRatingKey == "" && movie.HasFile
		},
		DownloadedNotWanted: func(movie *Movie) bool {
			return movie.HasFile && !movie.Wanted
		},
	}
}

func (m Movies) Find(fn FindFunc) (rtn []*Movie) {
	for _, movie := range m {
		if movie.Ignore {
			continue
		}

		if fn(movie) {
			rtn = append(rtn, movie)
		}
	}

	return
}
