<template>
  <div class="workflow-start-config">
    <div class="workflow-start-config__content">
      <section class="config-section">
        <div class="config-section__heading">
          <div>
            <span class="config-section__index">01</span>
            <h3>允许发起范围</h3>
          </div>
          <el-tag size="small" type="info">{{ scope === 'all' ? '全部用户' : '指定范围' }}</el-tag>
        </div>

        <el-form label-position="top" @submit.prevent>
          <el-form-item label="发起范围" required>
            <el-radio-group :model-value="scope" class="scope-mode" :disabled="readonly" @change="updateScope">
              <el-radio-button value="all">全部用户</el-radio-button>
              <el-radio-button value="specified">指定范围</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <div v-if="scope === 'specified'" class="config-grid">
            <el-form-item label="允许发起部门">
              <el-tree-select
                :model-value="departmentIds"
                :data="departments"
                :props="{ label: 'name', children: 'children' }"
                node-key="id"
                multiple
                check-strictly
                clearable
                collapse-tags
                collapse-tags-tooltip
                :render-after-expand="false"
                :disabled="readonly"
                placeholder="请选择允许发起该流程的部门"
                style="width: 100%"
                @update:model-value="updateDepartmentIds"
              />
            </el-form-item>
            <el-form-item label="额外允许用户">
              <WorkflowUserTreePicker
                :model-value="userIds"
                :departments="departments"
                :users="users"
                multiple
                :disabled="readonly"
                placeholder="请选择允许发起该流程的用户"
                @update:model-value="updateUserIds"
              />
            </el-form-item>
          </div>
        </el-form>
      </section>

      <section class="config-section">
        <div class="config-section__heading">
          <div>
            <span class="config-section__index">02</span>
            <h3>允许发起时间</h3>
          </div>
          <el-tag size="small" type="info">Asia/Shanghai</el-tag>
        </div>

        <el-form label-position="top" @submit.prevent>
          <el-form-item label="开放方式" required>
            <el-radio-group :model-value="availabilityMode" class="availability-mode" :disabled="readonly" @change="updateAvailabilityMode">
              <el-radio-button value="always">长期有效</el-radio-button>
              <el-radio-button value="fixed">指定时间段</el-radio-button>
              <el-radio-button value="weekly">每周周期开放</el-radio-button>
              <el-radio-button value="monthly">每月周期开放</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item v-if="availabilityMode === 'fixed'" label="允许发起时间段" required>
            <el-date-picker
              v-model="fixedRange"
              type="datetimerange"
              value-format="x"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              :disabled="readonly"
              style="width: 100%"
            />
          </el-form-item>

          <template v-else-if="availabilityMode === 'weekly' || availabilityMode === 'monthly'">
            <el-form-item v-if="availabilityMode === 'weekly'" label="每周开放日" required>
              <el-checkbox-group :model-value="weekdays" class="weekday-options" :disabled="readonly" @change="updateWeekdays">
                <el-checkbox-button v-for="day in weekdayOptions" :key="day.value" :value="day.value">{{ day.label }}</el-checkbox-button>
              </el-checkbox-group>
            </el-form-item>

            <el-form-item v-else label="每月开放日" required>
              <div class="month-day-row">
                <el-select
                  :model-value="monthDays"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  clearable
                  :disabled="readonly"
                  placeholder="请选择日期"
                  @update:model-value="updateMonthDays"
                >
                  <el-option v-for="day in monthDayOptions" :key="day" :label="`${day}日`" :value="day" />
                </el-select>
                <el-checkbox :model-value="lastDayOfMonth" :disabled="readonly" @change="updateLastDayOfMonth">最后一天</el-checkbox>
              </div>
            </el-form-item>

            <div class="config-grid config-grid--schedule">
              <el-form-item label="每日开放时段" required>
                <el-time-picker
                  v-model="dailyTimeRange"
                  is-range
                  value-format="HH:mm"
                  format="HH:mm"
                  range-separator="至"
                  start-placeholder="开始时间"
                  end-placeholder="结束时间"
                  :disabled="readonly"
                  style="width: 100%"
                />
              </el-form-item>
              <el-form-item label="周期生效日期">
                <el-date-picker
                  v-model="effectiveDateRange"
                  type="daterange"
                  value-format="YYYY-MM-DD"
                  range-separator="至"
                  start-placeholder="不限制开始日期"
                  end-placeholder="不限制结束日期"
                  clearable
                  :disabled="readonly"
                  style="width: 100%"
                />
              </el-form-item>
            </div>
          </template>
        </el-form>
      </section>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import WorkflowUserTreePicker from '../../components/WorkflowUserTreePicker.vue'
import type {
  WorkflowAssigneeUser,
  WorkflowDraft,
  WorkflowInitiatorScope,
  WorkflowStartAvailabilityConfig,
  WorkflowStartAvailabilityMode,
} from '../../types'

type DepartmentNode = {
  id: number
  name: string
  children?: DepartmentNode[]
}

const props = defineProps<{
  draft: WorkflowDraft
  departments: DepartmentNode[]
  users: WorkflowAssigneeUser[]
  readonly?: boolean
}>()

const emit = defineEmits<{ change: [] }>()
const timezone = 'Asia/Shanghai'
const weekdayOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 7 },
]
const monthDayOptions = Array.from({ length: 31 }, (_, index) => index + 1)

const startNode = computed(() => props.draft.nodes.find(node => node.type === 'start'))
const scope = computed<WorkflowInitiatorScope>(() => startNode.value?.initiator?.scope === 'specified' ? 'specified' : 'all')
const userIds = computed(() => normalizeIDs(startNode.value?.initiator?.userIds || []))
const departmentIds = computed(() => normalizeIDs(startNode.value?.initiator?.departmentIds || []))
const availabilityMode = computed<WorkflowStartAvailabilityMode>(() => {
  const mode = startNode.value?.availability?.mode
  return mode === 'fixed' || mode === 'weekly' || mode === 'monthly' ? mode : 'always'
})
const weekdays = computed(() => normalizeDays(startNode.value?.availability?.weekdays || [], 7))
const monthDays = computed(() => normalizeDays(startNode.value?.availability?.monthDays || [], 31))
const lastDayOfMonth = computed(() => Boolean(startNode.value?.availability?.lastDayOfMonth))

const fixedRange = computed<[number, number] | null>({
  get() {
    const config = startNode.value?.availability
    if (!config?.startsAt || !config?.endsAt) return null
    return [Number(config.startsAt), Number(config.endsAt)]
  },
  set(value) {
    if (!value || value.length !== 2) {
      updateAvailability({ ...baseAvailability('fixed'), startsAt: undefined, endsAt: undefined })
      return
    }
    updateAvailability({ ...baseAvailability('fixed'), startsAt: Number(value[0]), endsAt: Number(value[1]) })
  },
})

const dailyTimeRange = computed<[string, string] | null>({
  get() {
    const config = startNode.value?.availability
    if (!config?.dailyStartTime || !config?.dailyEndTime) return null
    return [config.dailyStartTime, config.dailyEndTime]
  },
  set(value) {
    updateAvailability({
      ...baseAvailability(availabilityMode.value),
      dailyStartTime: value?.[0] || '',
      dailyEndTime: value?.[1] || '',
    })
  },
})

const effectiveDateRange = computed<[string, string] | null>({
  get() {
    const config = startNode.value?.availability
    if (!config?.effectiveStartDate || !config?.effectiveEndDate) return null
    return [config.effectiveStartDate, config.effectiveEndDate]
  },
  set(value) {
    updateAvailability({
      ...baseAvailability(availabilityMode.value),
      effectiveStartDate: value?.[0] || undefined,
      effectiveEndDate: value?.[1] || undefined,
    })
  },
})

function updateScope(value: string | number | boolean | undefined) {
  const node = startNode.value
  if (!node || props.readonly) return
  node.initiator = value === 'specified'
    ? { scope: 'specified', userIds: userIds.value, departmentIds: departmentIds.value }
    : { scope: 'all' }
  emit('change')
}

function updateUserIds(values: number[]) {
  updateInitiatorIDs(normalizeIDs(values), departmentIds.value)
}

function updateDepartmentIds(values: unknown) {
  updateInitiatorIDs(userIds.value, normalizeIDs(Array.isArray(values) ? values : []))
}

function updateInitiatorIDs(nextUserIds: number[], nextDepartmentIds: number[]) {
  const node = startNode.value
  if (!node || props.readonly) return
  node.initiator = { scope: 'specified', userIds: nextUserIds, departmentIds: nextDepartmentIds }
  emit('change')
}

function updateAvailabilityMode(value: string | number | boolean | undefined) {
  if (props.readonly) return
  const mode: WorkflowStartAvailabilityMode = value === 'fixed' || value === 'weekly' || value === 'monthly' ? value : 'always'
  if (mode === 'always') {
    updateAvailability({ mode: 'always', timezone })
    return
  }
  if (mode === 'fixed') {
    updateAvailability({ mode: 'fixed', timezone })
    return
  }
  if (mode === 'weekly') {
    updateAvailability({ mode: 'weekly', timezone, weekdays: [1, 2, 3, 4, 5], dailyStartTime: '09:00', dailyEndTime: '18:00' })
    return
  }
  updateAvailability({ mode: 'monthly', timezone, monthDays: [1], dailyStartTime: '09:00', dailyEndTime: '18:00' })
}

function updateWeekdays(values: unknown) {
  updateAvailability({ ...baseAvailability('weekly'), weekdays: normalizeDays(Array.isArray(values) ? values : [], 7) })
}

function updateMonthDays(values: unknown) {
  updateAvailability({ ...baseAvailability('monthly'), monthDays: normalizeDays(Array.isArray(values) ? values : [], 31) })
}

function updateLastDayOfMonth(value: string | number | boolean) {
  updateAvailability({ ...baseAvailability('monthly'), lastDayOfMonth: Boolean(value) })
}

function baseAvailability(mode: WorkflowStartAvailabilityMode): WorkflowStartAvailabilityConfig {
  const current = startNode.value?.availability
  return {
    mode,
    timezone,
    effectiveStartDate: current?.effectiveStartDate,
    effectiveEndDate: current?.effectiveEndDate,
    weekdays: mode === 'weekly' ? normalizeDays(current?.weekdays || [], 7) : undefined,
    monthDays: mode === 'monthly' ? normalizeDays(current?.monthDays || [], 31) : undefined,
    lastDayOfMonth: mode === 'monthly' ? Boolean(current?.lastDayOfMonth) : undefined,
    dailyStartTime: mode === 'weekly' || mode === 'monthly' ? current?.dailyStartTime : undefined,
    dailyEndTime: mode === 'weekly' || mode === 'monthly' ? current?.dailyEndTime : undefined,
    startsAt: mode === 'fixed' ? current?.startsAt : undefined,
    endsAt: mode === 'fixed' ? current?.endsAt : undefined,
  }
}

function updateAvailability(config: WorkflowStartAvailabilityConfig) {
  const node = startNode.value
  if (!node || props.readonly) return
  node.availability = config
  emit('change')
}

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
}

function normalizeDays(values: unknown[], maximum: number) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value >= 1 && value <= maximum))).sort((left, right) => left - right)
}
</script>

<style scoped>
.workflow-start-config { width: 100%; min-width: 0; overflow: auto; background: #f7f9fc; }
.workflow-start-config__content { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); align-items: start; gap: 20px; width: min(1040px, 100%); margin: 10px; padding: 28px 0 48px; }
.config-section { padding: 24px 28px 28px; border: 1px solid #dfe5ec; border-radius: 8px; background: #fff; box-shadow: 0 2px 10px rgb(15 23 42 / 5%); }
.config-section__heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin: -2px 0 22px; padding-bottom: 16px; border-bottom: 1px solid #edf1f5; }
.config-section__heading > div { display: flex; align-items: center; gap: 10px; }
.config-section__heading h3 { margin: 0; color: #1f2937; font-size: 17px; font-weight: 650; }
.config-section__index { display: grid; width: 30px; height: 24px; place-items: center; border-radius: 4px; color: #0f766e; background: #dff5f1; font-size: 11px; font-weight: 700; }
.scope-mode, .availability-mode { display: flex; width: 100%; }
.scope-mode :deep(.el-radio-button), .availability-mode :deep(.el-radio-button) { flex: 1; }
.scope-mode :deep(.el-radio-button__inner), .availability-mode :deep(.el-radio-button__inner) { width: 100%; }
.config-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 24px; }
.config-grid--schedule { align-items: start; }
.weekday-options { display: flex; width: 100%; }
.weekday-options :deep(.el-checkbox-button) { flex: 1; }
.weekday-options :deep(.el-checkbox-button__inner) { width: 100%; padding-inline: 8px; }
.month-day-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 18px; width: 100%; }
.month-day-row .el-select { width: 100%; }
.config-section :deep(.el-form-item:last-child) { margin-bottom: 0; }
@media (max-width: 1100px) {
  .workflow-start-config__content { grid-template-columns: minmax(0, 500px); justify-content: start; width: 100%; }
}
@media (max-width: 900px) {
  .workflow-start-config__content { gap: 14px; padding: 16px 0 28px; }
  .config-section { padding: 20px 16px 22px; }
  .config-grid { grid-template-columns: 1fr; gap: 0; }
  .availability-mode { display: grid; grid-template-columns: 1fr 1fr; }
  .availability-mode :deep(.el-radio-button__inner) { border: 1px solid var(--el-border-color); }
}
</style>
