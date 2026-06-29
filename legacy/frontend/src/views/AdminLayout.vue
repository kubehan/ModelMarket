<script setup>
import { useRouter, useRoute } from 'vue-router'
import { computed } from 'vue'

const router = useRouter()
const route = useRoute()
const user = computed(() => localStorage.getItem('mm_user') || 'admin')
const active = computed(() => route.path)

function logout() {
  localStorage.removeItem('mm_token')
  localStorage.removeItem('mm_user')
  router.push('/admin/login')
}
</script>

<template>
  <el-container style="height: 100vh;">
    <el-aside width="220px" style="background:#1f1f2e; color:#fff;">
      <div style="padding:18px; font-weight:700; font-size:18px;">ModelMarket Admin</div>
      <el-menu :default-active="active" router background-color="#1f1f2e" text-color="#cdd0e0" active-text-color="#fff" style="border:none;">
        <el-menu-item index="/admin/vendors">厂商管理</el-menu-item>
        <el-menu-item index="/admin/models">模型管理</el-menu-item>
        <el-menu-item index="/admin/links">分销链接</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background:#fff; border-bottom:1px solid #eee; display:flex; align-items:center; justify-content:space-between;">
        <div>
          <router-link to="/" style="margin-right:16px; color:#1f6feb; text-decoration:none;">← 返回首页</router-link>
        </div>
        <div>
          <el-dropdown @command="(c) => c === 'logout' && logout()">
            <span>{{ user }} <el-icon><i-arrow-down /></el-icon></span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
