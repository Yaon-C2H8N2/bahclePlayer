# Bahcle Player

Bahcle Player is a music player that uses Twitch channel Rewards to add musics to a Playlist/Queue.
A running version can be found here: https://bahcleplayer.yaon.fr/player.

The project leverages the following APIs:

- [Twitch API](https://dev.twitch.tv/docs/api) to create Polls and subscriptions to channel events.
- [Twitch EventSub](https://dev.twitch.tv/docs/eventsub) to listen to channel events through websockets.
- [Youtube API](https://developers.google.com/youtube/v3) to retrieve videos metadata.

## Features

- [x] Login with your Twitch account.
- [x] Set up which channel Rewards you want to use and how (directly add to the queue or create a poll that needs to be
  validated) through the settings menu.
- [x] Auto refund the channel points if the song is not found.
- [x] Manage the queue and playlist (skip, remove, manually add songs).
- [ ] Customizable overlays to display the current song and the queue on your stream.

## Technologies
- [React](https://reactjs.org/) for the frontend.
- [Shadcn](https://shadcn.dev/) for the UI components.
- [Go](https://golang.org/) for the backend.
- [Gin](https://gin-gonic.com) for the backend Rest API.
- [Gorilla Websocket](https://gorilla.github.io) for the API websocket connections.
- [PostgreSQL](https://www.postgresql.org/) for the database.
- [Valkey](https://valkey.io) to enable Pub/Sub patterns for the backend and to schedule tasks.

## Installation

Coming soon... (Dockerfile and docker-compose.yml are available in the repository though)