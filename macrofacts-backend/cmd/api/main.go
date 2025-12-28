package main

import (
	"context"
	"embed"

	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/config"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/db"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/httpapi"
)

//go:embed ../../schema.sql
var schemaSQL string

var _ embed.FS

func main() {
	cfg := config.MustLoad()
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := db.ApplySchema(ctx, pool, schemaSQL); err != nil {
		panic(err)
	}

	r := httpapi.NewRouter(httpapi.App{
		DB:        pool,
		JWTSecret: []byte(cfg.JWTSecret),
	})

	_ = r.Run(":" + cfg.Port)
}
