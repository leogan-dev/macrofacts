package main

import (
	"context"
	"log/slog"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/auth"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/db"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/foods"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/httpapi"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/logs"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	port := envOr("PORT", "8080")
	pgURL := mustEnv("DATABASE_URL")

	mongoURI := mustEnv("MONGO_URI")
	offDB := envOr("OFF_DB", "openfoodfacts")
	offCollection := envOr("OFF_COLLECTION", "products")
	offSearchMode := envOr("OFF_SEARCH_MODE", "regex")

	jwtSecret := mustEnv("JWT_SECRET")

	if v := envOr("GIN_MODE", ""); v != "" {
		gin.SetMode(v)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pgPool, err := db.Connect(ctx, pgURL)
	if err != nil {
		slog.Error("postgres connect failed", "err", err)
		os.Exit(1)
	}
	defer pgPool.Close()

	// Apply schema on startup (simple dev-friendly behavior)
	schemaSQLBytes, err := os.ReadFile("cmd/api/schema.sql")
	if err == nil {
		if err := db.ApplySchema(ctx, pgPool, string(schemaSQLBytes)); err != nil {
			slog.Error("apply schema failed", "err", err)
			os.Exit(1)
		}
	} else {
		slog.Warn("could not read schema.sql, skipping auto-migrate", "err", err)
	}

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		slog.Error("mongo connect failed", "err", err)
		os.Exit(1)
	}
	defer func() { _ = mongoClient.Disconnect(context.Background()) }()

	offRepo := foods.NewRepoMongoOFF(mongoClient, offDB, offCollection, offSearchMode)
	customRepo := foods.NewRepoPostgresCustom(pgPool)

	foodsSvc := foods.NewService(offRepo, customRepo)
	foodsHandler := foods.NewHandler(foodsSvc)

	authSvc := auth.NewService(pgPool, []byte(jwtSecret))
	authHandler := auth.NewHandler(authSvc)

	logsRepo := logs.NewRepoPostgres(pgPool)
	logsSvc := logs.NewService(logsRepo, foodsSvc, authSvc)
	logsHandler := logs.NewHandler(logsSvc)

	// Optional: OFF index (safe no-op if not needed)
	{
		idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer idxCancel()
		if err := offRepo.EnsureTextIndex(idxCtx); err != nil {
			slog.Warn("OFF EnsureTextIndex failed", "err", err, "mode", offSearchMode)
		}
	}

	r := gin.New()
	r.Use(httpapi.RequestIDMiddleware())
	r.Use(httpapi.LoggerMiddleware(slog.Default()))
	r.Use(gin.Recovery())

	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	authRequired := authSvc.Middleware()

	api.GET("/me", authRequired, authHandler.Me)
	api.GET("/me/settings", authRequired, authHandler.MeSettings)
	api.PATCH("/me/settings", authRequired, authHandler.UpdateSettings)

	api.GET("/foods/search", foodsHandler.Search)
	api.GET("/foods/barcode/:code", foodsHandler.ByBarcode)
	api.POST("/foods", authRequired, foodsHandler.CreateCustom)
	api.POST("/foods/custom", authRequired, foodsHandler.CreateCustom)

	api.GET("/logs/today", authRequired, logsHandler.Today)
	api.POST("/logs/entries", authRequired, logsHandler.CreateEntry)

	slog.Info("api listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
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
