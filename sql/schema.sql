create table shows
(
    tvdb_id         integer
        constraint shows_pk
            primary key,
    title           text,
    rating          REAL,
    plex_rating_key text,
    continuing      integer,
    sonarr_id       integer,
    library_id      integer
);

create table episodes
(
    show_title      TEXT,
    title           text,
    season          integer,
    episode         integer,
    tvdb_id         integer not null
        constraint episodes_pk
            primary key,
    show_tvdb_id    integer
        constraint episodes_shows_tvdb_id_fk
            references shows,
    plex_rating_key text,
    sonarr_id       integer,
    sonarr_file_id  integer,
    downloading     integer,
    watched         integer,
    has_file        integer,
    downloaded_at   integer,
    last_viewed_at  integer,
    air_date        integer,
    duration        integer
);