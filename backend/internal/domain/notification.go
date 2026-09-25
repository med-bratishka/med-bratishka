package domain

import "encoding/json"

const (
	NotificationTopicChat        = "chat_notifications"
	NotificationEventChatCreated = "chat.message.created"
	WebSocketMessageAuthOK       = "auth_ok"
	WebSocketMessageSubscribed   = "subscribed"
	WebSocketMessageUnsubscribed = "unsubscribed"
	WebSocketMessageNotification = "notification"
	WebSocketMessageError        = "error"
	WebSocketCommandSubscribe    = "subscribe"
	WebSocketCommandUnsubscribe  = "unsubscribe"
)

type OutboxEvent struct {
	ID            string          `db:"id"`
	EventType     string          `db:"event_type"`
	AggregateType string          `db:"aggregate_type"`
	AggregateID   string          `db:"aggregate_id"`
	Payload       json.RawMessage `db:"payload"`
	Attempts      int             `db:"attempts"`
	CreatedAt     int64           `db:"created_at"`
}

type ChatNotificationPayload struct {
	ChatID      int64   `json:"chat_id"`
	MessageID   int64   `json:"message_id"`
	SenderID    int64   `json:"sender_id"`
	RecipientID int64   `json:"recipient_id"`
	Content     *string `json:"content,omitempty"`
	CreatedAt   int64   `json:"created_at"`
}

type WebSocketInboundMessage struct {
	Type  string `json:"type"`
	Topic string `json:"topic,omitempty"`
}

type WebSocketOutboundMessage struct {
	Type  string      `json:"type"`
	Topic string      `json:"topic,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}
