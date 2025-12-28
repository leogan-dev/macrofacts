package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/httpapi"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.BadRequest(c, "invalid json", nil)
		return
	}
	if err := h.svc.Register(req.Username, req.Password); err != nil {
		httpapi.BadRequest(c, err.Error(), nil)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.BadRequest(c, "invalid json", nil)
		return
	}

	token, uid, un, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		httpapi.WriteError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials", nil)
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user":  gin.H{"id": uid, "username": un},
	})
}

func (h *Handler) Me(c *gin.Context) {
	c.JSON(200, gin.H{
		"id":       c.GetString("userId"),
		"username": c.GetString("username"),
	})
}
