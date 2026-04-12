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

type DoctorRepository interface {
	CreateDoctorProfileTX(ctx context.Context, tx transaction.Transaction, userID int64, specialization, licenseNumber, bio string, createdAt, updatedAt int64) error
	AddDoctorToClinicTX(ctx context.Context, tx transaction.Transaction, doctorID, clinicID, invitedByAdminID, createdAt, updatedAt int64) error
	ApproveDoctorInClinicTX(ctx context.Context, tx transaction.Transaction, doctorID, clinicID, updatedAt int64) error
	RejectDoctorInClinicTX(ctx context.Context, tx transaction.Transaction, doctorID, clinicID, updatedAt int64) error
	GetDoctorProfileTX(ctx context.Context, tx transaction.Transaction, userID int64) (*models.DoctorProfile, error)
	UpsertDoctorCodeTX(ctx context.Context, tx transaction.Transaction, userID int64, doctorCode string, createdAt, updatedAt int64) error
	FindDoctorIDByCodeTX(ctx context.Context, tx transaction.Transaction, doctorCode string) (int64, error)
	HasActiveClinicMembershipTX(ctx context.Context, tx transaction.Transaction, doctorID int64) (bool, error)
}

type pgDoctorRepository struct {
	client *db.PostgresClient
}

func NewDoctorRepository(client *db.PostgresClient) DoctorRepository {
	return &pgDoctorRepository{client: client}
}

func (r *pgDoctorRepository) CreateDoctorProfileTX(ctx context.Context, tx transaction.Transaction, userID int64, specialization, licenseNumber, bio string, createdAt, updatedAt int64) error {
	query := `
		INSERT INTO doctor_profiles (user_id, specialization, license_number, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE
		SET specialization = $2, license_number = $3, bio = $4, updated_at = $6
	`

	_, err := tx.Txm().ExecContext(ctx, query, userID, specialization, licenseNumber, bio, createdAt, updatedAt)
	if err != nil {
		return fmt.Errorf("failed to create doctor profile: %w", err)
	}

	return nil
}

func (r *pgDoctorRepository) AddDoctorToClinicTX(ctx context.Context, tx transaction.Transaction, doctorID, clinicID, invitedByAdminID, createdAt, updatedAt int64) error {
	query := `
		INSERT INTO doctor_clinic_memberships (doctor_id, clinic_id, invited_by, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $5)
		ON CONFLICT (doctor_id, clinic_id) DO UPDATE
		SET status = 'pending', updated_at = $5
	`

	_, err := tx.Txm().ExecContext(ctx, query, doctorID, clinicID, invitedByAdminID, createdAt, updatedAt)
	if err != nil {
		return fmt.Errorf("failed to add doctor to clinic: %w", err)
	}

	return nil
}

func (r *pgDoctorRepository) ApproveDoctorInClinicTX(ctx context.Context, tx transaction.Transaction, doctorID, clinicID, updatedAt int64) error {
	query := `
		UPDATE doctor_clinic_memberships
		SET status = 'active', updated_at = $1
		WHERE doctor_id = $2 AND clinic_id = $3 AND status = 'pending'
	`

	_, err := tx.Txm().ExecContext(ctx, query, updatedAt, doctorID, clinicID)
	if err != nil {
		return fmt.Errorf("failed to approve doctor: %w", err)
	}

	return nil
}

func (r *pgDoctorRepository) RejectDoctorInClinicTX(ctx context.Context, tx transaction.Transaction, doctorID, clinicID, updatedAt int64) error {
	query := `
		UPDATE doctor_clinic_memberships
		SET status = 'rejected', updated_at = $1
		WHERE doctor_id = $2 AND clinic_id = $3
	`

	_, err := tx.Txm().ExecContext(ctx, query, updatedAt, doctorID, clinicID)
	if err != nil {
		return fmt.Errorf("failed to reject doctor: %w", err)
	}

	return nil
}

func (r *pgDoctorRepository) GetDoctorProfileTX(ctx context.Context, tx transaction.Transaction, userID int64) (*models.DoctorProfile, error) {
	query := `
		SELECT user_id, specialization, license_number, bio, doctor_code, created_at, updated_at
		FROM doctor_profiles
		WHERE user_id = $1
	`

	var profile models.DoctorProfile
	err := tx.Txm().GetContext(ctx, &profile, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("doctor profile not found")
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &profile, nil
}

func (r *pgDoctorRepository) UpsertDoctorCodeTX(ctx context.Context, tx transaction.Transaction, userID int64, doctorCode string, createdAt, updatedAt int64) error {
	query := `
		INSERT INTO doctor_profiles (user_id, doctor_code, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET doctor_code = EXCLUDED.doctor_code, updated_at = EXCLUDED.updated_at
	`

	if _, err := tx.Txm().ExecContext(ctx, query, userID, doctorCode, createdAt, updatedAt); err != nil {
		return fmt.Errorf("failed to upsert doctor code: %w", err)
	}
	return nil
}

func (r *pgDoctorRepository) FindDoctorIDByCodeTX(ctx context.Context, tx transaction.Transaction, doctorCode string) (int64, error) {
	query := `
		SELECT dp.user_id
		FROM doctor_profiles dp
		JOIN users u ON u.id = dp.user_id
		WHERE dp.doctor_code = $1 AND u.role = 'doctor' AND u.deleted_at IS NULL
		LIMIT 1
	`

	var doctorID int64
	if err := tx.Txm().GetContext(ctx, &doctorID, query, doctorCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("query error: %w", err)
	}

	return doctorID, nil
}

func (r *pgDoctorRepository) HasActiveClinicMembershipTX(ctx context.Context, tx transaction.Transaction, doctorID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM doctor_clinic_memberships
			WHERE doctor_id = $1 AND status = 'active'
		)
	`

	var exists bool
	if err := tx.Txm().GetContext(ctx, &exists, query, doctorID); err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}

	return exists, nil
}
