# ModelMarket — AI Coding Plan 对比站

国内主流 AI 平台 Coding Plan / Agent Plan / Token Plan 套餐对比，涵盖智谱、Kimi、MiniMax、阿里百炼、字节方舟、讯飞星火、百度千帆、腾讯云、京东云等。价格、模型、用量一表看清。

成本：**0 元**（域名除外）。Cloudflare Pages 免费额度 500 builds/月；Workers 免费 100k 请求/天；KV 免费 1k 写 / 100k 读每天。

---

## 目录

- [架构](#架构)
- [快速启动](#快速启动)
- [平台使用说明](#平台使用说明)
- [部署指南](#部署指南)
- [配置指南](#配置指南)
- [日常运营](#日常运营)
- [数据模型](#数据模型)
- [站点效果](#站点效果)
- [与 codingplan.fyi 的差异](#与-codingplanfyi-的差异)

---

## 架构

```
┌─────────────────────────┐         ┌────────────────────────┐
│ Cloudflare Pages         │         │ Cloudflare Worker      │
│   modelmarket.com        │ 跳转 →  │   go.modelmarket.com   │
│   (Astro 静态站)         │         │   (KV 短链 + 点击统计)  │
└─────────────────────────┘         └────────────────────────┘
        ↑
        │ git push
        │
  ┌─────────────┐        ┌──────────────┐
  │ GitHub Repo  │ ← 数据 │ web/data/    │
  │              │ PR 流程 │ *.json       │
  └─────────────┘        └──────────────┘
```

| 部分 | 技术 | 说明 |
|---|---|---|
| 站点 | Astro 4 + Tailwind 3 + TypeScript | 静态构建，零运行时框架开销 |
| 数据 | `web/data/plans.json` + `web/data/config.json` | 提 PR 改 JSON，CI 自动 Zod 校验 |
| 部署 | Cloudflare Pages | 全球 CDN，免费额度大方 |
| 短链 | Cloudflare Workers + KV | `/r/<slug>` 302 跳转 + 异步点击统计 |
| 主题 | CSS 变量 + 3 套换肤 | 深色/浅色/终端，localStorage 持久化 |
| CI | GitHub Actions | PR 校验数据，main 自动部署 |

---

## 快速启动

### 依赖

- Node.js 20+ / npm
- Cloudflare 账号（部署用，本地开发不需要）

### 1. 本地启动站点

```bash
cd web
npm install
npm run dev               # http://localhost:4321
```

修改 `web/data/plans.json` 后浏览器自动热更新。

### 2. 验证数据格式

```bash
cd web
npm run validate          # Zod 校验 + 推荐位引用完整性检查
```

### 3. 构建产物

```bash
cd web
npm run build             # 产物在 web/dist/
```

### 4. 短链 Worker 本地开发（可选）

```bash
cd worker
npm install
npx wrangler dev          # http://localhost:8787/r/zhipu
```

---

## 平台使用说明

### 浏览与筛选

页面打开后依次为：

1. **Hero 区域** — 数据更新时间、套餐/平台/模型总数统计条
2. **推荐位** — "综合首选""旗舰模型""高额度""随时可购"四个分组
3. **全部套餐对比表** — 16 维度的可滚动表格

**筛选操作**：

| 操作 | 位置 | 效果 |
|---|---|---|
| 文本搜索 | 左上搜索框 | 匹配平台名/套餐名/模型名/标签 |
| 类型切换 | 搜索框右侧按钮 | Coding / Agent / Token |
| 平台下拉 | 类型按钮右侧 | 按平台过滤 |
| 模型下拉 | 平台下拉右侧 | 按模型过滤 |
| 含已下线 | 复选框 | 显示已下架套餐 |
| 预设标签 | 筛选栏底部 | "模型强""性价比高""用量足" |
| 重置 | 重置按钮 | 清空所有筛选条件 |
| Cmd-K | `⌘K` / `Ctrl+K` | 打开命令面板 |

**命令面板（Cmd-K）** 支持结构化查询：

| 语法 | 示例 | 效果 |
|---|---|---|
| `price<100` | `price<50` | 每百万 token 成本低于 50 元 |
| `price=100` | `price=80` | 每百万 token 成本等于 80 元 |
| `model:名称` | `model:GLM` | 支持 GLM 系列模型的套餐 |
| `type:类型` | `type:coding` | 只看 Coding Plan |
| 自由文本 | `Kimi` | 模糊匹配名称/平台/模型/标签 |

**排序**：点击表头任意列（平台、评分、首月、月付、季付、年付、¥/1M Token、5h 请求、月 Token）切换升降序。

### 对比操作

1. 勾选表格左侧复选框（最多 4 个）
2. 底部弹出对比 Tray，显示已选套餐
3. 点击"并排对比" → 右侧滑出 Drawer，多行对比表
4. 点"清空"取消所有勾选

### 查看详情

点击表格最右侧眼睛图标 `👁` → 右侧滑出 Drawer，显示完整套餐信息（状态、模型列表、标签、月付价格）。

### 成本估算

点击筛选栏"成本估算"按钮 → 滑出计算面板：

- 拖动**每月 Token 用量**（10M–2B）
- 拖动**每月请求数**（500–50K）
- 自动按预估月成本升序排列 Top 10

计算公式：`预估月成本 = Token 用量(百万) × 每百万 Token 成本`

### 主题切换

导航栏右侧主题按钮（🌙 → ☀️ → 💻）循环切换：

| 主题 | `data-theme` | 特征 |
|---|---|---|
| 深色 | `dark`（默认） | 紫青极光暗影，`#05070d` 基色 |
| 浅色 | `light` | 白底紫蓝 accent |
| 终端 | `terminal` | GitHub Dark + 蓝绿 monospace 风格 |

选择自动保存到 `localStorage('mm-theme')`；首次访问遵循系统 `prefers-color-scheme`；初始化脚本在 `<head>` 中阻塞执行，**无闪白/闪黑**。

---

## 部署指南

### 1. Cloudflare Pages（站点）

**方式 A：Pages 直连 GitHub（推荐）**

1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Workers & Pages → Pages → **Create a project** → **Connect to Git**
3. 授权 GitHub，选择 `modelmarket` 仓库
4. 构建设置：

   | 字段 | 值 |
   |---|---|
   | Project name | `modelmarket` |
   | Production branch | `main` |
   | Framework preset | None |
   | Build command | `cd web && npm ci && node scripts/validate-data.mjs && npx astro build` |
   | Build output directory | `web/dist` |
   | Root directory | (留空) |

5. 环境变量：`NODE_VERSION` = `20`
6. 点击 **Save and Deploy**

首次部署后 Cloudflare 分配 `xxxx.pages.dev` 域名，可在 Pages 项目 → Custom domains 绑定自定义域名。

**方式 B：GitHub Actions 推送**

在 GitHub 仓库 Settings → Secrets and variables → Actions 中添加：

| Secret | 说明 |
|---|---|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API 令牌（Account · Cloudflare Pages · Edit）|
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare Dashboard 右侧栏复制 |

推送到 `main` 后 `.github/workflows/deploy.yml` 自动构建部署。

### 2. Cloudflare Worker + KV（短链）

```bash
cd worker

# 1. 登录
npx wrangler login

# 2. 创建 KV namespace
npx wrangler kv:namespace create LINKS           # 保存输出的 id
npx wrangler kv:namespace create LINKS --preview  # 保存 preview_id

# 3. 把 id / preview_id 填入 wrangler.toml
#    [[kv_namespaces]]
#    binding = "LINKS"
#    id = "刚才的id"
#    preview_id = "刚才的preview_id"

# 4. 导入短链接映射
npm run seed                  # 生产 KV
# 或 npm run seed -- --local  # 本地预览 KV

# 5. 修改 wrangler.toml 中的 SITE_URL
#    [vars]
#    SITE_URL = "https://你的主站域名"

# 6. 部署
npm run deploy                # → https://modelmarket-go.xxxx.workers.dev
```

Worker 部署后绑自定义域名（如 `go.yourdomain.com`），然后修改 `web/astro.config.mjs` 中的 `site`。

### 3. 短链接数据更新

修改 `worker/data/links.json` 后：

```bash
cd worker
node scripts/seed-kv.mjs      # 推送到 KV，不需要重新部署 Worker
```

### 4. CI/CD 工作流

| 事件 | 触发动作 |
|---|---|
| 提交 PR（修改 `web/data/*`） | `validate.yml` — Zod 校验数据完整性 |
| 推送到 `main` | `deploy.yml` — 校验 + 构建 Pages + 部署 |
| commit 含 `[worker]` | 额外部署 Worker |
| 手动触发 `workflow_dispatch` | 部署全部 |

---

## 配置指南

### 套餐数据（`web/data/plans.json`）

每个套餐对象完整字段：

```json
{
  "id": "zhipu-coding-pro",
  "vendor": "智谱AI",
  "vendorSlug": "zhipu",
  "vendorOfficialUrl": "https://bigmodel.cn",
  "vendorMonogram": "ZP",
  "name": "CodingPlan Pro",
  "type": "coding",
  "status": "active",
  "rating": 5,
  "tags": ["最新模型", "不用抢"],
  "firstMonthPrice": 199,
  "monthlyPrice": 199,
  "quarterlyPrice": 597,
  "yearlyPrice": 2388,
  "requestsPer5h": 100,
  "requestsPerWeek": 500,
  "requestsPerMonth": 2000,
  "tokenLimitM": 200,
  "measuredTokenM": 180,
  "models": ["GLM-5.2", "GLM-5.1"],
  "benefits": ["优先队列", "无限制对话"],
  "highlights": ["全场最低价 ¥199"],
  "actionSlug": "zhipu",
  "actionUrl": null,
  "lastVerifiedAt": "2026-06-29T00:00:00Z",
  "priceHistory": [{ "date": "2026-06-01", "desc": "¥199" }]
}
```

**字段参考**：

| 字段 | 必填 | 类型 | 说明 |
|---|---|---|---|
| `id` | ✓ | string | 唯一标识，用于推荐分组和 URL 参数 |
| `vendor` | ✓ | string | 平台显示名 |
| `vendorSlug` | ✓ | string | 英文标识，用于 monogram 色相计算 |
| `vendorOfficialUrl` | | string(url) | 官网链接 |
| `vendorMonogram` | | string | 头像文字（最多 2 字符，留空取 vendor 前 2 字）|
| `name` | ✓ | string | 套餐名 |
| `type` | ✓ | `coding` / `agent` / `token` | 影响类型徽章颜色 |
| `status` | | one of | `active`(绿) / `limited`(黄) / `sold_out`(红) / `offline`(灰) |
| `rating` | | number(0-5) | 编辑评分 |
| `tags` | | string[] | 关联标签，与 presetTags 配合筛选 |
| `firstMonthPrice` | | number | 首月优惠价，用于 sparkline 首点 |
| `monthlyPrice` | | number | 常规月价 |
| `quarterlyPrice` | | number | 季价 |
| `yearlyPrice` | | number | 年价 |
| `requestsPer5h` | | number | 5 小时请求上限 |
| `requestsPerWeek` | | number | 周请求上限 |
| `requestsPerMonth` | | number | 月请求上限 |
| `tokenLimitM` | | number | 月 token 上限（百万单位）|
| `measuredTokenM` | | number | 实测 token 上限 |
| `models` | | string[] | 支持模型列表 |
| `benefits` | | string[] | 权益列表 |
| `highlights` | | string[] | 推荐理由（支持 HTML 标签）|
| `actionSlug` | | string | 短链接 slug，对应 Worker KV `link:<slug>` |
| `actionUrl` | | string(url) | 兜底外链（actionSlug 未定义时使用）|
| `lastVerifiedAt` | | string(ISO8601) | 最近核对时间，影响"X 天前核对"徽章 |
| `priceHistory` | | array | `{date, desc}` 历史价格记录，用于 sparkline 趋势 + 价格变动检测 |

### 站点配置（`web/data/config.json`）

```json
{
  "site": {
    "name": "ModelMarket",
    "tagline": "AI Coding Plan 对比工具",
    "description": "国内主流 AI 平台...套餐对比...",
    "updatedNote": "目前整体趋势是..."
  },
  "recommendationGroups": [
    {
      "key": "top_rated",
      "title": "综合首选",
      "subtitle": "编辑评分最高，覆盖最广",
      "planIds": ["zhipu-coding-pro", "xunfei-spark-pro", "..."]
    }
  ],
  "presetTags": [
    { "label": "模型强", "tag": "最新模型" },
    { "label": "性价比高", "tag": "不用抢" }
  ],
  "community": {
    "feishuTitle": "加入飞书群讨论",
    "githubUrl": "https://github.com/yourname/modelmarket",
    "feedbackUrl": "https://github.com/yourname/modelmarket/issues"
  },
  "changelog": [
    { "date": "2026.6.29", "items": ["重构为静态站架构"] }
  ]
}
```

- **recommendationGroups**：最多 4 组，key 唯一；planIds 必须全部存在于 `plans.json`
- **presetTags**：FilterBar 底部的快速预设标签按钮
- **community**：导航栏 + 页脚链接
- **changelog**：更新日志页面内容

### 短链接映射（`worker/data/links.json`）

```json
{
  "links": {
    "zhipu": { "url": "https://bigmodel.cn/?ref=MODELMARKET", "note": "智谱AI" }
  }
}
```

建议 URL 尾部加 `?ref=MODELMARKET` 用于来源追踪。

### 站点配置（`web/astro.config.mjs`）

```js
export default defineConfig({
  site: 'https://你的域名.com',    // 改为 Pages 自定义域名
  integrations: [tailwind({ applyBaseStyles: false })],
  build: { inlineStylesheets: 'auto' },
  output: 'static',
  compressHTML: true,
  prefetch: { prefetchAll: true }
})
```

### Worker 配置（`worker/wrangler.toml`）

```toml
name = "modelmarket-go"
compatibility_date = "2025-01-01"

[[kv_namespaces]]
binding = "LINKS"
id = "你的KV_ID"

[vars]
SITE_URL = "https://你的域名.com"
TRACK_CLICKS = "1"          # 关闭设为 "0"
```

### 主题配置

三套主题的 CSS 变量在 `web/src/styles/global.css` 中定义：

| CSS 选择器 | 主题 | 基色 | Accent |
|---|---|---|---|
| `:root`（默认） | 深色 | `#05070d` | violet `#8b5cf6` + cyan `#06b6d4` |
| `[data-theme="light"]` | 浅色 | `#ffffff` | violet `#7c3aed` + cyan `#0e7490` |
| `[data-theme="terminal"]` | 终端 | `#0d1117` | blue `#58a6ff` + green `#3fb950` |

如需新增主题，在 `global.css` 中添加 `[data-theme="你的主题名"]` 块，重写对应的 `--*` CSS 变量即可。

---

## 日常运营

### 改数据流程

1. 在 GitHub 上 fork / 切分支
2. 编辑 `web/data/plans.json`：
   - **新增套餐**：往 `plans` 数组追加对象（参考已有项）
   - **改价格**：直接改对应字段
   - **下线套餐**：`"status"` 改为 `"offline"`
   - **新增短链**：同时在 `worker/data/links.json` 添加 slug → URL
3. 编辑 `web/data/config.json`：调整推荐位 `recommendationGroups[*].planIds`
4. 提交 PR → GitHub Actions 自动运行 `validate-data.mjs`（Zod schema 校验 + 推荐位完整性检查）
5. 校验通过后合并到 `main` → 自动部署

### 短链接独立更新

修改 `worker/data/links.json` 后：

```bash
cd worker
npm run seed    # 推送到 KV，不需要重新部署 Worker
```

### 数据关系

```
worker/data/links.json          web/data/config.json
       │                              │
       ▼                              ▼
  Cloudflare KV          Zod Schema 校验 ← GitHub Actions (validate.yml)
       │                              │
       ▼                              ▼
  Worker /r/:slug          Astro 构建 + Tailwind + TypeScript
       │                              │
       ▼                              ▼
  302 Redirect             Cloudflare Pages (dist/)
                                │
                                ▼
                         静态 HTML + CSS + JS
```

---

## 数据模型

详见 `web/src/lib/schema.ts`（Zod Schema 源文件）。

### 衍生计算字段（`web/src/lib/derive.ts`）

构建时自动计算，无需手动录入：

| 字段 | 类型 | 说明 |
|---|---|---|
| `pricePerMillion` | number \| null | ¥/1M token = 最优月价 ÷ tokenLimitM |
| `effectiveMonthly` | number \| null | 最优月度价（firstMonth 或 monthly）|
| `verifiedDaysAgo` | number \| null | 距今天数 |
| `recentlyChanged` | boolean | 30 天内有无价格变动 |
| `tokenPercentile` | number(0-100) | Token 限额在数据集内的百分位 |
| `valuePercentile` | number(0-100) | 性价比在数据集内的百分位 |

### 评分标准

基准 3 颗星，价格优势/劣势 ±1 分，模型优势或独占模型 +1 分，其他优势 +1 分。

---

## 站点效果

- **Hero** — 标题、副标题、数据更新日期、趋势提示、统计计数器
- **推荐位分组**（4 组卡片，`config.json` 自定义）：
  - 综合首选 / 旗舰模型 / 高额度 / 随时可购
- **筛选条**：搜索 + 类型切换 + 平台/模型下拉 + 预置标签 + 已下线开关 + 成本估算入口
- **对比表**：16 列字段，平台/套餐列冻结，inline meter 条，移动端横向滚动
- **Cmd-K 命令面板**：模糊搜索 + 结构化查询
- **对比 Tray + Drawer**：多选并排对比
- **成本估算面板**：Token 用量/请求数滑杆 → Top 10 性价比排序
- **主题切换**：深色/浅色/终端，防闪烁初始化
- **更新日志页**：从 `config.changelog` 自动生成
- **SEO**：sitemap.xml / robots.txt / OG meta / canonical URL 完备

---

## 与 codingplan.fyi 的差异

| 维度 | codingplan.fyi | 本项目 |
|---|---|---|
| 框架 | 无（裸 HTML+JS） | Astro 4（仍 0 运行时 JS）|
| 数据校验 | 无 | Zod schema + CI 检查 |
| 部署 | Tencent COS + Cloudflare | Cloudflare Pages |
| 短链 | 腾讯 SCF + OpenResty | Cloudflare Worker + KV |
| 数据修改 | PR 改 JSON | 同 |
| 主题 | 仅浅色 | 3 套主题（深色/浅色/终端）|
| 对比能力 | 基本表格 | 排序 + 筛选 + 多选对比 + Drawer |
| 成本估算 | 无 | 滑杆驱动动态排名 |
| Cmd-K | 无 | 结构化查询命令面板 |
| 视觉风格 | 奶油色圆角 pill | 深色玻璃态+方形圆角+极光 |

---

## License

MIT
