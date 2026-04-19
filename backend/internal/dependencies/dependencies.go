package dependencies

import (
	"fmt"
	"medbratishka/pkg/token"

	"medbratishka/internal/db"
	"medbratishka/internal/handler"
	"medbratishka/internal/repository"
	"medbratishka/internal/repository/transaction"
	"medbratishka/internal/service"
	"medbratishka/pkg/config"
	"medbratishka/pkg/logger"
	"medbratishka/pkg/s3"
	"medbratishka/pkg/time_manager"
)

type Dependencies struct {
	cfg *config.Config

	pgClient *db.PostgresClient

	txRepo       transaction.Repository
	usersRepo    repository.UsersRepository
	sessionsRepo repository.SessionsRepository
	clinicRepo   repository.ClinicRepository
	doctorRepo   repository.DoctorRepository
	patientRepo  repository.PatientRepository
	chatRepo     repository.ChatRepository

	timeManager time_manager.TimeManager
	logger      logger.Logger
	s3Storage   s3.Storage

	tokenManager    token.TokenManager
	hasher          service.PasswordHasher
	authService     service.AuthService
	bindingsService service.BindingsService
	chatService     service.ChatService
	authHandler     handler.Handler
	bindingsHandler handler.Handler
	chatHandler     handler.Handler
}

func New(cfg *config.Config) (*Dependencies, error) {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	pgClient, err := db.NewPostgresClient(dsn, cfg.Database.CertLoc)
	if err != nil {
		return nil, err
	}

	log, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	var storage s3.Storage
	if cfg.S3.Endpoint != "" && cfg.S3.AccessKey != "" && cfg.S3.SecretKey != "" && cfg.S3.Bucket != "" {
		s3Client, s3Err := s3.New(cfg.S3.Endpoint, cfg.S3.Region, cfg.S3.AccessKey, cfg.S3.SecretKey, cfg.S3.Bucket, cfg.S3.UseSSL)
		if s3Err != nil {
			return nil, s3Err
		}
		storage = s3Client
	}

	return &Dependencies{
		cfg:         cfg,
		pgClient:    pgClient,
		logger:      log,
		timeManager: time_manager.New(3),
		s3Storage:   storage,
	}, nil
}

func (d *Dependencies) Close() {
	if d.pgClient != nil {
		_ = d.pgClient.Close()
	}
	if d.logger != nil {
		d.logger.Sync()
	}
}
