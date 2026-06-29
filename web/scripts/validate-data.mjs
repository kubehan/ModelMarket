#!/usr/bin/env node
// 独立的 JSON 数据校验脚本（CI 用）
// 用法：node scripts/validate-data.mjs
import { readFileSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

// 在 ESM 中加载 ts schema 比较麻烦，直接重写一份等价的最小校验
import { z } from 'zod'

const __dirname = dirname(fileURLToPath(import.meta.url))
const root = resolve(__dirname, '..')

const PlanType = z.enum(['coding', 'agent', 'token'])
const PlanStatus = z.enum(['active', 'limited', 'sold_out', 'offline'])

const PlanSchema = z.object({
  id: z.string(),
  vendor: z.string(),
  vendorSlug: z.string(),
  vendorOfficialUrl: z.string().url().optional(),
  vendorLogoUrl: z.string().optional(),
  vendorMonogram: z.string().min(1).max(3).optional(),
  name: z.string(),
  type: PlanType,
  status: PlanStatus.default('active'),
  rating: z.number().int().min(0).max(5).default(0),
  tags: z.array(z.string()).default([]),
  firstMonthPrice: z.number().optional(),
  monthlyPrice: z.number().optional(),
  quarterlyPrice: z.number().optional(),
  yearlyPrice: z.number().optional(),
  requestsPer5h: z.number().optional(),
  requestsPerWeek: z.number().optional(),
  requestsPerMonth: z.number().optional(),
  tokenLimitM: z.number().optional(),
  measuredTokenM: z.number().optional(),
  models: z.array(z.string()).default([]),
  benefits: z.array(z.string()).default([]),
  highlights: z.array(z.string()).default([]),
  actionSlug: z.string().optional(),
  actionUrl: z.string().url().optional(),
  note: z.string().optional(),
  lastVerifiedAt: z.string().optional(),
  priceHistory: z.array(z.object({ date: z.string(), desc: z.string() })).default([])
})

const PlansFileSchema = z.object({
  updatedAt: z.string(),
  plans: z.array(PlanSchema)
})

const ConfigSchema = z.object({
  site: z.object({
    name: z.string(),
    tagline: z.string(),
    description: z.string(),
    updatedNote: z.string().optional()
  }),
  recommendationGroups: z.array(z.object({
    key: z.string(),
    title: z.string(),
    subtitle: z.string().optional(),
    planIds: z.array(z.string())
  })),
  presetTags: z.array(z.object({ label: z.string(), tag: z.string() })).default([]),
  community: z.object({}).passthrough().default({}),
  changelog: z.array(z.object({ date: z.string(), items: z.array(z.string()) })).default([])
})

function load(path) {
  return JSON.parse(readFileSync(resolve(root, path), 'utf8'))
}

let ok = true

try {
  const plansFile = PlansFileSchema.parse(load('data/plans.json'))
  console.log(`✓ data/plans.json   ${plansFile.plans.length} plans, updatedAt=${plansFile.updatedAt}`)

  // 唯一 id 校验
  const seen = new Set()
  for (const p of plansFile.plans) {
    if (seen.has(p.id)) {
      console.error(`✗ duplicate plan id: ${p.id}`)
      ok = false
    }
    seen.add(p.id)
  }

  const config = ConfigSchema.parse(load('data/config.json'))
  console.log(`✓ data/config.json  ${config.recommendationGroups.length} groups, ${config.changelog.length} changelog entries`)

  // 推荐位 planId 引用完整性
  for (const g of config.recommendationGroups) {
    for (const pid of g.planIds) {
      if (!seen.has(pid)) {
        console.error(`✗ recommendation group "${g.key}" references missing plan id: ${pid}`)
        ok = false
      }
    }
  }
} catch (err) {
  console.error('✗ schema validation failed:')
  console.error(err.errors || err)
  ok = false
}

if (!ok) {
  process.exit(1)
}
console.log('All data files are valid ✓')
