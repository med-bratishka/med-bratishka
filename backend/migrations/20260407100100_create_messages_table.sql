-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages
(
    id         BIGSERIAL PRIMARY KEY,
    chat_id    BIGINT NOT NULL REFERENCES chats (id) ON DELETE CASCADE,
    sender_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content    TEXT   NOT NULL,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT
);

CREATE INDEX idx_messages_chat_id ON messages (chat_id);
CREATE INDEX idx_messages_sender_id ON messages (sender_id);
CREATE INDEX idx_messages_created_at ON messages (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_messages_created_at;
DROP INDEX idx_messages_sender_id;
DROP INDEX idx_messages_chat_id;
DROP TABLE messages;
-- +goose StatementEnd

