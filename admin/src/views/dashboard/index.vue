<template>
  <div>
    <el-card>
      <template #header>控制台</template>
      <el-row :gutter="20">
        <el-col :span="6" v-for="(item, i) in stats" :key="i">
          <el-card shadow="hover">
            <div class="stat-num">{{ item.value }}</div>
            <div class="stat-label">{{ item.label }}</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
    <el-card style="margin-top:16px">
      <template #header>赛事活动</template>
      <el-row :gutter="20">
        <el-col :span="6" v-for="(item, i) in eventStats" :key="i">
          <el-card shadow="hover">
            <div class="stat-num">{{ item.value }}</div>
            <div class="stat-label">{{ item.label }}</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { adminApi } from '../../api'

const stats = ref([
  { label: '用户数', value: '-' },
  { label: '打卡项目', value: '-' },
  { label: '内容', value: '-' },
  { label: '管理员', value: '-' }
])
const eventStats = ref([
  { label: '赛事活动', value: '-' },
  { label: '参与人次', value: '-' }
])

onMounted(async () => {
  try {
    const res = await adminApi.home()
    const d = res.data || {}
    stats.value = [
      { label: '用户数', value: d.userCnt ?? '-' },
      { label: '打卡项目', value: d.enrollCnt ?? '-' },
      { label: '内容', value: d.newsCnt ?? '-' },
      { label: '管理员', value: d.mgrCnt ?? '-' }
    ]
    eventStats.value = [
      { label: '赛事活动', value: d.eventCnt ?? '-' },
      { label: '参与人次', value: d.eventUserCnt ?? '-' }
    ]
  } catch {}
})
</script>

<style scoped>
.stat-num { font-size: 36px; font-weight: bold; color: #409eff; text-align: center; }
.stat-label { text-align: center; color: #999; margin-top: 8px; font-size: 14px; }
</style>
