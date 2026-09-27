-- +goose Up
-- +goose StatementBegin
CREATE TABLE web_push_subscriptions
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    endpoint   TEXT   NOT NULL UNIQUE,
    p256dh     TEXT   NOT NULL,
    auth       TEXT   NOT NULL,
    user_agent TEXT   NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE INDEX idx_web_push_subscriptions_user
    ON web_push_subscriptions (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_web_push_subscriptions_user;
DROP TABLE IF EXISTS web_push_subscriptions;
-- +goose StatementEnd
