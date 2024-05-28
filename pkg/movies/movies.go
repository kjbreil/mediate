package movies

import "time"

type Movies map[int]*Movie

type Movie struct {
	Title            string
	PlexRatingKey    string
	Year             int
	RadarrID         int64
	Watched          bool
	Wanted           bool
	HasFile          bool
	DownloadedAt     time.Time
	LastViewedAt     *time.Time
	UpdatedAt        time.Time
	Rating           float64
	Ignore           bool
	Library          string
	TmdbID           int
	Path             string
	QualityProfileID int64
}

func (m *Movie) WatchedWithin(d time.Duration) bool {
	if m.LastViewedAt == nil {
		return false
	}
	return m.Watched && time.Since(*m.LastViewedAt) < d
}

func (m *Movie) WatchedOutside(d time.Duration) bool {
	if m.LastViewedAt == nil {
		return false
	}
	return m.Watched && time.Since(*m.LastViewedAt) > d
}

func (m *Movie) WatchedBeforeDownload() bool {

	if m.LastViewedAt == nil {
		return false
	}
	return m.LastViewedAt.Before(m.DownloadedAt)
}

func (m *Movie) AddedWithin(d time.Duration) bool {
	return time.Since(m.DownloadedAt) < d
}
