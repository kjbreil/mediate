module github.com/kjbreil/mediate

go 1.23

toolchain go1.24.4

require (
	github.com/kjbreil/go-plex v0.0.0-20240502194204-449fc6fa8eec
	github.com/mark3labs/mcp-go v0.32.0
	github.com/mattn/go-sqlite3 v1.14.22
	golift.io/starr v0.14.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/sqlite v1.5.5
	gorm.io/gorm v1.25.10
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/net v0.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	nhooyr.io/websocket v1.8.11 // indirect

)

replace golift.io/starr => /Users/kjell/dev/starr

replace github.com/jrudio/go-plex-client => /Users/kjell/dev/go-plex-client

replace github.com/kjbreil/go-plex => /Users/kjell/dev/go-plex
