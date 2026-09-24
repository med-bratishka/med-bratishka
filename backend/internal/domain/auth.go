package domain

import "time"

type User struct {
	ID         int64
	Login      string
	Email      string
	Phone      string
	Password   string
	Role       Role
	IsVerified bool
	FirstName  string
	LastName   string
	MiddleName string
	CreatedAt  int64
	UpdatedAt  int64
	DeletedAt  *int64
}

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleDoctor  Role = "doctor"
	RolePatient Role = "patient"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleDoctor || r == RolePatient
}

type TokenPurpose string

const (
	PurposeAccess  TokenPurpose = "access"
	PurposeRefresh TokenPurpose = "refresh"
)

type Session struct {
	ID        int64
	UserID    int64
	Role      Role
	Purpose   TokenPurpose
	Number    int
	Secret    string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

type TokenData struct {
	ID        int64
	Role      Role
	Login     string
	Purpose   TokenPurpose
	Number    int
	Secret    string
	ExpiresAt time.Time
}

type UserTokenContext struct {
	ID         int64
	Role       Role
	Number     int
	Authorized bool
}

type RegistrationInput struct {
	Login      string
	Email      string
	Phone      string
	Password   string
	FirstName  string
	LastName   string
	MiddleName string
	Role       Role
}

type AuthenticationInput struct {
	AccessParameter    string
	Password           string
	TrustedDeviceToken string
	DeviceName         string
	IPAddress          string
	UserAgent          string
}

type RefreshTokenInput struct {
	UserID int64
	Role   Role
	Number int
}

type TokenResponse struct {
	Token     string
	ExpiresAt int64
	Type      string
}

type UserResponse struct {
	ID         int64
	Login      string
	Email      string
	Phone      string
	Role       string
	IsVerified bool
	FirstName  string
	LastName   string
	MiddleName string
	CreatedAt  int64
	UpdatedAt  int64
}

type AuthResponse struct {
	AccessToken        TokenResponse
	RefreshToken       TokenResponse
	ServerTime         int64
	User               UserResponse
	TwoFactorRequired  bool
	TwoFactorChallenge string
	TwoFactorExpiresAt int64
	TrustedDeviceToken string
}

type TwoFactorSettings struct {
	UserID           int64
	Role             Role
	SecretCiphertext string
	Enabled          bool
	ConfirmedAt      *int64
	DisabledAt       *int64
	CreatedAt        int64
	UpdatedAt        int64
}

type AuthChallenge struct {
	ID             string
	UserID         int64
	Role           Role
	Purpose        string
	FailedAttempts int
	ExpiresAt      int64
	ConsumedAt     *int64
	CreatedAt      int64
}

type TrustedDevice struct {
	ID            int64
	UserID        int64
	Role          Role
	TokenHash     string
	DeviceName    string
	UserAgentHash string
	ExpiresAt     int64
	LastUsedAt    *int64
	RevokedAt     *int64
	CreatedAt     int64
}

type TwoFactorSetupResponse struct {
	Secret     string
	OTPAuthURL string
}

type TwoFactorVerifyInput struct {
	ChallengeID  string
	Code         string
	RecoveryCode string
	TrustDevice  bool
	DeviceName   string
	IPAddress    string
	UserAgent    string
}

type RecoveryCodesResponse struct {
	Codes []string
}
