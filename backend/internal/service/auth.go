package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	secretcrypto "medbratishka/pkg/crypto"
	"medbratishka/pkg/time_manager"
	"medbratishka/pkg/totp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserNotVerified    = errors.New("user not verified")
	ErrInvalidTokenSecret = errors.New("invalid token secret")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrSessionNotFound    = errors.New("session not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTwoFactorRequired  = errors.New("two factor required")
	ErrTwoFactorInvalid   = errors.New("invalid two factor code")
	ErrTwoFactorEnabled   = errors.New("two factor already enabled")
	ErrChallengeNotFound  = errors.New("challenge not found")
	ErrChallengeExpired   = errors.New("challenge expired")
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type TokenManager interface {
	GenerateToken(data *domain.TokenData) (string, error)
	ParseToken(tokenString string) (*domain.TokenData, error)
}

type AuthService interface {
	Register(ctx context.Context, input *domain.RegistrationInput) (*domain.AuthResponse, error)
	Login(ctx context.Context, input *domain.AuthenticationInput) (*domain.AuthResponse, error)
	Refresh(ctx context.Context, userCtx *domain.UserTokenContext) (*domain.AuthResponse, error)
	Logout(ctx context.Context, userCtx *domain.UserTokenContext) error
	LogoutAll(ctx context.Context, userID int64, role domain.Role) error
	ValidateToken(ctx context.Context, purpose domain.TokenPurpose, tokenString string) (*domain.UserTokenContext, error)
	SetupTOTP(ctx context.Context, userCtx *domain.UserTokenContext) (*domain.TwoFactorSetupResponse, error)
	ConfirmTOTP(ctx context.Context, userCtx *domain.UserTokenContext, code string) (*domain.RecoveryCodesResponse, error)
	VerifyTOTPChallenge(ctx context.Context, input *domain.TwoFactorVerifyInput) (*domain.AuthResponse, error)
	DisableTOTP(ctx context.Context, userCtx *domain.UserTokenContext, code string) error
	RegenerateRecoveryCodes(ctx context.Context, userCtx *domain.UserTokenContext, code string) (*domain.RecoveryCodesResponse, error)
}

type authService struct {
	usersRepo     repository.UsersRepository
	sessionsRepo  repository.SessionsRepository
	twoFARepo     repository.TwoFactorRepository
	txRepo        transaction.Repository
	tokenMgr      TokenManager
	hasher        PasswordHasher
	secretBox     *secretcrypto.SecretBox
	timeManager   time_manager.TimeManager
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
	trustedPepper string
	challengeTTL  time.Duration
	trustedTTL    time.Duration
}

func NewAuthService(
	usersRepo repository.UsersRepository,
	sessionsRepo repository.SessionsRepository,
	twoFARepo repository.TwoFactorRepository,
	txRepo transaction.Repository,
	tokenMgr TokenManager,
	hasher PasswordHasher,
	secretBox *secretcrypto.SecretBox,
	timeManager time_manager.TimeManager,
	accessTTL, refreshTTL, challengeTTL, trustedTTL time.Duration,
	issuer, trustedPepper string,
) AuthService {
	return &authService{
		usersRepo:     usersRepo,
		sessionsRepo:  sessionsRepo,
		twoFARepo:     twoFARepo,
		txRepo:        txRepo,
		tokenMgr:      tokenMgr,
		hasher:        hasher,
		secretBox:     secretBox,
		timeManager:   timeManager,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		issuer:        issuer,
		trustedPepper: trustedPepper,
		challengeTTL:  challengeTTL,
		trustedTTL:    trustedTTL,
	}
}

func (s *authService) Register(ctx context.Context, input *domain.RegistrationInput) (*domain.AuthResponse, error) {
	if !input.Role.IsValid() {
		return nil, newServiceError(CodeBadRequest, ErrInvalidRole, "INVALID_ROLE", "invalid role")
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("Register/StartTransaction", err)
	}
	defer tx.Rollback()

	exists, err := s.usersRepo.CheckUserExistsTX(ctx, tx, input.Login, input.Email, input.Phone)
	if err != nil {
		return nil, wrapInternal("Register/CheckUserExistsTX", err)
	}
	if exists {
		return nil, newServiceError(CodeConflict, ErrUserAlreadyExists, "USER_EXISTS", "user already exists")
	}

	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, wrapInternal("Register/Hash", err)
	}

	now := s.timeManager.Now().UnixMilli()
	user := &domain.User{
		Login:      input.Login,
		Email:      input.Email,
		Phone:      input.Phone,
		Password:   passwordHash,
		Role:       input.Role,
		IsVerified: false,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		MiddleName: input.MiddleName,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	user.ID, err = s.usersRepo.CreateTX(ctx, tx, user)
	if err != nil {
		return nil, wrapInternal("Register/CreateTX", err)
	}

	resp, err := s.createAuthResponseTX(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("Register/Commit", err)
	}

	return resp, nil
}

func (s *authService) Login(ctx context.Context, input *domain.AuthenticationInput) (*domain.AuthResponse, error) {
	user, err := s.usersRepo.GetByAccessParameter(ctx, input.AccessParameter)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeUnauthorized, ErrInvalidCredentials, "INVALID_CREDENTIALS", "invalid credentials")
		}
		return nil, wrapInternal("Login/GetByAccessParameter", err)
	}

	if !s.hasher.Compare(user.Password, input.Password) {
		_ = s.twoFARepo.CreateAudit(ctx, &user.ID, &user.Role, "login_password_failed", cleanIP(input.IPAddress), input.UserAgent, "", s.timeManager.Now().UnixMilli())
		return nil, newServiceError(CodeUnauthorized, ErrInvalidCredentials, "INVALID_CREDENTIALS", "invalid credentials")
	}

	settings, err := s.twoFARepo.GetTOTPSettings(ctx, user.ID, user.Role)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, wrapInternal("Login/GetTOTPSettings", err)
	}
	if settings != nil && settings.Enabled {
		if input.TrustedDeviceToken != "" && s.validateTrustedDevice(ctx, input.TrustedDeviceToken, user, input.UserAgent) {
			_ = s.twoFARepo.CreateAudit(ctx, &user.ID, &user.Role, "trusted_device_used", cleanIP(input.IPAddress), input.UserAgent, "", s.timeManager.Now().UnixMilli())
			return s.createAuthResponse(ctx, user)
		}
		return s.createTwoFactorChallenge(ctx, user, input)
	}

	return s.createAuthResponse(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, userCtx *domain.UserTokenContext) (*domain.AuthResponse, error) {
	user, err := s.usersRepo.GetByID(ctx, userCtx.ID, userCtx.Role)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeUnauthorized, ErrUserNotFound, "USER_NOT_FOUND", "user not found")
		}
		return nil, wrapInternal("Refresh/GetByID", err)
	}

	revokedAt := s.timeManager.Now().UnixMilli()
	if err := s.sessionsRepo.DeleteByNumber(ctx, user.ID, user.Role, userCtx.Number, revokedAt); err != nil {
		return nil, wrapInternal("Refresh/DeleteByNumber", err)
	}

	return s.createAuthResponse(ctx, user)
}

func (s *authService) Logout(ctx context.Context, userCtx *domain.UserTokenContext) error {
	revokedAt := s.timeManager.Now().UnixMilli()
	return s.sessionsRepo.DeleteByNumber(ctx, userCtx.ID, userCtx.Role, userCtx.Number, revokedAt)
}

func (s *authService) LogoutAll(ctx context.Context, userID int64, role domain.Role) error {
	revokedAt := s.timeManager.Now().UnixMilli()
	return s.sessionsRepo.DeleteAllByUser(ctx, userID, role, revokedAt)
}

func (s *authService) ValidateToken(ctx context.Context, purpose domain.TokenPurpose, tokenString string) (*domain.UserTokenContext, error) {
	data, err := s.tokenMgr.ParseToken(tokenString)
	if err != nil {
		return nil, newServiceError(CodeUnauthorized, fmt.Errorf("ValidateToken/ParseToken: %w", err), "UNAUTHORIZED", "unauthorized")
	}

	if data.Purpose != purpose {
		return nil, newServiceError(CodeUnauthorized, errors.New("invalid token purpose"), "UNAUTHORIZED", "unauthorized")
	}

	if s.timeManager.Now().After(data.ExpiresAt) {
		return nil, newServiceError(CodeUnauthorized, ErrTokenExpired, "TOKEN_EXPIRED", "token expired")
	}

	dbSecret, err := s.sessionsRepo.GetSecret(ctx, data.ID, data.Role, data.Number, purpose)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeUnauthorized, ErrSessionNotFound, "SESSION_NOT_FOUND", "session not found")
		}
		return nil, wrapInternal("ValidateToken/GetSecret", err)
	}

	if dbSecret != data.Secret {
		return nil, newServiceError(CodeUnauthorized, ErrInvalidTokenSecret, "UNAUTHORIZED", "unauthorized")
	}

	return &domain.UserTokenContext{ID: data.ID, Role: data.Role, Number: data.Number, Authorized: true}, nil
}

func (s *authService) SetupTOTP(ctx context.Context, userCtx *domain.UserTokenContext) (*domain.TwoFactorSetupResponse, error) {
	user, err := s.usersRepo.GetByID(ctx, userCtx.ID, userCtx.Role)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeUnauthorized, ErrUserNotFound, "USER_NOT_FOUND", "user not found")
		}
		return nil, wrapInternal("SetupTOTP/GetByID", err)
	}
	existing, err := s.twoFARepo.GetTOTPSettings(ctx, user.ID, user.Role)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, wrapInternal("SetupTOTP/GetTOTPSettings", err)
	}
	if existing != nil && existing.Enabled {
		return nil, newServiceError(CodeConflict, ErrTwoFactorEnabled, "TOTP_ALREADY_ENABLED", "totp already enabled")
	}

	secret, err := totp.GenerateSecret()
	if err != nil {
		return nil, wrapInternal("SetupTOTP/GenerateSecret", err)
	}
	ciphertext, err := s.secretBox.EncryptString(secret, totpAAD(user.ID, user.Role))
	if err != nil {
		return nil, wrapInternal("SetupTOTP/EncryptSecret", err)
	}

	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("SetupTOTP/StartTransaction", err)
	}
	defer tx.Rollback()

	now := s.timeManager.Now().UnixMilli()
	if err := s.twoFARepo.UpsertTOTPSettingsTX(ctx, tx, &domain.TwoFactorSettings{
		UserID: user.ID, Role: user.Role, SecretCiphertext: ciphertext, Enabled: false, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return nil, wrapInternal("SetupTOTP/UpsertTOTPSettingsTX", err)
	}
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "2fa_setup_started", "", "", "", now); err != nil {
		return nil, wrapInternal("SetupTOTP/CreateAuditTX", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("SetupTOTP/Commit", err)
	}

	return &domain.TwoFactorSetupResponse{
		Secret:     secret,
		OTPAuthURL: totp.BuildURL(s.issuer, user.Login, secret),
	}, nil
}

func (s *authService) ConfirmTOTP(ctx context.Context, userCtx *domain.UserTokenContext, code string) (*domain.RecoveryCodesResponse, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("ConfirmTOTP/StartTransaction", err)
	}
	defer tx.Rollback()

	settings, err := s.twoFARepo.GetTOTPSettingsTX(ctx, tx, userCtx.ID, userCtx.Role)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeBadRequest, ErrTwoFactorInvalid, "TOTP_NOT_SETUP", "totp is not setup")
		}
		return nil, wrapInternal("ConfirmTOTP/GetTOTPSettingsTX", err)
	}
	secret, err := s.secretBox.DecryptString(settings.SecretCiphertext, totpAAD(userCtx.ID, userCtx.Role))
	if err != nil {
		return nil, wrapInternal("ConfirmTOTP/DecryptSecret", err)
	}
	if !totp.Validate(secret, code, s.timeManager.Now(), 1) {
		_ = s.twoFARepo.CreateAuditTX(ctx, tx, &userCtx.ID, &userCtx.Role, "2fa_failed", "", "", "", s.timeManager.Now().UnixMilli())
		return nil, newServiceError(CodeUnauthorized, ErrTwoFactorInvalid, "INVALID_2FA_CODE", "invalid two factor code")
	}

	codes, hashes, err := s.generateRecoveryCodes()
	if err != nil {
		return nil, wrapInternal("ConfirmTOTP/GenerateRecoveryCodes", err)
	}
	now := s.timeManager.Now().UnixMilli()
	if err := s.twoFARepo.EnableTOTPTX(ctx, tx, userCtx.ID, userCtx.Role, now); err != nil {
		return nil, wrapInternal("ConfirmTOTP/EnableTOTPTX", err)
	}
	if err := s.twoFARepo.ReplaceRecoveryCodesTX(ctx, tx, userCtx.ID, userCtx.Role, hashes, now); err != nil {
		return nil, wrapInternal("ConfirmTOTP/ReplaceRecoveryCodesTX", err)
	}
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &userCtx.ID, &userCtx.Role, "2fa_enabled", "", "", "", now); err != nil {
		return nil, wrapInternal("ConfirmTOTP/CreateAuditTX", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("ConfirmTOTP/Commit", err)
	}
	return &domain.RecoveryCodesResponse{Codes: codes}, nil
}

func (s *authService) VerifyTOTPChallenge(ctx context.Context, input *domain.TwoFactorVerifyInput) (*domain.AuthResponse, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("VerifyTOTPChallenge/StartTransaction", err)
	}
	defer tx.Rollback()

	challenge, err := s.twoFARepo.GetChallengeTX(ctx, tx, input.ChallengeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, newServiceError(CodeUnauthorized, ErrChallengeNotFound, "CHALLENGE_NOT_FOUND", "challenge not found")
		}
		return nil, wrapInternal("VerifyTOTPChallenge/GetChallengeTX", err)
	}
	now := s.timeManager.Now().UnixMilli()
	if challenge.ConsumedAt != nil || challenge.Purpose != "login_2fa" {
		return nil, newServiceError(CodeUnauthorized, ErrChallengeNotFound, "CHALLENGE_NOT_FOUND", "challenge not found")
	}
	if challenge.ExpiresAt < now {
		return nil, newServiceError(CodeUnauthorized, ErrChallengeExpired, "CHALLENGE_EXPIRED", "challenge expired")
	}
	if challenge.FailedAttempts >= 5 {
		return nil, newServiceError(CodeTooManyRequests, ErrTwoFactorInvalid, "TOO_MANY_ATTEMPTS", "too many attempts")
	}

	user, err := s.usersRepo.GetByIDTX(ctx, tx, challenge.UserID, challenge.Role)
	if err != nil {
		return nil, wrapInternal("VerifyTOTPChallenge/GetByIDTX", err)
	}
	ok, err := s.validateTOTPOrRecoveryTX(ctx, tx, user, input.Code, input.RecoveryCode, now)
	if err != nil {
		return nil, err
	}
	if !ok {
		_ = s.twoFARepo.IncrementChallengeFailuresTX(ctx, tx, challenge.ID)
		_ = s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "2fa_failed", cleanIP(input.IPAddress), input.UserAgent, "", now)
		return nil, newServiceError(CodeUnauthorized, ErrTwoFactorInvalid, "INVALID_2FA_CODE", "invalid two factor code")
	}
	if err := s.twoFARepo.ConsumeChallengeTX(ctx, tx, challenge.ID, now); err != nil {
		return nil, wrapInternal("VerifyTOTPChallenge/ConsumeChallengeTX", err)
	}

	var trustedToken string
	if input.TrustDevice {
		var err error
		trustedToken, err = s.createTrustedDeviceTX(ctx, tx, user, input.DeviceName, input.UserAgent, now)
		if err != nil {
			return nil, err
		}
	}

	resp, err := s.createAuthResponseTX(ctx, tx, user)
	if err != nil {
		return nil, err
	}
	resp.TrustedDeviceToken = trustedToken
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "2fa_verified", cleanIP(input.IPAddress), input.UserAgent, "", now); err != nil {
		return nil, wrapInternal("VerifyTOTPChallenge/CreateAuditTX", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("VerifyTOTPChallenge/Commit", err)
	}
	return resp, nil
}

func (s *authService) DisableTOTP(ctx context.Context, userCtx *domain.UserTokenContext, code string) error {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return wrapInternal("DisableTOTP/StartTransaction", err)
	}
	defer tx.Rollback()
	user, err := s.usersRepo.GetByIDTX(ctx, tx, userCtx.ID, userCtx.Role)
	if err != nil {
		return wrapInternal("DisableTOTP/GetByIDTX", err)
	}
	ok, err := s.validateTOTPOrRecoveryTX(ctx, tx, user, code, code, s.timeManager.Now().UnixMilli())
	if err != nil {
		return err
	}
	if !ok {
		return newServiceError(CodeUnauthorized, ErrTwoFactorInvalid, "INVALID_2FA_CODE", "invalid two factor code")
	}
	now := s.timeManager.Now().UnixMilli()
	if err := s.twoFARepo.DisableTOTPTX(ctx, tx, user.ID, user.Role, now); err != nil {
		return wrapInternal("DisableTOTP/DisableTOTPTX", err)
	}
	if err := s.twoFARepo.RevokeTrustedDevicesTX(ctx, tx, user.ID, user.Role, now); err != nil {
		return wrapInternal("DisableTOTP/RevokeTrustedDevicesTX", err)
	}
	if err := s.sessionsRepo.DeleteAllByUserTX(ctx, tx, user.ID, user.Role, now); err != nil {
		return wrapInternal("DisableTOTP/DeleteAllByUserTX", err)
	}
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "2fa_disabled", "", "", "", now); err != nil {
		return wrapInternal("DisableTOTP/CreateAuditTX", err)
	}
	return tx.Commit()
}

func (s *authService) RegenerateRecoveryCodes(ctx context.Context, userCtx *domain.UserTokenContext, code string) (*domain.RecoveryCodesResponse, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("RegenerateRecoveryCodes/StartTransaction", err)
	}
	defer tx.Rollback()
	user, err := s.usersRepo.GetByIDTX(ctx, tx, userCtx.ID, userCtx.Role)
	if err != nil {
		return nil, wrapInternal("RegenerateRecoveryCodes/GetByIDTX", err)
	}
	ok, err := s.validateTOTPOrRecoveryTX(ctx, tx, user, code, code, s.timeManager.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, newServiceError(CodeUnauthorized, ErrTwoFactorInvalid, "INVALID_2FA_CODE", "invalid two factor code")
	}
	codes, hashes, err := s.generateRecoveryCodes()
	if err != nil {
		return nil, wrapInternal("RegenerateRecoveryCodes/generateRecoveryCodes", err)
	}
	now := s.timeManager.Now().UnixMilli()
	if err := s.twoFARepo.ReplaceRecoveryCodesTX(ctx, tx, user.ID, user.Role, hashes, now); err != nil {
		return nil, wrapInternal("RegenerateRecoveryCodes/ReplaceRecoveryCodesTX", err)
	}
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "recovery_codes_regenerated", "", "", "", now); err != nil {
		return nil, wrapInternal("RegenerateRecoveryCodes/CreateAuditTX", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("RegenerateRecoveryCodes/Commit", err)
	}
	return &domain.RecoveryCodesResponse{Codes: codes}, nil
}

func (s *authService) createTwoFactorChallenge(ctx context.Context, user *domain.User, input *domain.AuthenticationInput) (*domain.AuthResponse, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("createTwoFactorChallenge/StartTransaction", err)
	}
	defer tx.Rollback()
	now := s.timeManager.Now()
	challengeID := uuid.New().String()
	expiresAt := now.Add(s.challengeTTL).UnixMilli()
	if err := s.twoFARepo.CreateChallengeTX(ctx, tx, &domain.AuthChallenge{
		ID: challengeID, UserID: user.ID, Role: user.Role, Purpose: "login_2fa", ExpiresAt: expiresAt, CreatedAt: now.UnixMilli(),
	}); err != nil {
		return nil, wrapInternal("createTwoFactorChallenge/CreateChallengeTX", err)
	}
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "login_2fa_required", cleanIP(input.IPAddress), input.UserAgent, "", now.UnixMilli()); err != nil {
		return nil, wrapInternal("createTwoFactorChallenge/CreateAuditTX", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("createTwoFactorChallenge/Commit", err)
	}
	return &domain.AuthResponse{TwoFactorRequired: true, TwoFactorChallenge: challengeID, TwoFactorExpiresAt: expiresAt, ServerTime: now.UnixMilli()}, nil
}

func (s *authService) createAuthResponse(ctx context.Context, user *domain.User) (*domain.AuthResponse, error) {
	tx, err := s.txRepo.StartTransaction(ctx)
	if err != nil {
		return nil, wrapInternal("createAuthResponse/StartTransaction", err)
	}
	defer tx.Rollback()
	resp, err := s.createAuthResponseTX(ctx, tx, user)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, wrapInternal("createAuthResponse/Commit", err)
	}
	return resp, nil
}

func (s *authService) createAuthResponseTX(ctx context.Context, tx transaction.Transaction, user *domain.User) (*domain.AuthResponse, error) {
	number, err := s.sessionsRepo.GetNextNumberTX(ctx, tx, user.ID, user.Role)
	if err != nil {
		return nil, wrapInternal("createAuthResponseTX/GetNextNumberTX", err)
	}

	accessToken, err := s.generateAndSaveTokenTX(ctx, tx, user, number, domain.PurposeAccess, s.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndSaveTokenTX(ctx, tx, user, number, domain.PurposeRefresh, s.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: domain.TokenResponse{
			Token:     accessToken.Token,
			ExpiresAt: accessToken.ExpiresAt,
			Type:      "Bearer",
		},
		RefreshToken: domain.TokenResponse{
			Token:     refreshToken.Token,
			ExpiresAt: refreshToken.ExpiresAt,
			Type:      "Bearer",
		},
		ServerTime: s.timeManager.Now().UnixMilli(),
		User: domain.UserResponse{
			ID:         user.ID,
			Login:      user.Login,
			Email:      user.Email,
			Phone:      user.Phone,
			Role:       string(user.Role),
			IsVerified: user.IsVerified,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			MiddleName: user.MiddleName,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		},
	}, nil
}

func (s *authService) generateAndSaveToken(
	ctx context.Context,
	user *domain.User,
	number int,
	purpose domain.TokenPurpose,
	ttl time.Duration,
) (*tokenInfo, error) {
	secret := hex.EncodeToString([]byte(uuid.New().String() + uuid.New().String())[:32])

	now := s.timeManager.Now()
	expiresAt := now.Add(ttl)

	data := &domain.TokenData{
		ID:        user.ID,
		Role:      user.Role,
		Login:     user.Login,
		Purpose:   purpose,
		Number:    number,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	tokenStr, err := s.tokenMgr.GenerateToken(data)
	if err != nil {
		return nil, wrapInternal("generateAndSaveToken/GenerateToken", err)
	}

	session := &domain.Session{
		UserID:    user.ID,
		Role:      user.Role,
		Purpose:   purpose,
		Number:    number,
		Secret:    secret,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	if err := s.sessionsRepo.Create(ctx, session); err != nil {
		return nil, wrapInternal("generateAndSaveToken/CreateSession", err)
	}

	return &tokenInfo{
		Token:     tokenStr,
		ExpiresAt: expiresAt.UnixMilli(),
	}, nil
}

func (s *authService) generateAndSaveTokenTX(
	ctx context.Context,
	tx transaction.Transaction,
	user *domain.User,
	number int,
	purpose domain.TokenPurpose,
	ttl time.Duration,
) (*tokenInfo, error) {
	secret := hex.EncodeToString([]byte(uuid.New().String() + uuid.New().String())[:32])

	now := s.timeManager.Now()
	expiresAt := now.Add(ttl)

	data := &domain.TokenData{
		ID:        user.ID,
		Role:      user.Role,
		Login:     user.Login,
		Purpose:   purpose,
		Number:    number,
		Secret:    secret,
		ExpiresAt: expiresAt,
	}

	tokenStr, err := s.tokenMgr.GenerateToken(data)
	if err != nil {
		return nil, wrapInternal("generateAndSaveTokenTX/GenerateToken", err)
	}

	session := &domain.Session{
		UserID:    user.ID,
		Role:      user.Role,
		Purpose:   purpose,
		Number:    number,
		Secret:    secret,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	if err := s.sessionsRepo.CreateTX(ctx, tx, session); err != nil {
		return nil, wrapInternal("generateAndSaveTokenTX/CreateSessionTX", err)
	}

	return &tokenInfo{
		Token:     tokenStr,
		ExpiresAt: expiresAt.UnixMilli(),
	}, nil
}

func (s *authService) validateTOTPOrRecoveryTX(ctx context.Context, tx transaction.Transaction, user *domain.User, code, recoveryCode string, now int64) (bool, error) {
	settings, err := s.twoFARepo.GetTOTPSettingsTX(ctx, tx, user.ID, user.Role)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, nil
		}
		return false, wrapInternal("validateTOTPOrRecoveryTX/GetTOTPSettingsTX", err)
	}
	if code != "" {
		secret, err := s.secretBox.DecryptString(settings.SecretCiphertext, totpAAD(user.ID, user.Role))
		if err != nil {
			return false, wrapInternal("validateTOTPOrRecoveryTX/DecryptSecret", err)
		}
		if totp.Validate(secret, code, s.timeManager.Now(), 1) {
			return true, nil
		}
	}
	if recoveryCode == "" {
		return false, nil
	}
	codes, err := s.twoFARepo.GetUnusedRecoveryCodesTX(ctx, tx, user.ID, user.Role)
	if err != nil {
		return false, wrapInternal("validateTOTPOrRecoveryTX/GetUnusedRecoveryCodesTX", err)
	}
	for id, hash := range codes {
		if s.hasher.Compare(hash, normalizeRecoveryCode(recoveryCode)) {
			if err := s.twoFARepo.UseRecoveryCodeTX(ctx, tx, id, now); err != nil {
				return false, wrapInternal("validateTOTPOrRecoveryTX/UseRecoveryCodeTX", err)
			}
			_ = s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "recovery_code_used", "", "", "", now)
			return true, nil
		}
	}
	return false, nil
}

func (s *authService) validateTrustedDevice(ctx context.Context, rawToken string, user *domain.User, userAgent string) bool {
	tokenHash := s.hashTrustedDeviceToken(rawToken)
	device, err := s.twoFARepo.GetTrustedDeviceByHash(ctx, tokenHash)
	if err != nil || device.UserID != user.ID || device.Role != user.Role || device.RevokedAt != nil {
		return false
	}
	now := s.timeManager.Now().UnixMilli()
	if device.ExpiresAt < now {
		return false
	}
	if device.UserAgentHash != "" && device.UserAgentHash != hashString(userAgent) {
		return false
	}
	_ = s.twoFARepo.TouchTrustedDevice(ctx, device.ID, now)
	return true
}

func (s *authService) createTrustedDeviceTX(ctx context.Context, tx transaction.Transaction, user *domain.User, deviceName, userAgent string, now int64) (string, error) {
	raw, err := randomURLToken(32)
	if err != nil {
		return "", wrapInternal("createTrustedDeviceTX/randomURLToken", err)
	}
	device := &domain.TrustedDevice{
		UserID:        user.ID,
		Role:          user.Role,
		TokenHash:     s.hashTrustedDeviceToken(raw),
		DeviceName:    deviceName,
		UserAgentHash: hashString(userAgent),
		ExpiresAt:     s.timeManager.Now().Add(s.trustedTTL).UnixMilli(),
		LastUsedAt:    &now,
		CreatedAt:     now,
	}
	if err := s.twoFARepo.CreateTrustedDeviceTX(ctx, tx, device); err != nil {
		return "", wrapInternal("createTrustedDeviceTX/CreateTrustedDeviceTX", err)
	}
	if err := s.twoFARepo.CreateAuditTX(ctx, tx, &user.ID, &user.Role, "trusted_device_created", "", userAgent, "", now); err != nil {
		return "", wrapInternal("createTrustedDeviceTX/CreateAuditTX", err)
	}
	return raw, nil
}

func (s *authService) generateRecoveryCodes() ([]string, []string, error) {
	codes := make([]string, 10)
	hashes := make([]string, 10)
	for i := range codes {
		raw, err := randomRecoveryCode()
		if err != nil {
			return nil, nil, err
		}
		hash, err := s.hasher.Hash(normalizeRecoveryCode(raw))
		if err != nil {
			return nil, nil, err
		}
		codes[i] = raw
		hashes[i] = hash
	}
	return codes, hashes, nil
}

func (s *authService) hashTrustedDeviceToken(raw string) string {
	mac := hmac.New(sha256.New, []byte(s.trustedPepper))
	_, _ = mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}

func randomURLToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func randomRecoveryCode() (string, error) {
	buf := make([]byte, 10)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	encoded := strings.ToUpper(base32NoPad(buf))
	return encoded[:5] + "-" + encoded[5:10] + "-" + encoded[10:15], nil
}

func base32NoPad(raw []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
}

func normalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
}

func totpAAD(userID int64, role domain.Role) []byte {
	return []byte(fmt.Sprintf("totp:%d:%s", userID, role))
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func cleanIP(value string) string {
	if value == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	ip := net.ParseIP(value)
	if ip == nil {
		return value
	}
	return ip.String()
}

type tokenInfo struct {
	Token     string
	ExpiresAt int64
}

type BCryptHasher struct {
	cost int
}

func NewBCryptHasher(cost int) *BCryptHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &BCryptHasher{cost: cost}
}

func (h *BCryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BCryptHasher) Compare(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
