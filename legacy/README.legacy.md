# ModelMarket — 大模型聚合分销与比价平台

一个仿 [codingplan.fyi](https://www.codingplan.fyi/) 的全栈原型：

- **首页**：所有已接入大模型的厂商 / 模型 / 输入价 / 输出价 / 上下文 / ELO / 延迟 一表对比
- **后台**：管理员可动态配置每个厂商的**登录认证方式**与**推广链接获取方式**
- **分销**：每个模型自动生成唯一推广短链 `/r/<code>`，支持点击统计与 referrer 跟踪
- **缓存**：首页数据库级 TTL 缓存，后台一键失效
- **安全**：API Key 等敏感凭证用 AES-256-GCM（与 Fernet 等价）加密落盘

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.22 + Gin + GORM (SQLite / PostgreSQL) |
| 认证 | JWT (HS256) |
| 加密 | AES-256-GCM，密钥来自 `ENCRYPTION_KEY`（base64 或任意字符串经 SHA-256 派生）|
| 前端 | Vue 3 + Vite + Element Plus + Pinia + Vue Router |
| 日志 | logrus 多级别（trace/debug/info/warn/error）|

## 目录结构

```
ModelMarket/
├── backend/
│   ├── cmd/server/main.go          # 启动入口，自动建表、引导管理员、注入演示数据
│   ├── internal/
│   │   ├── config/                 # 读取 .env / 环境变量
│   │   ├── database/               # GORM 初始化 + 过期缓存清理协程
│   │   ├── models/                 # Vendor / Model / DistributionLink / ClickLog / AdminUser / CacheEntry
│   │   ├── cache/                  # 数据库级 TTL 缓存 (Get / Set / GetOrSet / InvalidatePrefix)
│   │   ├── utils/                  # AES-GCM 加密、bcrypt、随机 code
│   │   ├── middleware/             # JWT 校验、请求日志、CORS
│   │   ├── services/
│   │   │   ├── auth_adapter.go       # 5 种认证方式适配器 (api_key/oauth2/cookie/basic/custom_header)
│   │   │   ├── promo_strategy.go     # 3 种推广链接获取策略 (manual/api/scrape)
│   │   │   ├── vendor_service.go     # 厂商增删改 + /v1/models 调用 (mock/real)
│   │   │   └── distribution_service.go # 分销链接生成 + 解析跳转 + 点击日志
│   │   └── handlers/               # Gin Handler
│   └── pkg/logger/                 # logrus 单例
└── frontend/
    ├── src/
    │   ├── views/Home.vue              # 首页：价格对比看板
    │   ├── views/AdminLogin.vue        # 后台登录
    │   ├── views/AdminLayout.vue       # 后台框架
    │   └── views/admin/
    │       ├── Vendors.vue             # 厂商管理（动态表单）
    │       ├── Models.vue              # 模型管理
    │       └── Links.vue               # 分销链接管理
    └── ...
```

## 快速启动

### 0. 依赖

- Go 1.22+
- Node.js 18+ / npm
- （可选）PostgreSQL，否则默认 SQLite

### 1. 启动后端

```bash
cd backend
cp .env.example .env       # 按需修改配置
go mod tidy
go run ./cmd/server
```

服务监听 `http://localhost:8080`。首次启动会自动：

1. 创建 SQLite 文件 `modelmarket.db`（或连接 PostgreSQL）
2. 创建默认管理员 `admin` / `admin123`
3. 注入演示数据：OpenAI / Anthropic / Baidu Qianfan / Aliyun DashScope 四个厂商及其 mock 模型与默认推广链接

打开 `http://localhost:8080/api/v1/models/` 可直接看到 JSON 数据。

### 2. 启动前端

```bash
cd frontend
npm install
npm run dev
```

打开 `http://localhost:5173`：

- 首页 → 价格对比看板
- 右上角 *管理后台* → `admin / admin123` 登录

### 3. 生产构建

```bash
cd frontend && npm run build      # 产物在 frontend/dist
cd ../backend && go build ./cmd/server -o modelmarket
```

可用 Nginx 把 `frontend/dist` 与后端 8080 反代到同一域名。

## 配置项 (`backend/.env`)

| 变量 | 默认 | 说明 |
|---|---|---|
| `SERVER_PORT` | `8080` | 监听端口 |
| `GIN_MODE` | `debug` | `debug` / `release` |
| `DB_DRIVER` | `sqlite` | `sqlite` / `postgres` |
| `DB_DSN` | `modelmarket.db` | SQLite 文件路径 或 PG DSN |
| `JWT_SECRET` | `dev-secret-change-me` | JWT 签名 |
| `JWT_EXPIRES_HOURS` | `24` | Token 有效期 |
| `ENCRYPTION_KEY` | *(auto)* | 加密密钥；留空则每次启动随机生成（重启后旧密文无法解密！生产必须固定）|
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | `admin`/`admin123` | 首次启动引导账号 |
| `CACHE_TTL_SECONDS` | `3600` | 首页缓存 TTL |
| `VENDOR_API_MODE` | `mock` | `mock` 直接返回预置数据；`real` 实际请求各厂商 `/v1/models` |
| `LOG_LEVEL` | `info` | `trace` / `debug` / `info` / `warn` / `error` |

## 核心功能演示

### 厂商登录方式（动态可配）

后台 → 厂商管理 → 新增厂商，可选：

| auth_type | 字段 | 适用场景 |
|---|---|---|
| `api_key` | api_key / header / prefix | OpenAI、Aliyun DashScope 等 Bearer 令牌 |
| `oauth2` | client_id / secret / token_url / access_token / refresh_token | OAuth 流程 |
| `cookie` | login_url / cookies | 网页登录态（爬取场景）|
| `basic` | username / password | 老旧 Basic Auth API |
| `custom_header` | header_name / header_value | Anthropic（`x-api-key`）等非标 |

前端会根据 `GET /api/v1/admin/vendors/schemas` 返回的字段 schema **动态渲染表单**。

### 推广链接获取方式

| promo_source_type | 字段 | 作用 |
|---|---|---|
| `manual` | template，如 `https://x.com/?ref={code}` | 用 link_code 替换 `{code}` |
| `api` | endpoint / method / json_path | 调厂商分销接口拉取真实链接 |
| `scrape` | page_url / regex | 抓页面用正则提取邀请链接 |

`/r/<code>` 访问时按上述策略生成最终跳转 URL，同时记录点击次数与 referrer。

### 智能缓存

`GET /api/v1/models/` 使用 `cache.GetOrSet("models:public:list", ttl, ...)`：

- 命中缓存 → 直接返回，不查表
- 未命中 → 加载 DB → 写缓存
- 后台「同步模型」或「强制刷新首页缓存」会失效 `models:` 前缀的所有键

### 测试连接与同步

在厂商管理点击 **测试连接**：

- `VENDOR_API_MODE=mock`：返回该厂商对应的预置模型（无网调用）
- `VENDOR_API_MODE=real`：实际向 `{api_base}/v1/models` 发请求并按厂商认证适配器附加 Header

点击 **同步模型** 会把结果 upsert 到 `models` 表并刷新首页缓存。

## API 一览

公开：

| Method | Path | 说明 |
|---|---|---|
| POST | `/api/v1/auth/login` | 管理员登录，返回 JWT |
| GET  | `/api/v1/models/`    | 首页价格看板数据（带缓存）|
| GET  | `/r/:code`           | 推广短链 302 跳转 |
| GET  | `/healthz`           | 健康检查 |

后台（需 `Authorization: Bearer <token>`）：

```
GET    /api/v1/admin/auth/me
GET    /api/v1/admin/vendors/schemas
GET    /api/v1/admin/vendors
POST   /api/v1/admin/vendors
GET    /api/v1/admin/vendors/:id
PUT    /api/v1/admin/vendors/:id
DELETE /api/v1/admin/vendors/:id
POST   /api/v1/admin/vendors/:id/test
POST   /api/v1/admin/vendors/:id/sync

GET    /api/v1/admin/models
PUT    /api/v1/admin/models/:id
DELETE /api/v1/admin/models/:id
POST   /api/v1/admin/models/refresh

GET    /api/v1/admin/links
POST   /api/v1/admin/links
PUT    /api/v1/admin/links/:id
DELETE /api/v1/admin/links/:id
```

## 安全注意

1. **`ENCRYPTION_KEY` 必须固定**：生产部署前先固化到 `.env`，否则重启会导致已加密的 API Key 无法解密。可用 `openssl rand -base64 32` 生成。
2. **JWT_SECRET 不要使用默认值**。
3. **API Key 不明文回传**：后台 GET 接口返回脱敏后的字符串（`sk-****xxx`）。前端如未修改字段，PUT 时会保留原密文。
4. **CORS**：当前是 `*`，生产请收紧。

## 切换 PostgreSQL

```ini
DB_DRIVER=postgres
DB_DSN=host=localhost user=postgres password=xxx dbname=modelmarket port=5432 sslmode=disable
```

应用启动时会自动迁移所有表。

## 日志示例

```
INFO[2026-06-29 08:38:55] Admin login ok: admin
INFO[2026-06-29 08:38:55] [200] POST /api/v1/auth/login 103ms ip=::1
DEBU[2026-06-29 08:39:01] Cache hit: key=models:public:list
INFO[2026-06-29 08:39:30] Distribution redirect: code=a34944684b3779dc -> https://platform.openai.com/signup?ref=a34944684b3779dc
WARN[2026-06-29 08:39:31] TestConnection failed vendor=Foo: api_key is required
```
