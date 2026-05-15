package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"medbratishka/internal/repository/models"
	"medbratishka/internal/repository/transaction"

	"github.com/google/uuid"
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
	GetLatestMessageIDTX(ctx context.Context, tx transaction.Transaction, chatID int64) (int64, error)
	DeleteMessageTX(ctx context.Context, tx transaction.Transaction, messageID, deletedAt int64) error
	CloseChatTX(ctx context.Context, tx transaction.Transaction, chatID, closedAt int64) error
	GetChatByIDTX(ctx context.Context, tx transaction.Transaction, chatID int64) (*models.Chat, error)
	GetMessageByIDTX(ctx context.Context, tx transaction.Transaction, messageID int64) (*models.Message, error)
	CreateMessageNotificationTX(ctx context.Context, tx transaction.Transaction, chat *models.Chat, messageID, senderID int64, content *string, createdAt int64) error
	MarkChatReadTX(ctx context.Context, tx transaction.Transaction, chatID, userID, lastReadMessageID, updatedAt int64) error
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
			c.updated_at,
			lm.id as last_message_id,
			COALESCE(lm.content, CASE WHEN lm.attachment_url IS NOT NULL THEN 'Вложение' ELSE '' END) as last_message,
			lm.created_at as last_message_at,
			COALESCE(crs.unread_count, 0) as unread_count,
			crs.last_read_message_id as last_read_message_id
		FROM chats c
		JOIN users u ON (
			CASE WHEN c.doctor_id = $1 THEN u.id = c.patient_id ELSE u.id = c.doctor_id END
		)
		LEFT JOIN chat_read_state crs ON crs.chat_id = c.id AND crs.user_id = $1
		LEFT JOIN LATERAL (
			SELECT id, content, attachment_url, created_at
			FROM messages
			WHERE chat_id = c.id AND deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT 1
		) lm ON TRUE
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

func (r *pgChatRepository) GetLatestMessageIDTX(ctx context.Context, tx transaction.Transaction, chatID int64) (int64, error) {
	var id int64
	err := tx.Txm().QueryRowContext(ctx, `
		SELECT id
		FROM messages
		WHERE chat_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, chatID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get latest message id: %w", err)
	}
	return id, nil
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

func (r *pgChatRepository) CreateMessageNotificationTX(ctx context.Context, tx transaction.Transaction, chat *models.Chat, messageID, senderID int64, content *string, createdAt int64) error {
	recipientID := chat.PatientID
	if senderID == chat.PatientID {
		recipientID = chat.DoctorID
	}
	eventID := uuid.New().String()
	idempotencyKey := fmt.Sprintf("chat.message.created:%d:%d", messageID, recipientID)
	payload, err := json.Marshal(map[string]interface{}{
		"chat_id":      chat.ID,
		"message_id":   messageID,
		"sender_id":    senderID,
		"recipient_id": recipientID,
		"content":      content,
		"created_at":   createdAt,
	})
	if err != nil {
		return err
	}

	if _, err := tx.Txm().ExecContext(ctx, `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, idempotency_key, payload, created_at)
		VALUES ($1, 'chat', $2, 'chat.message.created', $3, $4::jsonb, $5)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, eventID, strconv.FormatInt(chat.ID, 10), idempotencyKey, string(payload), createdAt); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	if _, err := tx.Txm().ExecContext(ctx, `
		INSERT INTO chat_notification_deliveries (
			event_id, chat_id, message_id, recipient_id, sender_id, status, created_at, idempotency_key
		)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, eventID, chat.ID, messageID, recipientID, senderID, createdAt, idempotencyKey); err != nil {
		return fmt.Errorf("insert chat notification delivery: %w", err)
	}

	if _, err := tx.Txm().ExecContext(ctx, `
		INSERT INTO chat_read_state (chat_id, user_id, last_read_message_id, unread_count, updated_at)
		VALUES ($1, $2, $3, 0, $4)
		ON CONFLICT (chat_id, user_id) DO UPDATE
		SET last_read_message_id = GREATEST(COALESCE(chat_read_state.last_read_message_id, 0), EXCLUDED.last_read_message_id),
		    updated_at = EXCLUDED.updated_at
	`, chat.ID, senderID, messageID, createdAt); err != nil {
		return fmt.Errorf("upsert sender read state: %w", err)
	}

	if _, err := tx.Txm().ExecContext(ctx, `
		INSERT INTO chat_read_state (chat_id, user_id, unread_count, updated_at)
		VALUES ($1, $2, 1, $3)
		ON CONFLICT (chat_id, user_id) DO UPDATE
		SET unread_count = chat_read_state.unread_count + 1,
		    updated_at = EXCLUDED.updated_at
	`, chat.ID, recipientID, createdAt); err != nil {
		return fmt.Errorf("upsert recipient read state: %w", err)
	}

	return nil
}

func (r *pgChatRepository) MarkChatReadTX(ctx context.Context, tx transaction.Transaction, chatID, userID, lastReadMessageID, updatedAt int64) error {
	if _, err := tx.Txm().ExecContext(ctx, `
		INSERT INTO chat_read_state (chat_id, user_id, last_read_message_id, unread_count, updated_at)
		VALUES ($1, $2, $3, 0, $4)
		ON CONFLICT (chat_id, user_id) DO UPDATE
		SET last_read_message_id = GREATEST(COALESCE(chat_read_state.last_read_message_id, 0), EXCLUDED.last_read_message_id),
		    last_read_at = EXCLUDED.updated_at,
		    updated_at = EXCLUDED.updated_at
	`, chatID, userID, lastReadMessageID, updatedAt); err != nil {
		return fmt.Errorf("mark chat read state: %w", err)
	}
	if _, err := tx.Txm().ExecContext(ctx, `
		UPDATE chat_notification_deliveries d
		SET status = 'read', read_at = $1
		WHERE d.chat_id = $2
		  AND d.recipient_id = $3
		  AND d.read_at IS NULL
		  AND d.message_id <= $4
	`, updatedAt, chatID, userID, lastReadMessageID); err != nil {
		return fmt.Errorf("mark deliveries read: %w", err)
	}
	if _, err := tx.Txm().ExecContext(ctx, `
		UPDATE chat_read_state
		SET unread_count = (
			SELECT COUNT(1)
			FROM chat_notification_deliveries d
			WHERE d.chat_id = $1
			  AND d.recipient_id = $2
			  AND d.read_at IS NULL
		),
		    updated_at = $3
		WHERE chat_id = $1 AND user_id = $2
	`, chatID, userID, updatedAt); err != nil {
		return fmt.Errorf("recalculate unread count: %w", err)
	}
	return nil
}
