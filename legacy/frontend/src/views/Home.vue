<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../api'
import { ElMessage } from 'element-plus'

const rows = ref([])
const loading = ref(false)
const search = ref('')
const sortBy = ref('input_price')

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/models/')
    rows.value = data
  } finally {
    loading.value = false
  }
}

const filtered = computed(() => {
  let list = rows.value
  if (search.value.trim()) {
    const s = search.value.toLowerCase()
    list = list.filter(m =>
      (m.model_name || '').toLowerCase().includes(s) ||
      (m.vendor_name || '').toLowerCase().includes(s)
    )
  }
  return [...list].sort((a, b) => {
    if (sortBy.value === 'context_length') return (b.context_length || 0) - (a.context_length || 0)
    return (a[sortBy.value] || 0) - (b[sortBy.value] || 0)
  })
})

const vendors = computed(() => {
  const m = new Map()
  rows.value.forEach(r => {
    if (!m.has(r.vendor_id)) m.set(r.vendor_id, { id: r.vendor_id, name: r.vendor_name, logo: r.vendor_logo_url })
  })
  return Array.from(m.values())
})

function fmt(p) { return p == null ? '-' : '$' + Number(p).toFixed(2) }

function copyLink(row) {
  if (!row.distribution_url) return
  const url = location.origin + row.distribution_url
  navigator.clipboard.writeText(url)
  ElMessage.success('推广链接已复制：' + url)
}

onMounted(load)
</script>

<template>
  <div class="hero">
    <h1>大模型聚合分销与比价</h1>
    <p>覆盖 OpenAI / Anthropic / 百度千帆 / 阿里通义 等主流厂商，一站对比价格与性能</p>
  </div>

  <div class="page">
    <el-card shadow="never" style="margin-bottom: 16px;">
      <div style="display:flex; gap:24px; flex-wrap:wrap; align-items:center;">
        <div>
          <div class="muted">已接入厂商</div>
          <div style="font-size:24px; font-weight:600;">{{ vendors.length }}</div>
        </div>
        <div>
          <div class="muted">模型总数</div>
          <div style="font-size:24px; font-weight:600;">{{ rows.length }}</div>
        </div>
        <el-divider direction="vertical" style="height:48px;" />
        <div style="flex:1; min-width:200px;">
          <div class="muted" style="margin-bottom:6px;">厂商</div>
          <div>
            <el-tag v-for="v in vendors" :key="v.id" style="margin-right:8px; margin-bottom:4px;">
              <img v-if="v.logo" :src="v.logo" class="vendor-logo" />{{ v.name }}
            </el-tag>
          </div>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div style="display:flex; justify-content:space-between; align-items:center;">
          <span style="font-weight:600;">价格对比表 ($ / 1M tokens)</span>
          <div style="display:flex; gap:12px;">
            <el-input v-model="search" placeholder="搜索模型/厂商" style="width:240px;" clearable />
            <el-select v-model="sortBy" style="width:160px;">
              <el-option label="按输入价升序" value="input_price" />
              <el-option label="按输出价升序" value="output_price" />
              <el-option label="按上下文降序" value="context_length" />
            </el-select>
            <el-button @click="load" :loading="loading">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table :data="filtered" v-loading="loading" stripe style="width:100%;">
        <el-table-column label="厂商" min-width="160">
          <template #default="{ row }">
            <img v-if="row.vendor_logo_url" :src="row.vendor_logo_url" class="vendor-logo" />
            <span>{{ row.vendor_name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="model_name" label="模型" min-width="180" />
        <el-table-column label="上下文" min-width="100" align="right">
          <template #default="{ row }">
            <span class="price-cell">{{ row.context_length ? (row.context_length / 1000).toFixed(0) + 'K' : '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="输入价格" min-width="110" align="right">
          <template #default="{ row }">
            <span class="price-cell">{{ fmt(row.input_price) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="输出价格" min-width="110" align="right">
          <template #default="{ row }">
            <span class="price-cell">{{ fmt(row.output_price) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="ELO" min-width="80" align="right">
          <template #default="{ row }">{{ row.elo_score ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="延迟(ms)" min-width="90" align="right">
          <template #default="{ row }">{{ row.latency_ms ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" :disabled="!row.distribution_url"
              tag="a" :href="row.distribution_url" target="_blank">立即使用</el-button>
            <el-button size="small" @click="copyLink(row)" :disabled="!row.distribution_url">复制推广</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
