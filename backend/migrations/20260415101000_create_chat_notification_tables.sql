-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox_events
(
    id              UUID   PRIMARY KEY,
    aggregate_type  TEXT   NOT NULL,
    aggregate_id    TEXT   NOT NULL,
    event_type      TEXT   NOT NULL,
    idempotency_key TEXT   NOT NULL UNIQUE,
    payload         JSONB  NOT NULL,
    status          TEXT   NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'processed', 'failed')),
    attempts        INT    NOT NULL DEFAULT 0,
    next_attempt_at BIGINT,
    locked_until    BIGINT,
    created_at      BIGINT NOT NULL,
    updated_at      BIGINT NOT NULL DEFAULT 0,
    processed_at    BIGINT,
    last_error      TEXT
);

CREATE INDEX idx_outbox_events_pickup
    ON outbox_events (status, (COALESCE(next_attempt_at, 0)), created_at)
    WHERE status IN ('pending', 'processing');

CREATE INDEX idx_outbox_events_type_status
    ON outbox_events (event_type, status, created_at);

CREATE TABLE chat_notification_deliveries
(
    id              BIGSERIAL PRIMARY KEY,
    event_id        UUID   NOT NULL REFERENCES outbox_events (id) ON DELETE CASCADE,
    chat_id         BIGINT NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    message_id      BIGINT NOT NULL REFERENCES messages (id) ON DELETE CASCADE,
    recipient_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    sender_id       BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status          TEXT   NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'delivered', 'read', 'failed')),
    created_at      BIGINT NOT NULL,
    delivered_at    BIGINT,
    read_at         BIGINT,
    idempotency_key TEXT   NOT NULL UNIQUE,

    UNIQUE (message_id, recipient_id)
);

CREATE INDEX idx_chat_notification_deliveries_recipient_status
    ON chat_notification_deliveries (recipient_id, status, created_at);

CREATE INDEX idx_chat_notification_deliveries_chat_unread
    ON chat_notification_deliveries (chat_id, recipient_id, read_at);

CREATE TABLE chat_read_state
(
    id                   BIGSERIAL PRIMARY KEY,
    chat_id              BIGINT NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    user_id              BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    last_read_message_id BIGINT,
    last_read_at         BIGINT,
    unread_count         INT    NOT NULL DEFAULT 0,
    updated_at           BIGINT NOT NULL,

    UNIQUE (chat_id, user_id)
);

CREATE INDEX idx_chat_read_state_user
    ON chat_read_state (user_id, unread_count);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_chat_read_state_user;
DROP TABLE IF EXISTS chat_read_state;

DROP INDEX IF EXISTS idx_chat_notification_deliveries_chat_unread;
DROP INDEX IF EXISTS idx_chat_notification_deliveries_recipient_status;
DROP TABLE IF EXISTS chat_notification_deliveries;

DROP INDEX IF EXISTS idx_outbox_events_type_status;
DROP INDEX IF EXISTS idx_outbox_events_pickup;
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd
