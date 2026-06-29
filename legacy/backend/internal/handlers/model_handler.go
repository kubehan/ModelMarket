package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modelmarket/internal/cache"
	"modelmarket/internal/config"
	"modelmarket/internal/models"
)

type ModelHandler struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewModelHandler(db *gorm.DB, c *cache.Cache) *ModelHandler {
	return &ModelHandler{db: db, cache: c}
}

// PublicModelDTO 价格对比看板的展示结构
type PublicModelDTO struct {
	ID              uint     `json:"id"`
	ModelName       string   `json:"model_name"`
	DisplayName     string   `json:"display_name"`
	ContextLength   int      `json:"context_length"`
	InputPrice      float64  `json:"input_price"`
	OutputPrice     float64  `json:"output_price"`
	ELOScore        *int     `json:"elo_score,omitempty"`
	LatencyMS       *int     `json:"latency_ms,omitempty"`
	VendorID        uint     `json:"vendor_id"`
	VendorName      string   `json:"vendor_name"`
	VendorLogoURL   string   `json:"vendor_logo_url,omitempty"`
	VendorOfficial  string   `json:"vendor_official_url,omitempty"`
	DistributionURL string   `json:"distribution_url,omitempty"` // 跳转入口
}

// PublicList GET /api/v1/models/  价格对比看板用
func (h *ModelHandler) PublicList(c *gin.Context) {
	var out []PublicModelDTO
	cacheKey := "models:public:list"
	if err := h.cache.GetOrSet(cacheKey, config.Global.CacheTTLSeconds, &out, func() (any, error) {
		var rows []models.Model
		if err := h.db.Preload("Vendor").
			Where("is_active = ?", true).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		// 一次性查所有分销链接
		var links []models.DistributionLink
		h.db.Find(&links)
		linkByModel := map[uint]models.DistributionLink{}
		for _, l := range links {
			linkByModel[l.ModelID] = l
		}

		list := make([]PublicModelDTO, 0, len(rows))
		for _, m := range rows {
			if m.Vendor == nil || !m.Vendor.IsActive {
				continue
			}
			dto := PublicModelDTO{
				ID: m.ID, ModelName: m.ModelName, DisplayName: m.DisplayName,
				ContextLength: m.ContextLength,
				InputPrice:    m.InputPrice, OutputPrice: m.OutputPrice,
				ELOScore: m.ELOScore, LatencyMS: m.LatencyMS,
				VendorID: m.Vendor.ID, VendorName: m.Vendor.Name,
				VendorLogoURL: m.Vendor.LogoURL, VendorOfficial: m.Vendor.OfficialURL,
			}
			if link, ok := linkByModel[m.ID]; ok {
				dto.DistributionURL = "/r/" + link.LinkCode
			}
			list = append(list, dto)
		}
		return list, nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// Refresh POST /api/v1/admin/models/refresh  强制刷新缓存
func (h *ModelHandler) Refresh(c *gin.Context) {
	_ = h.cache.InvalidatePrefix("models:")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// AdminList GET /api/v1/admin/models
func (h *ModelHandler) AdminList(c *gin.Context) {
	var rows []models.Model
	q := h.db.Preload("Vendor")
	if vid := c.Query("vendor_id"); vid != "" {
		q = q.Where("vendor_id = ?", vid)
	}
	if err := q.Order("id desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type modelUpdateReq struct {
	DisplayName   *string  `json:"display_name"`
	ContextLength *int     `json:"context_length"`
	InputPrice    *float64 `json:"input_price"`
	OutputPrice   *float64 `json:"output_price"`
	ELOScore      *int     `json:"elo_score"`
	LatencyMS     *int     `json:"latency_ms"`
	IsActive      *bool    `json:"is_active"`
}

// AdminUpdate PUT /api/v1/admin/models/:id
func (h *ModelHandler) AdminUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m models.Model
	if err := h.db.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req modelUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.ContextLength != nil {
		updates["context_length"] = *req.ContextLength
	}
	if req.InputPrice != nil {
		updates["input_price"] = *req.InputPrice
	}
	if req.OutputPrice != nil {
		updates["output_price"] = *req.OutputPrice
	}
	if req.ELOScore != nil {
		updates["elo_score"] = *req.ELOScore
	}
	if req.LatencyMS != nil {
		updates["latency_ms"] = *req.LatencyMS
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if err := h.db.Model(&m).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.cache.InvalidatePrefix("models:")
	c.JSON(http.StatusOK, m)
}

// AdminDelete DELETE /api/v1/admin/models/:id
func (h *ModelHandler) AdminDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Delete(&models.Model{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.cache.InvalidatePrefix("models:")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
