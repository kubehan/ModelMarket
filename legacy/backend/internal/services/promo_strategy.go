package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"modelmarket/internal/models"
	"modelmarket/pkg/logger"
)

// PromoStrategy 推广链接获取策略
type PromoStrategy interface {
	// Fetch 根据配置取得最终的推广 URL
	Fetch(cfg map[string]any, vendor *models.Vendor, linkCode string) (string, error)
	Describe() PromoSchema
}

type PromoSchema struct {
	Type   string            `json:"type"`
	Label  string            `json:"label"`
	Fields []AuthFieldSchema `json:"fields"` // 复用 AuthFieldSchema 结构
}

// ----- manual：管理员手填 URL 模板 -----

type manualPromo struct{}

func (manualPromo) Fetch(cfg map[string]any, _ *models.Vendor, linkCode string) (string, error) {
	tpl, _ := cfg["template"].(string)
	if tpl == "" {
		return "", fmt.Errorf("template is required")
	}
	// 替换 {code} 占位符
	return strings.ReplaceAll(tpl, "{code}", linkCode), nil
}
func (manualPromo) Describe() PromoSchema {
	return PromoSchema{
		Type: models.PromoSourceManual, Label: "手动填写 URL 模板",
		Fields: []AuthFieldSchema{
			{Key: "template", Label: "URL 模板", Type: "text", Required: true,
				Hint: "占位符 {code} 会被替换为 link_code，例如 https://x.com/?ref={code}"},
		},
	}
}

// ----- api：调用厂商接口获取 -----

type apiPromo struct{}

func (apiPromo) Fetch(cfg map[string]any, vendor *models.Vendor, _ string) (string, error) {
	endpoint, _ := cfg["endpoint"].(string)
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	method := getStr(cfg, "method", "GET")
	jsonPath := getStr(cfg, "json_path", "link")

	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", err
	}

	// 复用 vendor 认证适配器
	if vendor != nil && vendor.AuthType != "" {
		adapter, err := GetAuthAdapter(vendor.AuthType)
		if err == nil && vendor.AuthConfig != nil {
			if err := adapter.Apply(req, vendor.AuthConfig); err != nil {
				logger.L().Warnf("PromoAPI auth apply failed: %v", err)
			}
		}
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("promo api returned %d: %s", resp.StatusCode, string(body))
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("decode json failed: %w", err)
	}
	val := extractJSONPath(data, jsonPath)
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("json_path %q did not yield string", jsonPath)
}
func (apiPromo) Describe() PromoSchema {
	return PromoSchema{
		Type: models.PromoSourceAPI, Label: "调用厂商 API 拉取",
		Fields: []AuthFieldSchema{
			{Key: "endpoint", Label: "API URL", Type: "text", Required: true},
			{Key: "method", Label: "HTTP 方法", Type: "text", Hint: "默认 GET"},
			{Key: "json_path", Label: "结果 JSON 路径", Type: "text", Hint: "如 data.affiliate.url"},
		},
	}
}

// ----- scrape：抓取网页正则提取 -----

type scrapePromo struct{}

func (scrapePromo) Fetch(cfg map[string]any, vendor *models.Vendor, _ string) (string, error) {
	pageURL, _ := cfg["page_url"].(string)
	pattern, _ := cfg["regex"].(string)
	if pageURL == "" || pattern == "" {
		return "", fmt.Errorf("page_url and regex required")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", err
	}
	if vendor != nil && vendor.AuthType != "" {
		adapter, err := GetAuthAdapter(vendor.AuthType)
		if err == nil && vendor.AuthConfig != nil {
			_ = adapter.Apply(req, vendor.AuthConfig)
		}
	}
	req.Header.Set("User-Agent", "ModelMarket/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) == 0 {
		return "", fmt.Errorf("regex did not match the page content")
	}
	if len(matches) > 1 {
		return matches[1], nil // 第一个捕获组
	}
	return matches[0], nil
}
func (scrapePromo) Describe() PromoSchema {
	return PromoSchema{
		Type: models.PromoSourceScrape, Label: "抓取网页正则提取",
		Fields: []AuthFieldSchema{
			{Key: "page_url", Label: "抓取页面 URL", Type: "text", Required: true},
			{Key: "regex", Label: "正则表达式", Type: "text", Required: true,
				Hint: "用第一个捕获组提取链接，如 href=\"(https://[^\"]+/ref/[^\"]+)\""},
		},
	}
}

// GetPromoStrategy 根据类型取策略
func GetPromoStrategy(t string) (PromoStrategy, error) {
	switch t {
	case models.PromoSourceManual, "":
		return manualPromo{}, nil
	case models.PromoSourceAPI:
		return apiPromo{}, nil
	case models.PromoSourceScrape:
		return scrapePromo{}, nil
	default:
		return nil, fmt.Errorf("unsupported promo_source_type: %s", t)
	}
}

// AllPromoSchemas 后台表单：所有推广来源 schema
func AllPromoSchemas() []PromoSchema {
	out := []PromoSchema{}
	for _, t := range []string{
		models.PromoSourceManual,
		models.PromoSourceAPI,
		models.PromoSourceScrape,
	} {
		s, _ := GetPromoStrategy(t)
		out = append(out, s.Describe())
	}
	return out
}

// extractJSONPath 极简 JSON path：用点号分段，如 "data.affiliate.url"
func extractJSONPath(data any, path string) any {
	if path == "" {
		return data
	}
	parts := strings.Split(path, ".")
	cur := data
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[p]
	}
	return cur
}
