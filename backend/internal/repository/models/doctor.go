package models

type DoctorProfile struct {
	UserID         int64   `db:"user_id"`
	Specialization *string `db:"specialization"`
	LicenseNumber  *string `db:"license_number"`
	Bio            *string `db:"bio"`
	DoctorCode     *string `db:"doctor_code"`
	CreatedAt      int64   `db:"created_at"`
	UpdatedAt      int64   `db:"updated_at"`
}

type DoctorClinicMembership struct {
	ID        int64  `db:"id"`
	DoctorID  int64  `db:"doctor_id"`
	ClinicID  int64  `db:"clinic_id"`
	InvitedBy int64  `db:"invited_by"`
	Status    string `db:"status"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

// Расширенная модель для выдачи вместе с данными клиники
type DoctorClinicWithDetails struct {
	ID            int64   `db:"id"`
	ClinicID      int64   `db:"clinic_id"`
	Status        string  `db:"status"`
	ClinicName    string  `db:"name"`
	ClinicAddress *string `db:"address"`
}
