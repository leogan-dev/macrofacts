package logs

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

func (h *Handler) Today(c *gin.Context) {
	uid := c.GetString("userId")
	resp, err := h.svc.Today(c.Request.Context(), uid)
	if err != nil {
		httpapi.BadRequest(c, err.Error(), nil)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateEntry(c *gin.Context) {
	uid := c.GetString("userId")

	var req CreateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.BadRequest(c, "invalid json", nil)
		return
	}

	id, err := h.svc.CreateEntry(c.Request.Context(), uid, req)
	if err != nil {
		httpapi.BadRequest(c, err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, CreateEntryResponse{ID: id})
}
