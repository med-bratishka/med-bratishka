-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats
(
    id          BIGSERIAL PRIMARY KEY,
    doctor_id   BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    patient_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at  BIGINT NOT NULL,
    updated_at  BIGINT NOT NULL,
    deleted_at  BIGINT,

    UNIQUE (doctor_id, patient_id)
);

CREATE INDEX idx_chats_doctor_id ON chats (doctor_id);
CREATE INDEX idx_chats_patient_id ON chats (patient_id);
CREATE INDEX idx_chats_active ON chats (doctor_id, patient_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_chats_active;
DROP INDEX idx_chats_patient_id;
DROP INDEX idx_chats_doctor_id;
DROP TABLE chats;
-- +goose StatementEnd

