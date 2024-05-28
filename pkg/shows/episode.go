package shows

import "time"

type EpisodesArr []*Episode

func (e EpisodesArr) SonarrIds() []int64 {
	rtn := make([]int64, len(e))
	for i, ep := range e {
		rtn[i] = ep.SonarrId
	}
	return rtn
}

func (e EpisodesArr) HasFile(hasFile bool) EpisodesArr {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.HasFile == hasFile {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

func (e EpisodesArr) InPlex(inPlex bool) EpisodesArr {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.PlexRatingKey == "" != inPlex {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

func (e EpisodesArr) Wanted(wanted bool) EpisodesArr {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.Wanted == wanted {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

func (e EpisodesArr) Aired(aired bool) EpisodesArr {
	now := time.Now()
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.AirDate != nil {
			if aired && ep.AirDate.Before(now) {
				rtn = append(rtn, ep)
			}
			if !aired && ep.AirDate.After(now) {
				rtn = append(rtn, ep)
			}
		}
	}
	return rtn
}

func (e EpisodesArr) Downloading(downloading bool) EpisodesArr {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.Downloading == downloading {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

type Episode struct {
	ShowTitle string
	Title     string
	Season    int
	Episode   int
	TvdbID    int

	PlexRatingKey string
	SonarrId      int64
	Watched       bool
	Wanted        bool
	HasFile       bool
	DownloadedAt  time.Time
	LastViewedAt  *time.Time
	UpdatedAt     time.Time
	SonarrFileId  int64
	AirDate       *time.Time
	Downloading   bool
	Duration      time.Duration
}

func (e *Episode) IsPilot() bool {
	return e.Season == 1 && e.Episode == 1
}

func (e *Episode) WatchedWithin(d time.Duration) bool {
	if e.LastViewedAt == nil {
		return false
	}
	return e.Watched && time.Since(*e.LastViewedAt) < d
}

func (e *Episode) WatchedOutside(d time.Duration) bool {
	if e.LastViewedAt == nil {
		return false
	}
	return e.Watched && time.Since(*e.LastViewedAt) > d
}

func (e *Episode) AddedWithin(d time.Duration) bool {
	return time.Since(e.DownloadedAt) < d
}

func (e *Episode) Keep() bool {
	return e.IsPilot()
}

func (e *Episode) HasAired() bool {
	return e.AirDate != nil && !e.AirDate.IsZero() && e.AirDate.Before(time.Now())
}

func (e *Episode) HasNotAired() bool {
	return e.AirDate != nil && !e.AirDate.IsZero() && e.AirDate.After(time.Now())
}
