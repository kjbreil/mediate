module github.com/kjbreil/mediate

go 1.24.0

toolchain go1.24.5

require (
	github.com/kjbreil/go-plex v0.0.0-20251213020936-07490c812f13
	github.com/mark3labs/mcp-go v0.43.2
	github.com/mattn/go-sqlite3 v1.14.32
	golift.io/starr v1.2.1
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/kjbreil/go-plex => ../go-plex
