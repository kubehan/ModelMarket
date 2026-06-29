package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"modelmarket/internal/config"
	"modelmarket/internal/models"
	"modelmarket/pkg/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接并迁移
func Init(cfg *config.Config) (*gorm.DB, error) {
	logger.L().Infof("Initializing database, driver=%s", cfg.DBDriver)

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	var (
		db  *gorm.DB
		err error
	)

	switch cfg.DBDriver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DBDSN), gormCfg)
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DBDSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.DBDriver)
	}
	if err != nil {
		logger.L().Errorf("Failed to connect database: %v", err)
		return nil, err
	}

	logger.L().Info("Database connected, running auto-migration...")
	if err := db.AutoMigrate(
		&models.Vendor{},
		&models.Model{},
		&models.Plan{},
		&models.DistributionLink{},
		&models.ClickLog{},
		&models.AdminUser{},
		&models.CacheEntry{},
	); err != nil {
		logger.L().Errorf("AutoMigrate failed: %v", err)
		return nil, err
	}
	logger.L().Info("Database migration completed")

	// 启动后台清理过期缓存的协程
	go cleanupCache(db)

	DB = db
	return db, nil
}

// cleanupCache 定期清理已过期的缓存条目
func cleanupCache(db *gorm.DB) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		res := db.Where("expires_at < ?", time.Now()).Delete(&models.CacheEntry{})
		if res.Error != nil {
			logger.L().Warnf("Cache cleanup error: %v", res.Error)
			continue
		}
		if res.RowsAffected > 0 {
			logger.L().Debugf("Cleaned %d expired cache entries", res.RowsAffected)
		}
	}
}
