package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopePlaylistReadPrivate, spotify.ScopePlaylistReadCollaborative, spotify.ScopePlaylistModifyPublic, spotify.ScopePlaylistModifyPrivate)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// Create client from token URL
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Authorized!")
	ch <- &client
}

func getPlaylistTracks(client *spotify.Client, playlistId spotify.ID) []spotify.ID {
	var allTracks []spotify.ID

	tracks, err := client.GetPlaylistTracks(playlistId)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Playlist has %d total tracks", tracks.Total)
	for page := 1; ; page++ {
		for _, track := range tracks.Tracks {
			allTracks = append(allTracks, track.Track.ID)
		}

		err = client.NextPage(tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	return allTracks
}

func contains(s []spotify.ID, str spotify.ID) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func main() {
	// Start HTTP
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Authenticate by visiting the following page in your browser:", url)

	// Stand by for authentication
	client := <-ch

	// Current user data
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Logged in as %s (%s)", user.ID, user.DisplayName)

	// Main Playlist
	playlistMain, err := client.GetPlaylist(spotify.ID(os.Getenv("SPOTIFY_PLAYLIST_ID")))
	if err != nil {
		log.Fatal(err)
	}

	// Holding Playlist
	playlistHolding, err := client.GetPlaylist(spotify.ID(os.Getenv("SPOTIFY_HOLDING_PLAYLIST_ID")))
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	playlistMainTracks := getPlaylistTracks(client, playlistMain.ID)
	playlistHoldingTracks := getPlaylistTracks(client, playlistHolding.ID)
	duration := time.Since(start)

	log.Printf("Loaded %d tracks in %v", len(playlistMainTracks)+len(playlistHoldingTracks), duration)

	var (
		newTracks     []spotify.ID
		maxTrackCount int
	)

	for _, trackId := range playlistMainTracks {
		if !contains(playlistHoldingTracks, trackId) {
			newTracks = append(newTracks, trackId)
		}
	}

	if len(newTracks) != 0 {
		log.Printf("Adding %d tracks to holding playlist", len(newTracks))
		start := time.Now()

		for len(newTracks) > 0 {
			if len(newTracks) < 100 {
				maxTrackCount = len(newTracks)
			} else {
				maxTrackCount = 99
			}

			client.AddTracksToPlaylist(playlistHolding.ID, newTracks[0:maxTrackCount]...)

			if len(newTracks) != maxTrackCount {
				newTracks = append(newTracks, newTracks[maxTrackCount+1:]...)
			} else {
				newTracks = []spotify.ID{}
			}
		}

		duration := time.Since(start)
		log.Printf("Added %d tracks to holding playlist in %v", len(newTracks), duration)
	} else {
		log.Println("No new tracks found")
	}
}
