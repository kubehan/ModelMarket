/**
 * ModelMarket — 客户端筛选/排序/CmdK/对比/计算器/详情/Drawer
 */
;(() => {
'use strict'

// ─── 工具 ────────────────────────────────────────────────────────────
const $ = <T=HTMLElement>(sel: string, ctx?: ParentNode) =>
  (ctx ?? document).querySelector<T>(sel)!
const $$ = <T=HTMLElement>(sel: string, ctx?: ParentNode) =>
  Array.from((ctx ?? document).querySelectorAll<T>(sel))

let debounceTimer: number
const debounce = (fn: () => void, ms = 200) => {
  clearTimeout(debounceTimer); debounceTimer = window.setTimeout(fn, ms)
}

// ─── 状态 ────────────────────────────────────────────────────────────
interface FilterState {
  q: string; type: string; vendor: string; model: string
  showOffline: boolean; tag: string
  sort: { col: string; dir: 'asc' | 'desc' } | null
}
const state: FilterState = {
  q: '', type: 'all', vendor: '', model: '',
  showOffline: false, tag: '',
  sort: null
}

// ─── DOM 引用 ────────────────────────────────────────────────────────
const tbody   = document.querySelector('#plan-table tbody')!
const rows    = $$<HTMLTableRowElement>('tr.plan-row', tbody)
const countEl = $('#filter-count')
const totalEl = $('#filter-total')
const emptyEl = $('#empty-state')
const tray    = $('#compare-tray')
const trayCount = $('#tray-count')
const trayList  = $('#tray-list')
let selectedIds = new Set<string>()

if (totalEl) totalEl.textContent = String(rows.length)

// ─── 核心筛选 ────────────────────────────────────────────────────────
function update() {
  const q = state.q.trim().toLowerCase()
  let shown = 0
  for (const row of rows) {
    const rType   = row.dataset.type || ''
    const rVendor = row.dataset.vendor || ''
    const rStatus = row.dataset.status || ''
    const rModels = (row.dataset.models || '').toLowerCase()
    const rTags   = (row.dataset.tags || '').toLowerCase()
    const rSearch = row.dataset.search || ''

    let v = true
    if (state.type !== 'all' && rType !== state.type) v = false
    if (state.vendor && rVendor !== state.vendor) v = false
    if (state.model && !rModels.includes(state.model.toLowerCase())) v = false
    if (state.tag && !rTags.includes(state.tag.toLowerCase())) v = false
    if (q && !rSearch.includes(q)) v = false
    if (rStatus === 'offline' && !state.showOffline) v = false
    row.style.display = v ? '' : 'none'
    if (v) shown++
  }
  if (countEl) countEl.textContent = String(shown)
  if (emptyEl) emptyEl.classList.toggle('hidden', shown !== 0)

  applySort()
  syncURL()
}

// ─── URL 同步 ────────────────────────────────────────────────────────
function syncURL() {
  const p = new URLSearchParams()
  if (state.q)        p.set('q', state.q)
  if (state.type !== 'all')   p.set('type', state.type)
  if (state.vendor)   p.set('vendor', state.vendor)
  if (state.model)    p.set('model', state.model)
  if (state.showOffline) p.set('offline', '1')
  if (state.tag)      p.set('tag', state.tag)
  const s = p.toString()
  history.replaceState(null, '', s ? `?${s}` : '/')
}

function loadURL() {
  const p = new URLSearchParams(location.search)
  if (p.has('q'))       state.q        = p.get('q')!
  if (p.has('type'))    state.type     = p.get('type')!
  if (p.has('vendor'))  state.vendor   = p.get('vendor')!
  if (p.has('model'))   state.model    = p.get('model')!
  if (p.has('offline')) state.showOffline = true
  if (p.has('tag'))     state.tag      = p.get('tag')!
  /* restore UI elements */
  const filterInput = $<HTMLInputElement>('#filter-search')
  filterInput && (filterInput.value = state.q)
  ;($<HTMLSelectElement>('#filter-vendor')).value = state.vendor
  ;($<HTMLSelectElement>('#filter-model')).value  = state.model
  ;($<HTMLInputElement>('#filter-show-offline')).checked = state.showOffline
  $$('.filter-type-btn').forEach(b => {
    b.setAttribute('data-active', b.dataset.type === state.type ? 'true' : 'false')
  })
  $$('.preset-tag-btn').forEach(b => {
    b.classList.toggle('is-active-pill', b.dataset.tag === state.tag)
  })
}

// ─── 排序 ─────────────────────────────────────────────────────────────
function applySort() {
  const s = state.sort
  if (!s) return
  const visible = rows.filter(r => r.style.display !== 'none')
  // 清除已有指示
  $$('th.sortable', tbody.closest('table')!).forEach(th => {
    th.classList.remove('sort-asc', 'sort-desc')
  })
  const activeTh = [...$$('th.sortable', tbody.closest('table')!)].find(
    th => th.dataset.sort === s.col
  )
  if (activeTh) activeTh.classList.add(s.dir === 'asc' ? 'sort-asc' : 'sort-desc')

  const c = s.col as keyof HTMLTableRowElement['dataset']
  visible.sort((a, b) => {
    const va = (a.dataset[c] ?? '').toString()
    const vb = (b.dataset[c] ?? '').toString()
    const na = parseFloat(va), nb = parseFloat(vb)
    const diff = (!isNaN(na) && !isNaN(nb)) ? na - nb : va.localeCompare(vb, 'zh')
    return s.dir === 'asc' ? diff : -diff
  })
  visible.forEach((row, i) => {
    const ref = visible[i + 1] || null
    tbody.insertBefore(row, ref)
  })
}

// ─── 筛选事件绑定 ────────────────────────────────────────────────────
$$('.filter-type-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    $$('.filter-type-btn').forEach(b => b.setAttribute('data-active', 'false'))
    btn.setAttribute('data-active', 'true')
    state.type = btn.dataset.type || 'all'
    update()
  })
})
$$('.preset-tag-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    const tag = btn.dataset.tag || ''
    const isActive = btn.classList.contains('is-active-pill')
    $$('.preset-tag-btn').forEach(b => b.classList.remove('is-active-pill'))
    if (!isActive) { btn.classList.add('is-active-pill'); state.tag = tag }
    else state.tag = ''
    update()
  })
})
$<HTMLInputElement>('#filter-search')?.addEventListener('input', (e) => {
  state.q = (e.target as HTMLInputElement).value
  debounce(update)
})
$<HTMLSelectElement>('#filter-vendor')?.addEventListener('change', (e) => {
  state.vendor = (e.target as HTMLSelectElement).value; update()
})
$<HTMLSelectElement>('#filter-model')?.addEventListener('change', (e) => {
  state.model = (e.target as HTMLSelectElement).value; update()
})
$<HTMLInputElement>('#filter-show-offline')?.addEventListener('change', (e) => {
  state.showOffline = (e.target as HTMLInputElement).checked; update()
})
$('#filter-reset')?.addEventListener('click', () => {
  state.q = ''; state.type = 'all'; state.vendor = ''; state.model = ''
  state.showOffline = false; state.tag = ''
  ;($<HTMLInputElement>('#filter-search')).value = ''
  ;($<HTMLSelectElement>('#filter-vendor')).value = ''
  ;($<HTMLSelectElement>('#filter-model')).value = ''
  ;($<HTMLInputElement>('#filter-show-offline')).checked = false
  $$('.filter-type-btn').forEach(b => b.setAttribute('data-active', b.dataset.type === 'all' ? 'true' : 'false'))
  $$('.preset-tag-btn').forEach(b => b.classList.remove('is-active-pill'))
  update()
})

// ─── 列排序 ──────────────────────────────────────────────────────────
$$('th.sortable').forEach(th => {
  th.addEventListener('click', () => {
    const col = th.dataset.sort || ''
    const same = state.sort?.col === col
    state.sort = { col, dir: same && state.sort!.dir === 'asc' ? 'desc' : 'asc' }
    update()
  })
})

// ─── 对比 Tray ──────────────────────────────────────────────────────
$$<HTMLInputElement>('input[data-cmp]').forEach(cb => {
  cb.addEventListener('change', () => {
    const id = cb.dataset.id || ''
    if (cb.checked) selectedIds.add(id)
    else selectedIds.delete(id)
    const row = cb.closest('tr')
    if (row) row.classList.toggle('selected', cb.checked)
    updateTray()
  })
})
function updateTray() {
  if (!tray || !trayCount || !trayList) return
  const n = selectedIds.size
  trayCount.textContent = String(n)
  tray.classList.toggle('show', n > 0)
  trayList.innerHTML = ''
  selectedIds.forEach(id => {
    const row = rows.find(r => r.dataset.id === id)
    if (!row) return
    const span = document.createElement('span')
    span.className = 'chip chip-violet text-[10px]'
    span.textContent = row.dataset.vendor + ' ' + row.dataset.id
    trayList.appendChild(span)
  })
}
$('#tray-clear')?.addEventListener('click', () => {
  selectedIds.clear()
  $$<HTMLInputElement>('input[data-cmp]').forEach(cb => { cb.checked = false })
  $$('tr.plan-row').forEach(r => r.classList.remove('selected'))
  updateTray()
})
$('#tray-compare')?.addEventListener('click', () => {
  if (selectedIds.size < 2) return
  openDrawer(Array.from(selectedIds))
})

// ─── 详情 Drawer ──────────────────────────────────────────────────
const detailBackdrop = $('#detail-backdrop')
const detailDrawer   = $('#detail-drawer')
const detailContent  = $('#detail-content')

$$('.detail-btn').forEach(btn => {
  btn.addEventListener('click', () => openDrawer([btn.dataset.id || '']))
})
$('#detail-backdrop')?.addEventListener('click', closeDrawer)

function openDrawer(ids: string[]) {
  if (!detailDrawer || !detailBackdrop || !detailContent) return
  detailBackdrop.classList.add('show')
  detailDrawer.classList.add('show')
  document.body.style.overflow = 'hidden'

  const datas = ids.map(id => {
    const row = rows.find(r => r.dataset.id === id)
    if (!row) return null
    return {
      id, vendor: row.dataset.vendor, name: row.dataset.id,
      type: row.dataset.type, status: row.dataset.status,
      models: row.dataset.models, tags: row.dataset.tags,
      price: row.dataset.monthlyPrice || row.dataset['firstMonthPrice']
    }
  }).filter(Boolean)
  detailContent.innerHTML = renderDrawerContent(datas)
}
function closeDrawer() {
  if (!detailDrawer || !detailBackdrop) return
  detailBackdrop.classList.remove('show')
  detailDrawer.classList.remove('show')
  document.body.style.overflow = ''
}
function renderDrawerContent(datas: any[]): string {
  if (datas.length === 1) {
    const d = datas[0]
    return `
      <div class="text-[10px] uppercase tracking-[0.15em] text-fg-3 font-bold mb-3">PLAN DETAIL</div>
      <h2 class="text-xl font-bold text-fg-1">${d.name}</h2>
      <p class="text-sm text-fg-3 mt-1">${d.vendor} · ${d.type}</p>
      <div class="divider-glow my-5"></div>
      <div class="space-y-3 text-sm text-fg-2">
        <div><span class="text-fg-3 w-20 inline-block">Status</span>${d.status}</div>
        <div><span class="text-fg-3 w-20 inline-block">Models</span>${d.models?.split(',').map((m:string)=>`<span class="chip chip-cyan text-[10px]">${m.trim()}</span>`).join(' ') || '-'}</div>
        <div><span class="text-fg-3 w-20 inline-block">Tags</span>${d.tags?.split(',').map((t:string)=>`<span class="chip chip-violet text-[10px]">${t.trim()}</span>`).join(' ') || '-'}</div>
        <div><span class="text-fg-3 w-20 inline-block">月付</span><span class="price-strong text-base">¥${d.price || '-'}</span></div>
      </div>`
  }
  if (datas.length > 1) {
    let html = `<div class="text-[10px] uppercase tracking-[0.15em] text-fg-3 font-bold mb-3">COMPARE</div>
      <table class="w-full text-sm"><thead><tr class="text-[10px] text-fg-3 tracking-wider uppercase">`
    html += datas.map(d => `<th class="px-2 py-2 font-semibold text-left">${d.name}</th>`).join('')
    html += '</tr></thead><tbody>'
    const fields = ['vendor', 'type', 'status', 'price', 'models']
    const labels = ['平台', '类型', '状态', '月付', '模型']
    for (let fi = 0; fi < fields.length; fi++) {
      html += `<tr class="border-b border-line-1"><td class="px-2 py-2 text-fg-3 text-[11px]">${labels[fi]}</td>`
      html += datas.map(d => {
        const v = d[fields[fi]]
        if (fields[fi] === 'models')
          return `<td class="px-2 py-2">${v?.split(', ').map((m:string)=>`<span class="chip chip-cyan text-[10px]">${m}</span>`).join(' ') || '-'}</td>`
        return `<td class="px-3 py-2">${v || '-'}</td>`
      }).join('')
      html += '</tr>'
    }
    html += '</tbody></table>'
    return html
  }
  return '<p class="text-fg-3 text-sm">无数据</p>'
}
// ESC 关闭
document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') closeDrawer()
})

// ─── Cmd-K ───────────────────────────────────────────────────────────
const cmdkBackdrop = $('#cmdk-backdrop')
const cmdkPanel    = $('#cmdk')
const cmdkInput    = $<HTMLInputElement>('#cmdk-input')
const cmdkResults  = $('#cmdk-results')
let cmdkData: any[] = []
try {
  const el = document.getElementById('cmdk-data') as HTMLScriptElement
  if (el) cmdkData = JSON.parse(el.textContent || '[]')
} catch {}
let cmdkIdx = -1

function openCmdK() {
  if (!cmdkBackdrop || !cmdkPanel) return
  cmdkBackdrop.classList.add('show')
  cmdkPanel.classList.add('show')
  cmdkInput?.focus()
  cmdkInput && (cmdkInput.value = '')
  cmdkIdx = -1
  filterCmdK('')
  document.body.style.overflow = 'hidden'
}
function closeCmdK() {
  if (!cmdkBackdrop || !cmdkPanel) return
  cmdkBackdrop.classList.remove('show')
  cmdkPanel.classList.remove('show')
  document.body.style.overflow = ''
}
function filterCmdK(q: string) {
  if (!cmdkResults) return
  const lower = q.toLowerCase().trim()
  // 结构化查询
  let struct: { price?: number; model?: string; type?: string } = {}
  const priceM = lower.match(/price\s*([<>=]+)\s*([\d.]+)/)
  if (priceM) struct.price = parseFloat(priceM[1])
  const modelM = lower.match(/model:(\w+)/)
  if (modelM) struct.model = modelM[1].toLowerCase()
  const typeM = lower.match(/type:(\w+)/)
  if (typeM) struct.type = typeM[1].toLowerCase()

  const matched = cmdkData.filter((d: any) => {
    if (!q) return true
    if (struct.price && (d.ppm == null || d.ppm > struct.price)) return false
    if (struct.model && !d.models?.join(',').toLowerCase().includes(struct.model)) return false
    if (struct.type && d.type !== struct.type) return false
    // 自由文本搜索
    if (!struct.price && !struct.model && !struct.type) {
      const haystack = [d.name, d.vendor, ...(d.models || []), ...(d.tags || [])].join(' ').toLowerCase()
      if (!haystack.includes(lower)) return false
    }
    return true
  }).slice(0, 12)

  cmdkIdx = -1
  cmdkResults.innerHTML = matched.length
    ? matched.map((d: any, i: number) => `
      <div class="cmdk-item ${i === cmdkIdx ? 'active' : ''}" data-idx="${i}" data-id="${d.id}">
        <div class="w-7 h-7 rounded-lg bg-violet-500/20 flex items-center justify-center font-bold text-xs text-violet-400">${d.monogram || d.vendor[0]}</div>
        <div class="flex-1 min-w-0">
          <div class="text-sm text-fg-1 truncate">${d.name}</div>
          <div class="text-[11px] text-fg-3">${d.vendor}${d.ppm ? ` · ¥${d.ppm}/1M` : ''}</div>
        </div>
        <span class="chip chip-violet text-[10px]">${d.type}</span>
      </div>`).join('')
    : '<div class="p-6 text-center text-fg-3 text-sm">无匹配</div>'
}
function selectCmdK(idx: number) {
  const items = $$('.cmdk-item', cmdkResults)
  const el = items[idx]
  if (!el) return
  const id = el.dataset.id
  if (!id) return
  closeCmdK()
  // 滚动到该行
  const row = rows.find(r => r.dataset.id === id)
  if (row) {
    row.scrollIntoView({ behavior: 'smooth', block: 'center' })
    row.style.outline = '2px solid rgba(139,92,246,.5)'
    setTimeout(() => row.style.outline = '', 2000)
    state.q = row.dataset.search?.split(' ').slice(0,3).join(' ') || ''
    const searchInput = $<HTMLInputElement>('#filter-search')
    if (searchInput) searchInput.value = state.q
    update()
  }
}
cmdkInput?.addEventListener('input', (e) => filterCmdK((e.target as HTMLInputElement).value))
cmdkInput?.addEventListener('keydown', (e) => {
  const items = $$('.cmdk-item', cmdkResults)
  if (e.key === 'ArrowDown') { e.preventDefault(); cmdkIdx = Math.min(cmdkIdx + 1, items.length - 1); highlightCmdK() }
  if (e.key === 'ArrowUp') { e.preventDefault(); cmdkIdx = Math.max(cmdkIdx - 1, 0); highlightCmdK() }
  if (e.key === 'Enter') { e.preventDefault(); selectCmdK(cmdkIdx) }
  if (e.key === 'Escape') closeCmdK()
})
function highlightCmdK() {
  $$('.cmdk-item', cmdkResults).forEach((el, i) => el.classList.toggle('active', i === cmdkIdx))
  const active = $$('.cmdk-item', cmdkResults)[cmdkIdx]
  active?.scrollIntoView({ block: 'nearest' })
}
cmdkResults?.addEventListener('click', (e) => {
  const item = (e.target as HTMLElement).closest('.cmdk-item') as HTMLElement
  if (!item) return
  const idx = parseInt(item.dataset.idx || '0')
  closeCmdK()
  const id = item.dataset.id
  const row = rows.find(r => r.dataset.id === id)
  if (row) { row.scrollIntoView({ behavior: 'smooth', block: 'center' })
    row.style.outline = '2px solid rgba(139,92,246,.5)'
    setTimeout(() => row.style.outline = '', 2000) }
})
$('#open-cmdk')?.addEventListener('click', openCmdK)
cmdkBackdrop?.addEventListener('click', closeCmdK)
document.addEventListener('keydown', (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') { e.preventDefault(); openCmdK() }
})

// ─── 成本估算 Drawer ──────────────────────────────────────────────
const calcBackdrop = $('#calc-backdrop')
const calcDrawer   = $('#calc-drawer')
const calcTokens   = $<HTMLInputElement>('#calc-tokens')
const calcReqs     = $<HTMLInputElement>('#calc-requests')
const calcTokensV  = $('#calc-tokens-v')
const calcReqsV    = $('#calc-requests-v')
const calcResults  = $('#calc-results')

;$('#open-calc')?.addEventListener('click', () => {
  if (!calcBackdrop || !calcDrawer) return
  calcBackdrop.classList.add('show')
  calcDrawer.classList.add('show')
  document.body.style.overflow = 'hidden'
  runCalculation()
})
;$('#calc-close')?.addEventListener('click', closeCalc)
calcBackdrop?.addEventListener('click', closeCalc)
function closeCalc() {
  if (!calcBackdrop || !calcDrawer) return
  calcBackdrop.classList.remove('show')
  calcDrawer.classList.remove('show')
  document.body.style.overflow = ''
}
calcTokens?.addEventListener('input', (e) => {
  const v = parseInt((e.target as HTMLInputElement).value)
  if (calcTokensV) calcTokensV.textContent = v.toLocaleString()
  debounce(runCalculation, 100)
})
calcReqs?.addEventListener('input', (e) => {
  const v = parseInt((e.target as HTMLInputElement).value)
  if (calcReqsV) calcReqsV.textContent = v.toLocaleString()
  debounce(runCalculation, 100)
})
function runCalculation() {
  if (!calcResults) return
  const tokenM = parseInt(calcTokens?.value || '200')
  const reqs   = parseInt(calcReqs?.value || '5000')
  // 取所有 plan data 中的 derive 字段（ppm, eff）对没有 ppm 的 plan 预估算
  const scored = cmdkData.map((d: any) => {
    if (!d.ppm) return { ...d, estCost: null, score: -1 }
    // 成本 = 月资源费 + 请求溢出（简化：tokenM * ppm）
    const est = Math.round(tokenM * d.ppm)
    return { ...d, estCost: est }
  }).filter(d => d.estCost != null)
    .sort((a: any, b: any) => a.estCost - b.estCost)
    .slice(0, 10)
  calcResults.innerHTML = scored.length
    ? scored.map((d: any, i: number) => `
      <li class="flex items-center justify-between gap-3 glass rounded-lg p-3">
        <div class="flex items-center gap-2.5 min-w-0">
          <span class="font-mono text-[11px] text-fg-3 w-5 text-right">${i + 1}</span>
          <div class="w-6 h-6 rounded bg-violet-500/20 flex items-center justify-center font-bold text-[10px] text-violet-400">${d.monogram || d.vendor[0]}</div>
          <div class="min-w-0"><div class="text-sm text-fg-1 truncate">${d.name}</div>
            <div class="text-[10px] text-fg-3">${d.vendor} · ¥${d.ppm}/1M</div></div>
        </div>
        <div class="text-right shrink-0">
          <div class="font-mono text-sm font-bold text-emerald-400">¥${d.estCost.toLocaleString()}</div>
          <div class="text-[10px] text-fg-3">/ 月</div>
        </div>
      </li>`).join('')
    : '<li class="text-fg-3 text-sm text-center py-4">暂无性价比数据</li>'
}

// ─── 初始化 ──────────────────────────────────────────────────────────
loadURL()
update()

})()
