package db

import (
	"context"
	"database/sql"
)

// Store wraps database connection and queries for easy transaction support
type Store struct {
	DB      *sql.DB
	Queries *Queries
}

// NewStore creates a new Store
func NewStore(database *sql.DB, queries *Queries) *Store {
	return &Store{
		DB:      database,
		Queries: queries,
	}
}

// ExecTx executes a function within a database transaction
func (s *Store) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := s.Queries.WithTx(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
