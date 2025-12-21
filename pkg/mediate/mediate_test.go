package mediate

import (
	"testing"

	"github.com/kjbreil/mediate/pkg/config"
	"github.com/kjbreil/mediate/pkg/movies"
)

func TestNew(t *testing.T) {
	// Create a temporary database file for testing
	tempDB := t.TempDir() + "/test.db"

	c := config.Config{
		Plex: config.Plex{
			URL:   "http://plex.example.com:32400",
			Token: "your-plex-token",
			Ignored: []string{
				"Kids TV Shows",
				"Kids Movies",
			},
		},
		Sonarr: config.Sonarr{
			APIKey: "your-sonarr-api-key",
			URL:    "http://sonarr.example.com:8989",
		},
		Radarr: config.Radarr{
			APIKey: "your-radarr-api-key",
			URL:    "http://radarr.example.com:7878",
		},
		Database: config.Database{
			Path: tempDB,
		},
		Automation: config.DefaultAutomation(),
	}

	m, err := New(c)
	if err != nil {
		t.Fatal(err)
	}

	moviesList := m.Movies.Find(movies.Finders[movies.DownloadedNotWanted])
	for _, e := range moviesList {
		t.Logf("Title: %s", e.Title)
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
