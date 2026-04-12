package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"medbratishka/internal/repository/models"
	"medbratishka/internal/repository/transaction"
)

type ChatRepository interface {
	CreateOrGetChatTX(ctx context.Context, tx transaction.Transaction, doctorID, patientID, createdAt, updatedAt int64) (int64, error)
	SendMessageTX(
		ctx context.Context,
		tx transaction.Transaction,
		chatID, senderID int64,
		content *string,
		attachmentURL, attachmentType, attachmentMimeType *string,
		createdAt, chatUpdatedAt int64,
	) (int64, error)
	GetChatMessagesTX(ctx context.Context, tx transaction.Transaction, chatID int64, limit, offset int) ([]models.ChatMessageDetail, error)
	GetChatMessagesTotalTX(ctx context.Context, tx transaction.Transaction, chatID int64) (int64, error)
	GetUserChatsTX(ctx context.Context, tx transaction.Transaction, userID int64, limit, offset int) ([]models.UserChatDetail, error)
	GetUserChatsTotalTX(ctx context.Context, tx transaction.Transaction, userID int64) (int64, error)
	DeleteMessageTX(ctx context.Context, tx transaction.Transaction, messageID, deletedAt int64) error
	CloseChatTX(ctx context.Context, tx transaction.Transaction, chatID, closedAt int64) error
	GetChatByIDTX(ctx context.Context, tx transaction.Transaction, chatID int64) (*models.Chat, error)
	GetMessageByIDTX(ctx context.Context, tx transaction.Transaction, messageID int64) (*models.Message, error)
}

type pgChatRepository struct {
}

func NewChatRepository() ChatRepository {
	return &pgChatRepository{}
}

func (r *pgChatRepository) CreateOrGetChatTX(ctx context.Context, tx transaction.Transaction, doctorID, patientID, createdAt, updatedAt int64) (int64, error) {
	query := `
		SELECT id FROM chats
		WHERE doctor_id = $1 AND patient_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var chat models.Chat
	err := tx.Txm().GetContext(ctx, &chat, query, doctorID, patientID)
	if err == nil {
		return chat.ID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("query error: %w", err)
	}

	insertQuery := `
		INSERT INTO chats (doctor_id, patient_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var chatID int64
	err = tx.Txm().QueryRowContext(ctx, insertQuery, doctorID, patientID, createdAt, updatedAt).Scan(&chatID)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat: %w", err)
	}

	return chatID, nil
}

func (r *pgChatRepository) SendMessageTX(
	ctx context.Context,
	tx transaction.Transaction,
	chatID, senderID int64,
	content *string,
	attachmentURL, attachmentType, attachmentMimeType *string,
	createdAt, chatUpdatedAt int64,
) (int64, error) {
	query := `
		INSERT INTO messages (chat_id, sender_id, content, attachment_url, attachment_type, attachment_mime_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var messageID int64
	err := tx.Txm().QueryRowContext(
		ctx,
		query,
		chatID,
		senderID,
		content,
		attachmentURL,
		attachmentType,
		attachmentMimeType,
		createdAt,
	).Scan(&messageID)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	updateChatQuery := `UPDATE chats SET updated_at = $1 WHERE id = $2`
	_, _ = tx.Txm().ExecContext(ctx, updateChatQuery, chatUpdatedAt, chatID)

	return messageID, nil
}

func (r *pgChatRepository) GetChatMessagesTX(ctx context.Context, tx transaction.Transaction, chatID int64, limit, offset int) ([]models.ChatMessageDetail, error) {
	query := `
		SELECT m.id, m.sender_id, u.login, u.first_name, u.last_name, m.content, m.attachment_url, m.attachment_type, m.attachment_mime_type, m.created_at
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.chat_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var details []models.ChatMessageDetail
	err := tx.Txm().SelectContext(ctx, &details, query, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return details, nil
}

func (r *pgChatRepository) GetChatMessagesTotalTX(ctx context.Context, tx transaction.Transaction, chatID int64) (int64, error) {
	query := `
		SELECT COUNT(1)
		FROM messages
		WHERE chat_id = $1 AND deleted_at IS NULL
	`

	var total int64
	if err := tx.Txm().GetContext(ctx, &total, query, chatID); err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	return total, nil
}

func (r *pgChatRepository) GetUserChatsTX(ctx context.Context, tx transaction.Transaction, userID int64, limit, offset int) ([]models.UserChatDetail, error) {
	query := `
		SELECT 
			c.id,
			c.doctor_id,
			c.patient_id,
			CASE WHEN c.doctor_id = $1 THEN c.patient_id ELSE c.doctor_id END as other_user_id,
			u.login as login,
			u.first_name as first_name,
			u.last_name as last_name,
			c.updated_at
		FROM chats c
		JOIN users u ON (
			CASE WHEN c.doctor_id = $1 THEN u.id = c.patient_id ELSE u.id = c.doctor_id END
		)
		WHERE (c.doctor_id = $1 OR c.patient_id = $1) AND c.deleted_at IS NULL
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var chats []models.UserChatDetail
	err := tx.Txm().SelectContext(ctx, &chats, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return chats, nil
}

func (r *pgChatRepository) GetUserChatsTotalTX(ctx context.Context, tx transaction.Transaction, userID int64) (int64, error) {
	query := `
		SELECT COUNT(1)
		FROM chats
		WHERE (doctor_id = $1 OR patient_id = $1) AND deleted_at IS NULL
	`

	var total int64
	if err := tx.Txm().GetContext(ctx, &total, query, userID); err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}
	return total, nil
}

func (r *pgChatRepository) DeleteMessageTX(ctx context.Context, tx transaction.Transaction, messageID, deletedAt int64) error {
	query := `
		UPDATE messages
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := tx.Txm().ExecContext(ctx, query, deletedAt, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (r *pgChatRepository) CloseChatTX(ctx context.Context, tx transaction.Transaction, chatID, closedAt int64) error {
	query := `
		UPDATE chats
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := tx.Txm().ExecContext(ctx, query, closedAt, chatID)
	if err != nil {
		return fmt.Errorf("failed to close chat: %w", err)
	}
	return nil
}

func (r *pgChatRepository) GetChatByIDTX(ctx context.Context, tx transaction.Transaction, chatID int64) (*models.Chat, error) {
	query := `
		SELECT id, doctor_id, patient_id, created_at, updated_at, deleted_at
		FROM chats
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var chat models.Chat
	if err := tx.Txm().GetContext(ctx, &chat, query, chatID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &chat, nil
}

func (r *pgChatRepository) GetMessageByIDTX(ctx context.Context, tx transaction.Transaction, messageID int64) (*models.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, attachment_url, attachment_type, attachment_mime_type, created_at, deleted_at
		FROM messages
		WHERE id = $1
		LIMIT 1
	`

	var message models.Message
	if err := tx.Txm().GetContext(ctx, &message, query, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &message, nil
}
