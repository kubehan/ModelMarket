import { z } from 'zod'

// 套餐类型
export const PlanType = z.enum(['coding', 'agent', 'token'])
export type PlanType = z.infer<typeof PlanType>

// 套餐状态
export const PlanStatus = z.enum(['active', 'limited', 'sold_out', 'offline'])
export type PlanStatus = z.infer<typeof PlanStatus>

// 单个套餐
export const PlanSchema = z.object({
  id: z.string(),                       // 唯一 id，例如 'zhipu-coding-pro'
  vendor: z.string(),                   // 平台名，例如 '智谱AI'
  vendorSlug: z.string(),               // 平台 slug
  vendorOfficialUrl: z.string().url().optional(),
  vendorLogoUrl: z.string().optional(),
  /** 平台简码（如 ZP / KM），用于生成 monogram 头像 */
  vendorMonogram: z.string().min(1).max(3).optional(),

  name: z.string(),                     // 套餐名
  type: PlanType,
  status: PlanStatus.default('active'),
  rating: z.number().int().min(0).max(5).default(0),
  tags: z.array(z.string()).default([]),

  // 价格（人民币元）。0/null 表示该周期不可选
  firstMonthPrice: z.number().optional(),
  monthlyPrice:    z.number().optional(),
  quarterlyPrice:  z.number().optional(),
  yearlyPrice:     z.number().optional(),

  // 用量
  requestsPer5h:   z.number().optional(),
  requestsPerWeek: z.number().optional(),
  requestsPerMonth: z.number().optional(),
  tokenLimitM:     z.number().optional(),
  measuredTokenM:  z.number().optional(),

  models: z.array(z.string()).default([]),
  benefits: z.array(z.string()).default([]),
  highlights: z.array(z.string()).default([]),

  actionSlug: z.string().optional(),
  actionUrl:  z.string().url().optional(),

  note: z.string().optional(),

  /** 上次人工核对该套餐数据的日期（ISO yyyy-mm-dd），用于 last-verified 徽章 */
  lastVerifiedAt: z.string().optional(),
  /** 简短价格变动日志：[{date,desc}]，UI 会显示最近一条 */
  priceHistory: z.array(z.object({
    date: z.string(),
    desc: z.string()
  })).default([])
})
export type Plan = z.infer<typeof PlanSchema>

export const PlansFileSchema = z.object({
  updatedAt: z.string(),                // ISO date
  plans: z.array(PlanSchema)
})
export type PlansFile = z.infer<typeof PlansFileSchema>

// 推荐位分组
export const RecommendationGroupSchema = z.object({
  key: z.string(),
  title: z.string(),
  subtitle: z.string().optional(),
  /** 套餐 id 列表，从 plans.json 取详细信息 */
  planIds: z.array(z.string())
})
export type RecommendationGroup = z.infer<typeof RecommendationGroupSchema>

// 站点配置
export const ConfigSchema = z.object({
  site: z.object({
    name: z.string(),
    tagline: z.string(),
    description: z.string(),
    updatedNote: z.string().optional()
  }),
  recommendationGroups: z.array(RecommendationGroupSchema),
  presetTags: z.array(z.object({
    label: z.string(),
    tag: z.string()
  })).default([]),
  community: z.object({
    feishuTitle: z.string().optional(),
    feishuUrl: z.string().url().optional(),
    githubUrl: z.string().url().optional(),
    feedbackUrl: z.string().url().optional()
  }).default({}),
  changelog: z.array(z.object({
    date: z.string(),
    items: z.array(z.string())
  })).default([])
})
export type Config = z.infer<typeof ConfigSchema>
