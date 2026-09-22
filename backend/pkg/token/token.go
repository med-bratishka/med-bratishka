package token

import (
	"errors"
	"fmt"
	"time"

	"medbratishka/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager interface {
	GenerateToken(data *domain.TokenData) (string, error)
	ParseToken(tokenString string) (*domain.TokenData, error)
}

type jwtTokenManager struct {
	signingKey []byte
}

func NewJWTTokenManager(secret string) TokenManager {
	return &jwtTokenManager{
		signingKey: []byte(secret),
	}
}

func (m *jwtTokenManager) GenerateToken(data *domain.TokenData) (string, error) {
	claims := jwt.MapClaims{
		"id":      data.ID,
		"role":    string(data.Role),
		"login":   data.Login,
		"purpose": string(data.Purpose),
		"number":  data.Number,
		"secret":  data.Secret,
		"exp":     data.ExpiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ParseToken парсит и валидирует JWT токен
func (m *jwtTokenManager) ParseToken(tokenString string) (*domain.TokenData, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.signingKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Извлечение данных из claims
	idFloat, ok := claims["id"].(float64)
	if !ok {
		return nil, errors.New("invalid 'id' in token")
	}

	numberFloat, ok := claims["number"].(float64)
	if !ok {
		return nil, errors.New("invalid 'number' in token")
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid 'exp' in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid 'role' in token")
	}

	login, ok := claims["login"].(string)
	if !ok {
		return nil, errors.New("invalid 'login' in token")
	}

	purpose, ok := claims["purpose"].(string)
	if !ok {
		return nil, errors.New("invalid 'purpose' in token")
	}

	secret, ok := claims["secret"].(string)
	if !ok {
		return nil, errors.New("invalid 'secret' in token")
	}

	return &domain.TokenData{
		ID:        int64(idFloat),
		Role:      domain.Role(role),
		Login:     login,
		Purpose:   domain.TokenPurpose(purpose),
		Number:    int(numberFloat),
		Secret:    secret,
		ExpiresAt: time.Unix(int64(expFloat), 0),
	}, nil
}
