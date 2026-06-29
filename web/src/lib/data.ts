/**
 * 数据加载器：在构建时把 data/*.json 读入并通过 Zod 校验。
 * 由 Astro 页面以 import 形式调用，保证类型安全 + 数据出错时构建直接失败。
 */
import plansRaw from '../../data/plans.json'
import configRaw from '../../data/config.json'
import { PlansFileSchema, ConfigSchema, type Plan, type Config } from './schema'

const plansFile = PlansFileSchema.parse(plansRaw)
const config: Config = ConfigSchema.parse(configRaw)

export const plans: Plan[] = plansFile.plans
export const updatedAt: string = plansFile.updatedAt
export { config }

// 工具：根据 id 索引
export const planById: Map<string, Plan> = new Map(plans.map(p => [p.id, p]))

/** 按推荐分组拼装 */
export function resolveRecommendation(key: string): Plan[] {
  const group = config.recommendationGroups.find(g => g.key === key)
  if (!group) return []
  return group.planIds.map(id => planById.get(id)).filter(Boolean) as Plan[]
}

/** 所有平台名列表（去重） */
export function allVendors(): string[] {
  return Array.from(new Set(plans.map(p => p.vendor)))
}

/** 所有标签（去重） */
export function allTags(): string[] {
  return Array.from(new Set(plans.flatMap(p => p.tags)))
}

/** 所有支持的模型（去重） */
export function allModels(): string[] {
  return Array.from(new Set(plans.flatMap(p => p.models)))
}
