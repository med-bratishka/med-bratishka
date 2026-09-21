-- +goose Up
-- +goose StatementBegin
CREATE TYPE doctor_clinic_status AS ENUM (
    'pending',
    'active',
    'suspended',
    'rejected'
);

CREATE TABLE doctor_profiles
(
    user_id        BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    specialization TEXT,
    license_number TEXT,
    bio            TEXT,
    created_at     BIGINT NOT NULL,
    updated_at     BIGINT NOT NULL
);

CREATE TABLE doctor_clinic_memberships
(
    id         BIGSERIAL PRIMARY KEY,
    doctor_id  BIGINT               NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    clinic_id  BIGINT               NOT NULL REFERENCES clinics (id) ON DELETE CASCADE,
    invited_by BIGINT               NOT NULL REFERENCES users (id),
    status     doctor_clinic_status NOT NULL DEFAULT 'pending',
    created_at BIGINT               NOT NULL,
    updated_at BIGINT               NOT NULL,

    UNIQUE (doctor_id, clinic_id)
);

CREATE INDEX idx_dcm_doctor_id ON doctor_clinic_memberships (doctor_id);
CREATE INDEX idx_dcm_clinic_id ON doctor_clinic_memberships (clinic_id);
CREATE INDEX idx_dcm_status ON doctor_clinic_memberships (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_dcm_status;
DROP INDEX idx_dcm_clinic_id;
DROP INDEX idx_dcm_doctor_id;
DROP TABLE doctor_clinic_memberships;
DROP TABLE doctor_profiles;
DROP TYPE doctor_clinic_status;
-- +goose StatementEnd