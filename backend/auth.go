package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CCLJ/playlisthome/internal/auth"
	"github.com/CCLJ/playlisthome/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

// AuthHandler holds dependencies for auth-related endpoints.
type AuthHandler struct {
	db      *pgxpool.Pool
	google  *oauth2.Config
	spotify *oauth2.Config
}

func NewAuthHandler(db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		db:      db,
		google:  auth.GoogleOAuthConfig(),
		spotify: auth.SpotifyOAuthConfig(),
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ── Initiate OAuth flow ───────────────────────────────────────────────────────

// GET /auth/google/login
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, HttpOnly: true, MaxAge: 300})
	http.Redirect(w, r, h.google.AuthCodeURL(state, oauth2.AccessTypeOffline), http.StatusTemporaryRedirect)
}

// GET /auth/spotify/login
func (h *AuthHandler) SpotifyLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, HttpOnly: true, MaxAge: 300})
	http.Redirect(w, r, h.spotify.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// ── OAuth callbacks ───────────────────────────────────────────────────────────

// GET /auth/google/callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, models.ProviderGoogle, h.google, fetchGoogleUserInfo)
}

// GET /auth/spotify/callback
func (h *AuthHandler) SpotifyCallback(w http.ResponseWriter, r *http.Request) {
	h.handleCallback(w, r, models.ProviderSpotify, h.spotify, fetchSpotifyUserInfo)
}

// providerUserInfo is the normalised info we get back from each provider.
type providerUserInfo struct {
	ProviderUserID string
	Email          string
	DisplayName    string
	AvatarURL      string
}

type fetchUserInfoFn func(ctx context.Context, token *oauth2.Token) (*providerUserInfo, error)

// handleCallback is the shared logic for both OAuth callbacks.
// It:
//  1. Validates state
//  2. Exchanges code for token
//  3. Fetches user info from provider
//  4. Upserts user + oauth_account rows
//  5. Creates a session + issues JWT
func (h *AuthHandler) handleCallback(
	w http.ResponseWriter,
	r *http.Request,
	provider string,
	cfg *oauth2.Config,
	fetchInfo fetchUserInfoFn,
) {
	// 1. Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	// 2. Exchange code
	token, err := cfg.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Fetch user info
	info, err := fetchInfo(r.Context(), token)
	if err != nil {
		http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Upsert user + oauth_account
	userID, err := h.upsertOAuthAccount(r.Context(), provider, info, token)
	if err != nil {
		http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Create session + issue JWT
	sessionID, err := h.createSession(r.Context(), userID, provider)
	if err != nil {
		http.Error(w, "session error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jwt, err := auth.IssueToken(userID, sessionID)
	if err != nil {
		http.Error(w, "jwt error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the JWT to the frontend (redirect with token in query, or set cookie)
	// Using a cookie here — adjust to match your frontend auth strategy
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwt,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
	})
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}

// ── Database helpers ──────────────────────────────────────────────────────────

func (h *AuthHandler) upsertOAuthAccount(
	ctx context.Context,
	provider string,
	info *providerUserInfo,
	token *oauth2.Token,
) (userID interface{ String() string }, err error) {

	tx, err := h.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Does this provider account already exist?
	var existingUserID *string
	err = tx.QueryRow(ctx,
		`SELECT user_id::text FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`,
		provider, info.ProviderUserID,
	).Scan(&existingUserID)

	var uid string

	if existingUserID == nil {
		// First time — create a new user
		err = tx.QueryRow(ctx,
			`INSERT INTO users (display_name, avatar_url) VALUES ($1, $2) RETURNING id::text`,
			info.DisplayName, info.AvatarURL,
		).Scan(&uid)
		if err != nil {
			return nil, fmt.Errorf("insert user: %w", err)
		}
	} else {
		uid = *existingUserID
	}

	// Upsert the oauth_account row
	_, err = tx.Exec(ctx, `
		INSERT INTO oauth_accounts
			(user_id, provider, provider_user_id, email, display_name, avatar_url,
			 access_token, refresh_token, token_expires_at, scopes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (provider, provider_user_id) DO UPDATE SET
			access_token    = EXCLUDED.access_token,
			refresh_token   = COALESCE(EXCLUDED.refresh_token, oauth_accounts.refresh_token),
			token_expires_at = EXCLUDED.token_expires_at,
			display_name    = EXCLUDED.display_name,
			avatar_url      = EXCLUDED.avatar_url,
			updated_at      = NOW()
	`,
		uid, provider, info.ProviderUserID,
		info.Email, info.DisplayName, info.AvatarURL,
		token.AccessToken, token.RefreshToken, token.Expiry,
		"", // scopes stored separately if needed
	)
	if err != nil {
		return nil, fmt.Errorf("upsert oauth_account: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Return as a simple string-wrapper so we don't import uuid here
	return stringID(uid), nil
}

type stringID string

func (s stringID) String() string { return string(s) }

func (h *AuthHandler) createSession(ctx context.Context, userID interface{ String() string }, provider string) (interface{ String() string }, error) {
	var sessionID string
	err := h.db.QueryRow(ctx,
		`INSERT INTO sessions (user_id, provider, expires_at)
		 VALUES ($1, $2, $3) RETURNING id::text`,
		userID.String(), provider, time.Now().Add(24*time.Hour),
	).Scan(&sessionID)
	if err != nil {
		return nil, err
	}
	return stringID(sessionID), nil
}

// ── Provider user-info fetchers ───────────────────────────────────────────────

func fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*providerUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &providerUserInfo{
		ProviderUserID: result.Sub,
		Email:          result.Email,
		DisplayName:    result.Name,
		AvatarURL:      result.Picture,
	}, nil
}

func fetchSpotifyUserInfo(ctx context.Context, token *oauth2.Token) (*providerUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Images      []struct {
			URL string `json:"url"`
		} `json:"images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	avatarURL := ""
	if len(result.Images) > 0 {
		avatarURL = result.Images[0].URL
	}

	return &providerUserInfo{
		ProviderUserID: result.ID,
		Email:          result.Email,
		DisplayName:    result.DisplayName,
		AvatarURL:      avatarURL,
	}, nil
}
