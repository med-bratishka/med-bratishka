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
	ChatID         int64   `json:"chat_id"`
	MessageID      int64   `json:"message_id"`
	SenderID       int64   `json:"sender_id"`
	RecipientID    int64   `json:"recipient_id"`
	RecipientRole  Role    `json:"recipient_role"`
	Content        *string `json:"content,omitempty"`
	AttachmentType *string `json:"attachment_type,omitempty"`
	CreatedAt      int64   `json:"created_at"`
}

type ChatNotificationMessage struct {
	Type           string  `json:"type"`
	ChatID         int64   `json:"chat_id"`
	MessageID      int64   `json:"message_id"`
	SenderID       int64   `json:"sender_id"`
	RecipientID    int64   `json:"recipient_id"`
	RecipientRole  Role    `json:"recipient_role"`
	Content        *string `json:"content,omitempty"`
	AttachmentType *string `json:"attachment_type,omitempty"`
	CreatedAt      int64   `json:"created_at"`
}

type WebPushSubscription struct {
	ID        int64  `db:"id"`
	UserID    int64  `db:"user_id"`
	Endpoint  string `db:"endpoint"`
	P256DH    string `db:"p256dh"`
	Auth      string `db:"auth"`
	UserAgent string `db:"user_agent"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

type WebPushSubscriptionInput struct {
	Endpoint string      `json:"endpoint"`
	Keys     WebPushKeys `json:"keys"`
}

type WebPushKeys struct {
	P256DH string `json:"p256dh"`
	Auth   string `json:"auth"`
}

type WebPushNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
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
