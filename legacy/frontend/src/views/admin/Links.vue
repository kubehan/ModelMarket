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
    const { data } = await api.get('/admin/links')
    rows.value = data
  } finally { loading.value = false }
}
onMounted(load)

function edit(row) {
  editing.value = { ...row }
  dialog.value = true
}

async function save() {
  await api.put(`/admin/links/${editing.value.id}`, {
    custom_url: editing.value.custom_url,
    is_active: editing.value.is_active
  })
  ElMessage.success('已保存')
  dialog.value = false
  await load()
}

async function remove(row) {
  await ElMessageBox.confirm(`删除分销链接 ${row.link_code}？`, '提示', { type: 'warning' })
  await api.delete(`/admin/links/${row.id}`)
  ElMessage.success('已删除')
  await load()
}

function fullURL(code) { return location.origin + '/r/' + code }
function copyLink(code) {
  navigator.clipboard.writeText(fullURL(code))
  ElMessage.success('已复制：' + fullURL(code))
}
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <span style="font-weight:600;">分销推广链接</span>
    </template>

    <el-table :data="rows" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="厂商 / 模型" min-width="220">
        <template #default="{ row }">
          <strong>{{ row.Model?.Vendor?.name || row.model?.vendor?.name || '-' }}</strong>
          <span class="muted"> / </span>
          {{ row.Model?.model_name || row.model?.model_name }}
        </template>
      </el-table-column>
      <el-table-column label="link_code" min-width="160">
        <template #default="{ row }">
          <el-tag>{{ row.link_code }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="完整 URL" min-width="260">
        <template #default="{ row }">
          <code style="font-size:12px;">{{ fullURL(row.link_code) }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="custom_url" label="自定义跳转 URL" min-width="240" show-overflow-tooltip />
      <el-table-column prop="clicks" label="点击" width="80" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
            {{ row.is_active ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="copyLink(row.link_code)">复制</el-button>
          <el-button size="small" @click="edit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="dialog" title="编辑推广链接" width="560">
    <el-form v-if="editing" label-position="top">
      <el-form-item label="link_code"><el-input :model-value="editing.link_code" disabled /></el-form-item>
      <el-form-item label="自定义跳转 URL（留空则按厂商推广配置生成）">
        <el-input v-model="editing.custom_url" placeholder="https://x.com/?ref={code}" />
        <div class="muted" style="margin-top:4px;">可使用 {code} 占位符</div>
      </el-form-item>
      <el-form-item><el-switch v-model="editing.is_active" active-text="启用" inactive-text="停用" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialog = false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>
