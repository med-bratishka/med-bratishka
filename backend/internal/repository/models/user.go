package models

type User struct {
	ID         int64   `db:"id"`
	Login      string  `db:"login"`
	Email      *string `db:"email"`
	Phone      *string `db:"phone"`
	Password   string  `db:"password_hash"`
	Role       string  `db:"role"`
	IsVerified bool    `db:"is_verified"`
	FirstName  string  `db:"first_name"`
	LastName   string  `db:"last_name"`
	MiddleName *string `db:"middle_name"`
	CreatedAt  int64   `db:"created_at"`
	UpdatedAt  int64   `db:"updated_at"`
	DeletedAt  *int64  `db:"deleted_at"`
}
