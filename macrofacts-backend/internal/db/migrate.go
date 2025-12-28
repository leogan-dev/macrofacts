package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ApplySchema(ctx context.Context, pool *pgxpool.Pool, schemaSQL string) error {
	_, err := pool.Exec(ctx, schemaSQL)
	return err
}
