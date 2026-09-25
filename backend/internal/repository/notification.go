package repository

import (
	"context"
	"fmt"
	"strings"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository/transaction"
)

type NotificationRepository interface {
	PickPendingEventsTX(ctx context.Context, tx transaction.Transaction, eventType string, now int64, limit int, lockUntil int64) ([]domain.OutboxEvent, error)
	MarkEventProcessedTX(ctx context.Context, tx transaction.Transaction, eventID string, delivered bool, deliveredAt int64) error
	MarkEventFailedTX(ctx context.Context, tx transaction.Transaction, eventID string, errMsg string, now int64, nextAttemptAt int64, maxAttempts int) error
}

type pgNotificationRepository struct{}

func NewNotificationRepository() NotificationRepository {
	return &pgNotificationRepository{}
}

func (r *pgNotificationRepository) PickPendingEventsTX(ctx context.Context, tx transaction.Transaction, eventType string, now int64, limit int, lockUntil int64) ([]domain.OutboxEvent, error) {
	query := `
		WITH candidates AS (
			SELECT id
			FROM outbox_events
			WHERE event_type = $1
			  AND (
				(status = 'pending' AND COALESCE(next_attempt_at, 0) <= $2)
				OR (status = 'processing' AND COALESCE(locked_until, 0) <= $2)
			  )
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT $3
		)
		UPDATE outbox_events e
		SET status = 'processing',
		    locked_until = $4,
		    attempts = e.attempts + 1,
		    updated_at = $2
		FROM candidates
		WHERE e.id = candidates.id
		RETURNING e.id, e.event_type, e.aggregate_type, e.aggregate_id, e.payload, e.attempts, e.created_at
	`
	var events []domain.OutboxEvent
	if err := tx.Txm().SelectContext(ctx, &events, query, eventType, now, limit, lockUntil); err != nil {
		return nil, fmt.Errorf("pick pending events: %w", err)
	}
	return events, nil
}

func (r *pgNotificationRepository) MarkEventProcessedTX(ctx context.Context, tx transaction.Transaction, eventID string, delivered bool, deliveredAt int64) error {
	if delivered {
		if _, err := tx.Txm().ExecContext(ctx, `
			UPDATE chat_notification_deliveries
			SET status = 'delivered',
			    delivered_at = COALESCE(delivered_at, $2)
			WHERE event_id = $1
			  AND status = 'pending'
			  AND read_at IS NULL
		`, eventID, deliveredAt); err != nil {
			return fmt.Errorf("mark delivery delivered: %w", err)
		}
	}

	if _, err := tx.Txm().ExecContext(ctx, `
		UPDATE outbox_events
		SET status = 'processed',
		    processed_at = $2,
		    locked_until = NULL,
		    next_attempt_at = NULL,
		    updated_at = $2
		WHERE id = $1
	`, eventID, deliveredAt); err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}
	return nil
}

func (r *pgNotificationRepository) MarkEventFailedTX(ctx context.Context, tx transaction.Transaction, eventID string, errMsg string, now int64, nextAttemptAt int64, maxAttempts int) error {
	errMsg = strings.TrimSpace(errMsg)
	if len(errMsg) > 2000 {
		errMsg = errMsg[:2000]
	}

	if _, err := tx.Txm().ExecContext(ctx, `
		UPDATE outbox_events
		SET status = CASE WHEN attempts >= $5 THEN 'failed' ELSE 'pending' END,
		    next_attempt_at = CASE WHEN attempts >= $5 THEN NULL ELSE $4 END,
		    locked_until = NULL,
		    last_error = $2,
		    updated_at = $3
		WHERE id = $1
	`, eventID, errMsg, now, nextAttemptAt, maxAttempts); err != nil {
		return fmt.Errorf("mark event failed: %w", err)
	}
	return nil
}
