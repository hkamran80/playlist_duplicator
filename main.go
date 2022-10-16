/*
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
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"log"

	"github.com/gtuk/discordwebhook"
	"github.com/joho/godotenv"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/schollz/progressbar/v3"

	"github.com/zmb3/spotify/v2"
)

func GetPlaylistTracks(client *spotify.Client, context *context.Context, playlistId spotify.ID, bar *progressbar.ProgressBar) []spotify.ID {
	var allTracks []spotify.ID

	items, err := client.GetPlaylistItems(*context, playlistId)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Playlist has %d total tracks", items.Total)
	for page := 1; ; page++ {
		for _, item := range items.Items {
			// Check if the item is an actual track
			if item.Track.Track != nil {
				allTracks = append(allTracks, item.Track.Track.ID)
				bar.Add(1)
			}
		}

		err = client.NextPage(*context, items)
		if err == spotify.ErrNoMorePages {
			break
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	return allTracks
}

func Contains(s []spotify.ID, str spotify.ID) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func SendNotification(content string) {
	discordWebhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhookUrl != "" {
		username := "Playlist Duplicator"

		title := "Playlist Duplicator"
		colour := "1947988"
		footerText := "Playlist Duplicator is an open-source program created by H. Kamran"

		footer := discordwebhook.Footer{Text: &footerText}
		embed := discordwebhook.Embed{Title: &title, Description: &content, Color: &colour, Footer: &footer}
		message := discordwebhook.Message{Username: &username, Embeds: &[]discordwebhook.Embed{embed}}

		err := discordwebhook.SendMessage(discordWebhookUrl, message)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func checkIfEnvVarsLoaded() bool {
	spotifyId := os.Getenv("SPOTIFY_ID")
	spotifySecret := os.Getenv("SPOTIFY_SECRET")
	spotifyPlaylistId := os.Getenv("SPOTIFY_PLAYLIST_ID")
	spotifyHoldingPlaylistId := os.Getenv("SPOTIFY_HOLDING_PLAYLIST_ID")

	return spotifyId != "" && spotifySecret != "" && spotifyPlaylistId != "" && spotifyHoldingPlaylistId != ""
}

func main() {
	envVarsLoaded := checkIfEnvVarsLoaded()
	if !envVarsLoaded {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	auth := spotifyauth.New(spotifyauth.WithRedirectURL("http://localhost:8080/callback"), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopePlaylistReadPrivate, spotifyauth.ScopePlaylistReadCollaborative, spotifyauth.ScopePlaylistModifyPublic, spotifyauth.ScopePlaylistModifyPrivate))
	state := strconv.FormatInt(time.Now().Unix(), 10)
	ctx := context.Background()

	client := Authenticate(ctx, *auth, state)

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Logged in as %s (%s)", user.ID, user.DisplayName)

	// Main Playlist
	playlistMain, err := client.GetPlaylist(ctx, spotify.ID(os.Getenv("SPOTIFY_PLAYLIST_ID")))
	if err != nil {
		log.Fatal(err)
	}

	// Holding Playlist
	playlistHolding, err := client.GetPlaylist(ctx, spotify.ID(os.Getenv("SPOTIFY_HOLDING_PLAYLIST_ID")))
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	mainBar := progressbar.NewOptions(playlistMain.Tracks.Total, progressbar.OptionShowBytes(false), progressbar.OptionShowCount(), progressbar.OptionSetDescription("Loading tracks from main playlist..."),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}))
	playlistMainTracks := GetPlaylistTracks(client, &ctx, playlistMain.ID, mainBar)
	mainBar.Close()

	holdingBar := progressbar.NewOptions(playlistHolding.Tracks.Total, progressbar.OptionShowBytes(false), progressbar.OptionShowCount(), progressbar.OptionSetDescription("Loading tracks from holding playlist..."),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}))
	playlistHoldingTracks := GetPlaylistTracks(client, &ctx, playlistHolding.ID, holdingBar)
	holdingBar.Close()

	duration := time.Since(start)

	log.Printf("Loaded %d tracks in %v", len(playlistMainTracks)+len(playlistHoldingTracks), duration)

	var (
		newTracks     []spotify.ID
		maxTrackCount int
	)

	for _, trackId := range playlistMainTracks {
		if !Contains(playlistHoldingTracks, trackId) {
			newTracks = append(newTracks, trackId)
		}
	}

	if len(newTracks) != 0 {
		log.Printf("Adding %d tracks to holding playlist", len(newTracks))
		start := time.Now()

		originalNewTracksCount := len(newTracks)
		bar := progressbar.NewOptions(originalNewTracksCount, progressbar.OptionShowBytes(false), progressbar.OptionShowCount(), progressbar.OptionSetDescription("Saving tracks to holding playlist..."),
			progressbar.OptionOnCompletion(func() {
				fmt.Printf("\n")
			}))

		for len(newTracks) > 0 {
			// Spotify enforces a maximum of 100 tracks per `AddTracksToPlaylist` call
			if len(newTracks) < 100 {
				maxTrackCount = len(newTracks)
			} else {
				maxTrackCount = 99
			}

			newTracksInstance := newTracks[0:maxTrackCount]

			_, err := client.AddTracksToPlaylist(ctx, playlistHolding.ID, newTracksInstance...)
			if err != nil {
				log.Panic(err)
			}

			bar.Add(len(newTracksInstance))

			if len(newTracks) != maxTrackCount {
				newTracks = newTracks[maxTrackCount+1:]
			} else {
				newTracks = []spotify.ID{}
			}
		}

		duration := time.Since(start)
		bar.Close()

		var trackWord string
		if originalNewTracksCount != 1 {
			trackWord = "tracks"
		} else {
			trackWord = "track"
		}

		finishedMessage := fmt.Sprintf("Added %d %s to holding playlist in %v", originalNewTracksCount, trackWord, duration)
		log.Println(finishedMessage)

		SendNotification(finishedMessage)
	} else {
		log.Println("No new tracks found")

		if sendEmptyNotifications := os.Getenv("SEND_EMPTY_NOTIFICATIONS"); sendEmptyNotifications != "false" {
			SendNotification("No new tracks found")
		}
	}
}
