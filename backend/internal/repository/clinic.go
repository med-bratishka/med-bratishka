package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"medbratishka/internal/db"
	"medbratishka/internal/repository/models"
	"medbratishka/internal/repository/transaction"
)

type ClinicRepository interface {
	CreateClinicTX(ctx context.Context, tx transaction.Transaction, name, description, address, phone, email string, createdAt, updatedAt int64) (int64, error)
	AddClinicAdminTX(ctx context.Context, tx transaction.Transaction, clinicID, userID, createdAt int64) error
	GetClinicByIDTX(ctx context.Context, tx transaction.Transaction, clinicID int64) (map[string]interface{}, error)
	IsClinicAdminTX(ctx context.Context, tx transaction.Transaction, clinicID, userID int64) (bool, error)
}

type pgClinicRepository struct {
	client *db.PostgresClient
}

func NewClinicRepository(client *db.PostgresClient) ClinicRepository {
	return &pgClinicRepository{client: client}
}

func (r *pgClinicRepository) CreateClinicTX(ctx context.Context, tx transaction.Transaction, name, description, address, phone, email string, createdAt, updatedAt int64) (int64, error) {
	query := `
		INSERT INTO clinics (name, description, address, phone, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var clinicID int64
	err := tx.Txm().QueryRowContext(
		ctx,
		query,
		name,
		description,
		address,
		phone,
		email,
		createdAt,
		updatedAt,
	).Scan(&clinicID)

	if err != nil {
		return 0, fmt.Errorf("failed to create clinic: %w", err)
	}

	return clinicID, nil
}

func (r *pgClinicRepository) AddClinicAdminTX(ctx context.Context, tx transaction.Transaction, clinicID, userID, createdAt int64) error {
	query := `
		INSERT INTO clinic_admins (clinic_id, user_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	_, err := tx.Txm().ExecContext(ctx, query, clinicID, userID, createdAt)
	if err != nil {
		return fmt.Errorf("failed to add clinic admin: %w", err)
	}

	return nil
}

func (r *pgClinicRepository) GetClinicByIDTX(ctx context.Context, tx transaction.Transaction, clinicID int64) (map[string]interface{}, error) {
	query := `
		SELECT id, name, description, address, phone, email, created_at, updated_at
		FROM clinics
		WHERE id = $1 AND deleted_at IS NULL
	`

	var clinic models.Clinic
	err := tx.Txm().GetContext(ctx, &clinic, query, clinicID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("clinic not found")
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	var description, address, phone, email string
	if clinic.Description != nil {
		description = *clinic.Description
	}
	if clinic.Address != nil {
		address = *clinic.Address
	}
	if clinic.Phone != nil {
		phone = *clinic.Phone
	}
	if clinic.Email != nil {
		email = *clinic.Email
	}

	return map[string]interface{}{
		"id":          clinic.ID,
		"name":        clinic.Name,
		"description": description,
		"address":     address,
		"phone":       phone,
		"email":       email,
		"created_at":  clinic.CreatedAt,
		"updated_at":  clinic.UpdatedAt,
	}, nil
}

func (r *pgClinicRepository) IsClinicAdminTX(ctx context.Context, tx transaction.Transaction, clinicID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM clinic_admins
			WHERE clinic_id = $1 AND user_id = $2
		)
	`

	var isAdmin bool
	if err := tx.Txm().GetContext(ctx, &isAdmin, query, clinicID, userID); err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}

	return isAdmin, nil
}
