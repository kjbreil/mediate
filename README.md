# Mediate

## Overview
Mediate is a Go-based media management system that integrates with Plex, Sonarr, and Radarr to automate the management of TV shows. The application monitors viewing habits, automatically downloads new episodes, and cleans up old content based on configurable rules.

## Features
- **Automatic Monitoring**: Tracks pilot episodes and sets appropriate monitoring status
- **Smart Downloads**: Automatically downloads the next few episodes when you start watching a show
- **Cleanup Management**: Removes watched episodes after 5 days and unwatched episodes after 30 days
- **Real-time Plex Integration**: Monitors Plex activity to trigger immediate downloads
- **Scheduled Jobs**: Configurable job system to manage various media tasks

## Project Structure
- `/cmd/mediate`: Contains the main application entry point
- `/model`: Database models and SQL queries
- `/pkg`: Core functionality modules
  - `/cli`: Command-line interface handling
  - `/config`: Configuration management
  - `/jobs`: Job definitions and scheduling
  - `/mediate`: Core business logic
  - `/movies`: Movie-related functionality
  - `/service`: Service management and scheduling
  - `/shows`: TV show management
  - `/store`: Data storage functionality

## Getting Started
To run the application locally:

```bash
go build -o mediate ./cmd/mediate
./mediate --log-level=debug --jobs=monitor,download,delete,refresh --watch-plex
```

Available jobs:
- `monitor`: Tracks and sets monitoring status for episodes
- `download`: Downloads new episodes for shows you're watching
- `delete`: Removes old episodes based on configured rules
- `refresh`: Updates metadata for shows and episodes

## Configuration
Currently, configuration is mostly hardcoded in the application. Future versions will support external configuration files.

## Docker (Future Implementation)

### Docker Considerations
For containerizing this application, the following should be addressed:
- Configuration should be moved to environment variables or a mounted config file
- Security-sensitive information (API keys, tokens) should not be hardcoded
- SQLite database should be mounted as a volume for persistence
- Network configuration to allow communication with Plex, Sonarr, and Radarr services

### Planned Docker Implementation
```dockerfile
# TO BE IMPLEMENTED
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o mediate ./cmd/mediate

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/mediate .
VOLUME /app/data
# Configuration will be provided via environment variables
CMD ["./mediate", "--config=/app/config/config.yaml"]
```

## Known Issues
- Configuration is currently hardcoded and needs to be extracted to a configuration file
- API keys and tokens are exposed in the code and should be moved to a secure configuration
- Some IP addresses and URLs are hardcoded and would need to be configurable for use in different environments
- The config file loading functionality is not yet implemented (marked with TODO in main.go)

## License
See the [LICENSE](LICENSE) file for details.