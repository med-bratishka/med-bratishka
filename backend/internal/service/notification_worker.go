package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	"medbratishka/pkg/logger"
	"medbratishka/pkg/time_manager"
)

type NotificationPublisher interface {
	PublishToUser(userID int64, topic string, payload interface{}) bool
}

type NotificationWorker struct {
	txRepo           transaction.Repository
	notificationRepo repository.NotificationRepository
	publisher        NotificationPublisher
	timeManager      time_manager.TimeManager
	log              logger.Logger
	interval         time.Duration
	batchSize        int
	lockTTL          time.Duration
	maxAttempts      int
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

func NewNotificationWorker(
	txRepo transaction.Repository,
	notificationRepo repository.NotificationRepository,
	publisher NotificationPublisher,
	timeManager time_manager.TimeManager,
	log logger.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		txRepo:           txRepo,
		notificationRepo: notificationRepo,
		publisher:        publisher,
		timeManager:      timeManager,
		log:              log,
		interval:         2 * time.Second,
		batchSize:        50,
		lockTTL:          30 * time.Second,
		maxAttempts:      10,
	}
}

func (w *NotificationWorker) Start(parent context.Context) {
	if w.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	w.cancel = cancel
	w.wg.Add(1)
	go w.run(ctx)
}

func (w *NotificationWorker) Stop() {
	if w.cancel == nil {
		return
	}
	w.cancel()
	w.wg.Wait()
	w.cancel = nil
}

func (w *NotificationWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.process(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *NotificationWorker) process(ctx context.Context) {
	for {
		processed, err := w.processBatch(ctx)
		if err != nil {
			w.log.Warningf("notification worker batch failed: %v", err)
			return
		}
		if processed == 0 {
			return
		}
	}
}

func (w *NotificationWorker) processBatch(ctx context.Context) (int, error) {
	now := w.timeManager.Now().UnixMilli()
	lockUntil := now + w.lockTTL.Milliseconds()

	tx, err := w.txRepo.StartTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback()

	events, err := w.notificationRepo.PickPendingEventsTX(ctx, tx, domain.NotificationEventChatCreated, now, w.batchSize, lockUntil)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit picked events: %w", err)
	}

	for _, event := range events {
		w.handleEvent(ctx, event)
	}
	return len(events), nil
}

func (w *NotificationWorker) handleEvent(ctx context.Context, event domain.OutboxEvent) {
	now := w.timeManager.Now().UnixMilli()

	var payload domain.ChatNotificationPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		w.markFailed(ctx, event.ID, fmt.Errorf("decode payload: %w", err), now)
		return
	}
	if payload.RecipientID == 0 || payload.ChatID == 0 || payload.MessageID == 0 {
		w.markFailed(ctx, event.ID, fmt.Errorf("invalid payload identifiers"), now)
		return
	}

	delivered := w.publisher.PublishToUser(payload.RecipientID, domain.NotificationTopicChat, map[string]interface{}{
		"type":         event.EventType,
		"chat_id":      payload.ChatID,
		"message_id":   payload.MessageID,
		"sender_id":    payload.SenderID,
		"recipient_id": payload.RecipientID,
		"content":      payload.Content,
		"created_at":   payload.CreatedAt,
	})

	tx, err := w.txRepo.StartTransaction(ctx)
	if err != nil {
		w.log.Warningf("notification worker start mark processed tx failed: event_id=%s err=%v", event.ID, err)
		return
	}
	defer tx.Rollback()

	if err := w.notificationRepo.MarkEventProcessedTX(ctx, tx, event.ID, delivered, now); err != nil {
		w.log.Warningf("notification worker mark processed failed: event_id=%s err=%v", event.ID, err)
		return
	}
	if err := tx.Commit(); err != nil {
		w.log.Warningf("notification worker commit mark processed failed: event_id=%s err=%v", event.ID, err)
		return
	}
}

func (w *NotificationWorker) markFailed(ctx context.Context, eventID string, cause error, now int64) {
	backoff := int64(time.Second.Milliseconds())
	nextAttemptAt := now + backoff

	tx, err := w.txRepo.StartTransaction(ctx)
	if err != nil {
		w.log.Warningf("notification worker start mark failed tx failed: event_id=%s err=%v", eventID, err)
		return
	}
	defer tx.Rollback()

	if err := w.notificationRepo.MarkEventFailedTX(ctx, tx, eventID, cause.Error(), now, nextAttemptAt, w.maxAttempts); err != nil {
		w.log.Warningf("notification worker mark failed failed: event_id=%s err=%v", eventID, err)
		return
	}
	if err := tx.Commit(); err != nil {
		w.log.Warningf("notification worker commit mark failed failed: event_id=%s err=%v", eventID, err)
	}
}
