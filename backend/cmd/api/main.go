package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/CCLJ/playlisthome/internal/db"
	"github.com/CCLJ/playlisthome/internal/handlers"
	"github.com/CCLJ/playlisthome/internal/middleware"
)

func main() {
	// Load .env in development (ignored in production if file doesn't exist)
	_ = godotenv.Load()

	ctx := context.Background()

	// Connect to Postgres
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	// Wire up handlers
	authHandler := handlers.NewAuthHandler(pool)
	meHandler := *handlers.NewMeHandler(pool)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// ── Public routes ──────────────────────────────────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// OAuth flows
	r.Get("/auth/google/login", authHandler.GoogleLogin)
	r.Get("/auth/google/callback", authHandler.GoogleCallback)
	r.Get("/auth/spotify/login", authHandler.SpotifyLogin)
	r.Get("/auth/spotify/callback", authHandler.SpotifyCallback)

	// ── Protected routes ───────────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate)

		r.Get("/api/me", meHandler.Me)
		r.Post("/api/logout", meHandler.Logout)

		// Connect a second provider to an existing account
		r.Get("/auth/google/connect", authHandler.GoogleLogin)   // reuses same flow
		r.Get("/auth/spotify/connect", authHandler.SpotifyLogin) // reuses same flow

		// Playlists (to be implemented)
		r.Get("/api/playlists", handlers.NotImplemented)
		r.Post("/api/playlists", handlers.NotImplemented)
		r.Get("/api/playlists/{id}", handlers.NotImplemented)
		r.Delete("/api/playlists/{id}", handlers.NotImplemented)

		// Sync from providers (to be implemented)
		r.Post("/api/sync/youtube", handlers.NotImplemented)
		r.Post("/api/sync/spotify", handlers.NotImplemented)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
