package service

import (
	"testing"

	"medbratishka/internal/domain"
)

func TestValidateWebPushSubscription(t *testing.T) {
	valid := domain.WebPushSubscriptionInput{
		Endpoint: "https://push.example.test/subscription",
		Keys: domain.WebPushKeys{
			P256DH: "p256dh",
			Auth:   "auth",
		},
	}
	if err := validateWebPushSubscription(valid); err != nil {
		t.Fatalf("valid subscription rejected: %v", err)
	}

	valid.Endpoint = "http://push.example.test/subscription"
	if err := validateWebPushSubscription(valid); err == nil {
		t.Fatal("expected non-https endpoint to be rejected")
	}

	valid.Endpoint = "https://127.0.0.1/subscription"
	if err := validateWebPushSubscription(valid); err == nil {
		t.Fatal("expected local endpoint to be rejected")
	}
}

func TestToWebPushNotificationUsesRecipientRouteAndAttachmentFallback(t *testing.T) {
	notification := toWebPushNotification(domain.ChatNotificationMessage{
		ChatID:         17,
		RecipientRole:  domain.RolePatient,
		AttachmentType: strPtr("audio"),
	})

	if notification.Body != "Новое голосовое сообщение" {
		t.Fatalf("unexpected body: %s", notification.Body)
	}
	if notification.URL != "/patient/chat?chat_id=17" {
		t.Fatalf("unexpected url: %s", notification.URL)
	}
	if notification.Tag != "chat-17" {
		t.Fatalf("unexpected tag: %s", notification.Tag)
	}
}

func TestToWebPushNotificationDoesNotExposeMessageContent(t *testing.T) {
	notification := toWebPushNotification(domain.ChatNotificationMessage{
		ChatID:        9,
		RecipientRole: domain.RoleDoctor,
		Content:       strPtr("private medical message"),
	})

	if notification.Body != "Новое сообщение в чате" {
		t.Fatalf("unexpected body: %s", notification.Body)
	}
	if notification.URL != "/doctor/messages?chat_id=9" {
		t.Fatalf("unexpected url: %s", notification.URL)
	}
}
