<template>
  <div class="survey-notify-page">
    <div class="page-header">
      <h3>站内通知</h3>
      <el-button size="small" @click="markAllRead" :disabled="unreadCount===0">全部标记已读</el-button>
    </div>
    <el-table :data="list" v-loading="loading" stripe style="width:100%">
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column prop="content" label="内容" min-width="400">
        <template #default="{ row }">
          <div style="white-space:pre-wrap;line-height:1.6">{{ row.content }}</div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.isRead" type="info" size="small">已读</el-tag>
          <el-tag v-else type="danger" size="small">未读</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ formatTime(row.addTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center">
        <template #default="{ row }">
          <el-button v-if="!row.isRead" text size="small" @click="markRead(row.id)">标为已读</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="page-footer" v-if="total>pageSize">
      <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev,pager,next" @current-change="load" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '../../api'

const list = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const unreadCount = ref(0)

async function load() {
  loading.value = true
  try {
    const r: any = await adminApi.surveyNotifyList({ page: page.value, pageSize: pageSize.value })
    list.value = r.data?.list || []
    total.value = r.data?.total || 0
  } catch { list.value = [] }
  finally { loading.value = false }
}

async function loadUnreadCount() {
  try {
    const r: any = await adminApi.surveyNotifyUnreadCount()
    unreadCount.value = r.data?.count || 0
  } catch {}
}

async function markRead(id: number) {
  await adminApi.surveyNotifyRead({ id })
  await load()
  await loadUnreadCount()
}

async function markAllRead() {
  await adminApi.surveyNotifyRead({ all: true })
  await load()
  await loadUnreadCount()
}

function formatTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { hour12: false })
}

onMounted(() => { load(); loadUnreadCount() })
</script>

<style scoped>
.survey-notify-page { padding: 20px; }
.page-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:16px; }
.page-header h3 { margin:0; font-size:16px; }
.page-footer { margin-top:16px; display:flex; justify-content:center; }
</style>
