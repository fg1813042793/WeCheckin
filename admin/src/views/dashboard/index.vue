<template>
  <div class="admin-page dashboard-page">
    <section class="dashboard-hero">
      <div>
        <p class="hero-kicker">WeCheckin 管理后台</p>
        <h2>今天从这里开始管理活动、用户和内容</h2>
        <p class="hero-desc">集中查看核心数据，快速进入高频模块，减少在菜单里来回寻找的时间。</p>
      </div>
      <div class="quick-actions">
        <el-button v-for="item in quickActions" :key="item.path" type="primary" plain @click="goRoute(item.path)">
          <el-icon><component :is="item.icon" /></el-icon>
          {{ item.label }}
        </el-button>
      </div>
    </section>

    <section class="metric-grid">
      <el-card v-for="item in metrics" :key="item.label" class="metric-card admin-card" shadow="never">
        <div class="metric-icon" :class="item.tone">
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <div>
          <div class="metric-value">{{ item.value }}</div>
          <div class="metric-label">{{ item.label }}</div>
        </div>
      </el-card>
    </section>

    <section class="dashboard-columns">
      <el-card class="admin-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>待办关注</span>
            <el-tag size="small" type="info" effect="plain">运营入口</el-tag>
          </div>
        </template>
        <div class="pending-list">
          <button v-for="item in pendingItems" :key="item.label" class="pending-item" type="button" @click="goRoute(item.path)">
            <span class="pending-dot" :class="item.tone" />
            <span>
              <strong>{{ item.label }}</strong>
              <em>{{ item.desc }}</em>
            </span>
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>
      </el-card>

      <el-card class="admin-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>模块状态</span>
            <el-tag size="small" type="success" effect="plain">可用</el-tag>
          </div>
        </template>
        <div class="module-health">
          <div v-for="item in moduleHealth" :key="item.label" class="module-health__row">
            <div>
              <strong>{{ item.label }}</strong>
              <span>{{ item.desc }}</span>
            </div>
            <el-tag size="small" :type="item.type" effect="light">{{ item.status }}</el-tag>
          </div>
        </div>
      </el-card>
    </section>
  </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminApi } from '../../api'

const router = useRouter()
const overview = ref<Record<string, any>>({})

const metricDefs = [
  { label: '用户数', key: 'userCnt', icon: 'User', tone: 'blue' },
  { label: '打卡项目', key: 'enrollCnt', icon: 'List', tone: 'green' },
  { label: '内容', key: 'newsCnt', icon: 'Document', tone: 'orange' },
  { label: '管理员', key: 'mgrCnt', icon: 'UserFilled', tone: 'purple' },
  { label: '赛事活动', key: 'eventCnt', icon: 'TrophyBase', tone: 'red' },
  { label: '参与人次', key: 'eventUserCnt', icon: 'TrendCharts', tone: 'cyan' }
]

const quickActions = [
  { label: '用户管理', path: '/user', icon: 'User' },
  { label: '问卷管理', path: '/survey', icon: 'List' },
  { label: '考试管理', path: '/exam/list', icon: 'EditPen' },
  { label: '系统配置', path: '/setup', icon: 'Setting' }
]

const pendingItems = [
  { label: '用户审核', desc: '查看待审核用户并处理账号状态', path: '/user', tone: 'warning' },
  { label: '问卷回收', desc: '检查答卷数据、统计报表和通知模版', path: '/survey/responses', tone: 'primary' },
  { label: '考试记录', desc: '查看考试提交记录和统计结果', path: '/exam/responses', tone: 'success' },
  { label: '活动维护', desc: '维护赛事活动和参与数据', path: '/event', tone: 'danger' }
]

const moduleHealth = [
  { label: '权限体系', desc: '角色、菜单、管理员入口已配置', status: '正常', type: 'success' },
  { label: '内容管理', desc: '新闻内容和静态资源配置可维护', status: '正常', type: 'success' },
  { label: '问卷考试', desc: '设计、填写、统计入口已接入', status: '活跃', type: 'primary' },
  { label: '系统设置', desc: '登录配置、首页配置、表单字段集中管理', status: '需谨慎', type: 'warning' }
]

const metrics = computed(() => metricDefs.map(item => ({
  ...item,
  value: overview.value[item.key] ?? '-'
})))

function goRoute(path: string) {
  router.push(path)
}

onMounted(async () => {
  try {
    const res = await adminApi.home()
    overview.value = res.data || {}
  } catch {
    overview.value = {}
  }
})
</script>

<style scoped>
.dashboard-page {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
  box-sizing: border-box;
}

.dashboard-page .admin-card {
  border: 1px solid var(--admin-border, #e5e7eb);
  border-radius: var(--admin-card-radius, 8px);
  box-shadow: none;
}

.dashboard-hero {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: center;
  padding: 24px;
  border-radius: 8px;
  background: linear-gradient(135deg, #ffffff 0%, #eef6ff 100%);
  border: 1px solid #dbeafe;
  box-sizing: border-box;
}

.hero-kicker {
  margin: 0 0 8px;
  color: #2563eb;
  font-size: 13px;
  font-weight: 600;
}

.dashboard-hero h2 {
  margin: 0;
  color: #111827;
  font-size: 24px;
  line-height: 1.35;
}

.hero-desc {
  margin: 8px 0 0;
  color: #6b7280;
  font-size: 14px;
}

.quick-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.quick-actions :deep(.el-button) {
  margin: 0;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.metric-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 76px;
  box-sizing: border-box;
}

.metric-icon {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.metric-icon.blue { color: #2563eb; background: #eff6ff; }
.metric-icon.green { color: #059669; background: #ecfdf5; }
.metric-icon.orange { color: #d97706; background: #fff7ed; }
.metric-icon.purple { color: #7c3aed; background: #f5f3ff; }
.metric-icon.red { color: #dc2626; background: #fef2f2; }
.metric-icon.cyan { color: #0891b2; background: #ecfeff; }

.metric-value {
  color: #111827;
  font-size: 28px;
  line-height: 1;
  font-weight: 700;
}

.metric-label {
  margin-top: 6px;
  color: #6b7280;
  font-size: 13px;
}

.dashboard-columns {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(320px, 0.9fr);
  gap: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.pending-list,
.module-health {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pending-item {
  width: 100%;
  border: 1px solid #eef2f7;
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  text-align: left;
  cursor: pointer;
  color: inherit;
  box-sizing: border-box;
}

.pending-item:hover {
  border-color: #bfdbfe;
  background: #f8fbff;
}

.pending-item strong,
.module-health__row strong {
  display: block;
  color: #111827;
  font-size: 14px;
}

.pending-item em,
.module-health__row span {
  display: block;
  margin-top: 3px;
  color: #6b7280;
  font-size: 12px;
  font-style: normal;
}

.pending-item .el-icon {
  margin-left: auto;
  color: #9ca3af;
}

.pending-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.pending-dot.warning { background: #f59e0b; }
.pending-dot.primary { background: #3b82f6; }
.pending-dot.success { background: #10b981; }
.pending-dot.danger { background: #ef4444; }

.module-health__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f1f5f9;
}

.module-health__row:last-child {
  border-bottom: 0;
}

@media (max-width: 920px) {
  .dashboard-hero,
  .dashboard-columns {
    grid-template-columns: 1fr;
  }

  .dashboard-hero {
    align-items: flex-start;
    flex-direction: column;
  }

  .quick-actions {
    justify-content: flex-start;
  }
}
</style>
