package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository/transaction"
)

func TestNotificationWorkerProcessesOnlineChatEvent(t *testing.T) {
	repo := &fakeNotificationRepo{}
	publisher := &fakeNotificationPublisher{delivered: true}
	worker := newTestNotificationWorker(repo, publisher)
	event := newChatOutboxEvent(t)

	worker.handleEvent(context.Background(), event)

	if publisher.userID != 22 {
		t.Fatalf("expected publish to user 22, got %d", publisher.userID)
	}
	if publisher.topic != domain.NotificationTopicChat {
		t.Fatalf("unexpected topic: %s", publisher.topic)
	}
	if repo.processedEventID != event.ID {
		t.Fatalf("expected event processed, got %s", repo.processedEventID)
	}
	if !repo.processedDelivered {
		t.Fatal("expected delivery status to be marked delivered")
	}
	if repo.failedEventID != "" {
		t.Fatalf("did not expect failed event, got %s", repo.failedEventID)
	}
}

func TestNotificationWorkerProcessesOfflineChatEventWithoutDeliveredStatus(t *testing.T) {
	repo := &fakeNotificationRepo{}
	publisher := &fakeNotificationPublisher{delivered: false}
	worker := newTestNotificationWorker(repo, publisher)
	event := newChatOutboxEvent(t)

	worker.handleEvent(context.Background(), event)

	if publisher.calls != 1 {
		t.Fatalf("expected one publish attempt, got %d", publisher.calls)
	}
	if repo.processedEventID != event.ID {
		t.Fatalf("expected event processed, got %s", repo.processedEventID)
	}
	if repo.processedDelivered {
		t.Fatal("did not expect offline delivery to be marked delivered")
	}
}

func TestNotificationWorkerMarksInvalidPayloadFailed(t *testing.T) {
	repo := &fakeNotificationRepo{}
	publisher := &fakeNotificationPublisher{delivered: true}
	worker := newTestNotificationWorker(repo, publisher)

	worker.handleEvent(context.Background(), domain.OutboxEvent{
		ID:        "bad-event",
		EventType: domain.NotificationEventChatCreated,
		Payload:   []byte(`{"chat_id":0}`),
	})

	if publisher.calls != 0 {
		t.Fatalf("did not expect publish attempt, got %d", publisher.calls)
	}
	if repo.failedEventID != "bad-event" {
		t.Fatalf("expected failed event bad-event, got %s", repo.failedEventID)
	}
	if repo.failedError == "" {
		t.Fatal("expected failure error text")
	}
	if repo.processedEventID != "" {
		t.Fatalf("did not expect processed event, got %s", repo.processedEventID)
	}
}

func TestNotificationWorkerProcessBatchPicksAndHandlesEvents(t *testing.T) {
	event := newChatOutboxEvent(t)
	repo := &fakeNotificationRepo{pickedEvents: []domain.OutboxEvent{event}}
	publisher := &fakeNotificationPublisher{delivered: true}
	worker := newTestNotificationWorker(repo, publisher)

	processed, err := worker.processBatch(context.Background())
	if err != nil {
		t.Fatalf("processBatch returned error: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected one processed event, got %d", processed)
	}
	if repo.pickEventType != domain.NotificationEventChatCreated {
		t.Fatalf("unexpected picked event type: %s", repo.pickEventType)
	}
	if repo.processedEventID != event.ID {
		t.Fatalf("expected processed event %s, got %s", event.ID, repo.processedEventID)
	}
}

func newTestNotificationWorker(repo *fakeNotificationRepo, publisher *fakeNotificationPublisher) *NotificationWorker {
	return &NotificationWorker{
		txRepo:           &fakeTxRepo{},
		notificationRepo: repo,
		publisher:        publisher,
		timeManager:      fakeTimeManager{now: time.Unix(1710000000, 0)},
		interval:         time.Second,
		batchSize:        10,
		lockTTL:          30 * time.Second,
		maxAttempts:      3,
	}
}

func newChatOutboxEvent(t *testing.T) domain.OutboxEvent {
	t.Helper()
	payload, err := json.Marshal(domain.ChatNotificationPayload{
		ChatID:      3,
		MessageID:   15,
		SenderID:    11,
		RecipientID: 22,
		Content:     strPtr("hello"),
		CreatedAt:   1710000000000,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return domain.OutboxEvent{
		ID:        "event-1",
		EventType: domain.NotificationEventChatCreated,
		Payload:   payload,
	}
}

func strPtr(value string) *string {
	return &value
}

type fakeNotificationRepo struct {
	pickedEvents  []domain.OutboxEvent
	pickEventType string

	processedEventID     string
	processedDelivered   bool
	processedDeliveredAt int64

	failedEventID       string
	failedError         string
	failedNextAttemptAt int64
}

func (r *fakeNotificationRepo) PickPendingEventsTX(ctx context.Context, tx transaction.Transaction, eventType string, now int64, limit int, lockUntil int64) ([]domain.OutboxEvent, error) {
	r.pickEventType = eventType
	return append([]domain.OutboxEvent(nil), r.pickedEvents...), nil
}

func (r *fakeNotificationRepo) MarkEventProcessedTX(ctx context.Context, tx transaction.Transaction, eventID string, delivered bool, deliveredAt int64) error {
	r.processedEventID = eventID
	r.processedDelivered = delivered
	r.processedDeliveredAt = deliveredAt
	return nil
}

func (r *fakeNotificationRepo) MarkEventFailedTX(ctx context.Context, tx transaction.Transaction, eventID string, errMsg string, now int64, nextAttemptAt int64, maxAttempts int) error {
	if eventID == "force-error" {
		return errors.New("forced error")
	}
	r.failedEventID = eventID
	r.failedError = errMsg
	r.failedNextAttemptAt = nextAttemptAt
	return nil
}

type fakeNotificationPublisher struct {
	delivered bool
	calls     int
	userID    int64
	topic     string
	payload   interface{}
}

func (p *fakeNotificationPublisher) PublishToUser(userID int64, topic string, payload interface{}) bool {
	p.calls++
	p.userID = userID
	p.topic = topic
	p.payload = payload
	return p.delivered
}
