package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modelmarket/internal/models"
	"modelmarket/internal/services"
	"modelmarket/pkg/logger"
)

type DistributionHandler struct {
	db  *gorm.DB
	svc *services.DistributionService
}

func NewDistributionHandler(db *gorm.DB, svc *services.DistributionService) *DistributionHandler {
	return &DistributionHandler{db: db, svc: svc}
}

// AdminList GET /api/v1/admin/links
func (h *DistributionHandler) AdminList(c *gin.Context) {
	var rows []models.DistributionLink
	if err := h.db.Preload("Model").Preload("Model.Vendor").Order("id desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type genReq struct {
	ModelID uint `json:"model_id" binding:"required"`
}

// Generate POST /api/v1/admin/links  为指定模型生成推广链接
func (h *DistributionHandler) Generate(c *gin.Context) {
	var req genReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	link, err := h.svc.GenerateForModel(req.ModelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, link)
}

type updateLinkReq struct {
	CustomURL *string `json:"custom_url"`
	IsActive  *bool   `json:"is_active"`
}

// AdminUpdate PUT /api/v1/admin/links/:id
func (h *DistributionHandler) AdminUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var link models.DistributionLink
	if err := h.db.First(&link, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req updateLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]any{}
	if req.CustomURL != nil {
		updates["custom_url"] = *req.CustomURL
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if err := h.db.Model(&link).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, link)
}

// AdminDelete DELETE /api/v1/admin/links/:id
func (h *DistributionHandler) AdminDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Delete(&models.DistributionLink{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Resolve GET /r/:code  公开跳转入口
func (h *DistributionHandler) Resolve(c *gin.Context) {
	code := c.Param("code")
	var link models.DistributionLink
	if err := h.db.Where("link_code = ?", code).First(&link).Error; err != nil {
		c.String(http.StatusNotFound, "link not found")
		return
	}
	if !link.IsActive {
		c.String(http.StatusGone, "link disabled")
		return
	}
	target, err := h.svc.Resolve(&link)
	if err != nil {
		logger.L().Errorf("Resolve link %s failed: %v", code, err)
		c.String(http.StatusInternalServerError, "resolve failed")
		return
	}
	go func() {
		if err := h.svc.RecordClick(link.ID, c.Request.Referer(), c.Request.UserAgent(), c.ClientIP()); err != nil {
			logger.L().Warnf("Record click failed: %v", err)
		}
	}()
	logger.L().Infof("Distribution redirect: code=%s -> %s", code, target)
	c.Redirect(http.StatusFound, target)
}
