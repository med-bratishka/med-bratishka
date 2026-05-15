-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_totp_settings
(
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role              user_role NOT NULL,
    secret_ciphertext TEXT      NOT NULL,
    enabled           BOOLEAN   NOT NULL DEFAULT FALSE,
    confirmed_at      BIGINT,
    disabled_at       BIGINT,
    created_at        BIGINT    NOT NULL,
    updated_at        BIGINT    NOT NULL,

    UNIQUE (user_id, role)
);

CREATE INDEX idx_user_totp_settings_enabled
    ON user_totp_settings (user_id, role)
    WHERE enabled = TRUE;

CREATE TABLE auth_challenges
(
    id              UUID      PRIMARY KEY,
    user_id         BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role            user_role NOT NULL,
    purpose         TEXT      NOT NULL,
    failed_attempts INT       NOT NULL DEFAULT 0,
    expires_at      BIGINT    NOT NULL,
    consumed_at     BIGINT,
    created_at      BIGINT    NOT NULL
);

CREATE INDEX idx_auth_challenges_user_role
    ON auth_challenges (user_id, role, created_at);

CREATE INDEX idx_auth_challenges_active
    ON auth_challenges (user_id, role, purpose, expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE recovery_codes
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role       user_role NOT NULL,
    code_hash  TEXT      NOT NULL,
    used_at    BIGINT,
    created_at BIGINT    NOT NULL
);

CREATE INDEX idx_recovery_codes_unused
    ON recovery_codes (user_id, role)
    WHERE used_at IS NULL;

CREATE TABLE trusted_devices
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT    NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role            user_role NOT NULL,
    token_hash      TEXT      NOT NULL UNIQUE,
    device_name     TEXT,
    user_agent_hash TEXT,
    ip              TEXT,
    expires_at      BIGINT    NOT NULL,
    last_used_at    BIGINT,
    revoked_at      BIGINT,
    created_at      BIGINT    NOT NULL
);

CREATE INDEX idx_trusted_devices_user_active
    ON trusted_devices (user_id, role, expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE auth_audit_log
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT REFERENCES users (id) ON DELETE SET NULL,
    role       user_role,
    event_type TEXT   NOT NULL,
    ip         TEXT,
    user_agent TEXT,
    metadata   JSONB,
    created_at BIGINT NOT NULL
);

CREATE INDEX idx_auth_audit_log_user_created
    ON auth_audit_log (user_id, role, created_at);

CREATE INDEX idx_auth_audit_log_event_type
    ON auth_audit_log (event_type, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_auth_audit_log_event_type;
DROP INDEX IF EXISTS idx_auth_audit_log_user_created;
DROP TABLE IF EXISTS auth_audit_log;

DROP INDEX IF EXISTS idx_trusted_devices_user_active;
DROP TABLE IF EXISTS trusted_devices;

DROP INDEX IF EXISTS idx_recovery_codes_unused;
DROP TABLE IF EXISTS recovery_codes;

DROP INDEX IF EXISTS idx_auth_challenges_active;
DROP INDEX IF EXISTS idx_auth_challenges_user_role;
DROP TABLE IF EXISTS auth_challenges;

DROP INDEX IF EXISTS idx_user_totp_settings_enabled;
DROP TABLE IF EXISTS user_totp_settings;
-- +goose StatementEnd
