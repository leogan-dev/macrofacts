package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/auth"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/foods"
)

type App struct {
	DB        *pgxpool.Pool
	JWTSecret []byte
}

func NewRouter(app App) *gin.Engine {
	r := gin.Default()

	// Health
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Server is running"})
	})

	authSvc := auth.NewService(app.DB, app.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)

	// Protected routes
	api := r.Group("/api")
	api.Use(authSvc.Middleware())

	api.GET("/me", authHandler.Me)

	foodsSvc := foods.NewService(app.DB)
	foodsHandler := foods.NewHandler(foodsSvc)

	api.GET("/foods/search", foodsHandler.Search)
	api.GET("/foods/barcode/:code", foodsHandler.ByBarcode)
	api.POST("/foods", foodsHandler.Create)

	return r
}
