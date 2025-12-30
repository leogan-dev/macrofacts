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
	token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		httpapi.BadRequest(c, err.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

func (h *Handler) Me(c *gin.Context) {
	uid := c.GetString("userId")
	un := c.GetString("username")
	c.JSON(http.StatusOK, MeResponse{ID: uid, Username: un})
}

func (h *Handler) MeSettings(c *gin.Context) {
	uid := c.GetString("userId")
	s, err := h.svc.GetSettings(uid)
	if err != nil {
		httpapi.BadRequest(c, err.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	uid := c.GetString("userId")

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.BadRequest(c, "invalid json", nil)
		return
	}

	s, err := h.svc.UpdateSettings(uid, req)
	if err != nil {
		httpapi.BadRequest(c, err.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, s)
}
