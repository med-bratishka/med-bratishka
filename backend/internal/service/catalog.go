package service

import (
	"context"
	"errors"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	repoModels "medbratishka/internal/repository/models"
	"medbratishka/internal/repository/transaction"
)

var ErrClinicNotFound = errors.New("clinic not found")

type CatalogService interface {
	GetClinicByID(ctx context.Context, clinicID int64) (*domain.Clinic, error)
	GetClinics(ctx context.Context) ([]domain.Clinic, error)
	GetDoctorByID(ctx context.Context, doctorID int64) (*domain.Doctor, error)
	GetDoctors(ctx context.Context) ([]domain.Doctor, error)
}

type catalogService struct {
	txRepo     transaction.Repository
	clinicRepo repository.ClinicRepository
	doctorRepo repository.DoctorRepository
}

func NewCatalogService(
	txRepo transaction.Repository,
	clinicRepo repository.ClinicRepository,
	doctorRepo repository.DoctorRepository,
) CatalogService {
	return &catalogService{
		txRepo:     txRepo,
		clinicRepo: clinicRepo,
		doctorRepo: doctorRepo,
	}
}

func (s *catalogService) GetClinicByID(ctx context.Context, clinicID int64) (*domain.Clinic, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("GetClinicByID/StartTransaction", err)
	}
	defer tx.Rollback()

	clinic, err := s.clinicRepo.GetClinicByIDTX(ctx, tx, clinicID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeNotFound, ErrClinicNotFound, "CLINIC_NOT_FOUND", "clinic not found")
		}
		return nil, wrapInternal("GetClinicByID/GetClinicByIDTX", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, wrapInternal("GetClinicByID/Commit", err)
	}

	return mapClinicModelToDomain(clinic), nil
}

func (s *catalogService) GetClinics(ctx context.Context) ([]domain.Clinic, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("GetClinics/StartTransaction", err)
	}
	defer tx.Rollback()

	rows, err := s.clinicRepo.GetClinicsTX(ctx, tx)
	if err != nil {
		return nil, wrapInternal("GetClinics/GetClinicsTX", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, wrapInternal("GetClinics/Commit", err)
	}

	clinics := make([]domain.Clinic, 0, len(rows))
	for i := range rows {
		clinics = append(clinics, *mapClinicModelToDomain(&rows[i]))
	}
	return clinics, nil
}

func (s *catalogService) GetDoctorByID(ctx context.Context, doctorID int64) (*domain.Doctor, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("GetDoctorByID/StartTransaction", err)
	}
	defer tx.Rollback()

	doctor, err := s.doctorRepo.GetDoctorByIDTX(ctx, tx, doctorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeNotFound, ErrDoctorNotFound, "DOCTOR_NOT_FOUND", "doctor not found")
		}
		return nil, wrapInternal("GetDoctorByID/GetDoctorByIDTX", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, wrapInternal("GetDoctorByID/Commit", err)
	}

	return mapDoctorModelToDomain(doctor), nil
}

func (s *catalogService) GetDoctors(ctx context.Context) ([]domain.Doctor, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("GetDoctors/StartTransaction", err)
	}
	defer tx.Rollback()

	rows, err := s.doctorRepo.GetDoctorsTX(ctx, tx)
	if err != nil {
		return nil, wrapInternal("GetDoctors/GetDoctorsTX", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, wrapInternal("GetDoctors/Commit", err)
	}

	doctors := make([]domain.Doctor, 0, len(rows))
	for i := range rows {
		doctors = append(doctors, *mapDoctorModelToDomain(&rows[i]))
	}
	return doctors, nil
}

func mapClinicModelToDomain(c *repoModels.Clinic) *domain.Clinic {
	return &domain.Clinic{
		ID:          c.ID,
		Name:        c.Name,
		Description: derefString(c.Description),
		Address:     derefString(c.Address),
		Phone:       derefString(c.Phone),
		Email:       derefString(c.Email),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func mapDoctorModelToDomain(d *repoModels.DoctorWithProfile) *domain.Doctor {
	return &domain.Doctor{
		ID:             d.ID,
		Login:          d.Login,
		Email:          derefString(d.Email),
		Phone:          derefString(d.Phone),
		IsVerified:     d.IsVerified,
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		MiddleName:     derefString(d.MiddleName),
		Specialization: derefString(d.Specialization),
		LicenseNumber:  derefString(d.LicenseNumber),
		Bio:            derefString(d.Bio),
		DoctorCode:     derefString(d.DoctorCode),
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
