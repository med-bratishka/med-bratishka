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

type PatientRepository interface {
	CreatePatientProfileTX(ctx context.Context, tx transaction.Transaction, userID int64, birthDate int64, gender string, createdAt, updatedAt int64) error
	AddPatientToDoctorTX(ctx context.Context, tx transaction.Transaction, doctorID, patientID, createdAt int64) error
	RemovePatientFromDoctorTX(ctx context.Context, tx transaction.Transaction, doctorID, patientID, deletedAt int64) error
	GetPatientProfileTX(ctx context.Context, tx transaction.Transaction, userID int64) (map[string]interface{}, error)
	IsPatientLinkedToDoctorTX(ctx context.Context, tx transaction.Transaction, patientID, doctorID int64) (bool, error)
}

type pgPatientRepository struct {
	client *db.PostgresClient
}

func NewPatientRepository(client *db.PostgresClient) PatientRepository {
	return &pgPatientRepository{client: client}
}

func (r *pgPatientRepository) CreatePatientProfileTX(ctx context.Context, tx transaction.Transaction, userID int64, birthDate int64, gender string, createdAt, updatedAt int64) error {
	query := `
		INSERT INTO patient_profiles (user_id, birth_date, gender, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET birth_date = $2, gender = $3, updated_at = $5
	`

	_, err := tx.Txm().ExecContext(ctx, query, userID, birthDate, gender, createdAt, updatedAt)
	if err != nil {
		return fmt.Errorf("failed to create patient profile: %w", err)
	}

	return nil
}

func (r *pgPatientRepository) AddPatientToDoctorTX(ctx context.Context, tx transaction.Transaction, doctorID, patientID, createdAt int64) error {
	query := `
		INSERT INTO doctor_patients (doctor_id, patient_id, created_at, deleted_at)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (doctor_id, patient_id) DO UPDATE
		SET deleted_at = NULL, created_at = EXCLUDED.created_at
	`

	_, err := tx.Txm().ExecContext(ctx, query, doctorID, patientID, createdAt)
	if err != nil {
		return fmt.Errorf("failed to add patient to doctor: %w", err)
	}

	return nil
}

func (r *pgPatientRepository) RemovePatientFromDoctorTX(ctx context.Context, tx transaction.Transaction, doctorID, patientID, deletedAt int64) error {
	query := `
		UPDATE doctor_patients
		SET deleted_at = $1
		WHERE doctor_id = $2 AND patient_id = $3 AND deleted_at IS NULL
	`

	_, err := tx.Txm().ExecContext(ctx, query, deletedAt, doctorID, patientID)
	if err != nil {
		return fmt.Errorf("failed to remove patient from doctor: %w", err)
	}

	return nil
}

func (r *pgPatientRepository) GetPatientProfileTX(ctx context.Context, tx transaction.Transaction, userID int64) (map[string]interface{}, error) {
	query := `
		SELECT user_id, birth_date, gender, created_at, updated_at
		FROM patient_profiles
		WHERE user_id = $1
	`

	var profile models.PatientProfile
	err := tx.Txm().GetContext(ctx, &profile, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("patient profile not found")
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	var gender string
	var birthDate int64
	if profile.Gender != nil {
		gender = *profile.Gender
	}
	if profile.BirthDate != nil {
		birthDate = *profile.BirthDate
	}

	return map[string]interface{}{
		"user_id":    profile.UserID,
		"birth_date": birthDate,
		"gender":     gender,
		"created_at": profile.CreatedAt,
		"updated_at": profile.UpdatedAt,
	}, nil
}

func (r *pgPatientRepository) IsPatientLinkedToDoctorTX(ctx context.Context, tx transaction.Transaction, patientID, doctorID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM doctor_patients
			WHERE doctor_id = $1 AND patient_id = $2 AND deleted_at IS NULL
		)
	`

	var linked bool
	if err := tx.Txm().GetContext(ctx, &linked, query, doctorID, patientID); err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}

	return linked, nil
}
