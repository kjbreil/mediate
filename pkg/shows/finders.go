package shows

import "time"

type FindFunc func(show *Show, episode *Episode) bool

//nolint:gochecknoglobals // Configuration variables used across finders
var (
	WindowDuration = time.Hour * 24 * 30
	MaxShowRating  = 3.0
)

type FinderKey int

const (
	WatchedCanDelete FinderKey = iota
	DownloadedAfterWatched
	MissingFromPlex
	NotWatchedNotWanted
	AllPilots
	RecentlyWatchedEndOfSeason
	NotWatchedCanDelete
)

//nolint:gochecknoglobals // Registry pattern for finder functions
var Finders = map[FinderKey]FindFunc{
	WatchedCanDelete: func(show *Show, episode *Episode) bool {
		return episode.HasFile &&
			episode.Watched &&
			episode.WatchedOutside(WindowDuration) &&
			show.Rating <= MaxShowRating &&
			!episode.IsPilot() &&
			!episode.AddedWithin(WindowDuration)
	},
	NotWatchedCanDelete: func(show *Show, episode *Episode) bool {
		return episode.HasFile &&
			!episode.Watched &&
			show.Rating <= MaxShowRating &&
			!episode.IsPilot() &&
			!episode.AddedWithin(WindowDuration)
	},
	DownloadedAfterWatched: func(show *Show, episode *Episode) bool {
		return episode.HasFile && episode.Watched && episode.LastViewedAt != nil &&
			episode.LastViewedAt.Before(episode.DownloadedAt) &&
			show.Rating <= MaxShowRating
	},
	MissingFromPlex: func(_ *Show, episode *Episode) bool {
		return episode.PlexRatingKey == "" && episode.HasFile
	},
	NotWatchedNotWanted: func(_ *Show, episode *Episode) bool {
		return !episode.Watched && !episode.Wanted && episode.HasFile
	},
	AllPilots: func(_ *Show, episode *Episode) bool {
		return episode.IsPilot()
	},
	RecentlyWatchedEndOfSeason: func(show *Show, episode *Episode) bool {
		if show.Title != "Workaholics" {
			return false
		}
		if episode.Season != 3 {
			return false
		}
		if episode.Episode != 16 {
			return false
		}
		return episode.Watched &&
			episode.LastViewedAt != nil &&
			episode.LastViewedAt.After(time.Now().Add(-WindowDuration)) &&
			!show.Episodes.XEpisodesAheadDownloaded(3, episode) &&
			episode.HasAired()
	},
}

// Show: Animal Control, Title: Bulls and Potbellies, Season: 2, Episode: 8 LastWatched: 2024-05-01T21:30:38-07:00
// Show: Animal Control, Title: Beagles and Lemurs, Season: 2, Episode: 9 LastWatched: 2024-05-14T16:14:47-07:00
// Show: Shark Tank, Title: VSEAT, Wedy, SORx, blinger, Season: 15, Episode: 20 LastWatched: 2024-05-05T20:56:40-07:00
// Show: Jersey Shore: Family Vacation, Title: Happy Birthday Snooki!, Season: 7, Episode: 16 LastWatched: 2024-05-20T21:17:35-07:00

func (s *Shows) Find(fn FindFunc) Episodes {
	rtn := make(Episodes, 0)
	for _, show := range *s {
		if show.Ignore {
			continue
		}

		for _, episode := range show.Episodes {
			// for _, episode := range season {
			if fn(show, episode) {
				rtn = append(rtn, episode)
				// }
			}
		}
	}

	return rtn
}
