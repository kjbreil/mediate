package shows

type Episodes map[int]map[int]*Episode

func (e Episodes) Count() int {
	count := 0
	for _, season := range e {
		for _, epi := range season {
			if epi.HasFile {
				count++
			}
		}
	}
	return count
}

func (e Episodes) OnlyPilot() *Episode {
	if e.Count() == 1 {
		for _, season := range e {
			for _, episode := range season {
				if episode.IsPilot() {
					return episode
				}
			}
		}
	}

	return nil
}

func (e Episodes) WithinLastXOfSeason(x int, episode *Episode) bool {
	for _, season := range e {
		for n, epi := range season {
			if epi.PlexRatingKey == episode.PlexRatingKey {
				if int(n)+x > len(season) {
					return true
				}
			}
		}
	}

	return false
}

func (e Episodes) XEpisodesAheadDownloaded(x int, episode *Episode) bool {
	episodes := e.GetEpisodesXAhead(x, episode)

	if len(episodes) == 0 {
		return false
	}

	for _, ep := range episodes {
		if !ep.HasFile {
			return false
		}
	}
	return true
}

func (e Episodes) GetEpisodesXAhead(x int, episode *Episode) EpisodesArr {
	var episodes EpisodesArr
	count := x

	seasonNumber := 1
	episodeNumber := 1
	var found bool
	for {
		if season, ok := e[seasonNumber]; ok {

			for {
				if epi, ok := season[episodeNumber]; ok {
					episodeNumber++

					if found {
						episodes = append(episodes, epi)
						count--
					}
					if count == 0 {
						return episodes
					}

					if epi.TvdbID == episode.TvdbID {
						found = true
					}
				} else {
					seasonNumber++
					episodeNumber = 1
					break
				}
			}

		} else {
			break
		}
	}

	return episodes
}
