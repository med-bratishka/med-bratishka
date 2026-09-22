package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"medbratishka/internal/db"
	"medbratishka/internal/domain"
	"medbratishka/internal/repository/transaction"
)

// SessionsRepository интерфейс для работы с сессиями
type SessionsRepository interface {
	// Методы с транзакциями
	CreateTX(ctx context.Context, tx transaction.Transaction, session *domain.Session) error
	GetNextNumberTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (int, error)
	GetSecretTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error)
	DeleteByNumberTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, number int, revokedAt int64) error
	DeleteAllByUserTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, revokedAt int64) error

	// Методы без транзакций
	Create(ctx context.Context, session *domain.Session) error
	GetNextNumber(ctx context.Context, userID int64, role domain.Role) (int, error)
	GetSecret(ctx context.Context, userID int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error)
	DeleteByNumber(ctx context.Context, userID int64, role domain.Role, number int, revokedAt int64) error
	DeleteAllByUser(ctx context.Context, userID int64, role domain.Role, revokedAt int64) error
}

type pgSessionsRepository struct {
	client *db.PostgresClient
}

func NewSessionsRepository(client *db.PostgresClient) SessionsRepository {
	return &pgSessionsRepository{client: client}
}

// CreateTX создает новую сессию в транзакции
func (r *pgSessionsRepository) CreateTX(ctx context.Context, tx transaction.Transaction, session *domain.Session) error {
	query := `
		INSERT INTO auth_tokens (
			user_id, role, purpose, session_number, secret, expires_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, role, purpose, session_number) DO UPDATE
		SET secret = $5, expires_at = $6, created_at = $7, revoked_at = NULL
	`

	_, err := tx.Txm().ExecContext(ctx,
		query,
		session.UserID,
		string(session.Role),
		string(session.Purpose),
		session.Number,
		session.Secret,
		session.ExpiresAt.UnixMilli(),
		session.CreatedAt.UnixMilli(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	return nil
}

// GetNextNumberTX получает следующий номер сессии в транзакции
func (r *pgSessionsRepository) GetNextNumberTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (int, error) {
	query := `
		SELECT COALESCE(MAX(session_number), 0) + 1
		FROM auth_tokens
		WHERE user_id = $1 AND role = $2 AND revoked_at IS NULL
	`

	var number int
	err := tx.Txm().QueryRowContext(ctx, query, userID, string(role)).Scan(&number)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	return number, nil
}

// GetSecretTX получает секрет сессии в транзакции
func (r *pgSessionsRepository) GetSecretTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error) {
	query := `
		SELECT secret
		FROM auth_tokens
		WHERE user_id = $1 AND role = $2 AND session_number = $3 AND purpose = $4
		AND revoked_at IS NULL
		LIMIT 1
	`

	var secret string
	err := tx.Txm().QueryRowContext(ctx, query, userID, string(role), number, string(purpose)).Scan(&secret)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}

	return secret, nil
}

// DeleteByNumberTX удаляет сессию по номеру в транзакции (мягкое удаление)
func (r *pgSessionsRepository) DeleteByNumberTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, number int, revokedAt int64) error {
	query := `
		UPDATE auth_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND role = $3 AND session_number = $4 AND revoked_at IS NULL
	`

	_, err := tx.Txm().ExecContext(ctx, query, revokedAt, userID, string(role), number)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

// DeleteAllByUserTX удаляет все сессии пользователя в транзакции (мягкое удаление)
func (r *pgSessionsRepository) DeleteAllByUserTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, revokedAt int64) error {
	query := `
		UPDATE auth_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND role = $3 AND revoked_at IS NULL
	`

	_, err := tx.Txm().ExecContext(ctx, query, revokedAt, userID, string(role))
	if err != nil {
		return fmt.Errorf("failed to revoke all sessions: %w", err)
	}

	return nil
}

// Create создает новую сессию (без транзакции)
func (r *pgSessionsRepository) Create(ctx context.Context, session *domain.Session) error {
	tx, err := r.client.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	err = r.CreateTX(ctx, &transaction.Tx{Tx: tx}, session)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetNextNumber получает следующий номер сессии (без транзакции)
func (r *pgSessionsRepository) GetNextNumber(ctx context.Context, userID int64, role domain.Role) (int, error) {
	tx, err := r.client.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	number, err := r.GetNextNumberTX(ctx, &transaction.Tx{Tx: tx}, userID, role)
	if err != nil {
		return 0, err
	}

	_ = tx.Commit()
	return number, nil
}

// GetSecret получает секрет сессии (без транзакции)
func (r *pgSessionsRepository) GetSecret(ctx context.Context, userID int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error) {
	tx, err := r.client.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	secret, err := r.GetSecretTX(ctx, &transaction.Tx{Tx: tx}, userID, role, number, purpose)
	if err != nil {
		return "", err
	}

	_ = tx.Commit()
	return secret, nil
}

// DeleteByNumber удаляет сессию по номеру (без транзакции)
func (r *pgSessionsRepository) DeleteByNumber(ctx context.Context, userID int64, role domain.Role, number int, revokedAt int64) error {
	tx, err := r.client.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	err = r.DeleteByNumberTX(ctx, &transaction.Tx{Tx: tx}, userID, role, number, revokedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteAllByUser удаляет все сессии пользователя (без транзакции)
func (r *pgSessionsRepository) DeleteAllByUser(ctx context.Context, userID int64, role domain.Role, revokedAt int64) error {
	tx, err := r.client.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	err = r.DeleteAllByUserTX(ctx, &transaction.Tx{Tx: tx}, userID, role, revokedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}
