package handler

import (
	"errors"
	"net/http"
	"strconv"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/models"
	"medbratishka/pkg/logger"

	"github.com/gorilla/mux"
)

type CatalogHandler struct {
	catalogService service.CatalogService
	log            logger.Logger
}

func NewCatalogHandler(catalogService service.CatalogService, log logger.Logger) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService, log: log}
}

func (h *CatalogHandler) FillHandlers(router *mux.Router) {
	clinics := router.PathPrefix("/clinics").Subrouter()
	clinics.HandleFunc("", h.GetClinics).Methods(http.MethodGet)
	clinics.HandleFunc("/{clinic_id}", h.GetClinicByID).Methods(http.MethodGet)

	doctors := router.PathPrefix("/doctors").Subrouter()
	doctors.HandleFunc("", h.GetDoctors).Methods(http.MethodGet)
	doctors.HandleFunc("/{doctor_id}", h.GetDoctorByID).Methods(http.MethodGet)
}

func (h *CatalogHandler) Shutdown() {
}

// GetClinics godoc
// @Summary Получить список клиник
// @Description Возвращает все клиники без пагинации
// @Tags clinic
// @Produce json
// @Success 200 {array} models.ClinicResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /clinics [get]
func (h *CatalogHandler) GetClinics(w http.ResponseWriter, r *http.Request) {
	clinics, err := h.catalogService.GetClinics(r.Context())
	if err != nil {
		h.handleCatalogError(w, r, err)
		return
	}

	resp := make(models.ClinicsResponse, 0, len(clinics))
	for i := range clinics {
		resp = append(resp, toClinicResponse(&clinics[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetClinicByID godoc
// @Summary Получить клинику по id
// @Description Возвращает клинику по id
// @Tags clinic
// @Produce json
// @Param clinic_id path int true "Clinic ID"
// @Success 200 {object} models.ClinicResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /clinics/{clinic_id} [get]
func (h *CatalogHandler) GetClinicByID(w http.ResponseWriter, r *http.Request) {
	clinicID, err := strconv.ParseInt(mux.Vars(r)["clinic_id"], 10, 64)
	if err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_CLINIC_ID", "invalid clinic id", err)
		return
	}

	clinic, err := h.catalogService.GetClinicByID(r.Context(), clinicID)
	if err != nil {
		h.handleCatalogError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toClinicResponse(clinic))
}

// GetDoctors godoc
// @Summary Получить список врачей
// @Description Возвращает всех врачей без пагинации
// @Tags doctor
// @Produce json
// @Success 200 {array} models.DoctorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /doctors [get]
func (h *CatalogHandler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := h.catalogService.GetDoctors(r.Context())
	if err != nil {
		h.handleCatalogError(w, r, err)
		return
	}

	resp := make(models.DoctorsResponse, 0, len(doctors))
	for i := range doctors {
		resp = append(resp, toDoctorResponse(&doctors[i]))
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetDoctorByID godoc
// @Summary Получить врача по id
// @Description Возвращает врача по id
// @Tags doctor
// @Produce json
// @Param doctor_id path int true "Doctor ID"
// @Success 200 {object} models.DoctorResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /doctors/{doctor_id} [get]
func (h *CatalogHandler) GetDoctorByID(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r)["doctor_id"], 10, 64)
	if err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_DOCTOR_ID", "invalid doctor id", err)
		return
	}

	doctor, err := h.catalogService.GetDoctorByID(r.Context(), doctorID)
	if err != nil {
		h.handleCatalogError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toDoctorResponse(doctor))
}

func (h *CatalogHandler) respondWithError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, cause error) {
	makeErrorResponse(w, r, h.log, statusCode, code, message, cause)
}

func (h *CatalogHandler) handleCatalogError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrClinicNotFound):
		h.respondWithError(w, r, http.StatusNotFound, "CLINIC_NOT_FOUND", "clinic not found", err)
	case errors.Is(err, service.ErrDoctorNotFound):
		h.respondWithError(w, r, http.StatusNotFound, "DOCTOR_NOT_FOUND", "doctor not found", err)
	default:
		h.respondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
	}
}

func toClinicResponse(c *domain.Clinic) *models.ClinicResponse {
	return &models.ClinicResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Address:     c.Address,
		Phone:       c.Phone,
		Email:       c.Email,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toDoctorResponse(d *domain.Doctor) *models.DoctorResponse {
	return &models.DoctorResponse{
		ID:             d.ID,
		Login:          d.Login,
		Email:          d.Email,
		Phone:          d.Phone,
		IsVerified:     d.IsVerified,
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		MiddleName:     d.MiddleName,
		Specialization: d.Specialization,
		LicenseNumber:  d.LicenseNumber,
		Bio:            d.Bio,
		DoctorCode:     d.DoctorCode,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
