-- name: GetShow :one
SELECT * FROM shows
WHERE tvdb_id = ? LIMIT 1;

-- name: ListShows :many
SELECT * FROM shows;

-- name: UpsertShow :exec
INSERT INTO shows (tvdb_id, title, rating, plex_rating_key, continuing, sonarr_id, library_id)
VALUES (@tvdb_id, @title, @rating, @plexRatingKey, @continuing, @sonarrId, @libraryId)
ON CONFLICT (tvdb_id)
    DO UPDATE SET title = @title,
                  rating = @rating,
                  plex_rating_key = @plexRatingKey,
                  continuing = @continuing,
                  sonarr_id = @sonarrId,
                  library_id = @libraryId;