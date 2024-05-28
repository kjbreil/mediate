package shows

import "sync"

// Shows is a map of TVDB ID to show info
type Shows struct {
	s map[int]*Show
	m sync.Mutex
}

func NewShows() *Shows {
	return &Shows{
		s: make(map[int]*Show),
		m: sync.Mutex{},
	}
}

func (s *Shows) NewShow(tvdbID int) *Show {
	s.m.Lock()
	defer s.m.Unlock()

	ns := &Show{
		TvdbID:   tvdbID,
		Episodes: make(map[int]map[int]*Episode),
		m:        sync.Mutex{},
	}
	s.s[tvdbID] = ns

	return ns
}

func (s *Shows) Slice() []*Show {
	s.m.Lock()
	defer s.m.Unlock()
	rtn := make([]*Show, 0)
	for _, show := range s.s {
		rtn = append(rtn, show)
	}
	return rtn
}

func (s *Shows) GetNextXEpisodes(x int, episode *Episode) EpisodesArr {
	s.m.Lock()
	defer s.m.Unlock()
	for _, show := range s.s {
		for _, season := range show.Episodes {
			for _, epi := range season {
				if epi.TvdbID == episode.TvdbID {
					return show.Episodes.GetEpisodesXAhead(x, episode)
				}
			}
		}
	}
	return nil
}

func (s *Shows) GetShow(tvdbid int) *Show {
	s.m.Lock()
	defer s.m.Unlock()
	if show, ok := s.s[tvdbid]; ok {
		return show
	}
	for _, show := range s.s {
		if show.TvdbID == tvdbid {
			return show
		}
		for _, season := range show.Episodes {
			for _, epi := range season {
				if epi.TvdbID == tvdbid {
					return show
				}
			}
		}
	}
	return nil
}

func (s *Shows) GetShowEpisodes(tvdbid int) EpisodesArr {
	s.m.Lock()
	defer s.m.Unlock()
	var episodes EpisodesArr
	for _, show := range s.s {
		if show.TvdbID == tvdbid {
			for _, season := range show.Episodes {
				for _, epi := range season {
					episodes = append(episodes, epi)
				}
			}
		}
	}
	return episodes
}

// GetEpisode matches to the TvdbID or PlexRatingKey or SonarrId in that order
func (s *Shows) GetEpisode(id any) *Episode {
	s.m.Lock()
	defer s.m.Unlock()
	for _, show := range s.s {
		for _, season := range show.Episodes {
			for _, epi := range season {
				if epi.TvdbID == id || epi.PlexRatingKey == id || epi.SonarrId == id {
					return epi
				}
			}
		}
	}
	return nil
}

func (s *Shows) MarkDownloading(tvdbids ...int) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, show := range s.s {
		for _, season := range show.Episodes {
			for _, epi := range season {
				epi.Downloading = false
				for _, i := range tvdbids {
					if epi.TvdbID == i {
						epi.Downloading = true
					}
				}
			}
		}
	}
}

func (s *Shows) SonarrIdToTvdbId(sonarrIds ...int) []int {
	s.m.Lock()
	defer s.m.Unlock()
	rtn := make([]int, 0)
	for _, show := range s.s {
		for _, season := range show.Episodes {
			for _, epi := range season {
				for _, i := range sonarrIds {
					if int(epi.SonarrId) == i {
						rtn = append(rtn, epi.TvdbID)
					}
				}
			}
		}
	}
	return rtn
}

// func (s Shows) FindEpisode(ratingKey string) (*Show, *Episode) {
//
// }
