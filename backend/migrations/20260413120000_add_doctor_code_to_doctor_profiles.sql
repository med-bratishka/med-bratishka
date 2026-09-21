-- +goose Up
-- +goose StatementBegin
ALTER TABLE doctor_profiles
    ADD COLUMN IF NOT EXISTS doctor_code TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_doctor_profiles_doctor_code_unique
    ON doctor_profiles (doctor_code)
    WHERE doctor_code IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_doctor_profiles_doctor_code_unique;

ALTER TABLE doctor_profiles
    DROP COLUMN IF EXISTS doctor_code;
-- +goose StatementEnd

