package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"medbratishka/internal/domain"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	"medbratishka/pkg/time_manager"

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
}

type authService struct {
	usersRepo    repository.UsersRepository
	sessionsRepo repository.SessionsRepository
	txRepo       transaction.Repository
	tokenMgr     TokenManager
	hasher       PasswordHasher
	timeManager  time_manager.TimeManager
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

func NewAuthService(
	usersRepo repository.UsersRepository,
	sessionsRepo repository.SessionsRepository,
	txRepo transaction.Repository,
	tokenMgr TokenManager,
	hasher PasswordHasher,
	timeManager time_manager.TimeManager,
	accessTTL, refreshTTL time.Duration,
) AuthService {
	return &authService{
		usersRepo:    usersRepo,
		sessionsRepo: sessionsRepo,
		txRepo:       txRepo,
		tokenMgr:     tokenMgr,
		hasher:       hasher,
		timeManager:  timeManager,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
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
		return nil, newServiceError(CodeUnauthorized, ErrInvalidCredentials, "INVALID_CREDENTIALS", "invalid credentials")
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

func (s *authService) createAuthResponse(ctx context.Context, user *domain.User) (*domain.AuthResponse, error) {
	number, err := s.sessionsRepo.GetNextNumber(ctx, user.ID, user.Role)
	if err != nil {
		return nil, wrapInternal("createAuthResponse/GetNextNumber", err)
	}

	accessToken, err := s.generateAndSaveToken(ctx, user, number, domain.PurposeAccess, s.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateAndSaveToken(ctx, user, number, domain.PurposeRefresh, s.refreshTTL)
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
