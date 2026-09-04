<template>
  <section class="forbidden-page">
    <el-result icon="warning" title="无权访问" :sub-title="subtitle">
      <template #extra>
        <el-button v-if="loadFailed" :loading="retrying" @click="retry">重新加载权限</el-button>
        <el-button type="primary" @click="router.replace('/')">返回可用首页</el-button>
      </template>
    </el-result>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { loadAdminAccessSnapshot } from '../../router/adminAccess'

const route = useRoute()
const router = useRouter()
const retrying = ref(false)
const loadFailed = computed(() => route.query.reason === 'load')
const subtitle = computed(() => loadFailed.value ? '权限信息加载失败，请检查网络后重试' : '当前账号没有该页面的菜单权限')

async function retry() {
  retrying.value = true
  try {
    await loadAdminAccessSnapshot(true)
    await router.replace('/')
  } finally {
    retrying.value = false
  }
}
</script>

<style scoped>
.forbidden-page {
  min-height: calc(100vh - 150px);
  display: grid;
  place-items: center;
}
</style>
