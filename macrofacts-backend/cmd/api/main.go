package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/auth"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/foods"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/httpapi"
)

func main() {
	port := envOr("PORT", "8080")

	// Postgres
	pgURL := mustEnv("DATABASE_URL")

	// Mongo OFF
	mongoURI := mustEnv("MONGO_URI")
	offDB := envOr("OFF_DB", "off")
	offCol := envOr("OFF_COLLECTION", "products")
	offSearchMode := envOr("OFF_SEARCH_MODE", "regex") // "regex" (default) or "text"

	jwtSecret := []byte(mustEnv("JWT_SECRET"))

	ctx := context.Background()

	// --- Postgres ---
	pg, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		panic(err)
	}
	defer pg.Close()

	// --- Mongo ---
	mc, err := foods.NewMongoClient(ctx, mongoURI)
	if err != nil {
		panic(err)
	}
	defer mc.Disconnect(context.Background())

	// Repos
	customRepo := foods.NewRepoPostgres(pg)
	offRepo := foods.NewRepoMongoOFF(mc, offDB, offCol, offSearchMode)

	// Service
	foodsSvc := foods.NewService(offRepo, customRepo)

	// Ensure OFF text index (safe to call repeatedly)
	{
		cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		if err := foodsSvc.EnsureOffIndexes(cctx); err != nil {
			panic(err)
		}
	}

	foodsHandler := foods.NewHandler(foodsSvc)
	authSvc := auth.NewService(pg, jwtSecret)
	authHandler := auth.NewHandler(authSvc)

	// --- HTTP ---
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	logger := slog.Default()
	r.Use(httpapi.RequestIDMiddleware())
	r.Use(httpapi.LoggerMiddleware(logger))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Server is running"})
	})

	api := r.Group("/api")

	// Auth
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.GET("/me", authSvc.Middleware(), authHandler.Me)

	// Foods: public read endpoints
	foodsHandler.RegisterRoutes(api)

	// Foods: protect create endpoint
	// (Handler already checks, but middleware makes it consistent.)
	api.POST("/foods", authSvc.Middleware(), foodsHandler.CreateCustom)

	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env var: " + k)
	}
	return v
}
