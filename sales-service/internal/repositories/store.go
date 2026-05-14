package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store пул и транзакции.
type Store struct {
	Pool *pgxpool.Pool
}

// NewStore конструктор.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{Pool: pool}
}

// DB интерфейс Pool/Tx.
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (s *Store) db(tx pgx.Tx) DB {
	if tx != nil {
		return tx
	}
	return s.Pool
}

// BeginTx начинает транзакцию.
func (s *Store) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.Pool.Begin(ctx)
}

