package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/spotify"
)

// GoogleOAuthConfig returns the OAuth2 config for Google / YouTube.
// Scopes requested:
//   - openid, email, profile  — for user identity
//   - youtube.readonly        — read playlists & videos
//   - youtube                 — create / modify playlists
func GoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

// SpotifyOAuthConfig returns the OAuth2 config for Spotify.
// Scopes requested:
//   - user-read-private, user-read-email  — identity
//   - playlist-read-private               — read private playlists
//   - playlist-modify-public/private      — create & edit playlists
func SpotifyOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SPOTIFY_REDIRECT_URL"),
		Scopes: []string{
			"user-read-private",
			"user-read-email",
			"playlist-read-private",
			"playlist-read-collaborative",
			"playlist-modify-public",
			"playlist-modify-private",
		},
		Endpoint: spotify.Endpoint,
	}
}
