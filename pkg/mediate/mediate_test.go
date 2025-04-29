package mediate

import (
	"fmt"
	"testing"

	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/movies"
)

func TestNew(t *testing.T) {
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

	m, err := New(c)
	if err != nil {
		t.Fatal(err)
	}

	moviesList := m.Movies.Find(movies.Finders[movies.DownloadedNotWanted])
	for _, e := range moviesList {
		fmt.Printf("Title: %s\n", e.Title)
	}
	// err = m.DeleteMovies(moviesList)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// fmt.Printf("\n\n\n")
	// moviesList = m.Movies.Find(movies.Finders[movies.DownloadedAfterWatched])
	// for _, e := range moviesList {
	// 	fmt.Printf("Title: %s\n", e.Title)
	// }
	// fmt.Printf("\n\n\n")
	// moviesList = m.Movies.Find(movies.Finders[movies.WatchedCanDelete])
	// for _, e := range moviesList {
	// 	fmt.Printf("Title: %s\n", e.Title)
	// }

	// episodes := m.Shows.Find(shows.Finders[shows.MissingFromPlex])
	// for _, e := range episodes {
	// 	fmt.Printf("Show: %s, Title: %s, Season: %d, Episode: %d\n", e.ShowTitle, e.Title, e.Season, e.Episode)
	// }
	// fmt.Printf("\n\n\n")
	// episodes = m.Shows.Find(shows.Finders[shows.DownloadedAfterWatched])
	// for _, e := range episodes {
	// 	fmt.Printf("Show: %s, Title: %s, Season: %d, Episode: %d\n", e.ShowTitle, e.Title, e.Season, e.Episode)
	// }
	// fmt.Printf("\n\n\n")
	// episodes = m.Shows.Find(shows.Finders[shows.WatchedCanDelete])
	// for _, e := range episodes {
	// 	fmt.Printf("Show: %s, Title: %s, Season: %d, Episode: %d\n", e.ShowTitle, e.Title, e.Season, e.Episode)
	// }
	// err = m.DeleteEpisodes(episodes)
	// if err != nil {
	// 	return
	// }

	// m.findViewed()
	// m.findUnwatched()
	// m.setMonitored()
	// m.MarkOnlyPilotUnwatched()

}
