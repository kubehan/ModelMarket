import type { APIRoute } from 'astro'

export const GET: APIRoute = ({ site }) => {
  const base = (site?.toString() || 'https://modelmarket.example.com').replace(/\/$/, '')
  const now = new Date().toISOString().split('T')[0]
  const pages = [
    { path: '/', priority: '1.0' },
    { path: '/changelog', priority: '0.6' }
  ]
  const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${pages.map(p => `  <url>
    <loc>${base}${p.path}</loc>
    <lastmod>${now}</lastmod>
    <changefreq>weekly</changefreq>
    <priority>${p.priority}</priority>
  </url>`).join('\n')}
</urlset>
`
  return new Response(body, {
    headers: { 'Content-Type': 'application/xml; charset=utf-8' }
  })
}
