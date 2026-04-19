-- +goose Up
-- +goose StatementBegin
CREATE TABLE clinics
(
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT   NOT NULL,
    description TEXT,
    address     TEXT,
    phone       TEXT,
    email       TEXT,
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL,
    deleted_at  BIGINT
);

CREATE TABLE clinic_admins
(
    clinic_id  BIGINT NOT NULL REFERENCES clinics (id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL,
    PRIMARY KEY (clinic_id, user_id)
);

CREATE INDEX idx_clinic_admins_user_id
    ON clinic_admins (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_clinic_admins_user_id;
DROP TABLE clinic_admins;
DROP TABLE clinics;
-- +goose StatementEnd