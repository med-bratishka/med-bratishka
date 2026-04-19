-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('admin', 'doctor', 'patient');

CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    login         TEXT      NOT NULL,
    email         TEXT,
    phone         TEXT,
    password_hash TEXT      NOT NULL,
    role          user_role NOT NULL,
    is_verified   BOOLEAN   NOT NULL DEFAULT FALSE,
    first_name    TEXT      NOT NULL,
    last_name     TEXT      NOT NULL,
    middle_name   TEXT,
    created_at    BIGINT    NOT NULL,
    updated_at    BIGINT    NOT NULL,
    deleted_at    BIGINT
);

CREATE UNIQUE INDEX idx_users_login_active ON users (login) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_users_email_active ON users (email) WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE UNIQUE INDEX idx_users_phone_active ON users (phone) WHERE deleted_at IS NULL AND phone IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_users_phone_active;
DROP INDEX idx_users_email_active;
DROP INDEX idx_users_login_active;
DROP TABLE users;
DROP TYPE user_role;
-- +goose StatementEnd
