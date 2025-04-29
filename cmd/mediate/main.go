package main

import (
	"fmt"
	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/mediate"
	"github.com/kjbreil/mediate/pkg/shows"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := config.Config{
		Plex: config.Plex{
			URL:   "http://plex1.kaygel.io:32400",
			Token: "-HacSX44mXL1WHVACUZ5",
			Ignored: []string{
				"Kids TV Shows",
				"Kids Movies",
			},
		},
		Sonarr: config.Sonarr{
			ApiKey: "67bd04cc551149188947a0024a7f5c1e",
			URL:    "http://10.0.1.22:8989/show/",
		},
		Radarr: config.Radarr{
			ApiKey: "e2eab479a088404387c7b1b48eab5287",
			URL:    "http://10.0.1.22:7878/film/",
		},
	}

	logger := slog.Default()

	m, err := mediate.New(
		c,
		mediate.WithLogger(logger),
	)
	defer m.Close()

	if err != nil {
		log.Fatal(err)
	}

	// m.OnPlexPlaying(func(pp *mediate.PlexPlaying) {
	//
	// 	ep := pp.Episode()
	// 	if ep == nil {
	// 		return
	// 	}
	//
	// 	if pp.Changed {
	// 		logger.Info("Plex reporting show being watched", "title", ep.Title, "season", ep.Season, "episode", ep.Episode, "left", pp.TimeLeft().Minutes())
	// 		m.UpdateEpisode(ep)
	// 		m.DownloadEpisodes(
	// 			m.DB.NextXEpisodes(3, ep).
	// 				HasFile(false).
	// 				Aired(true).
	// 				Downloading(false))
	// 	}
	// })

	shows.WindowDuration = time.Minute * 24 * 5

	episodes := m.GetShows().Find(shows.Finders[shows.WatchedCanDelete])
	for _, e := range episodes {
		fmt.Printf("Show: %s, Title: %s, Season: %d, Episode: %d\n", e.ShowTitle, e.Title, e.Season, e.Episode)
	}

	shows.WindowDuration = time.Minute * 24 * 30

	fmt.Printf("\n\n\n")

	episodes = m.GetShows().Find(shows.Finders[shows.NotWatchedCanDelete])
	for _, e := range episodes {
		fmt.Printf("Show: %s, Title: %s, Season: %d, Episode: %d\n", e.ShowTitle, e.Title, e.Season, e.Episode)
	}
	_ = m.DeleteEpisodes(episodes)

	// m.MonitorEpisodes(m.Shows.Find(shows.Finders[shows.AllPilots]).HasFile(false).Wanted(false), true)

	// m.MarkOnlyPilotUnwatched()
	// m.MarkOnlyPilotUnwatched()
	// m.SetMonitored()
	// episodes := m.RecentlyWatched()
	// m.RefreshShowsEpisodes(episodes)
	// for _, e := range episodes {
	// 	// if e.ShowTvdbID != 211751 {
	// 	// 	continue
	// 	// }
	// 	nextEpisodes := m.DB.NextXEpisodes(3, e).
	// 		Downloading(false).
	// 		Aired(true).
	// 		HasFile(false)
	//
	// 	if len(nextEpisodes) > 0 {
	// 		m.DownloadEpisodes(nextEpisodes)
	// 		m.RefreshShowsEpisodes(nextEpisodes)
	// 	}
	//
	// }
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-ctrlC

}
