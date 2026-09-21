-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
    ALTER COLUMN content DROP NOT NULL;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS attachment_url TEXT,
    ADD COLUMN IF NOT EXISTS attachment_type TEXT,
    ADD COLUMN IF NOT EXISTS attachment_mime_type TEXT;

CREATE INDEX IF NOT EXISTS idx_messages_attachment_type ON messages (attachment_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_attachment_type;

ALTER TABLE messages
    DROP COLUMN IF EXISTS attachment_mime_type,
    DROP COLUMN IF EXISTS attachment_type,
    DROP COLUMN IF EXISTS attachment_url;

ALTER TABLE messages
    ALTER COLUMN content SET NOT NULL;
-- +goose StatementEnd

