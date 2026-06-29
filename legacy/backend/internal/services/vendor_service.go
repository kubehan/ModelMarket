package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"modelmarket/internal/cache"
	"modelmarket/internal/config"
	"modelmarket/internal/models"
	"modelmarket/internal/utils"
	"modelmarket/pkg/logger"
)

// VendorService 负责厂商相关业务逻辑
type VendorService struct {
	db     *gorm.DB
	cache  *cache.Cache
	crypto *utils.Crypto
}

func NewVendorService(db *gorm.DB, c *cache.Cache, cr *utils.Crypto) *VendorService {
	return &VendorService{db: db, cache: c, crypto: cr}
}

// CreateOrUpdate 新建/更新厂商；处理 auth_config 加密、promo_config 序列化
func (s *VendorService) CreateOrUpdate(in *models.Vendor, authConfig map[string]any, promoConfig map[string]any) error {
	logger.L().Infof("Saving vendor: name=%s auth=%s promo=%s", in.Name, in.AuthType, in.PromoSourceType)

	// 加密 auth_config
	if authConfig != nil {
		raw, _ := json.Marshal(authConfig)
		enc, err := s.crypto.Encrypt(string(raw))
		if err != nil {
			return fmt.Errorf("encrypt auth_config: %w", err)
		}
		in.AuthConfigEnc = enc
	}

	// promo_config 不加密（不含敏感凭证）
	if promoConfig != nil {
		raw, _ := json.Marshal(promoConfig)
		in.PromoConfig = string(raw)
	}

	if in.ID == 0 {
		return s.db.Create(in).Error
	}
	return s.db.Save(in).Error
}

// LoadAuthConfig 解密 auth_config 到 map
func (s *VendorService) LoadAuthConfig(v *models.Vendor) (map[string]any, error) {
	if v.AuthConfigEnc == "" {
		return map[string]any{}, nil
	}
	plain, err := s.crypto.Decrypt(v.AuthConfigEnc)
	if err != nil {
		return nil, fmt.Errorf("decrypt auth_config: %w", err)
	}
	m := map[string]any{}
	if err := json.Unmarshal([]byte(plain), &m); err != nil {
		return nil, err
	}
	v.AuthConfig = m
	return m, nil
}

// MaskAuthConfig 把敏感字段打码，用于返回给前端
func MaskAuthConfig(c map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range c {
		s, ok := v.(string)
		if !ok {
			out[k] = v
			continue
		}
		lower := strings.ToLower(k)
		if strings.Contains(lower, "key") || strings.Contains(lower, "secret") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "token") ||
			strings.Contains(lower, "cookie") || lower == "header_value" {
			if len(s) > 6 {
				out[k] = s[:3] + "****" + s[len(s)-3:]
			} else if s != "" {
				out[k] = "****"
			} else {
				out[k] = ""
			}
		} else {
			out[k] = v
		}
	}
	return out
}

// PromoConfigMap 把 vendor.PromoConfig (JSON 字符串) 解到 map
func PromoConfigMap(v *models.Vendor) map[string]any {
	m := map[string]any{}
	if v.PromoConfig == "" {
		return m
	}
	_ = json.Unmarshal([]byte(v.PromoConfig), &m)
	return m
}

// VendorModelInfo /v1/models 拉取结果
type VendorModelInfo struct {
	ID            string  `json:"id"`
	ContextLength int     `json:"context_length,omitempty"`
	InputPrice    float64 `json:"input_price,omitempty"`
	OutputPrice   float64 `json:"output_price,omitempty"`
}

// TestConnection 测试连接；mock 模式直接返回预置数据，real 模式真实请求 /v1/models
func (s *VendorService) TestConnection(v *models.Vendor) ([]VendorModelInfo, error) {
	mode := config.Global.VendorAPIMode
	logger.L().Infof("TestConnection vendor=%s mode=%s", v.Name, mode)

	if mode == "mock" {
		return s.mockModels(v), nil
	}

	// 真实模式
	if v.APIBase == "" {
		return nil, fmt.Errorf("api_base is empty")
	}
	url := strings.TrimRight(v.APIBase, "/") + "/v1/models"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	authCfg, err := s.LoadAuthConfig(v)
	if err != nil {
		return nil, err
	}
	adapter, err := GetAuthAdapter(v.AuthType)
	if err != nil {
		return nil, err
	}
	if err := adapter.Apply(req, authCfg); err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.L().Errorf("TestConnection http error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("vendor returned %d: %s", resp.StatusCode, string(body))
	}

	// OpenAI 兼容响应 { data: [{id: "..."}] }
	var parsed struct {
		Data []struct {
			ID            string  `json:"id"`
			ContextLength int     `json:"context_length"`
			InputPrice    float64 `json:"input_price"`
			OutputPrice   float64 `json:"output_price"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	out := make([]VendorModelInfo, 0, len(parsed.Data))
	for _, d := range parsed.Data {
		out = append(out, VendorModelInfo{
			ID: d.ID, ContextLength: d.ContextLength,
			InputPrice: d.InputPrice, OutputPrice: d.OutputPrice,
		})
	}
	return out, nil
}

// mockModels 提供预置的模拟模型数据
func (s *VendorService) mockModels(v *models.Vendor) []VendorModelInfo {
	preset := map[string][]VendorModelInfo{
		"openai": {
			{ID: "gpt-4o", ContextLength: 128000, InputPrice: 5.00, OutputPrice: 15.00},
			{ID: "gpt-4o-mini", ContextLength: 128000, InputPrice: 0.15, OutputPrice: 0.60},
			{ID: "gpt-4-turbo", ContextLength: 128000, InputPrice: 10.00, OutputPrice: 30.00},
		},
		"anthropic": {
			{ID: "claude-3-5-sonnet", ContextLength: 200000, InputPrice: 3.00, OutputPrice: 15.00},
			{ID: "claude-3-opus", ContextLength: 200000, InputPrice: 15.00, OutputPrice: 75.00},
			{ID: "claude-3-haiku", ContextLength: 200000, InputPrice: 0.25, OutputPrice: 1.25},
		},
		"baidu": {
			{ID: "ernie-4.0", ContextLength: 8192, InputPrice: 16.80, OutputPrice: 16.80},
			{ID: "ernie-3.5", ContextLength: 8192, InputPrice: 1.68, OutputPrice: 1.68},
		},
		"aliyun": {
			{ID: "qwen-max", ContextLength: 8192, InputPrice: 2.80, OutputPrice: 8.40},
			{ID: "qwen-plus", ContextLength: 32768, InputPrice: 0.56, OutputPrice: 1.68},
			{ID: "qwen-turbo", ContextLength: 8192, InputPrice: 0.28, OutputPrice: 0.56},
		},
	}
	key := strings.ToLower(v.Name)
	for k, list := range preset {
		if strings.Contains(key, k) {
			return list
		}
	}
	// 默认通用
	return []VendorModelInfo{
		{ID: v.Name + "-default", ContextLength: 8192, InputPrice: 1.00, OutputPrice: 3.00},
	}
}

// SyncModels 把 /v1/models 结果同步到本地 Model 表
func (s *VendorService) SyncModels(v *models.Vendor) (int, error) {
	infos, err := s.TestConnection(v)
	if err != nil {
		now := time.Now()
		v.LastTestedAt = &now
		v.LastTestStatus = "failed"
		v.LastTestMessage = err.Error()
		s.db.Save(v)
		return 0, err
	}

	count := 0
	for _, info := range infos {
		var m models.Model
		err := s.db.Where("vendor_id = ? AND model_name = ?", v.ID, info.ID).First(&m).Error
		if err == gorm.ErrRecordNotFound {
			m = models.Model{
				VendorID:      v.ID,
				ModelName:     info.ID,
				DisplayName:   info.ID,
				ContextLength: info.ContextLength,
				InputPrice:    info.InputPrice,
				OutputPrice:   info.OutputPrice,
				IsActive:      true,
			}
			if err := s.db.Create(&m).Error; err != nil {
				logger.L().Warnf("Create model failed: %v", err)
				continue
			}
			count++
		} else if err == nil {
			// 更新 context_length / 价格（如果厂商返回了非零值）
			updates := map[string]any{}
			if info.ContextLength > 0 {
				updates["context_length"] = info.ContextLength
			}
			if info.InputPrice > 0 {
				updates["input_price"] = info.InputPrice
			}
			if info.OutputPrice > 0 {
				updates["output_price"] = info.OutputPrice
			}
			if len(updates) > 0 {
				s.db.Model(&m).Updates(updates)
			}
		}
	}

	now := time.Now()
	v.LastTestedAt = &now
	v.LastTestStatus = "ok"
	v.LastTestMessage = fmt.Sprintf("synced %d models", len(infos))
	s.db.Save(v)

	// 失效首页价格缓存
	_ = s.cache.InvalidatePrefix("models:")
	return count, nil
}
