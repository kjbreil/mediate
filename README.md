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
Configuration is managed via YAML files. See `config.yaml.example` for available options.

```bash
# Create default config
make config
```

## Docker

### Pull from GitHub Container Registry

```bash
docker pull ghcr.io/kjbreil/mediate:latest
```

### Run with Docker

```bash
docker run -d \
  --name mediate \
  -v /path/to/config.yaml:/app/config/config.yaml \
  -v /path/to/data:/app/data \
  ghcr.io/kjbreil/mediate:latest
```

### Build Locally

```bash
make docker
```

### Release a New Version

Version tags trigger the CI pipeline to build and push Docker images:

```bash
make patch  # v1.0.0 -> v1.0.1
make minor  # v1.0.0 -> v1.1.0
make major  # v1.0.0 -> v2.0.0
```

## Known Issues
- Some features are still in development

## License
See the [LICENSE](LICENSE) file for details.