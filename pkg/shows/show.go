package shows

import (
	"github.com/kjbreil/go-plex/library"
	"golift.io/starr/sonarr"
	"sync"
	"time"
)

type Show struct {
	Title         string
	PlexRatingKey string
	Rating        float64
	Ignore        bool
	SonarrId      int64
	Continuing    bool
	Episodes      Episodes
	Library       *library.Library
	TvdbID        int

	m sync.Mutex
}

func (s *Show) GetEpisodes() EpisodesArr {
	rtn := make(EpisodesArr, 0)
	for _, season := range s.Episodes {
		for _, epi := range season {
			rtn = append(rtn, epi)
		}
	}
	return rtn
}

func (s *Show) AddSonarrEpisode(ep *sonarr.Episode) *Episode {
	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.Episodes[int(ep.SeasonNumber)]; !ok {
		s.Episodes[int(ep.SeasonNumber)] = make(map[int]*Episode)
	}

	var airDate *time.Time
	if parse, err := time.Parse("2006-01-02", ep.AirDate); err == nil {
		airDate = &parse
	}

	e, ok := s.Episodes[int(ep.SeasonNumber)][int(ep.EpisodeNumber)]
	if !ok {
		e = &Episode{}
	}

	e.ShowTitle = s.Title
	e.Title = ep.Title
	e.Season = int(ep.SeasonNumber)
	e.Episode = int(ep.EpisodeNumber)
	e.SonarrId = ep.ID
	e.SonarrFileId = ep.EpisodeFileID
	e.Wanted = ep.Monitored
	e.HasFile = ep.HasFile
	e.AirDate = airDate
	e.TvdbID = int(ep.TvdbID)
	if e.HasFile {
		e.Downloading = false
	}

	s.Episodes[int(ep.SeasonNumber)][int(ep.EpisodeNumber)] = e
	return e
}

func (s *Show) LibraryTitle() string {
	if s.Library == nil {
		return ""
	}
	return s.Library.Title
}
