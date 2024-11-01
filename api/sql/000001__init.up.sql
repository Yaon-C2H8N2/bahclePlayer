create table users
(
    user_id   serial primary key,
    username  varchar,
    twitch_id varchar not null unique
);

create type video_type as enum ('PLAYLIST', 'QUEUE');

create table users_videos
(
    video_id   serial primary key,
    user_id    integer references users (user_id),
    youtube_id varchar    not null unique,
    url        varchar    not null,
    title      varchar    not null,
    duration   varchar    not null,
    type       video_type not null,
    created_at timestamp default now()
);

create type config_type as enum ('PLAYLIST_REDEMPTION', 'QUEUE_REDEMPTION');

create table users_config
(
    config_id serial primary key,
    user_id integer references users (user_id),
    config  config_type not null,
    value   varchar     not null
)