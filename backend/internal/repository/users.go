package repository

import (
	"context"
	"database/sql"
	"fmt"

	"medbratishka/internal/db"
	"medbratishka/internal/domain"
	"medbratishka/internal/repository/models"
	"medbratishka/internal/repository/transaction"
)

type UsersRepository interface {
	CreateTX(ctx context.Context, tx transaction.Transaction, user *domain.User) (int64, error)
	GetByAccessParameterTX(ctx context.Context, tx transaction.Transaction, accessParam string) (*domain.User, error)
	GetByIDTX(ctx context.Context, tx transaction.Transaction, id int64, role domain.Role) (*domain.User, error)
	CheckUserExistsTX(ctx context.Context, tx transaction.Transaction, login, email, phone string) (bool, error)
	GetByAccessParameter(ctx context.Context, accessParam string) (*domain.User, error)
	GetByID(ctx context.Context, id int64, role domain.Role) (*domain.User, error)
}

type pgUsersRepository struct {
	client *db.PostgresClient
}

func NewUsersRepository(client *db.PostgresClient) UsersRepository {
	return &pgUsersRepository{client: client}
}

func (r *pgUsersRepository) CreateTX(ctx context.Context, tx transaction.Transaction, user *domain.User) (int64, error) {
	query := `
		INSERT INTO users (
			login, email, phone, password_hash, role, is_verified,
			first_name, last_name, middle_name, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var userID int64
	err := tx.Txm().QueryRowContext(ctx,
		query,
		user.Login,
		user.Email,
		user.Phone,
		user.Password,
		string(user.Role),
		user.IsVerified,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return userID, nil
}

func (r *pgUsersRepository) GetByAccessParameterTX(ctx context.Context, tx transaction.Transaction, accessParam string) (*domain.User, error) {
	query := `
		SELECT id, login, email, phone, role, is_verified, password_hash,
		       first_name, last_name, middle_name, created_at, updated_at, deleted_at
		FROM users
		WHERE (login = $1 OR email = $1 OR phone = $1)
		AND deleted_at IS NULL
		LIMIT 1
	`

	var userModel models.User
	err := tx.Txm().GetContext(ctx, &userModel, query, accessParam)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return mapUserModelToDomain(&userModel), nil
}

func (r *pgUsersRepository) GetByIDTX(ctx context.Context, tx transaction.Transaction, id int64, role domain.Role) (*domain.User, error) {
	query := `
		SELECT id, login, email, phone, role, is_verified, password_hash,
		       first_name, last_name, middle_name, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND role = $2 AND deleted_at IS NULL
	`

	var userModel models.User
	err := tx.Txm().GetContext(ctx, &userModel, query, id, string(role))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	return mapUserModelToDomain(&userModel), nil
}

func (r *pgUsersRepository) CheckUserExistsTX(ctx context.Context, tx transaction.Transaction, login, email, phone string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE (login = $1 OR email = $2 OR phone = $3)
			AND deleted_at IS NULL
		)
	`
	var exists bool
	err := tx.Txm().QueryRowContext(ctx, query, login, email, phone).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}
	return exists, nil
}

func (r *pgUsersRepository) GetByAccessParameter(ctx context.Context, accessParam string) (*domain.User, error) {
	tx, err := r.client.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := r.GetByAccessParameterTX(ctx, &transaction.Tx{Tx: tx}, accessParam)
	if err != nil {
		return nil, err
	}

	_ = tx.Commit()
	return user, nil
}

func (r *pgUsersRepository) GetByID(ctx context.Context, id int64, role domain.Role) (*domain.User, error) {
	tx, err := r.client.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	user, err := r.GetByIDTX(ctx, &transaction.Tx{Tx: tx}, id, role)
	if err != nil {
		return nil, err
	}

	_ = tx.Commit()
	return user, nil
}

func mapUserModelToDomain(u *models.User) *domain.User {
	var email, phone, middleName string
	if u.Email != nil {
		email = *u.Email
	}
	if u.Phone != nil {
		phone = *u.Phone
	}
	if u.MiddleName != nil {
		middleName = *u.MiddleName
	}
	return &domain.User{
		ID:         u.ID,
		Login:      u.Login,
		Email:      email,
		Phone:      phone,
		Password:   u.Password,
		Role:       domain.Role(u.Role),
		IsVerified: u.IsVerified,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: middleName,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		DeletedAt:  u.DeletedAt,
	}
}
