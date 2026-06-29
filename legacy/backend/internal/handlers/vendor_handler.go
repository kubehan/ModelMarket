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

type VendorHandler struct {
	db   *gorm.DB
	svc  *services.VendorService
}

func NewVendorHandler(db *gorm.DB, svc *services.VendorService) *VendorHandler {
	return &VendorHandler{db: db, svc: svc}
}

// Schemas GET /api/v1/admin/vendors/schemas
// 返回认证方式 + 推广来源的字段 schema，前端按此动态渲染表单
func (h *VendorHandler) Schemas(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"auth_schemas":  services.AllAuthSchemas(),
		"promo_schemas": services.AllPromoSchemas(),
	})
}

// List GET /api/v1/admin/vendors
func (h *VendorHandler) List(c *gin.Context) {
	var vendors []models.Vendor
	if err := h.db.Order("id desc").Find(&vendors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 解密 auth_config，但返回脱敏版
	out := make([]map[string]any, 0, len(vendors))
	for i := range vendors {
		v := &vendors[i]
		cfg, _ := h.svc.LoadAuthConfig(v)
		out = append(out, vendorToMap(v, services.MaskAuthConfig(cfg)))
	}
	c.JSON(http.StatusOK, out)
}

// Get GET /api/v1/admin/vendors/:id
func (h *VendorHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var v models.Vendor
	if err := h.db.First(&v, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	cfg, _ := h.svc.LoadAuthConfig(&v)
	c.JSON(http.StatusOK, vendorToMap(&v, services.MaskAuthConfig(cfg)))
}

type vendorReq struct {
	ID              uint           `json:"id"`
	Name            string         `json:"name" binding:"required"`
	OfficialURL     string         `json:"official_url"`
	APIBase         string         `json:"api_base"`
	LogoURL         string         `json:"logo_url"`
	Description     string         `json:"description"`
	AuthType        string         `json:"auth_type" binding:"required"`
	AuthConfig      map[string]any `json:"auth_config"`
	PromoSourceType string         `json:"promo_source_type"`
	PromoConfig     map[string]any `json:"promo_config"`
	IsActive        *bool          `json:"is_active"`
}

// Create POST /api/v1/admin/vendors
func (h *VendorHandler) Create(c *gin.Context) {
	var req vendorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v := &models.Vendor{
		Name: req.Name, OfficialURL: req.OfficialURL, APIBase: req.APIBase,
		LogoURL: req.LogoURL, Description: req.Description,
		AuthType: req.AuthType, PromoSourceType: req.PromoSourceType,
		IsActive: true,
	}
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}
	if err := h.svc.CreateOrUpdate(v, req.AuthConfig, req.PromoConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vendorToMap(v, services.MaskAuthConfig(req.AuthConfig)))
}

// Update PUT /api/v1/admin/vendors/:id
func (h *VendorHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var v models.Vendor
	if err := h.db.First(&v, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req vendorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v.Name = req.Name
	v.OfficialURL = req.OfficialURL
	v.APIBase = req.APIBase
	v.LogoURL = req.LogoURL
	v.Description = req.Description
	v.AuthType = req.AuthType
	v.PromoSourceType = req.PromoSourceType
	if req.IsActive != nil {
		v.IsActive = *req.IsActive
	}
	// 若 auth_config 字段全部为空/未传，则保留原密文，不重新加密
	saveAuth := req.AuthConfig
	if hasOnlyMaskedValues(req.AuthConfig) {
		logger.L().Debugf("Vendor %d: keep existing auth_config (masked input)", v.ID)
		saveAuth = nil
	}
	if err := h.svc.CreateOrUpdate(&v, saveAuth, req.PromoConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cfg, _ := h.svc.LoadAuthConfig(&v)
	c.JSON(http.StatusOK, vendorToMap(&v, services.MaskAuthConfig(cfg)))
}

// Delete DELETE /api/v1/admin/vendors/:id
func (h *VendorHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Delete(&models.Vendor{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// TestConnection POST /api/v1/admin/vendors/:id/test
func (h *VendorHandler) TestConnection(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var v models.Vendor
	if err := h.db.First(&v, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if _, err := h.svc.LoadAuthConfig(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	infos, err := h.svc.TestConnection(&v)
	if err != nil {
		logger.L().Warnf("TestConnection failed vendor=%s: %v", v.Name, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "models": infos, "count": len(infos)})
}

// SyncModels POST /api/v1/admin/vendors/:id/sync
func (h *VendorHandler) SyncModels(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var v models.Vendor
	if err := h.db.First(&v, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if _, err := h.svc.LoadAuthConfig(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.svc.SyncModels(&v)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "newly_added": n})
}

func vendorToMap(v *models.Vendor, maskedAuth map[string]any) map[string]any {
	return map[string]any{
		"id":                 v.ID,
		"name":               v.Name,
		"official_url":       v.OfficialURL,
		"api_base":           v.APIBase,
		"logo_url":           v.LogoURL,
		"description":        v.Description,
		"auth_type":          v.AuthType,
		"auth_config":        maskedAuth,
		"promo_source_type":  v.PromoSourceType,
		"promo_config":       services.PromoConfigMap(v),
		"is_active":          v.IsActive,
		"last_tested_at":     v.LastTestedAt,
		"last_test_status":   v.LastTestStatus,
		"last_test_message":  v.LastTestMessage,
		"created_at":         v.CreatedAt,
		"updated_at":         v.UpdatedAt,
	}
}

// hasOnlyMaskedValues 检查 auth_config 是否全部是脱敏占位符（含 ****）
// 若是，说明前端未真正修改 -> 后端保留原值。
func hasOnlyMaskedValues(c map[string]any) bool {
	if len(c) == 0 {
		return true
	}
	hasReal := false
	for _, v := range c {
		s, ok := v.(string)
		if !ok || s == "" {
			continue
		}
		if !containsMaskMarker(s) {
			hasReal = true
			break
		}
	}
	return !hasReal
}

func containsMaskMarker(s string) bool {
	for i := 0; i+3 < len(s); i++ {
		if s[i] == '*' && s[i+1] == '*' && s[i+2] == '*' && s[i+3] == '*' {
			return true
		}
	}
	return false
}
