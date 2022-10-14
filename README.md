# Playlist Duplicator

A simple way to duplicate the contents of one Spotify playlist into another.

**Note:** Due to a limitation with the [Spotify API library being used](https://github.com/zmb3/spotify/issues/180), podcast episodes in a playlist will not be transferred at this time.

## Prerequisites

1. Go to the [Spotify developer dashboard](https://developer.spotify.com/dashboard/applications) and create an application
    - Copy the client ID and secret
    - Click "Edit Settings" and `http://localhost:8080/callback` to the "Redirect URIs" section
2. Set the `SPOTIFY_ID` and `SPOTIFY_SECRET` environment variables with the client ID and secret respectively
3. Set the `SPOTIFY_PLAYLIST_ID` environment variable with the playlist ID of the playlist you want to duplicate
    - This should just be the ID itself
    - E.g. `https://open.spotify.com/playlist/4GtQVhGjAwcHFz82UKy3Ca?si=32208c6432ca47c4` ⇒ `4GtQVhGjAwcHFz82UKy3Ca`
4. Set the `SPOTIFY_HOLDING_PLAYLIST_ID` environment variable with the playlist ID of the playlist you want to duplicate to
    - This should just be the ID itself
    - E.g. `https://open.spotify.com/playlist/4GtQVhGjAwcHFz82UKy3Ca?si=32208c6432ca47c4` ⇒ `4GtQVhGjAwcHFz82UKy3Ca`
5. **(Optional)** Set the `DISCORD_WEBHOOK_URL` to have Playlist Duplicator send you a notification via Discord
    1. Go to a Discord server
    2. Open the channel preferences (right-click > Edit Channel or hover/select the channel and click the settings icon)
    3. Go to Integrations > Webhooks > New Webhook
        - Give it a name, then copy the URL
    4. Paste the URL

Environment variables can be set directly or via a `.env` file

## Usage

There are two ways to run Playlist Duplicator: via the command-line or via Docker.

### Command-line

To use the program via the command-line, either clone the repo and build it yourself (TBA), or download the executable from the latest release. Then, make sure your environment variables are set, then run `./playlist-duplicator`.

On the first run, you will be prompted to copy a URL to your browser, then paste the link into the terminal. This is a one-time event to generate a token that Playlist Duplicator can use later on (stored in `token.json`).

### Docker

Using the Docker container requires cloning the repository, then building it with the following command:

```bash
docker build -t playlist-duplicator:v2.0 .
```

After that, either have a `token.json` file on hand (from a previous run) or create an empty `token.json`. Bind mount the `token.json` to the container, as well as adding your environment variables.

**IMPORTANT:** If you have an empty `token.json`, make sure to add the `-it` before the first `-v` (e.g. `docker run --rm -it -v ...`) so you are able to copy and paste the URL

#### With `.env` file

```bash
docker run --rm -v token.json:/token.json --env-file .env playlist-duplicator:v2.0
```

#### With environment variables as parameters

```bash
docker run --rm -v token.json:/token.json -e SPOTIFY_ID=xxx -e SPOTIFY_SECRET=xxx -e SPOTIFY_PLAYLIST_ID=xxx -e SPOTIFY_HOLDING_PLAYLIST_ID=xxx playlist-duplicator:v2.0
```

## License

```text
Playlist Duplicator - A simple way to duplicate the contents of one Spotify playlist into another
Copyright (C) 2022 H. Kamran (https://hkamran.com)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```
