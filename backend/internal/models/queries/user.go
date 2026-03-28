package queries

import (
	"context"
	"fmt"

	"github.com/CCLJ/playlisthome/internal/auth"
	"github.com/CCLJ/playlisthome/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetUserFromClaims(claims *auth.Claims, db *pgxpool.Pool, ctx context.Context) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(ctx, `SELECT id::text, display_name, avatar_url FROM users WHERE id = $1`, claims.UserID).Scan(
		&user.ID,
		&user.DisplayName,
		&user.AvatarURL,
	)
	if err != nil {
		return nil, fmt.Errorf("GetUserFromClaims: %w", err)
	}

	return user, nil
}

func GetOAuthProvidersFromUser(userId string, db *pgxpool.Pool, ctx context.Context) ([]string, error) {
	rows, err := db.Query(
		ctx,
		`SELECT provider FROM oauth_accounts WHERE user_id = $1`,
		userId,
	)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var providers []string
	for rows.Next() {
		var provider string
		if err := rows.Scan(&provider); err != nil {
			continue
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func DeleteUserSessionFromClaims(claims *auth.Claims, db *pgxpool.Pool, ctx context.Context) error {
	_, err := db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, claims.SessionID)
	if err != nil {
		return fmt.Errorf("Error deleting session with ID %s: %w", claims.SessionID, err)
	}

	return nil
}
