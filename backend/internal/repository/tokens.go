package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"medbratishka/internal/domain"
)

type TokensRepository struct {
	db *sql.DB
}

func NewTokensRepository(db *sql.DB) *TokensRepository {
	return &TokensRepository{db: db}
}

func (r *TokensRepository) InsertJwtToken(ctx context.Context, data domain.TokenData) error {
	query := `
		INSERT INTO auth_tokens (
			user_id, role, purpose, session_number, secret, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, role, purpose, session_number) DO UPDATE
		SET secret = $5, expires_at = $6, created_at = $7, revoked_at = NULL
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		data.ID,
		string(data.Role),
		string(data.Purpose),
		data.Number,
		data.Secret,
		data.ExpiresAt.UnixMilli(),
		time.Now().UnixMilli(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert token: %w", err)
	}

	return nil
}

func (r *TokensRepository) GetTokenNumber(ctx context.Context, id int64, role domain.Role, purpose domain.TokenPurpose) (int, error) {
	query := `
		SELECT COALESCE(MAX(session_number), 0) + 1
		FROM auth_tokens
		WHERE user_id = $1 AND role = $2 AND purpose = $3 AND revoked_at IS NULL
	`

	var number int
	err := r.db.QueryRowContext(ctx, query, id, string(role), string(purpose)).Scan(&number)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	return number, nil
}

func (r *TokensRepository) GetTokenSecret(ctx context.Context, id int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error) {
	query := `
		SELECT secret
		FROM auth_tokens
		WHERE user_id = $1 AND role = $2 AND session_number = $3 AND purpose = $4
		AND revoked_at IS NULL
		LIMIT 1
	`

	var secret string
	err := r.db.QueryRowContext(ctx, query, id, string(role), number, string(purpose)).Scan(&secret)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("token not found")
	}
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}

	return secret, nil
}

func (r *TokensRepository) DeleteTokenByNumber(ctx context.Context, id int64, role domain.Role, number int) error {
	query := `
		UPDATE auth_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND role = $3 AND session_number = $4 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now().UnixMilli(), id, string(role), number)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

func (r *TokensRepository) DeleteAllUserTokens(ctx context.Context, id int64, role domain.Role) error {
	query := `
		UPDATE auth_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND role = $3 AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now().UnixMilli(), id, string(role))
	if err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	return nil
}
