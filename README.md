# Playlist home

Manage your YouTube and Spotify playlists in one place.

**Stack:** Go · React (Vite) · PostgreSQL · Docker Compose

---

## Quick Start

### 1. Configure environment variables

```bash
cp .env.example .env
# Edit .env — fill in your Google and Spotify OAuth credentials
```

### 2. Create OAuth apps

**Google / YouTube**
1. Go to [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials)
2. Create an OAuth 2.0 Client ID (Web application)
3. Add authorised redirect URI: `http://localhost:8080/auth/google/callback`
4. Enable the **YouTube Data API v3** in your project

**Spotify**
1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create an app
3. Add redirect URI: `http://localhost:8080/auth/spotify/callback`

### 3. Start the stack

```bash
docker compose up --build
```

| Service  | URL                        |
|----------|----------------------------|
| Frontend | http://localhost:3000       |
| Backend  | http://localhost:8080       |
| Postgres | localhost:5432              |

---

## Database Schema Design

The schema is in `backend/migrations/001_initial_schema.sql`.

### Core idea: separate `users` from `oauth_accounts`

```
users (1) ──< oauth_accounts (many)
```

A **user** is a single human being.  
An **oauth_account** is one connected provider (Google or Spotify).

This allows two important flows:

**Flow A — Sign in for the first time**
1. User clicks "Login with YouTube"
2. Google returns a `sub` (stable user ID)
3. We look up `oauth_accounts` by `(provider='google', provider_user_id=sub)`
4. Not found → create a new `users` row, then an `oauth_accounts` row linked to it
5. Issue a JWT referencing the `users.id`

**Flow B — Connect the second provider**
1. Logged-in user clicks "Connect Spotify"
2. Same OAuth callback, but now we have a valid JWT in the request
3. We extract `user_id` from the JWT instead of creating a new user
4. Insert a second `oauth_accounts` row with `provider='spotify'` for the *same* `user_id`
5. The UNIQUE constraint `(user_id, provider)` ensures only one Spotify account per user

**Flow C — Login with the second provider later**
1. User clicks "Login with Spotify" from the login page (no JWT)
2. We find the existing `oauth_accounts` row by `(provider='spotify', provider_user_id=...)`
3. We load the linked `users` row — same user, regardless of which provider they used

### Tables at a glance

| Table | Purpose |
|---|---|
| `users` | One row per person |
| `oauth_accounts` | One row per (user × provider). Stores OAuth tokens |
| `sessions` | Server-side session so we can invalidate JWTs before expiry |
| `playlists` | Mirrors a playlist from YouTube/Spotify, or locally created |
| `tracks` | Individual songs/videos |
| `playlist_tracks` | Many-to-many join between playlists and tracks |

---

## Project Structure

```
playlisthome/
├── docker-compose.yml
├── .env.example
├── backend/
│   ├── Dockerfile
│   ├── .air.toml            # hot-reload config
│   ├── go.mod
│   ├── cmd/api/main.go      # entry point, router
│   ├── internal/
│   │   ├── auth/            # OAuth configs, JWT helpers
│   │   ├── db/              # connection pool
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/       # JWT auth middleware
│   │   └── models/          # domain types
│   └── migrations/
│       └── 001_initial_schema.sql
└── frontend/
    ├── Dockerfile
    ├── vite.config.js
    ├── package.json
    └── src/
        ├── main.jsx
        ├── lib/api.js
        ├── context/AuthContext.jsx
        └── pages/
            ├── LoginPage.jsx
            └── Dashboard.jsx
```

## Next Steps

- [ ] `/api/me` endpoint (return current user + connected providers)
- [ ] `/api/logout` endpoint (delete session)
- [ ] Sync playlists from YouTube (YouTube Data API v3)
- [ ] Sync playlists from Spotify (Spotify Web API)
- [ ] Create / edit playlists
- [ ] Token refresh logic (both providers)
- [ ] Encrypt OAuth tokens at rest
