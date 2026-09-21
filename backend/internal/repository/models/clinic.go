package models

type Clinic struct {
	ID          int64   `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
	Address     *string `db:"address"`
	Phone       *string `db:"phone"`
	Email       *string `db:"email"`
	CreatedAt   int64   `db:"created_at"`
	UpdatedAt   int64   `db:"updated_at"`
	DeletedAt   *int64  `db:"deleted_at"`
}

type ClinicAdmin struct {
	ClinicID  int64 `db:"clinic_id"`
	UserID    int64 `db:"user_id"`
	CreatedAt int64 `db:"created_at"`
}
