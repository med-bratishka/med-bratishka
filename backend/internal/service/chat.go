package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	repositoryModels "medbratishka/internal/repository/models"
	"medbratishka/internal/repository/transaction"
	"medbratishka/pkg/s3"
	"medbratishka/pkg/time_manager"
)

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrChatForbidden     = errors.New("chat forbidden")
	ErrMessageEmpty      = errors.New("message is empty")
	ErrMessageNotFound   = errors.New("message not found")
	ErrNoDoctorBinding   = errors.New("doctor has no active clinic binding")
	ErrNoPatientLink     = errors.New("patient is not linked to doctor")
	ErrAttachmentInvalid = errors.New("invalid attachment")
)

type ChatService interface {
	CreateOrGetChatWithDoctor(ctx context.Context, patientID, doctorID int64) (*domain.ChatSummary, error)
	SendMessage(ctx context.Context, senderCtx *domain.UserTokenContext, chatID int64, input *domain.SendMessageInput) (*domain.ChatMessage, error)
	DeleteMessage(ctx context.Context, senderCtx *domain.UserTokenContext, chatID, messageID int64) error
	CloseChat(ctx context.Context, userCtx *domain.UserTokenContext, chatID int64) error
	GetMyChats(ctx context.Context, userCtx *domain.UserTokenContext, limit, offset int) (*domain.PaginatedChats, error)
	GetChatMessages(ctx context.Context, userCtx *domain.UserTokenContext, chatID int64, limit, offset int) (*domain.PaginatedChatMessages, error)
}

type chatService struct {
	txRepo        transaction.Repository
	chatRepo      repository.ChatRepository
	doctorRepo    repository.DoctorRepository
	patientRepo   repository.PatientRepository
	timeManager   time_manager.TimeManager
	storage       s3.Storage
	maxUploadSize int64
}

func NewChatService(
	txRepo transaction.Repository,
	chatRepo repository.ChatRepository,
	doctorRepo repository.DoctorRepository,
	patientRepo repository.PatientRepository,
	timeManager time_manager.TimeManager,
	storage s3.Storage,
	maxUploadSizeMB int64,
) ChatService {
	return &chatService{
		txRepo:        txRepo,
		chatRepo:      chatRepo,
		doctorRepo:    doctorRepo,
		patientRepo:   patientRepo,
		timeManager:   timeManager,
		storage:       storage,
		maxUploadSize: maxUploadSizeMB * 1024 * 1024,
	}
}

func (s *chatService) CreateOrGetChatWithDoctor(ctx context.Context, patientID, doctorID int64) (*domain.ChatSummary, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("CreateOrGetChatWithDoctor/StartTransaction", err)
	}
	defer tx.Rollback()

	linked, err := s.patientRepo.IsPatientLinkedToDoctorTX(ctx, tx, patientID, doctorID)
	if err != nil {
		return nil, wrapInternal("CreateOrGetChatWithDoctor/IsPatientLinkedToDoctorTX", err)
	}
	if !linked {
		return nil, newServiceError(CodeForbidden, ErrNoPatientLink, "FORBIDDEN", "patient is not linked to doctor")
	}

	hasMembership, err := s.doctorRepo.HasActiveClinicMembershipTX(ctx, tx, doctorID)
	if err != nil {
		return nil, wrapInternal("CreateOrGetChatWithDoctor/HasActiveClinicMembershipTX", err)
	}
	if !hasMembership {
		return nil, newServiceError(CodeForbidden, ErrNoDoctorBinding, "FORBIDDEN", "doctor has no active clinic binding")
	}

	now := s.timeManager.Now().UnixMilli()
	chatID, err := s.chatRepo.CreateOrGetChatTX(ctx, tx, doctorID, patientID, now, now)
	if err != nil {
		return nil, wrapInternal("CreateOrGetChatWithDoctor/CreateOrGetChatTX", err)
	}

	chat, err := s.chatRepo.GetChatByIDTX(ctx, tx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeNotFound, ErrChatNotFound, "CHAT_NOT_FOUND", "chat not found")
		}
		return nil, wrapInternal("CreateOrGetChatWithDoctor/GetChatByIDTX", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("CreateOrGetChatWithDoctor/Commit", err)
	}

	return &domain.ChatSummary{
		ID:          chat.ID,
		DoctorID:    chat.DoctorID,
		PatientID:   chat.PatientID,
		OtherUserID: doctorID,
		UpdatedAt:   chat.UpdatedAt,
	}, nil
}

func (s *chatService) SendMessage(ctx context.Context, senderCtx *domain.UserTokenContext, chatID int64, input *domain.SendMessageInput) (*domain.ChatMessage, error) {
	if input == nil {
		return nil, newServiceError(CodeBadRequest, ErrMessageEmpty, "EMPTY_MESSAGE", "message is empty")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" && input.Attachment == nil {
		return nil, newServiceError(CodeBadRequest, ErrMessageEmpty, "EMPTY_MESSAGE", "message is empty")
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("SendMessage/StartTransaction", err)
	}
	defer tx.Rollback()

	chat, err := s.chatRepo.GetChatByIDTX(ctx, tx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeNotFound, ErrChatNotFound, "CHAT_NOT_FOUND", "chat not found")
		}
		return nil, wrapInternal("SendMessage/GetChatByIDTX", err)
	}

	if err := s.ensureChatAccessTX(ctx, tx, senderCtx, chat); err != nil {
		return nil, err
	}

	var (
		attachmentURL  *string
		attachmentType *string
		attachmentMime *string
	)
	if input.Attachment != nil {
		if s.storage == nil {
			return nil, newServiceError(CodeInternal, fmt.Errorf("s3 storage is not configured"), "INTERNAL_ERROR", "internal server error")
		}
		if err := s.validateAttachment(input.Attachment); err != nil {
			return nil, err
		}
		key := fmt.Sprintf("chats/%d/%d/%d_%s", chatID, senderCtx.ID, s.timeManager.Now().UnixMilli(), sanitizeFilename(input.Attachment.FileName))
		uploadedURL, err := s.storage.Upload(ctx, key, input.Attachment.Data, input.Attachment.MimeType)
		if err != nil {
			return nil, wrapInternal("SendMessage/UploadAttachment", err)
		}
		attachmentURL = &uploadedURL
		attachmentType = &input.Attachment.MediaType
		attachmentMime = &input.Attachment.MimeType
	}

	var contentPtr *string
	if content != "" {
		contentPtr = &content
	}

	now := s.timeManager.Now().UnixMilli()
	messageID, err := s.chatRepo.SendMessageTX(ctx, tx, chatID, senderCtx.ID, contentPtr, attachmentURL, attachmentType, attachmentMime, now, now)
	if err != nil {
		return nil, wrapInternal("SendMessage/SendMessageTX", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("SendMessage/Commit", err)
	}

	resp := &domain.ChatMessage{ID: messageID, SenderID: senderCtx.ID, Content: content, CreatedAt: now}
	if attachmentURL != nil {
		resp.AttachmentURL = *attachmentURL
	}
	if attachmentType != nil {
		resp.AttachmentType = *attachmentType
	}
	if attachmentMime != nil {
		resp.AttachmentMimeType = *attachmentMime
	}
	return resp, nil
}

func (s *chatService) DeleteMessage(ctx context.Context, senderCtx *domain.UserTokenContext, chatID, messageID int64) error {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("DeleteMessage/StartTransaction", err)
	}
	defer tx.Rollback()

	chat, err := s.chatRepo.GetChatByIDTX(ctx, tx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrChatNotFound, "CHAT_NOT_FOUND", "chat not found")
		}
		return wrapInternal("DeleteMessage/GetChatByIDTX", err)
	}
	if err := s.ensureChatAccessTX(ctx, tx, senderCtx, chat); err != nil {
		return err
	}

	msg, err := s.chatRepo.GetMessageByIDTX(ctx, tx, messageID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrMessageNotFound, "MESSAGE_NOT_FOUND", "message not found")
		}
		return wrapInternal("DeleteMessage/GetMessageByIDTX", err)
	}
	if msg.ChatID != chatID || msg.DeletedAt != nil {
		return newServiceError(CodeNotFound, ErrMessageNotFound, "MESSAGE_NOT_FOUND", "message not found")
	}
	if msg.SenderID != senderCtx.ID {
		return newServiceError(CodeForbidden, ErrAccessDenied, "FORBIDDEN", "forbidden")
	}

	now := s.timeManager.Now().UnixMilli()
	if err := s.chatRepo.DeleteMessageTX(ctx, tx, messageID, now); err != nil {
		return wrapInternal("DeleteMessage/DeleteMessageTX", err)
	}

	if err := tx.Commit(); err != nil {
		return wrapInternal("DeleteMessage/Commit", err)
	}
	return nil
}

func (s *chatService) CloseChat(ctx context.Context, userCtx *domain.UserTokenContext, chatID int64) error {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("CloseChat/StartTransaction", err)
	}
	defer tx.Rollback()

	chat, err := s.chatRepo.GetChatByIDTX(ctx, tx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return newServiceError(CodeNotFound, ErrChatNotFound, "CHAT_NOT_FOUND", "chat not found")
		}
		return wrapInternal("CloseChat/GetChatByIDTX", err)
	}
	if err := s.ensureChatAccessTX(ctx, tx, userCtx, chat); err != nil {
		return err
	}

	now := s.timeManager.Now().UnixMilli()
	if err := s.chatRepo.CloseChatTX(ctx, tx, chatID, now); err != nil {
		return wrapInternal("CloseChat/CloseChatTX", err)
	}

	if err := tx.Commit(); err != nil {
		return wrapInternal("CloseChat/Commit", err)
	}
	return nil
}

func (s *chatService) GetMyChats(ctx context.Context, userCtx *domain.UserTokenContext, limit, offset int) (*domain.PaginatedChats, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("GetMyChats/StartTransaction", err)
	}
	defer tx.Rollback()

	total, err := s.chatRepo.GetUserChatsTotalTX(ctx, tx, userCtx.ID)
	if err != nil {
		return nil, wrapInternal("GetMyChats/GetUserChatsTotalTX", err)
	}

	rows, err := s.chatRepo.GetUserChatsTX(ctx, tx, userCtx.ID, limit, offset)
	if err != nil {
		return nil, wrapInternal("GetMyChats/GetUserChatsTX", err)
	}

	res := make([]domain.ChatSummary, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ChatSummary{
			ID:          row.ID,
			DoctorID:    row.DoctorID,
			PatientID:   row.PatientID,
			OtherUserID: row.OtherUserID,
			OtherLogin:  row.OtherLogin,
			OtherName:   strings.TrimSpace(row.OtherFirstName + " " + row.OtherLastName),
			UpdatedAt:   row.UpdatedAt,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("GetMyChats/Commit", err)
	}

	return &domain.PaginatedChats{Items: res, Pagination: domain.Pagination{Limit: limit, Offset: offset, Total: total}}, nil
}

func (s *chatService) GetChatMessages(ctx context.Context, userCtx *domain.UserTokenContext, chatID int64, limit, offset int) (*domain.PaginatedChatMessages, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("GetChatMessages/StartTransaction", err)
	}
	defer tx.Rollback()

	chat, err := s.chatRepo.GetChatByIDTX(ctx, tx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeNotFound, ErrChatNotFound, "CHAT_NOT_FOUND", "chat not found")
		}
		return nil, wrapInternal("GetChatMessages/GetChatByIDTX", err)
	}
	if err := s.ensureChatAccessTX(ctx, tx, userCtx, chat); err != nil {
		return nil, err
	}

	total, err := s.chatRepo.GetChatMessagesTotalTX(ctx, tx, chatID)
	if err != nil {
		return nil, wrapInternal("GetChatMessages/GetChatMessagesTotalTX", err)
	}

	rows, err := s.chatRepo.GetChatMessagesTX(ctx, tx, chatID, limit, offset)
	if err != nil {
		return nil, wrapInternal("GetChatMessages/GetChatMessagesTX", err)
	}

	res := make([]domain.ChatMessage, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ChatMessage{
			ID:                 row.ID,
			SenderID:           row.SenderID,
			SenderName:         strings.TrimSpace(row.FirstName + " " + row.LastName),
			Content:            derefStr(row.Content),
			AttachmentURL:      derefStr(row.AttachmentURL),
			AttachmentType:     derefStr(row.AttachmentType),
			AttachmentMimeType: derefStr(row.AttachmentMimeType),
			CreatedAt:          row.CreatedAt,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("GetChatMessages/Commit", err)
	}

	return &domain.PaginatedChatMessages{Items: res, Pagination: domain.Pagination{Limit: limit, Offset: offset, Total: total}}, nil
}

func (s *chatService) ensureChatAccessTX(ctx context.Context, tx transaction.Transaction, userCtx *domain.UserTokenContext, chat *repositoryModels.Chat) error {
	if userCtx.Role == domain.RoleDoctor {
		if chat.DoctorID != userCtx.ID {
			return newServiceError(CodeForbidden, ErrAccessDenied, "FORBIDDEN", "forbidden")
		}
		hasMembership, err := s.doctorRepo.HasActiveClinicMembershipTX(ctx, tx, userCtx.ID)
		if err != nil {
			return wrapInternal("ensureChatAccessTX/HasActiveClinicMembershipTX", err)
		}
		if !hasMembership {
			return newServiceError(CodeForbidden, ErrNoDoctorBinding, "FORBIDDEN", "doctor has no active clinic binding")
		}
		return nil
	}

	if userCtx.Role == domain.RolePatient {
		if chat.PatientID != userCtx.ID {
			return newServiceError(CodeForbidden, ErrAccessDenied, "FORBIDDEN", "forbidden")
		}
		linked, err := s.patientRepo.IsPatientLinkedToDoctorTX(ctx, tx, chat.PatientID, chat.DoctorID)
		if err != nil {
			return wrapInternal("ensureChatAccessTX/IsPatientLinkedToDoctorTX", err)
		}
		if !linked {
			return newServiceError(CodeForbidden, ErrNoPatientLink, "FORBIDDEN", "patient is not linked to doctor")
		}
		return nil
	}

	return newServiceError(CodeForbidden, ErrChatForbidden, "FORBIDDEN", "forbidden")
}

func (s *chatService) validateAttachment(att *domain.AttachmentInput) error {
	if att == nil {
		return nil
	}
	if len(att.Data) == 0 || att.FileName == "" || att.MimeType == "" || att.MediaType == "" {
		return newServiceError(CodeBadRequest, ErrAttachmentInvalid, "INVALID_ATTACHMENT", "invalid attachment")
	}
	if s.maxUploadSize > 0 && int64(len(att.Data)) > s.maxUploadSize {
		return newServiceError(CodeBadRequest, ErrAttachmentInvalid, "ATTACHMENT_TOO_LARGE", "attachment too large")
	}
	if att.MediaType != "image" && att.MediaType != "audio" {
		return newServiceError(CodeBadRequest, ErrAttachmentInvalid, "INVALID_ATTACHMENT_TYPE", "invalid attachment type")
	}
	if att.MediaType == "image" && !strings.HasPrefix(att.MimeType, "image/") {
		return newServiceError(CodeBadRequest, ErrAttachmentInvalid, "INVALID_ATTACHMENT_MIME", "invalid attachment mime")
	}
	if att.MediaType == "audio" && !strings.HasPrefix(att.MimeType, "audio/") {
		return newServiceError(CodeBadRequest, ErrAttachmentInvalid, "INVALID_ATTACHMENT_MIME", "invalid attachment mime")
	}
	return nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, "..", "")
	if base == "" || base == "." || base == "/" {
		return "file"
	}
	return base
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
