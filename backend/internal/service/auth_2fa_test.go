package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	secretcrypto "medbratishka/pkg/crypto"

	"github.com/jmoiron/sqlx"
)

func TestDisableTOTPAcceptsRecoveryCode(t *testing.T) {
	svc, twoFA, sessions := newTestAuthService(t)
	userCtx := &domain.UserTokenContext{ID: 7, Role: domain.RolePatient, Authorized: true}

	err := svc.DisableTOTP(context.Background(), userCtx, "ABCDE-FGHIJ-KLMNO")
	if err != nil {
		t.Fatalf("DisableTOTP returned error: %v", err)
	}

	if !twoFA.disabled {
		t.Fatal("expected TOTP to be disabled")
	}
	if !twoFA.trustedDevicesRevoked {
		t.Fatal("expected trusted devices to be revoked")
	}
	if !sessions.deletedAll {
		t.Fatal("expected user sessions to be deleted")
	}
	if twoFA.usedRecoveryCodeID != 11 {
		t.Fatalf("expected recovery code 11 to be used, got %d", twoFA.usedRecoveryCodeID)
	}
	if !twoFA.hasAudit("2fa_disabled") {
		t.Fatal("expected 2fa_disabled audit event")
	}
}

func TestRegenerateRecoveryCodesAcceptsRecoveryCode(t *testing.T) {
	svc, twoFA, _ := newTestAuthService(t)
	userCtx := &domain.UserTokenContext{ID: 7, Role: domain.RolePatient, Authorized: true}

	resp, err := svc.RegenerateRecoveryCodes(context.Background(), userCtx, "ABCDE-FGHIJ-KLMNO")
	if err != nil {
		t.Fatalf("RegenerateRecoveryCodes returned error: %v", err)
	}

	if resp == nil || len(resp.Codes) != 10 {
		t.Fatalf("expected 10 recovery codes, got %#v", resp)
	}
	if twoFA.usedRecoveryCodeID != 11 {
		t.Fatalf("expected recovery code 11 to be used, got %d", twoFA.usedRecoveryCodeID)
	}
	if len(twoFA.replacedRecoveryHashes) != 10 {
		t.Fatalf("expected 10 replacement hashes, got %d", len(twoFA.replacedRecoveryHashes))
	}
	if !twoFA.hasAudit("recovery_codes_regenerated") {
		t.Fatal("expected recovery_codes_regenerated audit event")
	}
}

func TestDisableTOTPRejectsInvalidRecoveryCode(t *testing.T) {
	svc, twoFA, sessions := newTestAuthService(t)
	userCtx := &domain.UserTokenContext{ID: 7, Role: domain.RolePatient, Authorized: true}

	err := svc.DisableTOTP(context.Background(), userCtx, "WRONG-CODE")
	if !errors.Is(err, ErrTwoFactorInvalid) {
		t.Fatalf("expected ErrTwoFactorInvalid, got %v", err)
	}
	if twoFA.disabled {
		t.Fatal("did not expect TOTP to be disabled")
	}
	if sessions.deletedAll {
		t.Fatal("did not expect sessions to be deleted")
	}
}

func TestLoginWithEnabledTOTPReturnsChallengeWithoutTokens(t *testing.T) {
	svc, twoFA, sessions := newTestAuthService(t)

	resp, err := svc.Login(context.Background(), &domain.AuthenticationInput{
		AccessParameter: "patient@example.com",
		Password:        "correct-password",
		UserAgent:       "test-agent",
		IPAddress:       "127.0.0.1:12345",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if resp == nil || !resp.TwoFactorRequired {
		t.Fatalf("expected two factor challenge response, got %#v", resp)
	}
	if resp.TwoFactorChallenge == "" {
		t.Fatal("expected challenge id")
	}
	if resp.AccessToken.Token != "" || resp.RefreshToken.Token != "" {
		t.Fatalf("did not expect tokens before 2FA verification: %#v", resp)
	}
	if twoFA.createdChallenge == nil || twoFA.createdChallenge.ID != resp.TwoFactorChallenge {
		t.Fatalf("expected challenge to be persisted, got %#v", twoFA.createdChallenge)
	}
	if sessions.createdCount != 0 {
		t.Fatalf("did not expect sessions before 2FA verification, got %d", sessions.createdCount)
	}
	if !twoFA.hasAudit("login_2fa_required") {
		t.Fatal("expected login_2fa_required audit event")
	}
}

func TestLoginWrongPasswordReturnsInvalidCredentials(t *testing.T) {
	svc, twoFA, sessions := newTestAuthService(t)

	_, err := svc.Login(context.Background(), &domain.AuthenticationInput{
		AccessParameter: "patient@example.com",
		Password:        "wrong-password",
		UserAgent:       "test-agent",
		IPAddress:       "127.0.0.1:12345",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if twoFA.createdChallenge != nil {
		t.Fatalf("did not expect challenge for wrong password, got %#v", twoFA.createdChallenge)
	}
	if sessions.createdCount != 0 {
		t.Fatalf("did not expect sessions for wrong password, got %d", sessions.createdCount)
	}
	if !twoFA.hasAudit("login_password_failed") {
		t.Fatal("expected login_password_failed audit event")
	}
}

func newTestAuthService(t *testing.T) (AuthService, *fakeTwoFactorRepo, *fakeSessionsRepo) {
	t.Helper()

	box, err := secretcrypto.NewSecretBox("test-2fa-key")
	if err != nil {
		t.Fatalf("create secret box: %v", err)
	}

	user := &domain.User{
		ID:        7,
		Login:     "patient@example.com",
		Email:     "patient@example.com",
		Password:  "hash:correct-password",
		Role:      domain.RolePatient,
		FirstName: "Test",
		LastName:  "Patient",
	}
	ciphertext, err := box.EncryptString("JBSWY3DPEHPK3PXP", totpAAD(user.ID, user.Role))
	if err != nil {
		t.Fatalf("encrypt totp secret: %v", err)
	}

	hasher := &fakePasswordHasher{}
	recoveryHash, err := hasher.Hash(normalizeRecoveryCode("ABCDE-FGHIJ-KLMNO"))
	if err != nil {
		t.Fatalf("hash recovery code: %v", err)
	}

	twoFA := &fakeTwoFactorRepo{
		settings: &domain.TwoFactorSettings{
			UserID:           user.ID,
			Role:             user.Role,
			SecretCiphertext: ciphertext,
			Enabled:          true,
		},
		recoveryHashes: map[int64]string{11: recoveryHash},
	}
	sessions := &fakeSessionsRepo{}

	return NewAuthService(
		&fakeUsersRepo{user: user},
		sessions,
		twoFA,
		&fakeTxRepo{},
		&fakeTokenManager{},
		hasher,
		box,
		fakeTimeManager{now: time.Unix(1710000000, 0)},
		time.Hour,
		24*time.Hour,
		5*time.Minute,
		30*24*time.Hour,
		"Medbratishka",
		"trusted-pepper",
	), twoFA, sessions
}

type fakeTxRepo struct{}

func (r *fakeTxRepo) StartTransaction(ctx context.Context) (transaction.Transaction, error) {
	return &fakeTx{}, nil
}

func (r *fakeTxRepo) StartReadOnlyClientTransaction(ctx context.Context) (transaction.Transaction, error) {
	return &fakeTx{}, nil
}

type fakeTx struct {
	committed bool
}

func (t *fakeTx) Commit() error {
	t.committed = true
	return nil
}

func (t *fakeTx) Rollback() {}

func (t *fakeTx) Txm() *sqlx.Tx { return nil }

type fakeUsersRepo struct {
	user *domain.User
}

func (r *fakeUsersRepo) CreateTX(ctx context.Context, tx transaction.Transaction, user *domain.User) (int64, error) {
	return user.ID, nil
}

func (r *fakeUsersRepo) GetByAccessParameterTX(ctx context.Context, tx transaction.Transaction, accessParam string) (*domain.User, error) {
	return r.user, nil
}

func (r *fakeUsersRepo) GetByIDTX(ctx context.Context, tx transaction.Transaction, id int64, role domain.Role) (*domain.User, error) {
	if r.user == nil || r.user.ID != id || r.user.Role != role {
		return nil, repository.ErrNotFound
	}
	return r.user, nil
}

func (r *fakeUsersRepo) CheckUserExistsTX(ctx context.Context, tx transaction.Transaction, login, email, phone string) (bool, error) {
	return false, nil
}

func (r *fakeUsersRepo) GetByAccessParameter(ctx context.Context, accessParam string) (*domain.User, error) {
	return r.user, nil
}

func (r *fakeUsersRepo) GetByID(ctx context.Context, id int64, role domain.Role) (*domain.User, error) {
	return r.GetByIDTX(ctx, nil, id, role)
}

type fakeSessionsRepo struct {
	deletedAll   bool
	createdCount int
}

func (r *fakeSessionsRepo) CreateTX(ctx context.Context, tx transaction.Transaction, session *domain.Session) error {
	r.createdCount++
	return nil
}

func (r *fakeSessionsRepo) GetNextNumberTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (int, error) {
	return 1, nil
}

func (r *fakeSessionsRepo) GetSecretTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error) {
	return "secret", nil
}

func (r *fakeSessionsRepo) DeleteByNumberTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, number int, revokedAt int64) error {
	return nil
}

func (r *fakeSessionsRepo) DeleteAllByUserTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, revokedAt int64) error {
	r.deletedAll = true
	return nil
}

func (r *fakeSessionsRepo) Create(ctx context.Context, session *domain.Session) error {
	r.createdCount++
	return nil
}

func (r *fakeSessionsRepo) GetNextNumber(ctx context.Context, userID int64, role domain.Role) (int, error) {
	return 1, nil
}

func (r *fakeSessionsRepo) GetSecret(ctx context.Context, userID int64, role domain.Role, number int, purpose domain.TokenPurpose) (string, error) {
	return "secret", nil
}

func (r *fakeSessionsRepo) DeleteByNumber(ctx context.Context, userID int64, role domain.Role, number int, revokedAt int64) error {
	return nil
}

func (r *fakeSessionsRepo) DeleteAllByUser(ctx context.Context, userID int64, role domain.Role, revokedAt int64) error {
	r.deletedAll = true
	return nil
}

type fakeTwoFactorRepo struct {
	settings               *domain.TwoFactorSettings
	recoveryHashes         map[int64]string
	replacedRecoveryHashes []string
	usedRecoveryCodeID     int64
	createdChallenge       *domain.AuthChallenge
	disabled               bool
	trustedDevicesRevoked  bool
	auditEvents            []string
}

func (r *fakeTwoFactorRepo) UpsertTOTPSettingsTX(ctx context.Context, tx transaction.Transaction, settings *domain.TwoFactorSettings) error {
	r.settings = settings
	return nil
}

func (r *fakeTwoFactorRepo) GetTOTPSettingsTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (*domain.TwoFactorSettings, error) {
	if r.settings == nil {
		return nil, repository.ErrNotFound
	}
	return r.settings, nil
}

func (r *fakeTwoFactorRepo) GetTOTPSettings(ctx context.Context, userID int64, role domain.Role) (*domain.TwoFactorSettings, error) {
	return r.GetTOTPSettingsTX(ctx, nil, userID, role)
}

func (r *fakeTwoFactorRepo) EnableTOTPTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, confirmedAt int64) error {
	if r.settings != nil {
		r.settings.Enabled = true
	}
	return nil
}

func (r *fakeTwoFactorRepo) DisableTOTPTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, disabledAt int64) error {
	r.disabled = true
	if r.settings != nil {
		r.settings.Enabled = false
		r.settings.DisabledAt = &disabledAt
	}
	return nil
}

func (r *fakeTwoFactorRepo) CreateChallengeTX(ctx context.Context, tx transaction.Transaction, challenge *domain.AuthChallenge) error {
	cp := *challenge
	r.createdChallenge = &cp
	return nil
}

func (r *fakeTwoFactorRepo) GetChallengeTX(ctx context.Context, tx transaction.Transaction, id string) (*domain.AuthChallenge, error) {
	return nil, repository.ErrNotFound
}

func (r *fakeTwoFactorRepo) ConsumeChallengeTX(ctx context.Context, tx transaction.Transaction, id string, consumedAt int64) error {
	return nil
}

func (r *fakeTwoFactorRepo) IncrementChallengeFailuresTX(ctx context.Context, tx transaction.Transaction, id string) error {
	return nil
}

func (r *fakeTwoFactorRepo) ReplaceRecoveryCodesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, hashes []string, createdAt int64) error {
	r.replacedRecoveryHashes = append([]string(nil), hashes...)
	return nil
}

func (r *fakeTwoFactorRepo) GetUnusedRecoveryCodesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role) (map[int64]string, error) {
	res := make(map[int64]string, len(r.recoveryHashes))
	for id, hash := range r.recoveryHashes {
		res[id] = hash
	}
	return res, nil
}

func (r *fakeTwoFactorRepo) UseRecoveryCodeTX(ctx context.Context, tx transaction.Transaction, id int64, usedAt int64) error {
	r.usedRecoveryCodeID = id
	delete(r.recoveryHashes, id)
	return nil
}

func (r *fakeTwoFactorRepo) CreateTrustedDeviceTX(ctx context.Context, tx transaction.Transaction, d *domain.TrustedDevice) error {
	return nil
}

func (r *fakeTwoFactorRepo) GetTrustedDeviceByHash(ctx context.Context, tokenHash string) (*domain.TrustedDevice, error) {
	return nil, repository.ErrNotFound
}

func (r *fakeTwoFactorRepo) TouchTrustedDevice(ctx context.Context, id int64, touchedAt int64) error {
	return nil
}

func (r *fakeTwoFactorRepo) RevokeTrustedDevicesTX(ctx context.Context, tx transaction.Transaction, userID int64, role domain.Role, revokedAt int64) error {
	r.trustedDevicesRevoked = true
	return nil
}

func (r *fakeTwoFactorRepo) CreateAuditTX(ctx context.Context, tx transaction.Transaction, userID *int64, role *domain.Role, eventType, ip, userAgent, metadata string, createdAt int64) error {
	r.auditEvents = append(r.auditEvents, eventType)
	return nil
}

func (r *fakeTwoFactorRepo) CreateAudit(ctx context.Context, userID *int64, role *domain.Role, eventType, ip, userAgent, metadata string, createdAt int64) error {
	r.auditEvents = append(r.auditEvents, eventType)
	return nil
}

func (r *fakeTwoFactorRepo) hasAudit(eventType string) bool {
	for _, item := range r.auditEvents {
		if item == eventType {
			return true
		}
	}
	return false
}

type fakePasswordHasher struct{}

func (h *fakePasswordHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (h *fakePasswordHasher) Compare(hash, password string) bool {
	return hash == "hash:"+password
}

type fakeTokenManager struct{}

func (m *fakeTokenManager) GenerateToken(data *domain.TokenData) (string, error) {
	return "token", nil
}

func (m *fakeTokenManager) ParseToken(tokenString string) (*domain.TokenData, error) {
	return nil, errors.New("not implemented")
}

type fakeTimeManager struct {
	now time.Time
}

func (m fakeTimeManager) Now() time.Time {
	return m.now
}

func (m fakeTimeManager) MillisecondsToTime(v int64) time.Time {
	return time.UnixMilli(v)
}
