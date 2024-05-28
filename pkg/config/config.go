package config

type Config struct {
	Plex   Plex
	Sonarr Sonarr
	Radarr Radarr
}

type Plex struct {
	URL   string
	Token string

	Ignored []string
}

func (p *Plex) Ignore(toIgnore string) bool {
	for _, ig := range p.Ignored {
		if ig == toIgnore {
			return true
		}
	}
	return false
}

type Sonarr struct {
	URL    string
	ApiKey string
}

type Radarr struct {
	URL    string
	ApiKey string
}
