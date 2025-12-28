package foods

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, err := h.svc.Search(c.Request.Context(), q, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "search failed"})
		return
	}

	c.JSON(200, gin.H{"items": items})
}

func (h *Handler) ByBarcode(c *gin.Context) {
	code := c.Param("code")

	item, err := h.svc.ByBarcode(c.Request.Context(), code)
	if err != nil {
		c.JSON(500, gin.H{"error": "lookup failed"})
		return
	}
	if item == nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, item)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	uid := c.GetString("userId")
	dto, err := h.svc.Create(c.Request.Context(), uid, req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto)
}
