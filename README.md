# Mediate

Mediate is a media management tool that automates tasks between Plex, Sonarr, and Radarr. It runs as a background service, monitoring your media libraries and performing actions based on your viewing habits and configured settings.

## Key Features

- **Automated Downloads:** Automatically queues the next episodes in Sonarr when you finish watching an episode in Plex.
- **Smart Deletion:** Frees up disk space by automatically deleting watched episodes after a configurable period.
- **Library Monitoring:** Keeps your media libraries in sync and ensures that new shows are properly monitored in Sonarr.
- **Real-time Plex Monitoring:** Watches for real-time events in Plex and triggers actions accordingly.
- **Configurable Jobs:** Fine-tune the behavior of Mediate by enabling and configuring different jobs.

## How to Run

Mediate is a command-line application that runs as a service. You can specify which jobs to run and their intervals using command-line flags.

**Example:**

```bash
go run cmd/mediate/main.go --jobs=monitor,download,delete --monitor-interval=1h --download-interval=15m --delete-interval=24h
```

## Jobs

Mediate provides several jobs that can be enabled and configured:

- **`monitor`**: This job monitors pilot episodes of your shows. If a pilot episode is available but not yet monitored in Sonarr, this job will add it to Sonarr's monitoring list.
- **`download`**: When you watch an episode in Plex, this job will automatically download the next few unaired episodes, ensuring you always have something to watch.
- **`delete`**: To help manage disk space, this job will delete episodes that have been watched after a certain period. It can also delete unwatched episodes that have been on your drive for a long time.
- **`refresh`**: This job periodically refreshes the information about your shows and episodes, ensuring that Mediate's internal database is up-to-date with Plex and Sonarr.
- **`plex-watch`**: This job enables real-time monitoring of your Plex server. When someone starts watching a show, it can trigger immediate actions, such as downloading the next episodes. Use the `--watch-plex` flag to enable this job.
