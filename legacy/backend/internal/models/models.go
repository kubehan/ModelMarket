package models

import (
	"time"

	"gorm.io/gorm"
)

// AuthType 厂商认证方式
const (
	AuthTypeAPIKey       = "api_key"
	AuthTypeOAuth2       = "oauth2"
	AuthTypeCookie       = "cookie"
	AuthTypeBasic        = "basic"
	AuthTypeCustomHeader = "custom_header"
)

// PromoSourceType 推广链接获取方式
const (
	PromoSourceManual = "manual"
	PromoSourceAPI    = "api"
	PromoSourceScrape = "scrape"
)

// PlanType 套餐类型
const (
	PlanTypeCoding = "coding" // CodingPlan
	PlanTypeAgent  = "agent"  // AgentPlan
	PlanTypeToken  = "token"  // TokenPlan
)

// PlanStatus 套餐状态
const (
	PlanStatusActive   = "active"
	PlanStatusLimited  = "limited"  // 限购
	PlanStatusSoldOut  = "sold_out" // 售罄
	PlanStatusOffline  = "offline"  // 已下线
)

// Vendor 大模型平台（OpenAI / 智谱 / Kimi / 字节方舟 等）
type Vendor struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:64;uniqueIndex;not null" json:"name"`
	OfficialURL     string         `gorm:"size:255" json:"official_url"`
	APIBase         string         `gorm:"size:255" json:"api_base"`
	LogoURL         string         `gorm:"size:255" json:"logo_url"`
	Description     string         `gorm:"type:text" json:"description"`

	AuthType        string         `gorm:"size:32;not null;default:api_key" json:"auth_type"`
	AuthConfigEnc   string         `gorm:"type:text" json:"-"`
	AuthConfig      map[string]any `gorm:"-" json:"auth_config,omitempty"`

	PromoSourceType string         `gorm:"size:32;default:manual" json:"promo_source_type"`
	PromoConfig     string         `gorm:"type:text" json:"promo_config"`

	IsActive        bool           `gorm:"default:true" json:"is_active"`
	LastTestedAt    *time.Time     `json:"last_tested_at,omitempty"`
	LastTestStatus  string         `gorm:"size:32" json:"last_test_status,omitempty"`
	LastTestMessage string         `gorm:"type:text" json:"last_test_message,omitempty"`

	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Models          []Model        `gorm:"foreignKey:VendorID" json:"models,omitempty"`
	Plans           []Plan         `gorm:"foreignKey:VendorID" json:"plans,omitempty"`
}

// Model 单个大模型（保留：用于 token 单价对比 & Plan 关联）
type Model struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	VendorID      uint           `gorm:"index;not null" json:"vendor_id"`
	ModelName     string         `gorm:"size:128;not null" json:"model_name"`
	DisplayName   string         `gorm:"size:128" json:"display_name"`
	ContextLength int            `json:"context_length"`
	InputPrice    float64        `json:"input_price"`
	OutputPrice   float64        `json:"output_price"`
	ELOScore      *int           `json:"elo_score,omitempty"`
	LatencyMS     *int           `json:"latency_ms,omitempty"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`

	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Vendor        *Vendor        `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
}

// Plan AI 编码套餐（CodingPlan / AgentPlan / TokenPlan）—— 核心比价实体
type Plan struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	VendorID   uint   `gorm:"index;not null" json:"vendor_id"`
	Name       string `gorm:"size:128;not null" json:"name"`        // 套餐名，例：CodingPlan Pro
	PlanType   string `gorm:"size:32;index;not null" json:"plan_type"` // coding / agent / token

	// 价格（人民币）。0 表示不提供该周期
	FirstMonthPrice float64 `json:"first_month_price"` // 首月优惠价
	MonthlyPrice    float64 `json:"monthly_price"`     // 连续包月
	QuarterlyPrice  float64 `json:"quarterly_price"`   // 连续包季
	YearlyPrice     float64 `json:"yearly_price"`      // 连续包年

	// 用量限制（同时支持次数与 token）
	RequestsPer5h    int     `json:"requests_per_5h"`
	RequestsPerWeek  int     `json:"requests_per_week"`
	RequestsPerMonth int     `json:"requests_per_month"`
	TokenLimitM      float64 `json:"token_limit_m"`     // 月 token 上限（单位：百万）
	MeasuredTokenM   float64 `json:"measured_token_m"`  // 实测月 token（单位：百万）

	// 评级 & 标签
	Rating       int    `json:"rating"`                            // 1-5 星
	Highlights   string `gorm:"type:text" json:"highlights"`       // 多行亮点（用换行分隔，UI 渲染时拆开）
	Tags         string `gorm:"size:255" json:"tags"`              // 逗号分隔：量大管饱,不用抢,有GLM-5.2
	OtherRights  string `gorm:"type:text" json:"other_rights"`     // 其他权益描述
	Note         string `gorm:"type:text" json:"note"`             // 备注

	Status       string `gorm:"size:32;default:active" json:"status"` // active/limited/sold_out/offline

	// 跳转
	CustomLinkURL string `gorm:"size:512" json:"custom_link_url"` // 管理员手填的跳转 URL（与 DistributionLink 配合）

	IsActive  bool           `gorm:"default:true" json:"is_active"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Vendor *Vendor `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`

	// 套餐支持的模型（多对多）
	Models []Model `gorm:"many2many:plan_models;" json:"models,omitempty"`
}

// DistributionLink 分销推广链接（既可绑模型也可绑套餐）
type DistributionLink struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ModelID    *uint          `gorm:"index" json:"model_id,omitempty"`
	PlanID     *uint          `gorm:"index" json:"plan_id,omitempty"`
	LinkCode   string         `gorm:"size:32;uniqueIndex;not null" json:"link_code"`
	CustomURL  string         `gorm:"size:512" json:"custom_url"`
	Clicks     int64          `gorm:"default:0" json:"clicks"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`

	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Model      *Model         `gorm:"foreignKey:ModelID" json:"model,omitempty"`
	Plan       *Plan          `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

// ClickLog 点击日志
type ClickLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	LinkID     uint      `gorm:"index;not null" json:"link_id"`
	Referrer   string    `gorm:"size:255" json:"referrer"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	IP         string    `gorm:"size:64" json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

// AdminUser 管理员
type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CacheEntry 数据库级缓存
type CacheEntry struct {
	Key       string    `gorm:"primaryKey;size:255" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
