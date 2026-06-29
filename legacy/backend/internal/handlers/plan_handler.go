package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modelmarket/internal/cache"
	"modelmarket/internal/config"
	"modelmarket/internal/models"
	"modelmarket/internal/services"
	"modelmarket/pkg/logger"
)

type PlanHandler struct {
	db      *gorm.DB
	cache   *cache.Cache
	distSvc *services.DistributionService
}

func NewPlanHandler(db *gorm.DB, c *cache.Cache, ds *services.DistributionService) *PlanHandler {
	return &PlanHandler{db: db, cache: c, distSvc: ds}
}

// PublicPlanDTO 首页对比表行
type PublicPlanDTO struct {
	ID               uint     `json:"id"`
	VendorID         uint     `json:"vendor_id"`
	VendorName       string   `json:"vendor_name"`
	VendorLogoURL    string   `json:"vendor_logo_url"`
	Name             string   `json:"name"`
	PlanType         string   `json:"plan_type"`
	FirstMonthPrice  float64  `json:"first_month_price"`
	MonthlyPrice     float64  `json:"monthly_price"`
	QuarterlyPrice   float64  `json:"quarterly_price"`
	YearlyPrice      float64  `json:"yearly_price"`
	RequestsPer5h    int      `json:"requests_per_5h"`
	RequestsPerWeek  int      `json:"requests_per_week"`
	RequestsPerMonth int      `json:"requests_per_month"`
	TokenLimitM      float64  `json:"token_limit_m"`
	MeasuredTokenM   float64  `json:"measured_token_m"`
	Rating           int      `json:"rating"`
	Highlights       []string `json:"highlights"`
	Tags             []string `json:"tags"`
	OtherRights      string   `json:"other_rights"`
	Note             string   `json:"note"`
	Status           string   `json:"status"`
	SupportedModels  []string `json:"supported_models"`
	DistributionURL  string   `json:"distribution_url"`
}

// PublicList GET /api/v1/plans/  首页对比表（带缓存）
func (h *PlanHandler) PublicList(c *gin.Context) {
	var out []PublicPlanDTO
	cacheKey := "plans:public:list"
	if err := h.cache.GetOrSet(cacheKey, config.Global.CacheTTLSeconds, &out, func() (any, error) {
		var plans []models.Plan
		if err := h.db.Preload("Vendor").Preload("Models").
			Where("is_active = ?", true).
			Order("sort_order asc, id asc").
			Find(&plans).Error; err != nil {
			return nil, err
		}
		var links []models.DistributionLink
		h.db.Where("plan_id IS NOT NULL").Find(&links)
		linkByPlan := map[uint]models.DistributionLink{}
		for _, l := range links {
			if l.PlanID != nil {
				linkByPlan[*l.PlanID] = l
			}
		}

		list := make([]PublicPlanDTO, 0, len(plans))
		for _, p := range plans {
			if p.Vendor == nil || !p.Vendor.IsActive {
				continue
			}
			models := make([]string, 0, len(p.Models))
			for _, m := range p.Models {
				name := m.DisplayName
				if name == "" {
					name = m.ModelName
				}
				models = append(models, name)
			}
			dto := PublicPlanDTO{
				ID: p.ID, VendorID: p.Vendor.ID, VendorName: p.Vendor.Name, VendorLogoURL: p.Vendor.LogoURL,
				Name: p.Name, PlanType: p.PlanType,
				FirstMonthPrice: p.FirstMonthPrice, MonthlyPrice: p.MonthlyPrice,
				QuarterlyPrice: p.QuarterlyPrice, YearlyPrice: p.YearlyPrice,
				RequestsPer5h: p.RequestsPer5h, RequestsPerWeek: p.RequestsPerWeek, RequestsPerMonth: p.RequestsPerMonth,
				TokenLimitM: p.TokenLimitM, MeasuredTokenM: p.MeasuredTokenM,
				Rating: p.Rating, OtherRights: p.OtherRights, Note: p.Note,
				Status: p.Status, SupportedModels: models,
				Highlights: splitNonEmpty(p.Highlights, "\n"),
				Tags:       splitNonEmpty(p.Tags, ","),
			}
			if l, ok := linkByPlan[p.ID]; ok {
				dto.DistributionURL = "/r/" + l.LinkCode
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

// RecommendGroups GET /api/v1/plans/recommendations  推荐位分组
func (h *PlanHandler) RecommendGroups(c *gin.Context) {
	type Group struct {
		Key      string          `json:"key"`
		Title    string          `json:"title"`
		Subtitle string          `json:"subtitle"`
		Plans    []PublicPlanDTO `json:"plans"`
	}

	// 复用 PublicList 数据
	var all []PublicPlanDTO
	if err := h.cache.GetOrSet("plans:public:list", config.Global.CacheTTLSeconds, &all, func() (any, error) {
		// 重新构造一遍是最稳的
		return h.loadPublicListRaw()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasTag := func(p PublicPlanDTO, t string) bool {
		for _, x := range p.Tags {
			if x == t {
				return true
			}
		}
		return false
	}

	groups := []Group{
		{Key: "top_rated", Title: "⭐ 综合推荐", Subtitle: "各维度综合评估，最值得入手的几家"},
		{Key: "latest_model", Title: "🧠 支持最新模型", Subtitle: "GLM-5.2 / Claude / Doubao 等旗舰模型"},
		{Key: "volume", Title: "📦 量大管饱", Subtitle: "额度池子最深，适合重度使用"},
		{Key: "available", Title: "🟢 不用抢", Subtitle: "随时可下单，无需蹲点拼手速"},
	}

	for i := range groups {
		var picked []PublicPlanDTO
		switch groups[i].Key {
		case "top_rated":
			for _, p := range all {
				if p.Rating >= 5 {
					picked = append(picked, p)
				}
			}
		case "latest_model":
			for _, p := range all {
				if hasTag(p, "最新模型") || hasTag(p, "GLM-5.2") || hasTag(p, "Claude") {
					picked = append(picked, p)
				}
			}
		case "volume":
			for _, p := range all {
				if hasTag(p, "量大管饱") {
					picked = append(picked, p)
				}
			}
		case "available":
			for _, p := range all {
				if hasTag(p, "不用抢") {
					picked = append(picked, p)
				}
			}
		}
		if len(picked) > 5 {
			picked = picked[:5]
		}
		groups[i].Plans = picked
	}

	c.JSON(http.StatusOK, groups)
}

func (h *PlanHandler) loadPublicListRaw() ([]PublicPlanDTO, error) {
	var plans []models.Plan
	if err := h.db.Preload("Vendor").Preload("Models").
		Where("is_active = ?", true).
		Order("sort_order asc, id asc").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	var links []models.DistributionLink
	h.db.Where("plan_id IS NOT NULL").Find(&links)
	linkByPlan := map[uint]models.DistributionLink{}
	for _, l := range links {
		if l.PlanID != nil {
			linkByPlan[*l.PlanID] = l
		}
	}
	list := make([]PublicPlanDTO, 0, len(plans))
	for _, p := range plans {
		if p.Vendor == nil || !p.Vendor.IsActive {
			continue
		}
		mm := make([]string, 0, len(p.Models))
		for _, m := range p.Models {
			name := m.DisplayName
			if name == "" {
				name = m.ModelName
			}
			mm = append(mm, name)
		}
		dto := PublicPlanDTO{
			ID: p.ID, VendorID: p.Vendor.ID, VendorName: p.Vendor.Name, VendorLogoURL: p.Vendor.LogoURL,
			Name: p.Name, PlanType: p.PlanType,
			FirstMonthPrice: p.FirstMonthPrice, MonthlyPrice: p.MonthlyPrice,
			QuarterlyPrice: p.QuarterlyPrice, YearlyPrice: p.YearlyPrice,
			RequestsPer5h: p.RequestsPer5h, RequestsPerWeek: p.RequestsPerWeek, RequestsPerMonth: p.RequestsPerMonth,
			TokenLimitM: p.TokenLimitM, MeasuredTokenM: p.MeasuredTokenM,
			Rating: p.Rating, OtherRights: p.OtherRights, Note: p.Note,
			Status: p.Status, SupportedModels: mm,
			Highlights: splitNonEmpty(p.Highlights, "\n"),
			Tags:       splitNonEmpty(p.Tags, ","),
		}
		if l, ok := linkByPlan[p.ID]; ok {
			dto.DistributionURL = "/r/" + l.LinkCode
		}
		list = append(list, dto)
	}
	return list, nil
}

// ---- 后台 CRUD ----

// AdminList GET /api/v1/admin/plans
func (h *PlanHandler) AdminList(c *gin.Context) {
	var plans []models.Plan
	q := h.db.Preload("Vendor").Preload("Models")
	if vid := c.Query("vendor_id"); vid != "" {
		q = q.Where("vendor_id = ?", vid)
	}
	if pt := c.Query("plan_type"); pt != "" {
		q = q.Where("plan_type = ?", pt)
	}
	if err := q.Order("id desc").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

type planReq struct {
	ID               uint     `json:"id"`
	VendorID         uint     `json:"vendor_id" binding:"required"`
	Name             string   `json:"name" binding:"required"`
	PlanType         string   `json:"plan_type" binding:"required"`
	FirstMonthPrice  float64  `json:"first_month_price"`
	MonthlyPrice     float64  `json:"monthly_price"`
	QuarterlyPrice   float64  `json:"quarterly_price"`
	YearlyPrice      float64  `json:"yearly_price"`
	RequestsPer5h    int      `json:"requests_per_5h"`
	RequestsPerWeek  int      `json:"requests_per_week"`
	RequestsPerMonth int      `json:"requests_per_month"`
	TokenLimitM      float64  `json:"token_limit_m"`
	MeasuredTokenM   float64  `json:"measured_token_m"`
	Rating           int      `json:"rating"`
	Highlights       string   `json:"highlights"`
	Tags             string   `json:"tags"`
	OtherRights      string   `json:"other_rights"`
	Note             string   `json:"note"`
	Status           string   `json:"status"`
	CustomLinkURL    string   `json:"custom_link_url"`
	IsActive         *bool    `json:"is_active"`
	SortOrder        int      `json:"sort_order"`
	ModelIDs         []uint   `json:"model_ids"`
}

// Create POST /api/v1/admin/plans
func (h *PlanHandler) Create(c *gin.Context) {
	var req planReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := h.applyReq(&models.Plan{}, &req)
	if err := h.db.Create(p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.syncModels(p, req.ModelIDs); err != nil {
		logger.L().Warnf("sync models err: %v", err)
	}
	// 自动生成分销链接
	if _, err := h.distSvc.GenerateForPlan(p.ID); err != nil {
		logger.L().Warnf("auto distribution for plan %d failed: %v", p.ID, err)
	}
	_ = h.cache.InvalidatePrefix("plans:")
	c.JSON(http.StatusCreated, p)
}

// Update PUT /api/v1/admin/plans/:id
func (h *PlanHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.Plan
	if err := h.db.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req planReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.applyReq(&p, &req)
	if err := h.db.Save(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req.ModelIDs != nil {
		if err := h.syncModels(&p, req.ModelIDs); err != nil {
			logger.L().Warnf("sync models err: %v", err)
		}
	}
	_ = h.cache.InvalidatePrefix("plans:")
	c.JSON(http.StatusOK, p)
}

// Delete DELETE /api/v1/admin/plans/:id
func (h *PlanHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Delete(&models.Plan{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.cache.InvalidatePrefix("plans:")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Refresh 强制刷新缓存
func (h *PlanHandler) Refresh(c *gin.Context) {
	_ = h.cache.InvalidatePrefix("plans:")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *PlanHandler) applyReq(p *models.Plan, r *planReq) *models.Plan {
	p.VendorID = r.VendorID
	p.Name = r.Name
	p.PlanType = r.PlanType
	p.FirstMonthPrice = r.FirstMonthPrice
	p.MonthlyPrice = r.MonthlyPrice
	p.QuarterlyPrice = r.QuarterlyPrice
	p.YearlyPrice = r.YearlyPrice
	p.RequestsPer5h = r.RequestsPer5h
	p.RequestsPerWeek = r.RequestsPerWeek
	p.RequestsPerMonth = r.RequestsPerMonth
	p.TokenLimitM = r.TokenLimitM
	p.MeasuredTokenM = r.MeasuredTokenM
	p.Rating = r.Rating
	p.Highlights = r.Highlights
	p.Tags = r.Tags
	p.OtherRights = r.OtherRights
	p.Note = r.Note
	if r.Status != "" {
		p.Status = r.Status
	}
	p.CustomLinkURL = r.CustomLinkURL
	p.SortOrder = r.SortOrder
	if r.IsActive != nil {
		p.IsActive = *r.IsActive
	} else if p.ID == 0 {
		p.IsActive = true
	}
	return p
}

func (h *PlanHandler) syncModels(p *models.Plan, ids []uint) error {
	if ids == nil {
		return nil
	}
	var ms []models.Model
	if len(ids) > 0 {
		if err := h.db.Where("id IN ?", ids).Find(&ms).Error; err != nil {
			return err
		}
	}
	return h.db.Model(p).Association("Models").Replace(&ms)
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	out := []string{}
	for _, part := range strings.Split(s, sep) {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
