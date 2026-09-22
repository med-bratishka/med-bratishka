package service

import (
	"context"
	"errors"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	"medbratishka/pkg/time_manager"
)

var (
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidDoctorCode  = errors.New("invalid doctor code")
	ErrDoctorNotFound     = errors.New("doctor not found")
	ErrPatientNotFound    = errors.New("patient not found")
	ErrAdminAccessDenied  = errors.New("admin has no access to clinic")
	ErrDoctorCodeConflict = errors.New("doctor code already used")
)

type BindingsService interface {
	BindDoctorToClinicByAdmin(ctx context.Context, adminID, clinicID, doctorID int64) error
	BindPatientToDoctorByCode(ctx context.Context, patientID int64, doctorCode string) error
	UpsertDoctorCode(ctx context.Context, doctorID int64, doctorCode string) error
}

type bindingsService struct {
	txRepo      transaction.Repository
	usersRepo   repository.UsersRepository
	clinicRepo  repository.ClinicRepository
	doctorRepo  repository.DoctorRepository
	patientRepo repository.PatientRepository
	timeManager time_manager.TimeManager
}

func NewBindingsService(
	txRepo transaction.Repository,
	usersRepo repository.UsersRepository,
	clinicRepo repository.ClinicRepository,
	doctorRepo repository.DoctorRepository,
	patientRepo repository.PatientRepository,
	timeManager time_manager.TimeManager,
) BindingsService {
	return &bindingsService{
		txRepo:      txRepo,
		usersRepo:   usersRepo,
		clinicRepo:  clinicRepo,
		doctorRepo:  doctorRepo,
		patientRepo: patientRepo,
		timeManager: timeManager,
	}
}

func (s *bindingsService) BindDoctorToClinicByAdmin(ctx context.Context, adminID, clinicID, doctorID int64) error {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("BindDoctorToClinicByAdmin/StartTransaction", err)
	}
	defer tx.Rollback()

	if _, err = s.usersRepo.GetByIDTX(ctx, tx, adminID, domain.RoleAdmin); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeForbidden, ErrForbidden, "FORBIDDEN", "forbidden")
		}
		return wrapInternal("BindDoctorToClinicByAdmin/GetAdmin", err)
	}

	isClinicAdmin, err := s.clinicRepo.IsClinicAdminTX(ctx, tx, clinicID, adminID)
	if err != nil {
		return wrapInternal("BindDoctorToClinicByAdmin/IsClinicAdminTX", err)
	}
	if !isClinicAdmin {
		return newServiceError(CodeForbidden, ErrAdminAccessDenied, "FORBIDDEN", "forbidden")
	}

	if _, err = s.usersRepo.GetByIDTX(ctx, tx, doctorID, domain.RoleDoctor); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrDoctorNotFound, "DOCTOR_NOT_FOUND", "doctor not found")
		}
		return wrapInternal("BindDoctorToClinicByAdmin/GetDoctor", err)
	}

	now := s.timeManager.Now().UnixMilli()
	if err = s.doctorRepo.AddDoctorToClinicTX(ctx, tx, doctorID, clinicID, adminID, now, now); err != nil {
		return wrapInternal("BindDoctorToClinicByAdmin/AddDoctorToClinicTX", err)
	}
	if err = s.doctorRepo.ApproveDoctorInClinicTX(ctx, tx, doctorID, clinicID, now); err != nil {
		return wrapInternal("BindDoctorToClinicByAdmin/ApproveDoctorInClinicTX", err)
	}

	if err = tx.Commit(); err != nil {
		return wrapInternal("BindDoctorToClinicByAdmin/Commit", err)
	}

	return nil
}

func (s *bindingsService) BindPatientToDoctorByCode(ctx context.Context, patientID int64, doctorCode string) error {
	if doctorCode == "" {
		return newServiceError(CodeBadRequest, ErrInvalidDoctorCode, "INVALID_DOCTOR_CODE", "invalid doctor code")
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("BindPatientToDoctorByCode/StartTransaction", err)
	}
	defer tx.Rollback()

	if _, err = s.usersRepo.GetByIDTX(ctx, tx, patientID, domain.RolePatient); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrPatientNotFound, "PATIENT_NOT_FOUND", "patient not found")
		}
		return wrapInternal("BindPatientToDoctorByCode/GetPatient", err)
	}

	doctorID, err := s.doctorRepo.FindDoctorIDByCodeTX(ctx, tx, doctorCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrDoctorNotFound, "DOCTOR_NOT_FOUND", "doctor not found")
		}
		return wrapInternal("BindPatientToDoctorByCode/FindDoctorIDByCodeTX", err)
	}

	now := s.timeManager.Now().UnixMilli()
	if err = s.patientRepo.AddPatientToDoctorTX(ctx, tx, doctorID, patientID, now); err != nil {
		return wrapInternal("BindPatientToDoctorByCode/AddPatientToDoctorTX", err)
	}

	if err = tx.Commit(); err != nil {
		return wrapInternal("BindPatientToDoctorByCode/Commit", err)
	}

	return nil
}

func (s *bindingsService) UpsertDoctorCode(ctx context.Context, doctorID int64, doctorCode string) error {
	if doctorCode == "" {
		return newServiceError(CodeBadRequest, ErrInvalidDoctorCode, "INVALID_DOCTOR_CODE", "invalid doctor code")
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("UpsertDoctorCode/StartTransaction", err)
	}
	defer tx.Rollback()

	if _, err = s.usersRepo.GetByIDTX(ctx, tx, doctorID, domain.RoleDoctor); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrDoctorNotFound, "DOCTOR_NOT_FOUND", "doctor not found")
		}
		return wrapInternal("UpsertDoctorCode/GetDoctor", err)
	}

	now := s.timeManager.Now().UnixMilli()
	if err = s.doctorRepo.UpsertDoctorCodeTX(ctx, tx, doctorID, doctorCode, now, now); err != nil {
		return wrapInternal("UpsertDoctorCode/UpsertDoctorCodeTX", err)
	}

	if err = tx.Commit(); err != nil {
		return wrapInternal("UpsertDoctorCode/Commit", err)
	}

	return nil
}
