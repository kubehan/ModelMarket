<script setup>
import { RouterView, useRoute } from 'vue-router'
import { computed } from 'vue'
const route = useRoute()
const isAdmin = computed(() => route.path.startsWith('/admin'))
</script>

<template>
  <el-container style="min-height: 100vh">
    <el-header v-if="!isAdmin" style="background:#1f1f2e; display:flex; align-items:center; justify-content:space-between;">
      <router-link to="/" class="brand" style="text-decoration:none;">ModelMarket</router-link>
      <div>
        <router-link to="/" style="color:#fff; margin-right:24px; text-decoration:none;">价格对比</router-link>
        <router-link to="/admin" style="color:#fff; text-decoration:none;">管理后台</router-link>
      </div>
    </el-header>
    <el-main style="padding:0;">
      <router-view />
    </el-main>
  </el-container>
</template>
