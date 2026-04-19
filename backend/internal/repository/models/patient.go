package models

type PatientProfile struct {
	UserID    int64   `db:"user_id"`
	BirthDate *int64  `db:"birth_date"`
	Gender    *string `db:"gender"`
	CreatedAt int64   `db:"created_at"`
	UpdatedAt int64   `db:"updated_at"`
}

type DoctorPatient struct {
	ID        int64  `db:"id"`
	DoctorID  int64  `db:"doctor_id"`
	PatientID int64  `db:"patient_id"`
	CreatedAt int64  `db:"created_at"`
	DeletedAt *int64 `db:"deleted_at"`
}

type LinkedPatient struct {
	ID        int64   `db:"id"`
	UserID    int64   `db:"user_id"`
	Login     string  `db:"login"`
	Email     *string `db:"email"`
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	CreatedAt int64   `db:"created_at"`
}
