#!/usr/bin/env node
/**
 * 批量把 data/links.json 中的链接写入 Cloudflare KV
 * 用法：
 *   1. 先 `wrangler kv:namespace create LINKS` 并把 id 填入 wrangler.toml
 *   2. 运行 `node scripts/seed-kv.mjs`            -> 批量上传到远程
 *   3. 运行 `node scripts/seed-kv.mjs --local`    -> 写到本地 dev KV
 */
import { execSync } from 'node:child_process'
import { readFileSync, writeFileSync, mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const local = process.argv.includes('--local')
const raw = JSON.parse(readFileSync(new URL('../data/links.json', import.meta.url), 'utf8'))

const flag = local ? '--local' : '--remote'
const tmp = mkdtempSync(join(tmpdir(), 'mm-kv-'))

const entries = Object.entries(raw.links)
console.log(`Uploading ${entries.length} link(s) to KV ${local ? '(local dev)' : '(remote production)'}...`)

for (const [slug, payload] of entries) {
  const key = `link:${slug}`
  const value = JSON.stringify(payload)
  const valuePath = join(tmp, `${slug}.json`)
  writeFileSync(valuePath, value)
  const cmd = `npx wrangler kv key put --binding=LINKS ${flag} "${key}" --path="${valuePath}"`
  try {
    execSync(cmd, { stdio: 'inherit' })
  } catch (e) {
    console.error(`Failed to upload ${slug}:`, e.message)
    process.exit(1)
  }
}
console.log(`✓ Done. ${entries.length} entries written.`)
