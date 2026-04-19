-- +goose Up
-- +goose StatementBegin
CREATE TYPE token_purpose AS ENUM ('access', 'refresh');

CREATE TABLE auth_tokens
(
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role           user_role     NOT NULL,
    purpose        token_purpose NOT NULL,
    session_number INT           NOT NULL,
    secret         TEXT          NOT NULL,
    expires_at     BIGINT        NOT NULL,
    created_at     BIGINT        NOT NULL,
    revoked_at     BIGINT,

    UNIQUE (user_id, role, purpose, session_number)
);

CREATE INDEX idx_auth_tokens_lookup
    ON auth_tokens (user_id, role, purpose, session_number) WHERE revoked_at IS NULL;

CREATE INDEX idx_auth_tokens_user_role
    ON auth_tokens (user_id, role) WHERE revoked_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_auth_tokens_user_role;
DROP INDEX idx_auth_tokens_lookup;
DROP TABLE auth_tokens;
DROP TYPE token_purpose;
-- +goose StatementEnd