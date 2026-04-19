package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/models"
	"medbratishka/pkg/logger"

	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
)

type ChatHandler struct {
	authService service.AuthService
	chatService service.ChatService
	formats     strfmt.Registry
	log         logger.Logger
}

func NewChatHandler(authService service.AuthService, chatService service.ChatService, log logger.Logger) *ChatHandler {
	return &ChatHandler{authService: authService, chatService: chatService, formats: strfmt.Default, log: log}
}

func (h *ChatHandler) FillHandlers(router *mux.Router) {
	r := router.PathPrefix("/chats").Subrouter()
	r.Use(AuthMiddleware(h.authService, h.log))
	r.Use(RequireRolesMiddleware(h.log, domain.RoleDoctor, domain.RolePatient))

	r.HandleFunc("", h.GetMyChats).Methods(http.MethodGet)
	r.HandleFunc("", h.CreateChatWithDoctor).Methods(http.MethodPost)
	r.HandleFunc("/{chat_id}", h.CloseChat).Methods(http.MethodDelete)
	r.HandleFunc("/{chat_id}/messages", h.GetChatMessages).Methods(http.MethodGet)
	r.HandleFunc("/{chat_id}/messages", h.SendMessage).Methods(http.MethodPost)
	r.HandleFunc("/{chat_id}/messages/{message_id}", h.DeleteMessage).Methods(http.MethodDelete)
}

func (h *ChatHandler) Shutdown() {}

// CreateChatWithDoctor godoc
// @Summary Create chat with doctor
// @Description Patient creates or gets chat with a linked doctor
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateChatRequest true "doctor id"
// @Success 200 {object} models.ChatResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /chats [post]
func (h *ChatHandler) CreateChatWithDoctor(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}
	if userCtx.Role != domain.RolePatient {
		makeErrorResponse(w, r, h.log, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
		return
	}

	var req models.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := req.Validate(h.formats); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
		return
	}

	chat, err := h.chatService.CreateOrGetChatWithDoctor(r.Context(), userCtx.ID, *req.DoctorID)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}

	writeJSON(w, http.StatusOK, &models.ChatResponse{
		ID:          chat.ID,
		DoctorID:    chat.DoctorID,
		PatientID:   chat.PatientID,
		OtherUserID: chat.OtherUserID,
		OtherLogin:  chat.OtherLogin,
		OtherName:   chat.OtherName,
		UpdatedAt:   chat.UpdatedAt,
	})
}

// SendMessage godoc
// @Summary Send chat message
// @Description Send message to doctor-patient chat
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chat_id path int true "Chat ID"
// @Param request body models.ChatMessageRequest true "message payload"
// @Success 200 {object} models.ChatMessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /chats/{chat_id}/messages [post]
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	chatID, err := parseChatID(r)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_CHAT_ID", "invalid chat id", err)
		return
	}

	var req models.ChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body", err)
		return
	}
	if err := req.Validate(h.formats); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", err)
		return
	}

	input, err := toSendMessageInput(&req)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_ATTACHMENT", "invalid attachment", err)
		return
	}

	msg, err := h.chatService.SendMessage(r.Context(), userCtx, chatID, input)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}

	writeJSON(w, http.StatusOK, &models.ChatMessageResponse{
		ID:                 msg.ID,
		SenderID:           msg.SenderID,
		SenderName:         msg.SenderName,
		Content:            msg.Content,
		AttachmentURL:      msg.AttachmentURL,
		AttachmentType:     msg.AttachmentType,
		AttachmentMimeType: msg.AttachmentMimeType,
		CreatedAt:          msg.CreatedAt,
	})
}

// GetMyChats godoc
// @Summary Get my chats
// @Description Returns all chats for authorized doctor or patient
// @Tags chat
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} models.PaginatedChatsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /chats [get]
func (h *ChatHandler) GetMyChats(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	limit, offset := parsePagination(r)
	result, err := h.chatService.GetMyChats(r.Context(), userCtx, limit, offset)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}

	items := make([]*models.ChatResponse, 0, len(result.Items))
	for _, c := range result.Items {
		items = append(items, &models.ChatResponse{
			ID:          c.ID,
			DoctorID:    c.DoctorID,
			PatientID:   c.PatientID,
			OtherUserID: c.OtherUserID,
			OtherLogin:  c.OtherLogin,
			OtherName:   c.OtherName,
			UpdatedAt:   c.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, &models.PaginatedChatsResponse{
		Items: items,
		Pagination: &models.PaginationResponse{
			Limit:  int32(result.Pagination.Limit),
			Offset: int32(result.Pagination.Offset),
			Total:  result.Pagination.Total,
		},
	})
}

// GetChatMessages godoc
// @Summary Get chat messages
// @Description Returns chat messages for authorized participant
// @Tags chat
// @Produce json
// @Security BearerAuth
// @Param chat_id path int true "Chat ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} models.PaginatedChatMessagesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /chats/{chat_id}/messages [get]
func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	chatID, err := parseChatID(r)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_CHAT_ID", "invalid chat id", err)
		return
	}

	limit, offset := parsePagination(r)
	result, err := h.chatService.GetChatMessages(r.Context(), userCtx, chatID, limit, offset)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}

	items := make([]*models.ChatMessageResponse, 0, len(result.Items))
	for _, m := range result.Items {
		items = append(items, &models.ChatMessageResponse{
			ID:                 m.ID,
			SenderID:           m.SenderID,
			SenderName:         m.SenderName,
			Content:            m.Content,
			AttachmentURL:      m.AttachmentURL,
			AttachmentType:     m.AttachmentType,
			AttachmentMimeType: m.AttachmentMimeType,
			CreatedAt:          m.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, &models.PaginatedChatMessagesResponse{
		Items: items,
		Pagination: &models.PaginationResponse{
			Limit:  int32(result.Pagination.Limit),
			Offset: int32(result.Pagination.Offset),
			Total:  result.Pagination.Total,
		},
	})
}

// CloseChat godoc
// @Summary Close chat
// @Description Close chat for doctor/patient participant
// @Tags chat
// @Produce json
// @Security BearerAuth
// @Param chat_id path int true "Chat ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /chats/{chat_id} [delete]
func (h *ChatHandler) CloseChat(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	chatID, err := parseChatID(r)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_CHAT_ID", "invalid chat id", err)
		return
	}

	if err := h.chatService.CloseChat(r.Context(), userCtx, chatID); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}

	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "chat closed"})
}

// DeleteMessage godoc
// @Summary Delete message
// @Description Delete own message in chat
// @Tags chat
// @Produce json
// @Security BearerAuth
// @Param chat_id path int true "Chat ID"
// @Param message_id path int true "Message ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /chats/{chat_id}/messages/{message_id} [delete]
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserFromContext(r)
	if userCtx == nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	chatID, err := parseChatID(r)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_CHAT_ID", "invalid chat id", err)
		return
	}
	messageID, err := parseMessageID(r)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusBadRequest, "INVALID_MESSAGE_ID", "invalid message id", err)
		return
	}

	if err := h.chatService.DeleteMessage(r.Context(), userCtx, chatID, messageID); err != nil {
		makeErrorResponse(w, r, h.log, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
		return
	}

	writeJSON(w, http.StatusOK, &models.SuccessResponse{Success: true, Message: "message deleted"})
}

func parseChatID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["chat_id"], 10, 64)
}

func parseMessageID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["message_id"], 10, 64)
}

func parsePagination(r *http.Request) (int, int) {
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p <= 200 {
			limit = p
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p >= 0 {
			offset = p
		}
	}
	return limit, offset
}

func toSendMessageInput(req *models.ChatMessageRequest) (*domain.SendMessageInput, error) {
	input := &domain.SendMessageInput{}
	if req.Content != nil {
		input.Content = *req.Content
	}
	if req.AttachmentBase64 == nil || *req.AttachmentBase64 == "" {
		return input, nil
	}

	payload := *req.AttachmentBase64
	if idx := strings.Index(payload, ","); idx > 0 && strings.Contains(payload[:idx], "base64") {
		payload = payload[idx+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	att := &domain.AttachmentInput{Data: decoded}
	if req.AttachmentName != nil {
		att.FileName = *req.AttachmentName
	}
	if req.AttachmentMimeType != nil {
		att.MimeType = *req.AttachmentMimeType
	}
	if req.AttachmentType != nil {
		att.MediaType = *req.AttachmentType
	}
	input.Attachment = att
	return input, nil
}
