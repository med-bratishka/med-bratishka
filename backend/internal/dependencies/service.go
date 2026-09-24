package dependencies

import (
	"medbratishka/internal/service"
	secretcrypto "medbratishka/pkg/crypto"
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

func (d *Dependencies) SecretBox() *secretcrypto.SecretBox {
	if d.secretBox == nil {
		box, err := secretcrypto.NewSecretBox(d.cfg.Auth.TwoFactorEncryptionKey)
		if err != nil {
			panic(err)
		}
		d.secretBox = box
	}
	return d.secretBox
}

func (d *Dependencies) AuthService() service.AuthService {
	if d.authService == nil {
		d.authService = service.NewAuthService(
			d.UsersRepo(),
			d.SessionsRepo(),
			d.TwoFactorRepo(),
			d.TxRepo(),
			d.TokenManager(),
			d.Hasher(),
			d.SecretBox(),
			d.TimeManager(),
			d.cfg.Auth.AccessTTL,
			d.cfg.Auth.RefreshTTL,
			d.cfg.Auth.TwoFactorChallengeTTL,
			d.cfg.Auth.TrustedDeviceTTL,
			d.cfg.Auth.TwoFactorIssuer,
			d.cfg.Auth.TwoFactorEncryptionKey,
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

func (d *Dependencies) CatalogService() service.CatalogService {
	if d.catalogService == nil {
		d.catalogService = service.NewCatalogService(
			d.TxRepo(),
			d.ClinicRepo(),
			d.DoctorRepo(),
		)
	}
	return d.catalogService
}
