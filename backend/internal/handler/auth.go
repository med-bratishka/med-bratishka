package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/models"
	"medbratishka/pkg/logger"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

const (
	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "
)

type AuthHandler struct {
	authService service.AuthService
	formats     strfmt.Registry
	log         logger.Logger
}

func NewAuthHandler(authService service.AuthService, log logger.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, formats: strfmt.Default, log: log}
}

func (h *AuthHandler) FillHandlers(router *mux.Router) {
	base := "/auth"
	r := router.PathPrefix(base).Subrouter()

	r.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", h.Login).Methods(http.MethodPost)

	// В будущем можно добавить middleware авторизации в router
	r.HandleFunc("/refresh", h.Refresh).Methods(http.MethodPost)
	r.HandleFunc("/logout", h.Logout).Methods(http.MethodPost)
	r.HandleFunc("/logout-all", h.LogoutAll).Methods(http.MethodPost)
}

func (h *AuthHandler) Shutdown() {
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает новый аккаунт пользователя в системе и выдает токены доступа
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} models.AuthResponse "Успешная регистрация"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 409 {object} models.ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := req.Validate(h.formats); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
		return
	}

	input := &domain.RegistrationInput{
		Login:      deref(req.Login),
		Email:      req.Email.String(),
		Phone:      req.Phone,
		Password:   deref(req.Password),
		FirstName:  deref(req.FirstName),
		LastName:   deref(req.LastName),
		MiddleName: req.MiddleName,
		Role:       domain.Role(deref(req.Role)),
	}

	resp, err := h.authService.Register(r.Context(), input)
	if err != nil {
		h.handleAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toSwaggerAuthResponse(resp))
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентифицирует пользователя и выдает токены доступа
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Учетные данные"
// @Success 200 {object} models.AuthResponse "Успешный вход"
// @Failure 400 {object} models.ErrorResponse "Неверный формат запроса"
// @Failure 401 {object} models.ErrorResponse "Неверные учетные данные"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := req.Validate(h.formats); err != nil {
		h.respondWithError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
		return
	}

	input := &domain.AuthenticationInput{
		AccessParameter: deref(req.AccessParameter),
		Password:        deref(req.Password),
	}

	resp, err := h.authService.Login(r.Context(), input)
	if err != nil {
		h.handleAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSwaggerAuthResponse(resp))
}

// Refresh godoc
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены используя текущую сессию
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.AuthResponse "Новые токены"
// @Failure 401 {object} models.ErrorResponse "Неверный токен"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		h.respondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	resp, err := h.authService.Refresh(r.Context(), userCtx)
	if err != nil {
		h.handleAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSwaggerAuthResponse(resp))
}

// Logout godoc
// @Summary Выход из текущей сессии
// @Description Завершает текущую сессию пользователя
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse "Успешный выход"
// @Failure 401 {object} models.ErrorResponse "Неверный токен"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		h.respondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}
	if err := h.authService.Logout(r.Context(), userCtx); err != nil {
		h.handleAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "logged out successfully"})
}

// LogoutAll godoc
// @Summary Выход из всех сессий
// @Description Завершает все сессии пользователя на всех устройствах
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse "Успешный выход"
// @Failure 401 {object} models.ErrorResponse "Неверный токен"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		h.respondWithError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}
	if err := h.authService.LogoutAll(r.Context(), userCtx.ID, userCtx.Role); err != nil {
		h.handleAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "logged out from all sessions"})
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, cause error) {
	makeErrorResponse(w, r, h.log, statusCode, code, message, cause)
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrUserAlreadyExists):
		h.respondWithError(w, r, http.StatusConflict, "USER_EXISTS", "user already exists", err)
	case errors.Is(err, service.ErrInvalidCredentials):
		h.respondWithError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials", err)
	case errors.Is(err, service.ErrUserNotFound):
		h.respondWithError(w, r, http.StatusUnauthorized, "USER_NOT_FOUND", "user not found", err)
	case errors.Is(err, service.ErrInvalidRole):
		h.respondWithError(w, r, http.StatusBadRequest, "INVALID_ROLE", "invalid role", err)
	case errors.Is(err, service.ErrSessionNotFound):
		h.respondWithError(w, r, http.StatusUnauthorized, "SESSION_NOT_FOUND", "session not found", err)
	case errors.Is(err, service.ErrTokenExpired):
		h.respondWithError(w, r, http.StatusUnauthorized, "TOKEN_EXPIRED", "token expired", err)
	default:
		h.respondWithError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
	}
}

func toSwaggerAuthResponse(in *domain.AuthResponse) *models.AuthResponse {
	if in == nil {
		return &models.AuthResponse{}
	}
	return &models.AuthResponse{
		AccessToken: &models.TokenResponse{
			Token:     in.AccessToken.Token,
			ExpiresAt: in.AccessToken.ExpiresAt,
			Type:      in.AccessToken.Type,
		},
		RefreshToken: &models.TokenResponse{
			Token:     in.RefreshToken.Token,
			ExpiresAt: in.RefreshToken.ExpiresAt,
			Type:      in.RefreshToken.Type,
		},
		ServerTime: in.ServerTime,
		User: &models.UserResponse{
			ID:         in.User.ID,
			Login:      in.User.Login,
			Email:      in.User.Email,
			Phone:      in.User.Phone,
			Role:       in.User.Role,
			IsVerified: in.User.IsVerified,
			FirstName:  in.User.FirstName,
			LastName:   in.User.LastName,
			MiddleName: in.User.MiddleName,
			CreatedAt:  in.User.CreatedAt,
			UpdatedAt:  in.User.UpdatedAt,
		},
	}
}

func getErrorType(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	default:
		return "internal_error"
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
