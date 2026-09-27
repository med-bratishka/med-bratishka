package handler

import (
	"encoding/json"
	"net/http"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/models"
	"medbratishka/pkg/logger"

	"github.com/gorilla/mux"
)

type WebPushHandler struct {
	authService    service.AuthService
	webPushService service.WebPushService
	log            logger.Logger
}

const webPushRequestMaxBytes = 32 * 1024

func NewWebPushHandler(authService service.AuthService, webPushService service.WebPushService, log logger.Logger) *WebPushHandler {
	return &WebPushHandler{authService: authService, webPushService: webPushService, log: log}
}

func (h *WebPushHandler) FillHandlers(router *mux.Router) {
	r := router.PathPrefix("/notifications/web-push").Subrouter()
	r.Use(AuthMiddleware(h.authService, h.log))
	r.Use(RequireRolesMiddleware(h.log, domain.RoleDoctor, domain.RolePatient))

	r.HandleFunc("/config", h.GetConfig).Methods(http.MethodGet)
	r.HandleFunc("/subscriptions", h.Subscribe).Methods(http.MethodPost)
	r.HandleFunc("/subscriptions", h.Unsubscribe).Methods(http.MethodDelete)
}

func (h *WebPushHandler) Shutdown() {}

func (h *WebPushHandler) GetConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":    h.webPushService.Enabled(),
		"public_key": h.webPushService.PublicKey(),
	})
}

func (h *WebPushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, webPushRequestMaxBytes)
	var input domain.WebPushSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := h.webPushService.Subscribe(r.Context(), userCtx.ID, input, r.UserAgent()); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}
	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "web push subscription saved"})
}

func (h *WebPushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, webPushRequestMaxBytes)
	var input struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := h.webPushService.Unsubscribe(r.Context(), userCtx.ID, input.Endpoint); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}
	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "web push subscription deleted"})
}
