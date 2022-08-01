# Playlist Duplicator

A simple way to duplicate the contents of one Spotify playlist into another.

## Usage

1. Download the `playlist_duplicator` executable OR clone the repository and build it
    - To build, run `go build`
2. Go to the [Spotify developer dashboard](https://developer.spotify.com/dashboard/applications) and create an application
    - Copy the client ID and secret
    - Click "Edit Settings" and `http://localhost:8080/callback` to the "Redirect URIs" section
3. Set the `SPOTIFY_ID` and `SPOTIFY_SECRET` environment variables with the client ID and secret respectively
4. Set the `SPOTIFY_PLAYLIST_ID` environment variable with the playlist ID of the playlist you want to duplicate
    - This should just be the ID itself
    - E.g. `https://open.spotify.com/playlist/4GtQVhGjAwcHFz82UKy3Ca?si=32208c6432ca47c4` ⇒ `4GtQVhGjAwcHFz82UKy3Ca`
5. Run the program with `./playlist_duplicator`

**Note:** Not tested on Windows
