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

type TwoFactorRepository interface {
	UpsertTOTPSettingsTX(ctx context.Context, tx transaction.Transaction, settings *domain.TwoFactorSettings) error
	GetTOTPSettingsTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (*domain.TwoFactorSettings, error)
	GetTOTPSettings(ctx context.Context, userID int64, role domain.Role) (*domain.TwoFactorSettings, error)
	EnableTOTPTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, confirmedAt int64) error
	DisableTOTPTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, disabledAt int64) error

	CreateChallengeTX(ctx context.Context, tx transaction.Transaction, challenge *domain.AuthChallenge) error
	GetChallengeTX(ctx context.Context, tx transaction.Transaction, id string) (*domain.AuthChallenge, error)
	ConsumeChallengeTX(ctx context.Context, tx transaction.Transaction, id string, consumedAt int64) error
	IncrementChallengeFailuresTX(ctx context.Context, tx transaction.Transaction, id string) error

	ReplaceRecoveryCodesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, hashes []string, createdAt int64) error
	GetUnusedRecoveryCodesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (map[int64]string, error)
	UseRecoveryCodeTX(ctx context.Context, tx transaction.Transaction, id int64, usedAt int64) error

	CreateTrustedDeviceTX(ctx context.Context, tx transaction.Transaction, device *domain.TrustedDevice) error
	GetTrustedDeviceByHash(ctx context.Context, tokenHash string) (*domain.TrustedDevice, error)
	TouchTrustedDevice(ctx context.Context, id int64, touchedAt int64) error
	RevokeTrustedDevicesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, revokedAt int64) error

	CreateAuditTX(ctx context.Context, tx transaction.Transaction, userID *int64, role *domain.Role, eventType, ip, userAgent, metadata string, createdAt int64) error
	CreateAudit(ctx context.Context, userID *int64, role *domain.Role, eventType, ip, userAgent, metadata string, createdAt int64) error
}

type pgTwoFactorRepository struct {
	client *db.PostgresClient
}

func NewTwoFactorRepository(client *db.PostgresClient) TwoFactorRepository {
	return &pgTwoFactorRepository{client: client}
}

func (r *pgTwoFactorRepository) UpsertTOTPSettingsTX(ctx context.Context, tx transaction.Transaction, s *domain.TwoFactorSettings) error {
	query := `
		INSERT INTO user_totp_settings (user_id, role, secret_ciphertext, enabled, confirmed_at, disabled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, role) DO UPDATE
		SET secret_ciphertext = EXCLUDED.secret_ciphertext,
		    enabled = EXCLUDED.enabled,
		    confirmed_at = EXCLUDED.confirmed_at,
		    disabled_at = EXCLUDED.disabled_at,
		    updated_at = EXCLUDED.updated_at
	`
	_, err := tx.Txm().ExecContext(ctx, query, s.UserID, string(s.Role), s.SecretCiphertext, s.Enabled, s.ConfirmedAt, s.DisabledAt, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert totp settings: %w", err)
	}
	return nil
}

func (r *pgTwoFactorRepository) GetTOTPSettingsTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (*domain.TwoFactorSettings, error) {
	query := `
		SELECT user_id, role, secret_ciphertext, enabled, confirmed_at, disabled_at, created_at, updated_at
		FROM user_totp_settings
		WHERE user_id = $1 AND role = $2
	`
	var row struct {
		UserID           int64         `db:"user_id"`
		Role             string        `db:"role"`
		SecretCiphertext string        `db:"secret_ciphertext"`
		Enabled          bool          `db:"enabled"`
		ConfirmedAt      sql.NullInt64 `db:"confirmed_at"`
		DisabledAt       sql.NullInt64 `db:"disabled_at"`
		CreatedAt        int64         `db:"created_at"`
		UpdatedAt        int64         `db:"updated_at"`
	}
	if err := tx.Txm().GetContext(ctx, &row, query, userID, string(role)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get totp settings: %w", err)
	}
	return &domain.TwoFactorSettings{
		UserID: row.UserID, Role: domain.Role(row.Role), SecretCiphertext: row.SecretCiphertext, Enabled: row.Enabled,
		ConfirmedAt: nullInt64Ptr(row.ConfirmedAt), DisabledAt: nullInt64Ptr(row.DisabledAt), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *pgTwoFactorRepository) GetTOTPSettings(ctx context.Context, userID int64, role domain.Role) (*domain.TwoFactorSettings, error) {
	tx, err := r.client.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	settings, err := r.GetTOTPSettingsTX(ctx, &transaction.Tx{Tx: tx}, userID, role)
	if err != nil {
		return nil, err
	}
	_ = tx.Commit()
	return settings, nil
}

func (r *pgTwoFactorRepository) EnableTOTPTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, confirmedAt int64) error {
	_, err := tx.Txm().ExecContext(ctx, `UPDATE user_totp_settings SET enabled = TRUE, confirmed_at = $1, disabled_at = NULL, updated_at = $1 WHERE user_id = $2 AND role = $3`, confirmedAt, userID, string(role))
	return err
}

func (r *pgTwoFactorRepository) DisableTOTPTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, disabledAt int64) error {
	_, err := tx.Txm().ExecContext(ctx, `UPDATE user_totp_settings SET enabled = FALSE, disabled_at = $1, updated_at = $1 WHERE user_id = $2 AND role = $3`, disabledAt, userID, string(role))
	return err
}

func (r *pgTwoFactorRepository) CreateChallengeTX(ctx context.Context, tx transaction.Transaction, c *domain.AuthChallenge) error {
	_, err := tx.Txm().ExecContext(ctx, `INSERT INTO auth_challenges (id, user_id, role, purpose, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		c.ID, c.UserID, string(c.Role), c.Purpose, c.ExpiresAt, c.CreatedAt)
	return err
}

func (r *pgTwoFactorRepository) GetChallengeTX(ctx context.Context, tx transaction.Transaction, id string) (*domain.AuthChallenge, error) {
	var row struct {
		ID             string        `db:"id"`
		UserID         int64         `db:"user_id"`
		Role           string        `db:"role"`
		Purpose        string        `db:"purpose"`
		FailedAttempts int           `db:"failed_attempts"`
		ExpiresAt      int64         `db:"expires_at"`
		ConsumedAt     sql.NullInt64 `db:"consumed_at"`
		CreatedAt      int64         `db:"created_at"`
	}
	if err := tx.Txm().GetContext(ctx, &row, `SELECT id, user_id, role, purpose, failed_attempts, expires_at, consumed_at, created_at FROM auth_challenges WHERE id = $1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &domain.AuthChallenge{ID: row.ID, UserID: row.UserID, Role: domain.Role(row.Role), Purpose: row.Purpose, FailedAttempts: row.FailedAttempts, ExpiresAt: row.ExpiresAt, ConsumedAt: nullInt64Ptr(row.ConsumedAt), CreatedAt: row.CreatedAt}, nil
}

func (r *pgTwoFactorRepository) ConsumeChallengeTX(ctx context.Context, tx transaction.Transaction, id string, consumedAt int64) error {
	_, err := tx.Txm().ExecContext(ctx, `UPDATE auth_challenges SET consumed_at = $1 WHERE id = $2 AND consumed_at IS NULL`, consumedAt, id)
	return err
}

func (r *pgTwoFactorRepository) IncrementChallengeFailuresTX(ctx context.Context, tx transaction.Transaction, id string) error {
	_, err := tx.Txm().ExecContext(ctx, `UPDATE auth_challenges SET failed_attempts = failed_attempts + 1 WHERE id = $1 AND consumed_at IS NULL`, id)
	return err
}

func (r *pgTwoFactorRepository) ReplaceRecoveryCodesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, hashes []string, createdAt int64) error {
	if _, err := tx.Txm().ExecContext(ctx, `UPDATE recovery_codes SET used_at = $1 WHERE user_id = $2 AND role = $3 AND used_at IS NULL`, createdAt, userID, string(role)); err != nil {
		return err
	}
	for _, h := range hashes {
		if _, err := tx.Txm().ExecContext(ctx, `INSERT INTO recovery_codes (user_id, role, code_hash, created_at) VALUES ($1, $2, $3, $4)`, userID, string(role), h, createdAt); err != nil {
			return err
		}
	}
	return nil
}

func (r *pgTwoFactorRepository) GetUnusedRecoveryCodesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (map[int64]string, error) {
	rows, err := tx.Txm().QueryxContext(ctx, `SELECT id, code_hash FROM recovery_codes WHERE user_id = $1 AND role = $2 AND used_at IS NULL`, userID, string(role))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]string{}
	for rows.Next() {
		var id int64
		var hash string
		if err := rows.Scan(&id, &hash); err != nil {
			return nil, err
		}
		res[id] = hash
	}
	return res, rows.Err()
}

func (r *pgTwoFactorRepository) UseRecoveryCodeTX(ctx context.Context, tx transaction.Transaction, id int64, usedAt int64) error {
	_, err := tx.Txm().ExecContext(ctx, `UPDATE recovery_codes SET used_at = $1 WHERE id = $2 AND used_at IS NULL`, usedAt, id)
	return err
}

func (r *pgTwoFactorRepository) CreateTrustedDeviceTX(ctx context.Context, tx transaction.Transaction, d *domain.TrustedDevice) error {
	_, err := tx.Txm().ExecContext(ctx, `INSERT INTO trusted_devices (user_id, role, token_hash, device_name, user_agent_hash, expires_at, last_used_at, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		d.UserID, string(d.Role), d.TokenHash, d.DeviceName, d.UserAgentHash, d.ExpiresAt, d.LastUsedAt, d.CreatedAt)
	return err
}

func (r *pgTwoFactorRepository) GetTrustedDeviceByHash(ctx context.Context, tokenHash string) (*domain.TrustedDevice, error) {
	var row struct {
		ID            int64          `db:"id"`
		UserID        int64          `db:"user_id"`
		Role          string         `db:"role"`
		TokenHash     string         `db:"token_hash"`
		DeviceName    sql.NullString `db:"device_name"`
		UserAgentHash sql.NullString `db:"user_agent_hash"`
		ExpiresAt     int64          `db:"expires_at"`
		LastUsedAt    sql.NullInt64  `db:"last_used_at"`
		RevokedAt     sql.NullInt64  `db:"revoked_at"`
		CreatedAt     int64          `db:"created_at"`
	}
	if err := r.client.DB.GetContext(ctx, &row, `SELECT id, user_id, role, token_hash, device_name, user_agent_hash, expires_at, last_used_at, revoked_at, created_at FROM trusted_devices WHERE token_hash = $1`, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &domain.TrustedDevice{ID: row.ID, UserID: row.UserID, Role: domain.Role(row.Role), TokenHash: row.TokenHash, DeviceName: row.DeviceName.String, UserAgentHash: row.UserAgentHash.String, ExpiresAt: row.ExpiresAt, LastUsedAt: nullInt64Ptr(row.LastUsedAt), RevokedAt: nullInt64Ptr(row.RevokedAt), CreatedAt: row.CreatedAt}, nil
}

func (r *pgTwoFactorRepository) TouchTrustedDevice(ctx context.Context, id int64, touchedAt int64) error {
	_, err := r.client.DB.ExecContext(ctx, `UPDATE trusted_devices SET last_used_at = $1 WHERE id = $2 AND revoked_at IS NULL`, touchedAt, id)
	return err
}

func (r *pgTwoFactorRepository) RevokeTrustedDevicesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, revokedAt int64) error {
	_, err := tx.Txm().ExecContext(ctx, `UPDATE trusted_devices SET revoked_at = $1 WHERE user_id = $2 AND role = $3 AND revoked_at IS NULL`, revokedAt, userID, string(role))
	return err
}

func (r *pgTwoFactorRepository) CreateAuditTX(ctx context.Context, tx transaction.Transaction, userID *int64, role *domain.Role, eventType, ip, userAgent, metadata string, createdAt int64) error {
	var roleVal *string
	if role != nil {
		s := string(*role)
		roleVal = &s
	}
	_, err := tx.Txm().ExecContext(ctx, `INSERT INTO auth_audit_log (user_id, role, event_type, ip, user_agent, metadata, created_at) VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::jsonb, $7)`,
		userID, roleVal, eventType, ip, userAgent, metadata, createdAt)
	return err
}

func (r *pgTwoFactorRepository) CreateAudit(ctx context.Context, userID *int64, role *domain.Role, eventType, ip, userAgent, metadata string, createdAt int64) error {
	_, err := r.client.DB.ExecContext(ctx, `INSERT INTO auth_audit_log (user_id, role, event_type, ip, user_agent, metadata, created_at) VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::jsonb, $7)`,
		userID, roleStringPtr(role), eventType, ip, userAgent, metadata, createdAt)
	return err
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func roleStringPtr(role *domain.Role) *string {
	if role == nil {
		return nil
	}
	s := string(*role)
	return &s
}
