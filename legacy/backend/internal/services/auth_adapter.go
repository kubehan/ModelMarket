package services

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"modelmarket/internal/models"
)

// AuthAdapter 厂商认证适配器：把 vendor 的认证配置应用到 http.Request 上
type AuthAdapter interface {
	Apply(req *http.Request, config map[string]any) error
	// Describe 返回该认证方式所需的字段（用于后台动态表单 schema）
	Describe() AuthSchema
}

// AuthFieldSchema 单个字段描述
type AuthFieldSchema struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`     // text | password | textarea
	Required bool   `json:"required"`
	Hint     string `json:"hint,omitempty"`
}

// AuthSchema 一种认证方式的字段列表
type AuthSchema struct {
	Type   string            `json:"type"`
	Label  string            `json:"label"`
	Fields []AuthFieldSchema `json:"fields"`
}

// ----- 各适配器 -----

type apiKeyAdapter struct{}

func (apiKeyAdapter) Apply(req *http.Request, c map[string]any) error {
	key, _ := c["api_key"].(string)
	if key == "" {
		return fmt.Errorf("api_key is required")
	}
	header := getStr(c, "header", "Authorization")
	prefix := getStr(c, "prefix", "Bearer ")
	req.Header.Set(header, prefix+key)
	return nil
}
func (apiKeyAdapter) Describe() AuthSchema {
	return AuthSchema{
		Type: models.AuthTypeAPIKey, Label: "API Key / Bearer Token",
		Fields: []AuthFieldSchema{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true},
			{Key: "header", Label: "Header 名称", Type: "text", Hint: "默认 Authorization"},
			{Key: "prefix", Label: "Header 前缀", Type: "text", Hint: "默认 'Bearer '"},
		},
	}
}

type oauth2Adapter struct{}

func (oauth2Adapter) Apply(req *http.Request, c map[string]any) error {
	tok, _ := c["access_token"].(string)
	if tok == "" {
		return fmt.Errorf("access_token is required (refresh-flow not yet implemented)")
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	return nil
}
func (oauth2Adapter) Describe() AuthSchema {
	return AuthSchema{
		Type: models.AuthTypeOAuth2, Label: "OAuth2",
		Fields: []AuthFieldSchema{
			{Key: "client_id", Label: "Client ID", Type: "text"},
			{Key: "client_secret", Label: "Client Secret", Type: "password"},
			{Key: "token_url", Label: "Token URL", Type: "text"},
			{Key: "access_token", Label: "Access Token", Type: "password", Required: true},
			{Key: "refresh_token", Label: "Refresh Token", Type: "password"},
		},
	}
}

type cookieAdapter struct{}

func (cookieAdapter) Apply(req *http.Request, c map[string]any) error {
	cookies, _ := c["cookies"].(string)
	if cookies == "" {
		return fmt.Errorf("cookies is required")
	}
	req.Header.Set("Cookie", cookies)
	return nil
}
func (cookieAdapter) Describe() AuthSchema {
	return AuthSchema{
		Type: models.AuthTypeCookie, Label: "Cookie 登录态",
		Fields: []AuthFieldSchema{
			{Key: "login_url", Label: "登录页 URL", Type: "text"},
			{Key: "cookies", Label: "Cookie 字符串", Type: "textarea", Required: true, Hint: "形如 a=1; b=2"},
		},
	}
}

type basicAdapter struct{}

func (basicAdapter) Apply(req *http.Request, c map[string]any) error {
	u, _ := c["username"].(string)
	p, _ := c["password"].(string)
	if u == "" || p == "" {
		return fmt.Errorf("username/password required")
	}
	cred := base64.StdEncoding.EncodeToString([]byte(u + ":" + p))
	req.Header.Set("Authorization", "Basic "+cred)
	return nil
}
func (basicAdapter) Describe() AuthSchema {
	return AuthSchema{
		Type: models.AuthTypeBasic, Label: "Basic Auth",
		Fields: []AuthFieldSchema{
			{Key: "username", Label: "用户名", Type: "text", Required: true},
			{Key: "password", Label: "密码", Type: "password", Required: true},
		},
	}
}

type customHeaderAdapter struct{}

func (customHeaderAdapter) Apply(req *http.Request, c map[string]any) error {
	name, _ := c["header_name"].(string)
	val, _ := c["header_value"].(string)
	if name == "" || val == "" {
		return fmt.Errorf("header_name/header_value required")
	}
	req.Header.Set(name, val)
	return nil
}
func (customHeaderAdapter) Describe() AuthSchema {
	return AuthSchema{
		Type: models.AuthTypeCustomHeader, Label: "自定义 Header",
		Fields: []AuthFieldSchema{
			{Key: "header_name", Label: "Header 名", Type: "text", Required: true},
			{Key: "header_value", Label: "Header 值", Type: "password", Required: true},
		},
	}
}

// GetAuthAdapter 根据 auth_type 取适配器
func GetAuthAdapter(authType string) (AuthAdapter, error) {
	switch authType {
	case models.AuthTypeAPIKey:
		return apiKeyAdapter{}, nil
	case models.AuthTypeOAuth2:
		return oauth2Adapter{}, nil
	case models.AuthTypeCookie:
		return cookieAdapter{}, nil
	case models.AuthTypeBasic:
		return basicAdapter{}, nil
	case models.AuthTypeCustomHeader:
		return customHeaderAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported auth_type: %s", authType)
	}
}

// AllAuthSchemas 后台表单用：列出所有认证方式 schema
func AllAuthSchemas() []AuthSchema {
	out := []AuthSchema{}
	for _, t := range []string{
		models.AuthTypeAPIKey,
		models.AuthTypeOAuth2,
		models.AuthTypeCookie,
		models.AuthTypeBasic,
		models.AuthTypeCustomHeader,
	} {
		a, _ := GetAuthAdapter(t)
		out = append(out, a.Describe())
	}
	return out
}

func getStr(m map[string]any, k, def string) string {
	if v, ok := m[k].(string); ok && v != "" {
		return v
	}
	return def
}
