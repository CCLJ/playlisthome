package models

import (
	"time"

	"github.com/google/uuid"
)

// Provider constants — used everywhere provider strings appear.
const (
	ProviderGoogle  = "google"
	ProviderSpotify = "spotify"
)

// ──────────────────────────────────────────────────────────────────────────────
// User
// ──────────────────────────────────────────────────────────────────────────────

type User struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	DisplayName *string   `db:"display_name" json:"display_name,omitempty"`
	AvatarURL   *string   `db:"avatar_url"   json:"avatar_url,omitempty"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// OAuthAccount
// ──────────────────────────────────────────────────────────────────────────────

type OAuthAccount struct {
	ID             uuid.UUID  `db:"id"               json:"id"`
	UserID         uuid.UUID  `db:"user_id"           json:"user_id"`
	Provider       string     `db:"provider"          json:"provider"`
	ProviderUserID string     `db:"provider_user_id"  json:"provider_user_id"`
	Email          *string    `db:"email"             json:"email,omitempty"`
	DisplayName    *string    `db:"display_name"      json:"display_name,omitempty"`
	AvatarURL      *string    `db:"avatar_url"        json:"avatar_url,omitempty"`
	AccessToken    string     `db:"access_token"      json:"-"` // never expose in JSON
	RefreshToken   *string    `db:"refresh_token"     json:"-"`
	TokenExpiresAt *time.Time `db:"token_expires_at"  json:"-"`
	Scopes         *string    `db:"scopes"            json:"scopes,omitempty"`
	CreatedAt      time.Time  `db:"created_at"        json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"        json:"updated_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Session
// ──────────────────────────────────────────────────────────────────────────────

type Session struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	Provider  string    `db:"provider"   json:"provider"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Playlist
// ──────────────────────────────────────────────────────────────────────────────

type Playlist struct {
	ID                 uuid.UUID  `db:"id"                   json:"id"`
	UserID             uuid.UUID  `db:"user_id"              json:"user_id"`
	Provider           *string    `db:"provider"             json:"provider,omitempty"`
	ProviderPlaylistID *string    `db:"provider_playlist_id" json:"provider_playlist_id,omitempty"`
	Title              string     `db:"title"                json:"title"`
	Description        *string    `db:"description"          json:"description,omitempty"`
	ThumbnailURL       *string    `db:"thumbnail_url"        json:"thumbnail_url,omitempty"`
	IsPublic           bool       `db:"is_public"            json:"is_public"`
	LastSyncedAt       *time.Time `db:"last_synced_at"       json:"last_synced_at,omitempty"`
	CreatedAt          time.Time  `db:"created_at"           json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"           json:"updated_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Track
// ──────────────────────────────────────────────────────────────────────────────

type Track struct {
	ID             uuid.UUID `db:"id"               json:"id"`
	Provider       string    `db:"provider"         json:"provider"`
	ProviderTrackID string   `db:"provider_track_id" json:"provider_track_id"`
	Title          string    `db:"title"            json:"title"`
	Artist         *string   `db:"artist"           json:"artist,omitempty"`
	Album          *string   `db:"album"            json:"album,omitempty"`
	DurationMs     *int      `db:"duration_ms"      json:"duration_ms,omitempty"`
	ThumbnailURL   *string   `db:"thumbnail_url"    json:"thumbnail_url,omitempty"`
	CreatedAt      time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"       json:"updated_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// PlaylistTrack  (join)
// ──────────────────────────────────────────────────────────────────────────────

type PlaylistTrack struct {
	ID         uuid.UUID `db:"id"          json:"id"`
	PlaylistID uuid.UUID `db:"playlist_id" json:"playlist_id"`
	TrackID    uuid.UUID `db:"track_id"    json:"track_id"`
	Position   int       `db:"position"    json:"position"`
	AddedAt    time.Time `db:"added_at"    json:"added_at"`
}

// TrackWithPosition is returned when listing tracks within a playlist.
type TrackWithPosition struct {
	Track
	Position int       `json:"position"`
	AddedAt  time.Time `json:"added_at"`
}
