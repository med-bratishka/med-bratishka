package repository

import (
	"context"
	"fmt"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository/transaction"
)

type WebPushSubscriptionRepository interface {
	UpsertWebPushSubscriptionTX(ctx context.Context, tx transaction.Transaction, userID int64, input domain.WebPushSubscriptionInput, userAgent string, now int64) error
	DeleteWebPushSubscriptionTX(ctx context.Context, tx transaction.Transaction, userID int64, endpoint string) error
	DeleteWebPushSubscriptionByEndpointTX(ctx context.Context, tx transaction.Transaction, endpoint string) error
	GetWebPushSubscriptionsTX(ctx context.Context, tx transaction.Transaction, userID int64) ([]domain.WebPushSubscription, error)
}

func NewWebPushSubscriptionRepository() WebPushSubscriptionRepository {
	return &pgNotificationRepository{}
}

func (r *pgNotificationRepository) UpsertWebPushSubscriptionTX(ctx context.Context, tx transaction.Transaction, userID int64, input domain.WebPushSubscriptionInput, userAgent string, now int64) error {
	if _, err := tx.Txm().ExecContext(ctx, `
		INSERT INTO web_push_subscriptions (user_id, endpoint, p256dh, auth, user_agent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (endpoint) DO UPDATE
		SET user_id = EXCLUDED.user_id,
		    p256dh = EXCLUDED.p256dh,
		    auth = EXCLUDED.auth,
		    user_agent = EXCLUDED.user_agent,
		    updated_at = EXCLUDED.updated_at
	`, userID, input.Endpoint, input.Keys.P256DH, input.Keys.Auth, userAgent, now); err != nil {
		return fmt.Errorf("upsert web push subscription: %w", err)
	}
	return nil
}

func (r *pgNotificationRepository) DeleteWebPushSubscriptionTX(ctx context.Context, tx transaction.Transaction, userID int64, endpoint string) error {
	if _, err := tx.Txm().ExecContext(ctx, `
		DELETE FROM web_push_subscriptions
		WHERE user_id = $1 AND endpoint = $2
	`, userID, endpoint); err != nil {
		return fmt.Errorf("delete web push subscription: %w", err)
	}
	return nil
}

func (r *pgNotificationRepository) DeleteWebPushSubscriptionByEndpointTX(ctx context.Context, tx transaction.Transaction, endpoint string) error {
	if _, err := tx.Txm().ExecContext(ctx, `
		DELETE FROM web_push_subscriptions
		WHERE endpoint = $1
	`, endpoint); err != nil {
		return fmt.Errorf("delete stale web push subscription: %w", err)
	}
	return nil
}

func (r *pgNotificationRepository) GetWebPushSubscriptionsTX(ctx context.Context, tx transaction.Transaction, userID int64) ([]domain.WebPushSubscription, error) {
	var subscriptions []domain.WebPushSubscription
	if err := tx.Txm().SelectContext(ctx, &subscriptions, `
		SELECT id, user_id, endpoint, p256dh, auth, user_agent, created_at, updated_at
		FROM web_push_subscriptions
		WHERE user_id = $1
		ORDER BY id
	`, userID); err != nil {
		return nil, fmt.Errorf("get web push subscriptions: %w", err)
	}
	return subscriptions, nil
}
