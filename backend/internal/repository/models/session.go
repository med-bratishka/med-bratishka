package models

type AuthToken struct {
	ID            int64  `db:"id"`
	UserID        int64  `db:"user_id"`
	Role          string `db:"role"`
	SessionNumber int    `db:"session_number"`
	Purpose       string `db:"purpose"`
	Secret        string `db:"secret"`
	ExpiresAt     int64  `db:"expires_at"`
	CreatedAt     int64  `db:"created_at"`
	RevokedAt     *int64 `db:"revoked_at"`
}
