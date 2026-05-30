-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS attachment_name TEXT,
    ADD COLUMN IF NOT EXISTS attachment_key TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages
    DROP COLUMN IF EXISTS attachment_key,
    DROP COLUMN IF EXISTS attachment_name;
-- +goose StatementEnd
