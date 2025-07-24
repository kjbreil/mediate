module github.com/kjbreil/mediate

go 1.23.0

toolchain go1.24.5

require (
	github.com/kjbreil/go-plex v0.0.0-20240502194204-449fc6fa8eec
	github.com/mark3labs/mcp-go v0.34.0
	github.com/mattn/go-sqlite3 v1.14.29
	golift.io/starr v1.1.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	nhooyr.io/websocket v1.8.17 // indirect

)

//replace golift.io/starr => /Users/kjell/dev/starr

replace github.com/jrudio/go-plex-client => /Users/kjell/dev/go-plex-client

replace github.com/kjbreil/go-plex => /Users/kjell/dev/go-plex
