-- ============================================================
-- Migration: 001_initial_schema.sql
-- Run automatically by postgres on first container start
-- ============================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- USERS
-- One row per human being. Created the first time they log in
-- via either OAuth provider.
-- ============================================================
CREATE TABLE users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    display_name TEXT,
    avatar_url  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- OAUTH ACCOUNTS
-- Each row represents one connected OAuth provider account.
-- A user can have ONE google account and ONE spotify account.
-- This table is the key to supporting:
--   1. Login with Google   → creates / finds user, upserts this row
--   2. Login with Spotify  → same
--   3. "Connect Spotify"   → inserts a second row for same user_id
-- ============================================================
CREATE TABLE oauth_accounts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 'google' | 'spotify'
    provider        TEXT        NOT NULL,

    -- The stable ID returned by the provider (Google sub, Spotify user id)
    provider_user_id TEXT       NOT NULL,

    -- Human-readable info from the provider (denormalised for convenience)
    email           TEXT,
    display_name    TEXT,
    avatar_url      TEXT,

    -- OAuth tokens — store encrypted at rest in production!
    access_token    TEXT        NOT NULL,
    refresh_token   TEXT,
    token_expires_at TIMESTAMPTZ,

    -- Scopes that were granted (space-separated, as returned by provider)
    scopes          TEXT,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One account per provider per user
    UNIQUE (user_id, provider),

    -- A provider account can only belong to one user
    UNIQUE (provider, provider_user_id)
);

-- ============================================================
-- SESSIONS
-- Server-side sessions. The JWT references session_id so we
-- can invalidate without waiting for JWT expiry.
-- ============================================================
CREATE TABLE sessions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Which provider was used to create this session
    provider    TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ============================================================
-- PLAYLISTS
-- Mirrors a playlist that exists on a provider.
-- source = 'youtube' | 'spotify' | 'local' (manually created)
-- ============================================================
CREATE TABLE playlists (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- The external provider playlist ID (null for local-only playlists)
    provider            TEXT,               -- 'youtube' | 'spotify' | NULL
    provider_playlist_id TEXT,

    title               TEXT        NOT NULL,
    description         TEXT,
    thumbnail_url       TEXT,
    is_public           BOOLEAN     NOT NULL DEFAULT false,

    -- When we last synced this playlist from the provider
    last_synced_at      TIMESTAMPTZ,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, provider, provider_playlist_id)
);

CREATE INDEX idx_playlists_user_id ON playlists(user_id);

-- ============================================================
-- TRACKS
-- Mirrors a track/video that exists on a provider.
-- ============================================================
CREATE TABLE tracks (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    provider            TEXT        NOT NULL,   -- 'youtube' | 'spotify'
    provider_track_id   TEXT        NOT NULL,

    title               TEXT        NOT NULL,
    artist              TEXT,
    album               TEXT,
    duration_ms         INTEGER,
    thumbnail_url       TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (provider, provider_track_id)
);

-- ============================================================
-- PLAYLIST TRACKS  (join table)
-- ============================================================
CREATE TABLE playlist_tracks (
    id              UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    playlist_id     UUID    NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id        UUID    NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    position        INTEGER NOT NULL DEFAULT 0,
    added_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (playlist_id, track_id)
);

CREATE INDEX idx_playlist_tracks_playlist_id ON playlist_tracks(playlist_id);

-- ============================================================
-- updated_at trigger (reusable)
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_oauth_accounts_updated_at
    BEFORE UPDATE ON oauth_accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_playlists_updated_at
    BEFORE UPDATE ON playlists
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_tracks_updated_at
    BEFORE UPDATE ON tracks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
