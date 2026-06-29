package cache

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"modelmarket/internal/models"
	"modelmarket/pkg/logger"
)

// Cache 数据库支持的 TTL 缓存
// 用法（"装饰器"模式）：
//
//	var result []ModelDTO
//	err := cache.GetOrSet(db, "models:list", 3600, &result, func() (any, error) {
//	    return service.ListModelsFromDB(), nil
//	})
type Cache struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Cache {
	return &Cache{db: db}
}

// Get 读取缓存。命中返回 true。
func (c *Cache) Get(key string, out any) (bool, error) {
	var entry models.CacheEntry
	if err := c.db.Where("key = ?", key).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	if time.Now().After(entry.ExpiresAt) {
		logger.L().Debugf("Cache expired: key=%s", key)
		_ = c.db.Delete(&entry).Error
		return false, nil
	}
	if err := json.Unmarshal([]byte(entry.Value), out); err != nil {
		logger.L().Warnf("Cache unmarshal failed key=%s: %v", key, err)
		return false, err
	}
	logger.L().Debugf("Cache hit: key=%s", key)
	return true, nil
}

// Set 写入缓存
func (c *Cache) Set(key string, value any, ttlSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	entry := models.CacheEntry{
		Key:       key,
		Value:     string(data),
		ExpiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}
	// upsert
	if err := c.db.Save(&entry).Error; err != nil {
		return err
	}
	logger.L().Debugf("Cache set: key=%s ttl=%ds", key, ttlSeconds)
	return nil
}

// Invalidate 主动失效一个 key
func (c *Cache) Invalidate(key string) error {
	logger.L().Infof("Cache invalidate: key=%s", key)
	return c.db.Where("key = ?", key).Delete(&models.CacheEntry{}).Error
}

// InvalidatePrefix 失效所有指定前缀 key（用于刷新一类数据）
func (c *Cache) InvalidatePrefix(prefix string) error {
	logger.L().Infof("Cache invalidate prefix: %s", prefix)
	return c.db.Where("key LIKE ?", prefix+"%").Delete(&models.CacheEntry{}).Error
}

// GetOrSet 装饰器风格：命中即用，否则调用 loader 并写缓存
func (c *Cache) GetOrSet(key string, ttlSeconds int, out any, loader func() (any, error)) error {
	if ok, _ := c.Get(key, out); ok {
		return nil
	}
	logger.L().Debugf("Cache miss, loading: key=%s", key)
	v, err := loader()
	if err != nil {
		return err
	}
	if err := c.Set(key, v, ttlSeconds); err != nil {
		logger.L().Warnf("Cache set failed key=%s: %v", key, err)
	}
	// 将 v 重新填回 out
	raw, _ := json.Marshal(v)
	return json.Unmarshal(raw, out)
}
