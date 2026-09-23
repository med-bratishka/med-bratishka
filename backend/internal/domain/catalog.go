package domain

type Clinic struct {
	ID          int64
	Name        string
	Description string
	Address     string
	Phone       string
	Email       string
	CreatedAt   int64
	UpdatedAt   int64
}

type Doctor struct {
	ID             int64
	Login          string
	Email          string
	Phone          string
	IsVerified     bool
	FirstName      string
	LastName       string
	MiddleName     string
	Specialization string
	LicenseNumber  string
	Bio            string
	DoctorCode     string
	CreatedAt      int64
	UpdatedAt      int64
}
