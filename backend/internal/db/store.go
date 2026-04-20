package db

import (
	"fmt"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	dbsqlc "github.com/DamianAcri/DevAssistant/internal/db/sqlc"
)

type Store struct {
	Pool *pgxpool.Pool // so we can use the connection pool for transactions and queries
	Queries *dbsqlc.Queries // So we can use the generated queries
}

// NewStore creates a new Store with the given database connection pool
func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	//create pgx connection pool
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	//verify connection
	err = pool.Ping(ctx) //ping justs checks if we can connect to the database, it doesn't actually do anything else
	if err != nil {
		pool.Close() //close the pool if we can't connect to the database
		return nil, err
	}

	queries := dbsqlc.New(pool) //create a new Queries struct with the connection pool
	
	return &Store{
		Pool: pool,
		Queries: queries,
	}, nil

}

//close shuts down the connection
func (s *Store) Close() {
	s.Pool.Close()
}