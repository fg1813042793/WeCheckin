<template>
  <div class="formkit-report">
    <div class="rep-header">
      <h2>答题报表 — {{ title }}</h2>
      <div class="rep-actions">
        <el-button @click="loadReport" :loading="loading">刷新</el-button>
        <el-button type="primary" @click="downloadCsv">导出 CSV</el-button>
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else>
      <div class="rep-summary">
        <el-tag size="large" type="info">总答题数: {{ count }}</el-tag>
        <el-tag size="large">字段数: {{ stats.length }}</el-tag>
      </div>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="数据表" name="table">
          <el-table :data="tableRows" border stripe style="width: 100%" max-height="500">
            <el-table-column prop="userId" label="用户ID" width="180" />
            <el-table-column prop="addTime" label="提交时间" width="180" />
            <el-table-column
              v-for="(h, i) in dynamicHeaders"
              :key="i"
              :prop="`v${i}`"
              :label="h"
              min-width="120"
            />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="字段统计" name="stats">
          <div v-for="s in stats" :key="s.questionId" class="stat-block">
            <h4>
              {{ s.title || s.questionId }}
              <el-tag size="small" :type="s.type === 'number' ? 'success' : 'info'">{{ s.type }}</el-tag>
            </h4>
            <div class="stat-row">
              <span>答题人数: {{ s.nonEmpty }} / {{ s.totalCount }}</span>
              <span v-if="s.empty > 0">未填: {{ s.empty }}</span>
            </div>
            <div v-if="s.numericStat" class="stat-row">
              <span>求和: {{ s.numericStat.sum.toFixed(2) }}</span>
              <span>平均: {{ s.numericStat.avg.toFixed(2) }}</span>
              <span>最小: {{ s.numericStat.min }}</span>
              <span>最大: {{ s.numericStat.max }}</span>
            </div>
            <div v-if="s.dist" class="stat-dist">
              <div v-for="(c, k) in s.dist" :key="k" class="dist-bar">
                <span class="dist-label">{{ k }}</span>
                <el-progress :percentage="distPct(c, s.nonEmpty)" :show-text="false" />
                <span class="dist-count">{{ c }}</span>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'

const route = useRoute()
const type = computed(() => (route.query.type as string) || 'enroll')
const id = computed(() => (route.query.id as string) || '')
const title = ref('')
const count = ref(0)
const tableRows = ref<any[]>([])
const stats = ref<any[]>([])
const loading = ref(false)
const activeTab = ref('table')

const dynamicHeaders = computed(() => {
  if (tableRows.value.length === 0) return []
  const first = tableRows.value[0]
  return first ? first._headers || [] : []
})

const distPct = (c: number, total: number) => {
  if (total === 0) return 0
  return Math.round((c / total) * 100)
}

const loadReport = async () => {
  if (!id.value) {
    ElMessage.error('缺少 id 参数')
    return
  }
  loading.value = true
  try {
    const api = type.value === 'event' ? adminApi.formkitReportEvent : adminApi.formkitReportEnroll
    const res = await api(id.value)
    const data = res.data
    title.value = data.title || '报表'
    count.value = data.count || 0
    stats.value = data.stats || []
    const headers = (data.table && data.table.headers) || []
    const rows = (data.table && data.table.rows) || []
    tableRows.value = rows.map((r: any) => {
      const obj: any = { userId: r.userId, addTime: r.addTime, _headers: headers.slice(2) }
      for (let i = 0; i < r.values.length; i++) {
        obj[`v${i}`] = r.values[i]
      }
      return obj
    })
  } catch (e: any) {
    ElMessage.error(e?.msg || '加载失败')
  } finally {
    loading.value = false
  }
}

const downloadCsv = () => {
  if (!id.value) return
  const baseUrl = (import.meta as any).env?.VITE_API_BASE || ''
  const url = baseUrl + `/admin/survey/export/${type.value}?${type.value}Id=${id.value}`
  const token = localStorage.getItem('admin_token') || ''
  fetch(url, { headers: { Authorization: token } })
    .then((r) => r.blob())
    .then((b) => {
      const a = document.createElement('a')
      a.href = URL.createObjectURL(b)
      a.download = `${type.value}_${id.value}.csv`
      a.click()
    })
    .catch(() => ElMessage.error('下载失败'))
}

onMounted(loadReport)
watch(() => route.query, loadReport)
</script>

<style scoped>
.formkit-report { padding: 20px; }
.rep-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.rep-header h2 { margin: 0; font-size: 18px; }
.rep-actions { display: flex; gap: 8px; }
.rep-summary { display: flex; gap: 12px; margin-bottom: 16px; }
.loading { text-align: center; padding: 40px; color: #909399; }
.stat-block { background: #fff; border: 1px solid #e4e7ed; border-radius: 8px; padding: 16px; margin-bottom: 12px; }
.stat-block h4 { margin: 0 0 12px 0; display: flex; align-items: center; gap: 8px; }
.stat-row { display: flex; gap: 24px; font-size: 13px; color: #606266; margin-bottom: 8px; }
.stat-dist { display: flex; flex-direction: column; gap: 8px; }
.dist-bar { display: grid; grid-template-columns: 120px 1fr 40px; gap: 12px; align-items: center; font-size: 12px; }
.dist-label { color: #303133; }
.dist-count { color: #909399; text-align: right; }
</style>
