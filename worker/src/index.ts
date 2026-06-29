/**
 * ModelMarket 推广短链 Worker
 *
 * 路由：
 *   GET  /r/:slug      -> 302 跳转到 KV 中存的 url；记录一次点击
 *   GET  /api/stats    -> 列出所有 slug 的点击数（管理接口；上线请加鉴权或删除）
 *   GET  /healthz      -> 健康检查
 *   其他              -> 跳回 SITE_URL
 *
 * KV 数据结构：
 *   link:<slug>       -> { url: "https://...", note?: "..." }
 *   click:<slug>      -> 累计点击数（字符串数字）
 */

interface Env {
  LINKS: KVNamespace
  SITE_URL: string
  TRACK_CLICKS: string
}

interface LinkEntry {
  url: string
  note?: string
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url)

    // 健康检查
    if (url.pathname === '/healthz') {
      return json({ ok: true, time: new Date().toISOString() })
    }

    // 跳转
    const m = url.pathname.match(/^\/r\/([\w\-]+)\/?$/)
    if (m) {
      const slug = m[1].toLowerCase()
      return handleRedirect(slug, request, env, ctx)
    }

    // 简单 stats（生产请加鉴权）
    if (url.pathname === '/api/stats') {
      return handleStats(env)
    }

    // 兜底回主站
    return Response.redirect(env.SITE_URL, 302)
  }
}

async function handleRedirect(
  slug: string,
  request: Request,
  env: Env,
  ctx: ExecutionContext
): Promise<Response> {
  const raw = await env.LINKS.get(`link:${slug}`)
  if (!raw) {
    return new Response(`Link "${slug}" not found`, {
      status: 404,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' }
    })
  }

  let entry: LinkEntry
  try {
    entry = JSON.parse(raw) as LinkEntry
  } catch {
    // 也兼容只存 url 字符串的极简模式
    entry = { url: raw }
  }
  if (!entry.url) {
    return new Response('Link malformed', { status: 500 })
  }

  // 异步记录点击，不阻塞跳转
  if (env.TRACK_CLICKS === '1') {
    ctx.waitUntil(recordClick(slug, request, env))
  }

  // 防止重定向循环
  return new Response(null, {
    status: 302,
    headers: {
      Location: entry.url,
      'Cache-Control': 'private, no-cache'
    }
  })
}

async function recordClick(slug: string, request: Request, env: Env): Promise<void> {
  try {
    const key = `click:${slug}`
    const cur = await env.LINKS.get(key)
    const n = cur ? parseInt(cur, 10) : 0
    await env.LINKS.put(key, String(n + 1))

    // 详细日志（保留 30 天）：含 referrer / ua / 国家 / 时间
    const cf = (request as any).cf || {}
    const detail = {
      t: Date.now(),
      ref: request.headers.get('referer') || '',
      ua: request.headers.get('user-agent') || '',
      country: cf.country || '',
      city: cf.city || ''
    }
    await env.LINKS.put(
      `log:${slug}:${detail.t}`,
      JSON.stringify(detail),
      { expirationTtl: 60 * 60 * 24 * 30 }
    )
  } catch (e) {
    console.error('recordClick error', e)
  }
}

async function handleStats(env: Env): Promise<Response> {
  const list = await env.LINKS.list({ prefix: 'click:' })
  const out: Record<string, number> = {}
  for (const k of list.keys) {
    const slug = k.name.slice('click:'.length)
    const v = await env.LINKS.get(k.name)
    out[slug] = v ? parseInt(v, 10) : 0
  }
  return json({ clicks: out, total: Object.values(out).reduce((a, b) => a + b, 0) })
}

function json(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data, null, 2), {
    status,
    headers: { 'Content-Type': 'application/json; charset=utf-8' }
  })
}
