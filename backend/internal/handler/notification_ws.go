package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/service"
	"medbratishka/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	websocketWriteWait      = 10 * time.Second
	websocketPongWait       = 60 * time.Second
	websocketPingPeriod     = (websocketPongWait * 9) / 10
	websocketMaxMessageSize = 4096
	websocketSendBufferSize = 32
)

type NotificationWSHandler struct {
	authService service.AuthService
	hub         *NotificationHub
	log         logger.Logger
	upgrader    websocket.Upgrader
}

func NewNotificationWSHandler(authService service.AuthService, hub *NotificationHub, log logger.Logger) *NotificationWSHandler {
	return &NotificationWSHandler{
		authService: authService,
		hub:         hub,
		log:         log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     allowWebSocketOrigin,
		},
	}
}

func (h *NotificationWSHandler) FillHandlers(router *mux.Router) {
	router.HandleFunc("/ws/notifications", h.Connect).Methods(http.MethodGet)
}

func (h *NotificationWSHandler) Shutdown() {}

// Connect godoc
// @Summary Subscribe to realtime notifications
// @Description Upgrades HTTP to WebSocket. Pass access token in query parameter `token` from browser clients, then send `{ "type": "subscribe", "topic": "chat_notifications" }`.
// @Tags notifications
// @Produce json
// @Param token query string true "Access token"
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} models.ErrorResponse
// @Router /ws/notifications [get]
func (h *NotificationWSHandler) Connect(w http.ResponseWriter, r *http.Request) {
	tokenString := extractWebSocketToken(r)
	if tokenString == "" {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
		return
	}

	userCtx, err := h.authService.ValidateToken(r.Context(), domain.PurposeAccess, tokenString)
	if err != nil {
		makeErrorResponse(w, r, h.log, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", err)
		return
	}
	if userCtx.Role != domain.RoleDoctor && userCtx.Role != domain.RolePatient {
		makeErrorResponse(w, r, h.log, http.StatusForbidden, "FORBIDDEN", "forbidden", nil)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warningf("notifications websocket upgrade failed: %v", err)
		return
	}

	client := newNotificationClient(h.hub, conn, userCtx, h.log)
	h.hub.register(client)
	client.send <- domain.WebSocketOutboundMessage{Type: domain.WebSocketMessageAuthOK}

	go client.writePump()
	go client.readPump()
}

type NotificationHub struct {
	mu      sync.RWMutex
	clients map[int64]map[*notificationClient]struct{}
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{clients: make(map[int64]map[*notificationClient]struct{})}
}

func (h *NotificationHub) PublishToUser(userID int64, topic string, payload interface{}) bool {
	msg := domain.WebSocketOutboundMessage{
		Type:  domain.WebSocketMessageNotification,
		Topic: topic,
		Data:  payload,
	}

	h.mu.RLock()
	clients := h.clients[userID]
	targets := make([]*notificationClient, 0, len(clients))
	for client := range clients {
		if client.isSubscribed(topic) {
			targets = append(targets, client)
		}
	}
	h.mu.RUnlock()

	delivered := false
	for _, client := range targets {
		select {
		case client.send <- msg:
			delivered = true
		default:
			client.close()
		}
	}
	return delivered
}

func (h *NotificationHub) register(client *notificationClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.user.ID] == nil {
		h.clients[client.user.ID] = make(map[*notificationClient]struct{})
	}
	h.clients[client.user.ID][client] = struct{}{}
}

func (h *NotificationHub) unregister(client *notificationClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.clients[client.user.ID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.user.ID)
		}
	}
}

type notificationClient struct {
	hub    *NotificationHub
	conn   *websocket.Conn
	user   *domain.UserTokenContext
	send   chan domain.WebSocketOutboundMessage
	done   chan struct{}
	topics map[string]struct{}
	mu     sync.RWMutex
	once   sync.Once
	log    logger.Logger
}

func newNotificationClient(hub *NotificationHub, conn *websocket.Conn, user *domain.UserTokenContext, log logger.Logger) *notificationClient {
	return &notificationClient{
		hub:    hub,
		conn:   conn,
		user:   user,
		send:   make(chan domain.WebSocketOutboundMessage, websocketSendBufferSize),
		done:   make(chan struct{}),
		topics: map[string]struct{}{domain.NotificationTopicChat: {}},
		log:    log,
	}
}

func (c *notificationClient) readPump() {
	defer c.close()

	c.conn.SetReadLimit(websocketMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(websocketPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(websocketPongWait))
	})

	for {
		var msg domain.WebSocketInboundMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.log.Debugf("notifications websocket read closed: %v", err)
			}
			return
		}
		c.handleMessage(msg)
	}
}

func (c *notificationClient) writePump() {
	ticker := time.NewTicker(websocketPingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-c.done:
			return
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *notificationClient) handleMessage(msg domain.WebSocketInboundMessage) {
	topic := strings.TrimSpace(msg.Topic)
	if topic == "" {
		topic = domain.NotificationTopicChat
	}
	if topic != domain.NotificationTopicChat {
		c.sendError("unsupported topic")
		return
	}

	switch msg.Type {
	case domain.WebSocketCommandSubscribe:
		c.mu.Lock()
		c.topics[topic] = struct{}{}
		c.mu.Unlock()
		c.send <- domain.WebSocketOutboundMessage{Type: domain.WebSocketMessageSubscribed, Topic: topic}
	case domain.WebSocketCommandUnsubscribe:
		c.mu.Lock()
		delete(c.topics, topic)
		c.mu.Unlock()
		c.send <- domain.WebSocketOutboundMessage{Type: domain.WebSocketMessageUnsubscribed, Topic: topic}
	default:
		c.sendError("unsupported message type")
	}
}

func (c *notificationClient) isSubscribed(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.topics[topic]
	return ok
}

func (c *notificationClient) sendError(message string) {
	select {
	case c.send <- domain.WebSocketOutboundMessage{Type: domain.WebSocketMessageError, Error: message}:
	default:
		c.close()
	}
}

func (c *notificationClient) close() {
	c.once.Do(func() {
		c.hub.unregister(c)
		close(c.done)
		_ = c.conn.Close()
	})
}

func extractWebSocketToken(r *http.Request) string {
	if tokenString := strings.TrimSpace(r.URL.Query().Get("token")); tokenString != "" {
		return tokenString
	}
	return extractToken(r)
}

func allowWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
		return true
	}
	return strings.Contains(origin, r.Host)
}
