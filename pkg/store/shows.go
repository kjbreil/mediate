package store

import (
	"time"

	"github.com/kjbreil/mediate/pkg/shows"
	"golift.io/starr/sonarr"
)

func (s *Store) AddSonarrSeries(ser *sonarr.Series) (*shows.Show, error) {
	var sh shows.Show

	s.First(&sh, ser.TvdbID)

	sh.Title = ser.Title
	sh.SonarrID = ser.ID
	sh.Continuing = ser.Status == "continuing"
	sh.TvdbID = int(ser.TvdbID)

	s.Save(&sh)

	return &sh, nil
}

func (s *Store) AddSonarrEpisode(sh *shows.Show, ep *sonarr.Episode) (*shows.Episode, error) {
	var e shows.Episode

	var airDate *time.Time
	if parse, err := time.Parse("2006-01-02", ep.AirDate); err == nil {
		airDate = &parse
	}

	s.First(&e, ep.TvdbID)
	e.ShowTitle = sh.Title
	e.ShowTvdbID = sh.TvdbID
	e.Title = ep.Title
	e.Season = ep.SeasonNumber
	e.SeasonEpisode = ep.SeasonNumber*1000 + ep.EpisodeNumber
	e.Episode = ep.EpisodeNumber
	e.SonarrID = ep.ID
	e.SonarrFileID = ep.EpisodeFileID
	e.Wanted = ep.Monitored
	e.HasFile = ep.HasFile
	e.AirDate = airDate
	e.TvdbID = int(ep.TvdbID)
	if e.HasFile {
		e.Downloading = false
	}

	s.Save(&e)

	return &e, nil
}

func (s *Store) GetShow(tvdbID int) *shows.Show {
	var show shows.Show

	result := s.Debug().Preload("Episodes").First(&show, tvdbID)
	if result.RowsAffected != 0 {
		show.Library = s.libraries[show.LibraryUUID]
		return &show
	}
	result = s.Joins("JOIN episodes ON episodes.show_tvdb_id = shows.tvdb_id AND episodes.tvdb_id = ?", tvdbID).
		Find(&show)
	if result.RowsAffected != 0 {
		show.Library = s.libraries[show.LibraryUUID]
		return &show
	}

	return nil
}

func (s *Store) GetShowFromRatingKey(ratingKey string) *shows.Show {
	var show shows.Show

	// result := s.Debug().Preload("Episodes").First(&show, tvdbID)
	// if result.RowsAffected != 0 {
	// 	show.Library = s.libraries[show.LibraryUUID]
	// 	return &show
	// }
	result := s.Joins("JOIN episodes ON episodes.show_tvdb_id = shows.tvdb_id AND episodes.plex_rating_key =?", ratingKey).
		Find(&show)
	if result.RowsAffected != 0 {
		show.Library = s.libraries[show.LibraryUUID]
		return &show
	}

	return nil
}

func (s *Store) GetEpisodeFromRatingKey(ratingKey string) *shows.Episode {
	var episode shows.Episode

	result := s.Where("plex_rating_key =?", ratingKey).Find(&episode)
	if result.RowsAffected != 0 {
		return &episode
	}

	return nil
}

func (s *Store) GetShows() *shows.Shows {
	var shows shows.Shows
	s.Preload("Episodes").Find(&shows)
	for _, sh := range shows {
		sh.Library = s.libraries[sh.LibraryUUID]
	}
	return &shows
}

func (s *Store) GetEpisodes(showTvdbID int) shows.Episodes {
	var episodes shows.Episodes

	s.Where("show_tvdb_id =?", showTvdbID).Find(&episodes)

	return episodes
}

func (s *Store) MarkDownloading(ids ...int) {
	s.Model(&shows.Episode{}).
		Where("tvdb_id IN (?)", ids).
		Update("downloading", true)
	s.Model(&shows.Episode{}).
		Where("tvdb_id NOT IN (?)", ids).
		Update("downloading", false)
}

func (s *Store) SonarrIDs(ids ...int) shows.Episodes {
	var episodes shows.Episodes
	s.Where("sonarr_id IN (?)", ids).Find(&episodes)
	return episodes
}

func (s *Store) NextXEpisodes(x int, ep *shows.Episode) shows.Episodes {
	var episodes shows.Episodes
	s.Where("show_tvdb_id =? AND  season_episode >?", ep.ShowTvdbID, ep.SeasonEpisode).
		Limit(x).
		Order("season ASC, episode ASC").
		Find(&episodes)
	return episodes
}

func (s *Store) AllEpisodesAfter(ep *shows.Episode) shows.Episodes {
	var episodes shows.Episodes
	s.Where("show_tvdb_id = ? AND season_episode > ?", ep.ShowTvdbID, ep.SeasonEpisode).
		Order("season ASC, episode ASC").
		Find(&episodes)
	return episodes
}
