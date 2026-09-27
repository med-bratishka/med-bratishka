package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	"medbratishka/pkg/logger"
	"medbratishka/pkg/time_manager"

	webpush "github.com/SherClockHolmes/webpush-go"
)

var ErrWebPushSubscriptionInvalid = errors.New("invalid web push subscription")

type WebPushService interface {
	PublicKey() string
	Enabled() bool
	Subscribe(ctx context.Context, userID int64, input domain.WebPushSubscriptionInput, userAgent string) error
	Unsubscribe(ctx context.Context, userID int64, endpoint string) error
	PublishToUser(userID int64, topic string, payload interface{}) bool
}

type webPushService struct {
	txRepo      transaction.Repository
	repo        repository.WebPushSubscriptionRepository
	timeManager time_manager.TimeManager
	log         logger.Logger
	publicKey   string
	privateKey  string
	subject     string
	httpClient  *http.Client
}

func NewWebPushService(
	txRepo transaction.Repository,
	repo repository.WebPushSubscriptionRepository,
	timeManager time_manager.TimeManager,
	log logger.Logger,
	publicKey, privateKey, subject string,
) WebPushService {
	return &webPushService{
		txRepo:      txRepo,
		repo:        repo,
		timeManager: timeManager,
		log:         log,
		publicKey:   strings.TrimSpace(publicKey),
		privateKey:  strings.TrimSpace(privateKey),
		subject:     strings.TrimSpace(subject),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *webPushService) PublicKey() string {
	return s.publicKey
}

func (s *webPushService) Enabled() bool {
	return s.publicKey != "" && s.privateKey != "" && s.subject != ""
}

func (s *webPushService) Subscribe(ctx context.Context, userID int64, input domain.WebPushSubscriptionInput, userAgent string) error {
	if !s.Enabled() {
		return newServiceError(CodeInternal, fmt.Errorf("web push is not configured"), "WEB_PUSH_DISABLED", "web push is disabled")
	}
	input.Endpoint = strings.TrimSpace(input.Endpoint)
	input.Keys.P256DH = strings.TrimSpace(input.Keys.P256DH)
	input.Keys.Auth = strings.TrimSpace(input.Keys.Auth)
	if err := validateWebPushSubscription(input); err != nil {
		return newServiceError(CodeBadRequest, err, "INVALID_WEB_PUSH_SUBSCRIPTION", "invalid web push subscription")
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("WebPushSubscribe/StartTransaction", err)
	}
	defer tx.Rollback()

	if err := s.repo.UpsertWebPushSubscriptionTX(ctx, tx, userID, input, strings.TrimSpace(userAgent), s.timeManager.Now().UnixMilli()); err != nil {
		return wrapInternal("WebPushSubscribe/Upsert", err)
	}
	if err := tx.Commit(); err != nil {
		return wrapInternal("WebPushSubscribe/Commit", err)
	}
	return nil
}

func (s *webPushService) Unsubscribe(ctx context.Context, userID int64, endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return newServiceError(CodeBadRequest, ErrWebPushSubscriptionInvalid, "INVALID_WEB_PUSH_SUBSCRIPTION", "invalid web push subscription")
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("WebPushUnsubscribe/StartTransaction", err)
	}
	defer tx.Rollback()

	if err := s.repo.DeleteWebPushSubscriptionTX(ctx, tx, userID, endpoint); err != nil {
		return wrapInternal("WebPushUnsubscribe/Delete", err)
	}
	if err := tx.Commit(); err != nil {
		return wrapInternal("WebPushUnsubscribe/Commit", err)
	}
	return nil
}

func (s *webPushService) PublishToUser(userID int64, topic string, payload interface{}) bool {
	if !s.Enabled() || topic != domain.NotificationTopicChat {
		return false
	}

	message, ok := payload.(domain.ChatNotificationMessage)
	if !ok {
		s.log.Warningf("web push skipped unsupported payload: %T", payload)
		return false
	}

	notification := toWebPushNotification(message)
	body, err := json.Marshal(notification)
	if err != nil {
		s.log.Warningf("web push marshal failed: %v", err)
		return false
	}

	subscriptions, err := s.getSubscriptions(context.Background(), userID)
	if err != nil {
		s.log.Warningf("web push subscriptions lookup failed: user_id=%d err=%v", userID, err)
		return false
	}

	delivered := false
	for _, subscription := range subscriptions {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, sendErr := webpush.SendNotificationWithContext(ctx, body, &webpush.Subscription{
			Endpoint: subscription.Endpoint,
			Keys: webpush.Keys{
				P256dh: subscription.P256DH,
				Auth:   subscription.Auth,
			},
		}, &webpush.Options{
			HTTPClient:      s.httpClient,
			Subscriber:      s.subject,
			VAPIDPublicKey:  s.publicKey,
			VAPIDPrivateKey: s.privateKey,
			TTL:             24 * 60 * 60,
			Topic:           fmt.Sprintf("chat-%d", message.ChatID),
		})
		cancel()
		if sendErr != nil {
			s.log.Warningf("web push send failed: user_id=%d err=%v", userID, sendErr)
			continue
		}
		_ = resp.Body.Close()

		switch {
		case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
			delivered = true
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
			if err := s.deleteStaleSubscription(context.Background(), subscription.Endpoint); err != nil {
				s.log.Warningf("web push stale subscription cleanup failed: %v", err)
			}
		default:
			s.log.Warningf("web push rejected: user_id=%d status=%d", userID, resp.StatusCode)
		}
	}
	return delivered
}

func (s *webPushService) getSubscriptions(ctx context.Context, userID int64) ([]domain.WebPushSubscription, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	subscriptions, err := s.repo.GetWebPushSubscriptionsTX(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (s *webPushService) deleteStaleSubscription(ctx context.Context, endpoint string) error {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.DeleteWebPushSubscriptionByEndpointTX(ctx, tx, endpoint); err != nil {
		return err
	}
	return tx.Commit()
}

func validateWebPushSubscription(input domain.WebPushSubscriptionInput) error {
	const maxSubscriptionValueLength = 4096
	rawEndpoint := strings.TrimSpace(input.Endpoint)
	p256dh := strings.TrimSpace(input.Keys.P256DH)
	auth := strings.TrimSpace(input.Keys.Auth)
	if len(rawEndpoint) > maxSubscriptionValueLength || len(p256dh) > maxSubscriptionValueLength || len(auth) > maxSubscriptionValueLength {
		return ErrWebPushSubscriptionInvalid
	}

	endpoint, err := url.ParseRequestURI(rawEndpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" {
		return ErrWebPushSubscriptionInvalid
	}
	host := strings.ToLower(endpoint.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return ErrWebPushSubscriptionInvalid
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()) {
		return ErrWebPushSubscriptionInvalid
	}
	if p256dh == "" || auth == "" {
		return ErrWebPushSubscriptionInvalid
	}
	return nil
}

func toWebPushNotification(message domain.ChatNotificationMessage) domain.WebPushNotification {
	body := "Новое сообщение в чате"
	switch derefStr(message.AttachmentType) {
	case "image":
		body = "Новое изображение"
	case "audio":
		body = "Новое голосовое сообщение"
	case "file":
		body = "Новое вложение"
	}

	path := "/doctor/messages"
	if message.RecipientRole == domain.RolePatient {
		path = "/patient/chat"
	}
	return domain.WebPushNotification{
		Title: "Новое сообщение",
		Body:  body,
		URL:   fmt.Sprintf("%s?chat_id=%d", path, message.ChatID),
		Tag:   fmt.Sprintf("chat-%d", message.ChatID),
	}
}
