package transaction

import (
	"context"

	"github.com/jmoiron/sqlx"
)

//go:generate  mockgen -source interface.go -destination ../../mocks/repository/transactions/interface.go

type (
	Repository interface {
		StartTransaction(ctx context.Context) (Transaction, error)
		StartReadOnlyClientTransaction(ctx context.Context) (Transaction, error)
	}

	Transaction interface {
		Commit() error
		Rollback()
		Txm() *sqlx.Tx
	}
)
