<script setup>
import { ref, onMounted } from 'vue'
import api from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const rows = ref([])
const loading = ref(false)
const editing = ref(null)
const dialog = ref(false)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/admin/models')
    rows.value = data
  } finally { loading.value = false }
}
onMounted(load)

function edit(row) {
  editing.value = { ...row }
  dialog.value = true
}

async function save() {
  await api.put(`/admin/models/${editing.value.id}`, {
    display_name: editing.value.display_name,
    context_length: Number(editing.value.context_length) || 0,
    input_price: Number(editing.value.input_price) || 0,
    output_price: Number(editing.value.output_price) || 0,
    elo_score: editing.value.elo_score == null ? null : Number(editing.value.elo_score),
    latency_ms: editing.value.latency_ms == null ? null : Number(editing.value.latency_ms),
    is_active: editing.value.is_active
  })
  ElMessage.success('已保存')
  dialog.value = false
  await load()
}

async function remove(row) {
  await ElMessageBox.confirm(`删除模型 ${row.model_name}？`, '提示', { type: 'warning' })
  await api.delete(`/admin/models/${row.id}`)
  ElMessage.success('已删除')
  await load()
}

async function refreshCache() {
  await api.post('/admin/models/refresh')
  ElMessage.success('已强制刷新首页缓存')
}
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div style="display:flex; justify-content:space-between; align-items:center;">
        <span style="font-weight:600;">模型管理</span>
        <el-button @click="refreshCache">强制刷新首页缓存</el-button>
      </div>
    </template>

    <el-table :data="rows" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="厂商" min-width="140">
        <template #default="{ row }">{{ row.Vendor?.name || row.vendor?.name || '-' }}</template>
      </el-table-column>
      <el-table-column prop="model_name" label="模型名" min-width="160" />
      <el-table-column prop="display_name" label="显示名" min-width="140" />
      <el-table-column prop="context_length" label="上下文" width="100" />
      <el-table-column label="输入价格" width="110">
        <template #default="{ row }">${{ Number(row.input_price).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="输出价格" width="110">
        <template #default="{ row }">${{ Number(row.output_price).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column prop="elo_score" label="ELO" width="80" />
      <el-table-column prop="latency_ms" label="延迟" width="80" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
            {{ row.is_active ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="edit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialog" title="编辑模型" width="560">
    <el-form v-if="editing" label-position="top">
      <el-form-item label="显示名"><el-input v-model="editing.display_name" /></el-form-item>
      <el-row :gutter="12">
        <el-col :span="12"><el-form-item label="上下文长度"><el-input-number v-model="editing.context_length" :min="0" style="width:100%;" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="输入价格 ($/1M)"><el-input-number v-model="editing.input_price" :step="0.01" :min="0" style="width:100%;" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="输出价格 ($/1M)"><el-input-number v-model="editing.output_price" :step="0.01" :min="0" style="width:100%;" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="ELO 评分"><el-input-number v-model="editing.elo_score" style="width:100%;" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="延迟 (ms)"><el-input-number v-model="editing.latency_ms" :min="0" style="width:100%;" /></el-form-item></el-col>
      </el-row>
      <el-form-item><el-switch v-model="editing.is_active" active-text="启用" inactive-text="停用" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>
