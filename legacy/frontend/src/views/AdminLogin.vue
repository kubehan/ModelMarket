<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const form = ref({ username: 'admin', password: 'admin123' })
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    const { data } = await api.post('/auth/login', form.value)
    localStorage.setItem('mm_token', data.token)
    localStorage.setItem('mm_user', data.username)
    ElMessage.success('登录成功')
    router.push('/admin/vendors')
  } catch (e) {
    // 已由拦截器报错
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div style="display:flex; align-items:center; justify-content:center; min-height: calc(100vh - 60px);">
    <el-card style="width: 360px;">
      <h2 style="margin-top: 0;">管理后台登录</h2>
      <el-form @keyup.enter="submit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="submit" style="width: 100%;">
          登录
        </el-button>
      </el-form>
      <p class="muted" style="margin-top:12px;">默认账号 admin / admin123（可通过 .env 修改）</p>
    </el-card>
  </div>
</template>
