package handlers

import (
	"net/http"

	"github.com/CCLJ/playlisthome/internal/auth"
	"github.com/CCLJ/playlisthome/internal/middleware"
	"github.com/CCLJ/playlisthome/internal/models/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MeHandler struct {
	db *pgxpool.Pool
}

type meResponse struct {
	ID                 string   `json:"id"`
	DisplayName        *string  `json:"display_name,omitempty"`
	AvatarURL          *string  `json:"avatar_url,omitempty"`
	ConnectedProviders []string `json:"connected_providers"`
}

func NewMeHandler(db *pgxpool.Pool) *MeHandler {
	return &MeHandler{db: db}
}

func (h *MeHandler) Me(w http.ResponseWriter, r *http.Request) {
	// get jwt token claims from context
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// get user using claims userId
	user, err := queries.GetUserFromClaims(claims, h.db, r.Context())
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	providers, err := queries.GetOAuthProvidersFromUser(user.ID.String(), h.db, r.Context())
	if err != nil {
		http.Error(w, "Providers not found", http.StatusInternalServerError)
		return
	}

	// build and return response
	resp := meResponse{
		ID:                 user.ID.String(),
		DisplayName:        user.DisplayName,
		AvatarURL:          user.AvatarURL,
		ConnectedProviders: providers,
	}

	writeJSON(w, http.StatusOK, resp)
}
