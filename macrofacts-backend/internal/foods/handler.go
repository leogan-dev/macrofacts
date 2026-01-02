package foods

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leogan-dev/macrofacts/macrofacts-backend/internal/httpapi"
)

type SearchResponse struct {
	Items      []FoodDTO `json:"items"`
	NextCursor *string   `json:"next_cursor,omitempty"`
}

type ItemResponse struct {
	Item *FoodDTO `json:"item"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/foods/search", h.Search)
	r.GET("/foods/barcode/:code", h.ByBarcode)

	// ✅ add this
	r.POST("/foods/custom", h.CreateCustom)
}

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	limit := parseLimit(c.Query("limit"), 25)
	cursor := strings.TrimSpace(c.Query("cursor"))

	out, next, err := h.svc.Search(c.Request.Context(), q, limit, cursor)
	if err != nil {
		httpapi.Internal(c, "search failed")
		return
	}
	c.JSON(http.StatusOK, SearchResponse{Items: out, NextCursor: next})
}

func (h *Handler) ByBarcode(c *gin.Context) {
	code := c.Param("code")

	dto, err := h.svc.ByBarcode(c.Request.Context(), code)
	if err != nil {
		httpapi.Internal(c, "lookup failed")
		return
	}
	if dto == nil {
		httpapi.NotFound(c, "not found", map[string]any{"barcode": code})
		return
	}

	c.JSON(http.StatusOK, ItemResponse{Item: dto})
}

func (h *Handler) CreateCustom(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		httpapi.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.BadRequest(c, "invalid json", nil)
		return
	}

	dto, err := h.svc.CreateCustom(c.Request.Context(), userID, req)
	if err != nil {
		httpapi.Internal(c, "create failed")
		return
	}

	c.JSON(http.StatusCreated, ItemResponse{Item: &dto})
}

func parseLimit(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 || n > 50 {
		return def
	}
	return n
}
