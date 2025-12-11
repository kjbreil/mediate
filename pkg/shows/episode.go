package shows

import "time"

func (e Episodes) SonarrIDs() []int64 {
	rtn := make([]int64, len(e))
	for i, ep := range e {
		rtn[i] = ep.SonarrID
	}
	return rtn
}

func (e Episodes) HasFile(hasFile bool) Episodes {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.HasFile == hasFile {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

func (e Episodes) InPlex(inPlex bool) Episodes {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.PlexRatingKey == "" != inPlex {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

func (e Episodes) Wanted(wanted bool) Episodes {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.Wanted == wanted {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

func (e Episodes) Aired(aired bool) Episodes {
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

func (e Episodes) Downloading(downloading bool) Episodes {
	rtn := make([]*Episode, 0)
	for _, ep := range e {
		if ep.Downloading == downloading {
			rtn = append(rtn, ep)
		}
	}
	return rtn
}

type Episode struct {
	ShowTitle     string
	Title         string
	Season        int
	Episode       int
	SeasonEpisode int
	TvdbID        int `gorm:"primaryKey"`
	ShowTvdbID    int

	PlexRatingKey string
	SonarrID      int64
	Watched       bool
	Wanted        bool
	HasFile       bool
	DownloadedAt  time.Time
	LastViewedAt  *time.Time
	UpdatedAt     time.Time
	SonarrFileID  int64
	AirDate       *time.Time
	Downloading   bool
	Duration      time.Duration

	// Enhanced viewing analytics
	ViewCount      int `gorm:"default:0"`
	FirstViewedAt  *time.Time
	TotalWatchTime time.Duration
	CompletionRate float64 `gorm:"default:0"` // 0-100% how much of episode was watched
	SkipCount      int     `gorm:"default:0"` // How many times user skipped this episode
	Rating         float64 `gorm:"default:0"` // User rating if available
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
