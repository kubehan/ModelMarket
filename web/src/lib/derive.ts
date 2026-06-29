import type { Plan } from './schema'

/** 套餐计算字段 */
export interface PlanDerived {
  /** ¥ / 1M token：综合月成本（取最优的月度价）/ 月 token 上限。无数据返回 null */
  pricePerMillion: number | null
  /** 最优月度价：firstMonth 或 monthly */
  effectiveMonthly: number | null
  /** 上次核对距今天数 */
  verifiedDaysAgo: number | null
  /** 近 30 天内是否有价格变动 */
  recentlyChanged: boolean
  /** 数据集内此 token 上限占的百分位（0-100） */
  tokenPercentile: number
  /** 数据集内此性价比占的百分位（0-100，越高越优） */
  valuePercentile: number
}

/** 一次计算所有套餐的派生字段（百分位需要全集） */
export function deriveAll(plans: Plan[]): Map<string, PlanDerived> {
  const now = Date.now()
  const map = new Map<string, PlanDerived>()

  // 先算每条 plan 的基础字段
  const tokens: number[] = []
  const values: number[] = []  // value 越大越好：tokenLimit/月费

  const basics = plans.map(p => {
    const eff = (p.firstMonthPrice && p.firstMonthPrice > 0)
      ? p.firstMonthPrice
      : (p.monthlyPrice && p.monthlyPrice > 0 ? p.monthlyPrice : null)
    const ppm = (eff != null && p.tokenLimitM && p.tokenLimitM > 0)
      ? +(eff / p.tokenLimitM).toFixed(2)
      : null
    const verifiedDays = p.lastVerifiedAt
      ? Math.max(0, Math.floor((now - new Date(p.lastVerifiedAt).getTime()) / 86400000))
      : null
    const recent = p.priceHistory?.some(h => {
      const ts = new Date(h.date).getTime()
      return !isNaN(ts) && (now - ts) < 30 * 86400000
    }) ?? false

    if (p.tokenLimitM) tokens.push(p.tokenLimitM)
    if (ppm != null) values.push(1 / ppm)  // 1/ppm 表示性价比正向

    return { p, eff, ppm, verifiedDays, recent }
  })

  // 计算百分位
  const tokenSorted = [...tokens].sort((a, b) => a - b)
  const valueSorted = [...values].sort((a, b) => a - b)

  for (const b of basics) {
    const tp = b.p.tokenLimitM
      ? Math.round((tokenSorted.indexOf(b.p.tokenLimitM) / Math.max(1, tokenSorted.length - 1)) * 100)
      : 0
    const v = b.ppm != null ? 1 / b.ppm : 0
    const vp = v > 0
      ? Math.round((valueSorted.indexOf(v) / Math.max(1, valueSorted.length - 1)) * 100)
      : 0

    map.set(b.p.id, {
      pricePerMillion: b.ppm,
      effectiveMonthly: b.eff,
      verifiedDaysAgo: b.verifiedDays,
      recentlyChanged: b.recent,
      tokenPercentile: tp,
      valuePercentile: vp
    })
  }
  return map
}

/** 生成 vendor monogram 头像 svg 字符串（用品牌色彩虹） */
export function monogramSVG(text: string, size = 36): string {
  const t = text.slice(0, 2).toUpperCase()
  const fontSize = Math.floor(size * 0.45)
  // 用 vendorSlug hash 选起始色相
  let hash = 0
  for (let i = 0; i < text.length; i++) hash = (hash * 31 + text.charCodeAt(i)) | 0
  const h1 = Math.abs(hash) % 360
  const h2 = (h1 + 60) % 360
  return `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${size} ${size}" width="${size}" height="${size}">
  <defs>
    <linearGradient id="g${hash}" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="hsl(${h1}, 70%, 60%)"/>
      <stop offset="100%" stop-color="hsl(${h2}, 70%, 50%)"/>
    </linearGradient>
  </defs>
  <rect width="${size}" height="${size}" rx="${size * 0.25}" fill="url(#g${hash})" opacity="0.95"/>
  <text x="50%" y="55%" font-family="Inter, sans-serif" font-size="${fontSize}" font-weight="800" fill="white" text-anchor="middle" dominant-baseline="middle">${t}</text>
</svg>`.trim()
}

/** 生成多周期价格 sparkline svg（首月/月/季均化/年均化） */
export function pricesSparkline(plan: Plan, w = 80, h = 24): string {
  const series: number[] = []
  if (plan.firstMonthPrice && plan.firstMonthPrice > 0) series.push(plan.firstMonthPrice)
  if (plan.monthlyPrice && plan.monthlyPrice > 0)       series.push(plan.monthlyPrice)
  if (plan.quarterlyPrice && plan.quarterlyPrice > 0)   series.push(plan.quarterlyPrice / 3)
  if (plan.yearlyPrice && plan.yearlyPrice > 0)         series.push(plan.yearlyPrice / 12)
  if (series.length < 2) return ''

  const min = Math.min(...series)
  const max = Math.max(...series)
  const range = max - min || 1
  const step = w / (series.length - 1)
  const pts = series.map((v, i) => {
    const x = i * step
    const y = h - ((v - min) / range) * h * 0.85 - 2
    return `${x.toFixed(1)},${y.toFixed(1)}`
  }).join(' ')
  const last = pts.split(' ').pop()?.split(',') || ['0', '0']
  const trend = series[series.length - 1] <= series[0] ? '#34d399' : '#fb7185'
  return `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${w} ${h}" width="${w}" height="${h}" aria-hidden="true">
  <polyline points="${pts}" fill="none" stroke="${trend}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="${last[0]}" cy="${last[1]}" r="2" fill="${trend}"/>
</svg>`.trim()
}

export function fmtRelativeDays(d: number | null): string {
  if (d == null) return ''
  if (d <= 1) return '今日核对'
  if (d < 7)  return `${d} 天前核对`
  if (d < 30) return `${Math.floor(d / 7)} 周前核对`
  return `${Math.floor(d / 30)} 个月前`
}
