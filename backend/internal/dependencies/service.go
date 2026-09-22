package dependencies

import (
	"medbratishka/internal/service"
	"medbratishka/pkg/token"
)

func (d *Dependencies) TokenManager() token.TokenManager {
	if d.tokenManager == nil {
		d.tokenManager = token.NewJWTTokenManager(d.cfg.Auth.JWTSecret)
	}
	return d.tokenManager
}

func (d *Dependencies) Hasher() service.PasswordHasher {
	if d.hasher == nil {
		d.hasher = service.NewBCryptHasher(10)
	}
	return d.hasher
}

func (d *Dependencies) AuthService() service.AuthService {
	if d.authService == nil {
		d.authService = service.NewAuthService(
			d.UsersRepo(),
			d.SessionsRepo(),
			d.TxRepo(),
			d.TokenManager(),
			d.Hasher(),
			d.TimeManager(),
			d.cfg.Auth.AccessTTL,
			d.cfg.Auth.RefreshTTL,
		)
	}
	return d.authService
}

func (d *Dependencies) BindingsService() service.BindingsService {
	if d.bindingsService == nil {
		d.bindingsService = service.NewBindingsService(
			d.TxRepo(),
			d.UsersRepo(),
			d.ClinicRepo(),
			d.DoctorRepo(),
			d.PatientRepo(),
			d.TimeManager(),
		)
	}
	return d.bindingsService
}

func (d *Dependencies) ChatService() service.ChatService {
	if d.chatService == nil {
		d.chatService = service.NewChatService(
			d.TxRepo(),
			d.ChatRepo(),
			d.DoctorRepo(),
			d.PatientRepo(),
			d.TimeManager(),
			d.s3Storage,
			d.cfg.S3.MaxUploadSizeMB,
		)
	}
	return d.chatService
}
