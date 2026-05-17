package handler

import (
	"testing"
	"time"

	"medbratishka/internal/domain"
)

func TestNotificationHubPublishToSubscribedUser(t *testing.T) {
	hub := NewNotificationHub()
	client := &notificationClient{
		hub:    hub,
		user:   &domain.UserTokenContext{ID: 42, Role: domain.RolePatient},
		send:   make(chan domain.WebSocketOutboundMessage, 1),
		topics: map[string]struct{}{domain.NotificationTopicChat: {}},
	}
	hub.register(client)

	delivered := hub.PublishToUser(42, domain.NotificationTopicChat, map[string]int{"message_id": 7})
	if !delivered {
		t.Fatal("expected notification to be delivered")
	}

	select {
	case msg := <-client.send:
		if msg.Type != domain.WebSocketMessageNotification {
			t.Fatalf("unexpected message type: %s", msg.Type)
		}
		if msg.Topic != domain.NotificationTopicChat {
			t.Fatalf("unexpected topic: %s", msg.Topic)
		}
	case <-time.After(time.Second):
		t.Fatal("expected notification message")
	}
}

func TestNotificationHubSkipsUnsubscribedUser(t *testing.T) {
	hub := NewNotificationHub()
	client := &notificationClient{
		hub:    hub,
		user:   &domain.UserTokenContext{ID: 42, Role: domain.RolePatient},
		send:   make(chan domain.WebSocketOutboundMessage, 1),
		topics: map[string]struct{}{},
	}
	hub.register(client)

	delivered := hub.PublishToUser(42, domain.NotificationTopicChat, map[string]int{"message_id": 7})
	if delivered {
		t.Fatal("expected notification to be skipped")
	}
}
