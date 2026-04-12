package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostgresClient struct {
	DB              *sqlx.DB
	ReadOnlyDB      *sqlx.DB
	ReadOnlyAllowed bool
}

func prepareCfg(cfg *pgx.ConnConfig) {
	if _, ok := cfg.RuntimeParams["idle_in_transaction_session_timeout"]; !ok {
		cfg.RuntimeParams["idle_in_transaction_session_timeout"] = fmt.Sprintf("%d", (30 * time.Second).Milliseconds())
	}
	if _, ok := cfg.RuntimeParams["statement_timeout"]; !ok {
		cfg.RuntimeParams["statement_timeout"] = fmt.Sprintf("%d", (30 * time.Second).Milliseconds())
	}
}

func NewPostgresClient(DSN string, pgCertLoc string) (*PostgresClient, error) {
	var client *sqlx.DB
	if pgCertLoc != "" && !strings.Contains(DSN, "disable") {
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(pgCertLoc)
		if err != nil {
			return nil, fmt.Errorf("error reading cert: %w", err)
		}
		rootCertPool.AppendCertsFromPEM(pem)
		connCfg, err := pgx.ParseConfig(DSN)
		if err != nil {
			return nil, fmt.Errorf("error parsing postgres cdn: %w", err)
		}

		prepareCfg(connCfg)
		connCfg.TLSConfig = &tls.Config{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: true,
		}
		db := stdlib.OpenDB(*connCfg)
		client = sqlx.NewDb(db, "pgx")
	} else {
		var err error
		client, err = sqlx.Connect("pgx", DSN)
		if err != nil {
			return nil, fmt.Errorf("error while connecting to postgres %w", err)
		}
	}
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("error while ping to postgres %w", err)
	}

	client.SetMaxOpenConns(100)
	client.SetMaxIdleConns(100)
	client.SetConnMaxLifetime(30 * time.Second)

	return &PostgresClient{
		DB:              client,
		ReadOnlyDB:      client,
		ReadOnlyAllowed: false,
	}, nil
}

func NewPostgresqlClientWithReadWriteSplit(readOnlyDSN, readWriteDSN, pgCertLoc string) (*PostgresClient, error) {

	readOnlyClient, err := NewPostgresClient(readOnlyDSN, pgCertLoc)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the postgresql database: %w", err)
	}

	readWriteClient, err := NewPostgresClient(readWriteDSN, pgCertLoc)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the postgresql database: %w", err)
	}

	return &PostgresClient{
		DB:              readWriteClient.DB,
		ReadOnlyDB:      readOnlyClient.ReadOnlyDB,
		ReadOnlyAllowed: true,
	}, nil
}

func (client *PostgresClient) Close() error {
	err := client.DB.Close()
	if client.ReadOnlyAllowed {
		err = client.ReadOnlyDB.Close()
	}
	return err
}
