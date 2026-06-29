<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import api from '../../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const vendors = ref([])
const loading = ref(false)
const authSchemas = ref([])     // [{type, label, fields:[...]}]
const promoSchemas = ref([])

const dialogVisible = ref(false)
const editing = ref(null)        // 当前编辑对象
const isNew = ref(true)

async function loadSchemas() {
  const { data } = await api.get('/admin/vendors/schemas')
  authSchemas.value = data.auth_schemas
  promoSchemas.value = data.promo_schemas
}

async function loadVendors() {
  loading.value = true
  try {
    const { data } = await api.get('/admin/vendors')
    vendors.value = data
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadSchemas()
  await loadVendors()
})

// 找到选中类型的 schema
const currentAuthSchema = computed(() =>
  authSchemas.value.find(s => s.type === editing.value?.auth_type) || { fields: [] }
)
const currentPromoSchema = computed(() =>
  promoSchemas.value.find(s => s.type === editing.value?.promo_source_type) || { fields: [] }
)

// 当切换 auth_type / promo_source_type 时，确保字段存在
watch(() => editing.value?.auth_type, (t) => {
  if (!editing.value) return
  if (!editing.value.auth_config) editing.value.auth_config = {}
  const s = authSchemas.value.find(x => x.type === t)
  if (s) s.fields.forEach(f => {
    if (editing.value.auth_config[f.key] === undefined) editing.value.auth_config[f.key] = ''
  })
})
watch(() => editing.value?.promo_source_type, (t) => {
  if (!editing.value) return
  if (!editing.value.promo_config) editing.value.promo_config = {}
  const s = promoSchemas.value.find(x => x.type === t)
  if (s) s.fields.forEach(f => {
    if (editing.value.promo_config[f.key] === undefined) editing.value.promo_config[f.key] = ''
  })
})

function openCreate() {
  isNew.value = true
  editing.value = {
    name: '', official_url: '', api_base: '', logo_url: '', description: '',
    auth_type: 'api_key', auth_config: {},
    promo_source_type: 'manual', promo_config: {},
    is_active: true
  }
  dialogVisible.value = true
}

function openEdit(row) {
  isNew.value = false
  editing.value = JSON.parse(JSON.stringify(row))
  if (!editing.value.auth_config) editing.value.auth_config = {}
  if (!editing.value.promo_config) editing.value.promo_config = {}
  dialogVisible.value = true
}

async function save() {
  try {
    const payload = { ...editing.value }
    if (isNew.value) {
      await api.post('/admin/vendors', payload)
      ElMessage.success('已创建')
    } else {
      await api.put(`/admin/vendors/${payload.id}`, payload)
      ElMessage.success('已保存')
    }
    dialogVisible.value = false
    await loadVendors()
  } catch (e) { /* */ }
}

async function remove(row) {
  await ElMessageBox.confirm(`确定删除厂商 ${row.name}？`, '提示', { type: 'warning' })
  await api.delete(`/admin/vendors/${row.id}`)
  ElMessage.success('已删除')
  await loadVendors()
}

async function testConnection(row) {
  try {
    const { data } = await api.post(`/admin/vendors/${row.id}/test`)
    ElMessageBox.alert(
      `成功！拉取到 ${data.count} 个模型：\n${data.models.map(m => '• ' + m.id).join('\n')}`,
      '测试连接成功',
      { type: 'success' }
    )
    await loadVendors()
  } catch (e) { await loadVendors() }
}

async function syncModels(row) {
  try {
    const { data } = await api.post(`/admin/vendors/${row.id}/sync`)
    ElMessage.success(`同步完成，新增 ${data.newly_added} 个模型`)
    await loadVendors()
  } catch (e) { /* */ }
}
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div style="display:flex; justify-content:space-between; align-items:center;">
        <span style="font-weight:600;">厂商管理</span>
        <el-button type="primary" @click="openCreate">新增厂商</el-button>
      </div>
    </template>

    <el-table :data="vendors" v-loading="loading" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="名称" min-width="160">
        <template #default="{ row }">
          <img v-if="row.logo_url" :src="row.logo_url" class="vendor-logo" />
          <strong>{{ row.name }}</strong>
        </template>
      </el-table-column>
      <el-table-column label="认证方式" width="130">
        <template #default="{ row }">
          <el-tag size="small">{{ row.auth_type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="推广来源" width="120">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.promo_source_type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="API Base" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ row.api_base }}</template>
      </el-table-column>
      <el-table-column label="最后测试" width="150">
        <template #default="{ row }">
          <el-tag v-if="row.last_test_status === 'ok'" size="small" type="success">OK</el-tag>
          <el-tag v-else-if="row.last_test_status === 'failed'" size="small" type="danger">FAIL</el-tag>
          <span v-else class="muted">未测试</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'" size="small">
            {{ row.is_active ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="320" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="testConnection(row)">测试连接</el-button>
          <el-button size="small" type="primary" @click="syncModels(row)">同步模型</el-button>
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <!-- 编辑/新增弹窗：动态表单 -->
  <el-dialog v-model="dialogVisible" :title="isNew ? '新增厂商' : '编辑厂商'" width="720" destroy-on-close>
    <el-form v-if="editing" label-position="top">
      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="厂商名称" required>
            <el-input v-model="editing.name" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="官网 URL">
            <el-input v-model="editing.official_url" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="API Base">
            <el-input v-model="editing.api_base" placeholder="https://api.openai.com" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="Logo URL">
            <el-input v-model="editing.logo_url" />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="描述">
            <el-input v-model="editing.description" type="textarea" :rows="2" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">认证 / 登录方式</el-divider>
      <el-form-item label="认证方式">
        <el-radio-group v-model="editing.auth_type">
          <el-radio-button v-for="s in authSchemas" :key="s.type" :label="s.type">
            {{ s.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-row :gutter="16">
        <el-col v-for="f in currentAuthSchema.fields" :key="f.key" :span="f.type === 'textarea' ? 24 : 12">
          <el-form-item :label="f.label" :required="f.required">
            <el-input
              v-if="f.type === 'textarea'"
              v-model="editing.auth_config[f.key]"
              type="textarea" :rows="3"
              :placeholder="f.hint"
            />
            <el-input
              v-else
              v-model="editing.auth_config[f.key]"
              :type="f.type === 'password' ? 'password' : 'text'"
              :show-password="f.type === 'password'"
              :placeholder="f.hint"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">推广链接获取方式</el-divider>
      <el-form-item label="推广来源">
        <el-radio-group v-model="editing.promo_source_type">
          <el-radio-button v-for="s in promoSchemas" :key="s.type" :label="s.type">
            {{ s.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-row :gutter="16">
        <el-col v-for="f in currentPromoSchema.fields" :key="f.key" :span="24">
          <el-form-item :label="f.label" :required="f.required">
            <el-input
              v-model="editing.promo_config[f.key]"
              :type="f.type === 'password' ? 'password' : 'text'"
              :placeholder="f.hint"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item>
        <el-switch v-model="editing.is_active" active-text="启用" inactive-text="停用" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>
