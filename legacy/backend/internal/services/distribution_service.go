package services

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"modelmarket/internal/cache"
	"modelmarket/internal/models"
	"modelmarket/internal/utils"
	"modelmarket/pkg/logger"
)

// DistributionService 推广分销链接业务
type DistributionService struct {
	db    *gorm.DB
	cache *cache.Cache
	vsvc  *VendorService
}

func NewDistributionService(db *gorm.DB, c *cache.Cache, v *VendorService) *DistributionService {
	return &DistributionService{db: db, cache: c, vsvc: v}
}

// GenerateForModel 为某模型生成（或获取已有）推广链接
func (s *DistributionService) GenerateForModel(modelID uint) (*models.DistributionLink, error) {
	var existing models.DistributionLink
	err := s.db.Where("model_id = ?", modelID).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	mid := modelID
	link := &models.DistributionLink{
		ModelID:  &mid,
		LinkCode: utils.RandomCode(8),
		IsActive: true,
	}
	if err := s.db.Create(link).Error; err != nil {
		return nil, err
	}
	logger.L().Infof("Generated distribution link (model): model_id=%d code=%s", modelID, link.LinkCode)
	return link, nil
}

// GenerateForPlan 为某套餐生成（或获取已有）推广链接
func (s *DistributionService) GenerateForPlan(planID uint) (*models.DistributionLink, error) {
	var existing models.DistributionLink
	err := s.db.Where("plan_id = ?", planID).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	pid := planID
	link := &models.DistributionLink{
		PlanID:   &pid,
		LinkCode: utils.RandomCode(8),
		IsActive: true,
	}
	if err := s.db.Create(link).Error; err != nil {
		return nil, err
	}
	logger.L().Infof("Generated distribution link (plan): plan_id=%d code=%s", planID, link.LinkCode)
	return link, nil
}

// Resolve 解析推广链接 -> 目标 URL
//
// 优先级：
//  1. link.custom_url 存在则直接用（管理员覆盖）
//  2. 套餐链接：plan.custom_link_url -> vendor.official_url
//  3. 模型链接：vendor.promo_source_type 对应策略
func (s *DistributionService) Resolve(link *models.DistributionLink) (string, error) {
	if link.CustomURL != "" {
		return strings.ReplaceAll(link.CustomURL, "{code}", link.LinkCode), nil
	}

	// Plan 链接
	if link.PlanID != nil && *link.PlanID > 0 {
		var p models.Plan
		if err := s.db.Preload("Vendor").First(&p, *link.PlanID).Error; err != nil {
			return "", err
		}
		if p.CustomLinkURL != "" {
			return strings.ReplaceAll(p.CustomLinkURL, "{code}", link.LinkCode), nil
		}
		if p.Vendor != nil {
			// 走 vendor 推广策略
			if _, err := s.vsvc.LoadAuthConfig(p.Vendor); err != nil {
				logger.L().Warnf("Resolve plan: load auth_config failed: %v", err)
			}
			strategy, err := GetPromoStrategy(p.Vendor.PromoSourceType)
			if err == nil {
				cfg := PromoConfigMap(p.Vendor)
				if u, err := strategy.Fetch(cfg, p.Vendor, link.LinkCode); err == nil {
					return u, nil
				} else {
					logger.L().Warnf("Promo strategy failed for plan %d: %v", p.ID, err)
				}
			}
			if p.Vendor.OfficialURL != "" {
				return p.Vendor.OfficialURL, nil
			}
		}
		return "", fmt.Errorf("plan %d has no resolvable URL", p.ID)
	}

	// Model 链接（兼容旧）
	if link.ModelID != nil && *link.ModelID > 0 {
		var m models.Model
		if err := s.db.Preload("Vendor").First(&m, *link.ModelID).Error; err != nil {
			return "", err
		}
		if m.Vendor == nil {
			return "", fmt.Errorf("vendor not found for model %d", *link.ModelID)
		}
		if _, err := s.vsvc.LoadAuthConfig(m.Vendor); err != nil {
			logger.L().Warnf("Resolve model: load auth_config failed: %v", err)
		}
		strategy, err := GetPromoStrategy(m.Vendor.PromoSourceType)
		if err != nil {
			return "", err
		}
		cfg := PromoConfigMap(m.Vendor)
		url, err := strategy.Fetch(cfg, m.Vendor, link.LinkCode)
		if err != nil {
			logger.L().Warnf("Promo strategy failed (%v), fallback to official_url", err)
			if m.Vendor.OfficialURL != "" {
				return m.Vendor.OfficialURL, nil
			}
			return "", err
		}
		return url, nil
	}

	return "", fmt.Errorf("link has neither plan nor model")
}

// RecordClick 累加点击 + 写日志
func (s *DistributionService) RecordClick(linkID uint, referrer, ua, ip string) error {
	if err := s.db.Model(&models.DistributionLink{}).
		Where("id = ?", linkID).
		UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error; err != nil {
		return err
	}
	log := models.ClickLog{LinkID: linkID, Referrer: referrer, UserAgent: ua, IP: ip}
	return s.db.Create(&log).Error
}
