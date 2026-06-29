package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"modelmarket/internal/cache"
	"modelmarket/internal/config"
	"modelmarket/internal/database"
	"modelmarket/internal/handlers"
	"modelmarket/internal/middleware"
	"modelmarket/internal/models"
	"modelmarket/internal/services"
	"modelmarket/internal/utils"
	"modelmarket/pkg/logger"
)

func main() {
	cfg := config.Load()
	logger.SetLevel(cfg.LogLevel)
	logger.L().Info("=== ModelMarket starting ===")

	// 准备加密 key（首次启动若未配置则生成一个并提示）
	if cfg.EncryptionKey == "" {
		k := utils.GenerateKey()
		logger.L().Warnf("ENCRYPTION_KEY not set in env. Auto-generated key for THIS RUN ONLY: %s", k)
		logger.L().Warn("!!! Persist this key in .env to keep encrypted data readable on next restart !!!")
		cfg.EncryptionKey = k
	}
	crypto, err := utils.NewCrypto(cfg.EncryptionKey)
	if err != nil {
		logger.L().Fatalf("Init crypto failed: %v", err)
	}

	// 数据库
	db, err := database.Init(cfg)
	if err != nil {
		logger.L().Fatalf("Init database failed: %v", err)
	}

	// 引导管理员 & 演示数据
	if err := bootstrapAdmin(db, cfg); err != nil {
		logger.L().Errorf("Bootstrap admin failed: %v", err)
	}

	// 缓存
	cc := cache.New(db)

	// Services
	vendorSvc := services.NewVendorService(db, cc, crypto)
	distSvc := services.NewDistributionService(db, cc, vendorSvc)

	// 演示数据（仅当库为空时）
	seedDemoData(db, vendorSvc, distSvc)

	// Handlers
	authH := handlers.NewAuthHandler(db)
	vendorH := handlers.NewVendorHandler(db, vendorSvc)
	modelH := handlers.NewModelHandler(db, cc)
	distH := handlers.NewDistributionHandler(db, distSvc)

	// Gin
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// 公开跳转
	r.GET("/r/:code", distH.Resolve)

	api := r.Group("/api/v1")
	{
		// 公开
		api.POST("/auth/login", authH.Login)
		api.GET("/models/", modelH.PublicList)

		// 后台
		admin := api.Group("/admin", middleware.AuthRequired())
		admin.GET("/auth/me", authH.Me)

		admin.GET("/vendors/schemas", vendorH.Schemas)
		admin.GET("/vendors", vendorH.List)
		admin.POST("/vendors", vendorH.Create)
		admin.GET("/vendors/:id", vendorH.Get)
		admin.PUT("/vendors/:id", vendorH.Update)
		admin.DELETE("/vendors/:id", vendorH.Delete)
		admin.POST("/vendors/:id/test", vendorH.TestConnection)
		admin.POST("/vendors/:id/sync", vendorH.SyncModels)

		admin.GET("/models", modelH.AdminList)
		admin.PUT("/models/:id", modelH.AdminUpdate)
		admin.DELETE("/models/:id", modelH.AdminDelete)
		admin.POST("/models/refresh", modelH.Refresh)

		admin.GET("/links", distH.AdminList)
		admin.POST("/links", distH.Generate)
		admin.PUT("/links/:id", distH.AdminUpdate)
		admin.DELETE("/links/:id", distH.AdminDelete)
	}

	addr := ":" + cfg.ServerPort
	logger.L().Infof("HTTP server listening on %s (vendor mode=%s)", addr, cfg.VendorAPIMode)
	if err := r.Run(addr); err != nil {
		logger.L().Fatalf("Server exited: %v", err)
	}
}

// bootstrapAdmin 首次启动创建默认管理员
func bootstrapAdmin(db *gorm.DB, cfg *config.Config) error {
	var count int64
	db.Model(&models.AdminUser{}).Count(&count)
	if count > 0 {
		logger.L().Debug("Admin user already exists, skip bootstrap")
		return nil
	}
	hash, err := utils.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}
	u := models.AdminUser{Username: cfg.AdminUsername, PasswordHash: hash}
	if err := db.Create(&u).Error; err != nil {
		return err
	}
	logger.L().Warnf("Bootstrap admin created: username=%s password=%s (please change in production)",
		cfg.AdminUsername, cfg.AdminPassword)
	return nil
}

// seedDemoData 仅当 vendor 表为空时插入演示厂商 + 同步 mock 模型
func seedDemoData(db *gorm.DB, vsvc *services.VendorService, dsvc *services.DistributionService) {
	var count int64
	db.Model(&models.Vendor{}).Count(&count)
	if count > 0 {
		return
	}
	logger.L().Info("Seeding demo vendors (mock data)...")
	demos := []struct {
		v          *models.Vendor
		authConfig map[string]any
		promoCfg   map[string]any
	}{
		{
			v: &models.Vendor{
				Name: "OpenAI", OfficialURL: "https://openai.com", APIBase: "https://api.openai.com",
				LogoURL:         "https://openai.com/favicon.ico",
				Description:     "ChatGPT 系列模型提供商",
				AuthType:        models.AuthTypeAPIKey,
				PromoSourceType: models.PromoSourceManual,
				IsActive:        true,
			},
			authConfig: map[string]any{"api_key": "sk-demo-openai", "header": "Authorization", "prefix": "Bearer "},
			promoCfg:   map[string]any{"template": "https://platform.openai.com/signup?ref={code}"},
		},
		{
			v: &models.Vendor{
				Name: "Anthropic", OfficialURL: "https://anthropic.com", APIBase: "https://api.anthropic.com",
				LogoURL: "https://www.anthropic.com/favicon.ico", Description: "Claude 系列模型",
				AuthType: models.AuthTypeCustomHeader, PromoSourceType: models.PromoSourceManual,
				IsActive: true,
			},
			authConfig: map[string]any{"header_name": "x-api-key", "header_value": "demo-anthropic"},
			promoCfg:   map[string]any{"template": "https://console.anthropic.com/?ref={code}"},
		},
		{
			v: &models.Vendor{
				Name: "Baidu Qianfan", OfficialURL: "https://cloud.baidu.com/product/wenxinworkshop",
				APIBase: "https://qianfan.baidubce.com", LogoURL: "",
				Description:     "百度文心系列",
				AuthType:        models.AuthTypeAPIKey,
				PromoSourceType: models.PromoSourceManual, IsActive: true,
			},
			authConfig: map[string]any{"api_key": "demo-baidu-key"},
			promoCfg:   map[string]any{"template": "https://cloud.baidu.com/?invite={code}"},
		},
		{
			v: &models.Vendor{
				Name: "Aliyun DashScope", OfficialURL: "https://dashscope.aliyun.com",
				APIBase: "https://dashscope.aliyuncs.com", LogoURL: "",
				Description:     "通义千问系列",
				AuthType:        models.AuthTypeAPIKey, PromoSourceType: models.PromoSourceManual,
				IsActive:        true,
			},
			authConfig: map[string]any{"api_key": "demo-aliyun-key"},
			promoCfg:   map[string]any{"template": "https://dashscope.aliyun.com/?ref={code}"},
		},
	}
	for _, d := range demos {
		if err := vsvc.CreateOrUpdate(d.v, d.authConfig, d.promoCfg); err != nil {
			logger.L().Warnf("Seed vendor %s failed: %v", d.v.Name, err)
			continue
		}
		// 加载 auth_config 以便 TestConnection 使用
		_, _ = vsvc.LoadAuthConfig(d.v)
		if _, err := vsvc.SyncModels(d.v); err != nil {
			logger.L().Warnf("Seed sync models for %s failed: %v", d.v.Name, err)
		}
	}
	// 为每个 model 生成默认分销链接
	var allModels []models.Model
	db.Find(&allModels)
	for _, m := range allModels {
		if _, err := dsvc.GenerateForModel(m.ID); err != nil {
			logger.L().Warnf("Seed distribution link for model %d failed: %v", m.ID, err)
		}
	}
	logger.L().Infof("Demo seed completed: %d models", len(allModels))
}
