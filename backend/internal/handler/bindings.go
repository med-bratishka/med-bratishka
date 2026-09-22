package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/models"
	"medbratishka/pkg/logger"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

type BindingsHandler struct {
	authService     service.AuthService
	bindingsService service.BindingsService
	formats         strfmt.Registry
	log             logger.Logger
}

func NewBindingsHandler(authService service.AuthService, bindingsService service.BindingsService, log logger.Logger) *BindingsHandler {
	return &BindingsHandler{authService: authService, bindingsService: bindingsService, formats: strfmt.Default, log: log}
}

func (h *BindingsHandler) FillHandlers(router *mux.Router) {
	admin := router.PathPrefix("/clinics").Subrouter()
	admin.Use(AuthMiddleware(h.authService, h.log))
	admin.Use(RequireRolesMiddleware(h.log, domain.RoleAdmin))
	admin.HandleFunc("/{clinic_id}/doctors/{doctor_id}/bind", h.BindDoctorToClinicByAdmin).Methods(http.MethodPost)

	doctor := router.PathPrefix("/doctors").Subrouter()
	doctor.Use(AuthMiddleware(h.authService, h.log))
	doctor.Use(RequireRolesMiddleware(h.log, domain.RoleDoctor))
	doctor.HandleFunc("/me/code", h.UpsertDoctorCode).Methods(http.MethodPut)

	patient := router.PathPrefix("/patients").Subrouter()
	patient.Use(AuthMiddleware(h.authService, h.log))
	patient.Use(RequireRolesMiddleware(h.log, domain.RolePatient))
	patient.HandleFunc("/me/bind-doctor", h.BindPatientToDoctorByCode).Methods(http.MethodPost)
}

func (h *BindingsHandler) Shutdown() {
}

// BindDoctorToClinicByAdmin godoc
// @Summary Привязать доктора к клинике
// @Description Администратор клиники привязывает доктора к своей клинике
// @Tags clinic doctor
// @Produce json
// @Security BearerAuth
// @Param clinic_id path int true "Clinic ID"
// @Param doctor_id path int true "Doctor User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /clinics/{clinic_id}/doctors/{doctor_id}/bind [post]
func (h *BindingsHandler) BindDoctorToClinicByAdmin(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		h.respondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	vars := mux.Vars(r)
	clinicID, err := strconv.ParseInt(vars["clinic_id"], 10, 64)
	if err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_CLINIC_ID", "invalid clinic id", err)
		return
	}
	doctorID, err := strconv.ParseInt(vars["doctor_id"], 10, 64)
	if err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_DOCTOR_ID", "invalid doctor id", err)
		return
	}

	if err = h.bindingsService.BindDoctorToClinicByAdmin(r.Context(), userCtx.ID, clinicID, doctorID); err != nil {
		h.handleBindingsError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "doctor bound to clinic"})
}

// UpsertDoctorCode godoc
// @Summary Сохранить код доктора
// @Description Доктор сохраняет или обновляет свой уникальный код для привязки пациентов
// @Tags doctor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.DoctorCodeRequest true "Doctor code payload"
// @Success 200 {object} models.DoctorCodeResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /doctors/me/code [put]
func (h *BindingsHandler) UpsertDoctorCode(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		h.respondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	var req models.DoctorCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := req.Validate(h.formats); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
		return
	}

	doctorCode := deref(req.DoctorCode)
	if err := h.bindingsService.UpsertDoctorCode(r.Context(), userCtx.ID, doctorCode); err != nil {
		h.handleBindingsError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, &models.DoctorCodeResponse{DoctorCode: doctorCode})
}

// BindPatientToDoctorByCode godoc
// @Summary Привязаться к доктору по коду
// @Description Пациент вводит doctor code и привязывается к доктору
// @Tags patient doctor
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.DoctorCodeRequest true "Doctor code payload"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /patients/me/bind-doctor [post]
func (h *BindingsHandler) BindPatientToDoctorByCode(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		h.respondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	var req models.DoctorCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := req.Validate(h.formats); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
		return
	}

	if err := h.bindingsService.BindPatientToDoctorByCode(r.Context(), userCtx.ID, deref(req.DoctorCode)); err != nil {
		h.handleBindingsError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "patient bound to doctor"})
}

func (h *BindingsHandler) respondWithError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, cause error) {
	makeErrorResponse(w, r, h.log, statusCode, code, message, cause)
}

func (h *BindingsHandler) handleBindingsError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrAdminAccessDenied), errors.Is(err, service.ErrForbidden):
		h.respondWithError(w, r, http.StatusForbidden, "FORBIDDEN", "forbidden", err)
	case errors.Is(err, service.ErrDoctorNotFound):
		h.respondWithError(w, r, http.StatusNotFound, "DOCTOR_NOT_FOUND", "doctor not found", err)
	case errors.Is(err, service.ErrPatientNotFound):
		h.respondWithError(w, r, http.StatusNotFound, "PATIENT_NOT_FOUND", "patient not found", err)
	case errors.Is(err, service.ErrInvalidDoctorCode):
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_DOCTOR_CODE", "invalid doctor code", err)
	default:
		h.respondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
	}
}
