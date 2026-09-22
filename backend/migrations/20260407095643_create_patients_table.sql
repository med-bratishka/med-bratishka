-- +goose Up
-- +goose StatementBegin
CREATE TABLE patient_profiles
(
    user_id    BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    birth_date BIGINT,
    gender     TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE TABLE doctor_patients
(
    id         BIGSERIAL PRIMARY KEY,
    doctor_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    patient_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT,

    UNIQUE (doctor_id, patient_id)
);

CREATE INDEX idx_doctor_patients_doctor_id
    ON doctor_patients (doctor_id);

CREATE INDEX idx_doctor_patients_patient_id
    ON doctor_patients (patient_id);

CREATE INDEX idx_doctor_patients_active
    ON doctor_patients (doctor_id, patient_id) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_doctor_patients_active;
DROP INDEX idx_doctor_patients_patient_id;
DROP INDEX idx_doctor_patients_doctor_id;
DROP TABLE doctor_patients;
DROP TABLE patient_profiles;
-- +goose StatementEnd