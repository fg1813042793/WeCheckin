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

          <div :class="['config-grid', { 'config-grid--single': scope === 'all' }]">
            <el-form-item v-if="scope === 'specified'" label="允许发起范围" required>
              <WorkflowUserTreePicker
                :model-value="userIds"
                :department-model-value="departmentIds"
                :departments="departments"
                :users="users"
                multiple
                select-department-rules
                :disabled="readonly"
                placeholder="请选择允许发起的部门或用户"
                @update:model-value="updateUserIds"
                @update:department-model-value="updateDepartmentIds"
              />
            </el-form-item>
            <el-form-item label="排除用户">
              <WorkflowUserTreePicker
                :model-value="excludedUserIds"
                :departments="departments"
                :users="users"
                multiple
                :disabled="readonly"
                placeholder="请选择不允许发起该流程的用户"
                @update:model-value="updateExcludedUserIds"
              />
              <div class="form-help">排除用户优先级最高，在全部用户和指定范围下都会生效。</div>
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
              <el-radio-button class="availability-mode__wide" value="monthly_window">每月连续区间</el-radio-button>
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

          <template v-else-if="availabilityMode === 'weekly' || availabilityMode === 'monthly' || availabilityMode === 'monthly_window'">
            <el-form-item v-if="availabilityMode === 'weekly'" label="每周开放日" required>
              <el-checkbox-group :model-value="weekdays" class="weekday-options" :disabled="readonly" @change="updateWeekdays">
                <el-checkbox-button v-for="day in weekdayOptions" :key="day.value" :value="day.value">{{ day.label }}</el-checkbox-button>
              </el-checkbox-group>
            </el-form-item>

            <el-form-item v-else-if="availabilityMode === 'monthly'" label="每月开放日" required>
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

            <template v-else>
              <div class="config-grid">
                <el-form-item label="开始日期" required>
                  <div class="window-anchor">
                    <span>本月倒数第</span>
                    <el-input-number
                      :model-value="windowStartDayFromEnd"
                      :min="1"
                      :max="28"
                      :step="1"
                      controls-position="right"
                      :disabled="readonly"
                      @change="updateWindowStartDayFromEnd"
                    />
                    <span>天</span>
                  </div>
                </el-form-item>
                <el-form-item label="结束日期" required>
                  <div class="window-anchor">
                    <span>下月第</span>
                    <el-input-number
                      :model-value="windowEndDay"
                      :min="1"
                      :max="31"
                      :step="1"
                      controls-position="right"
                      :disabled="readonly"
                      @change="updateWindowEndDay"
                    />
                    <span>日</span>
                  </div>
                </el-form-item>
              </div>
              <div class="config-grid config-grid--schedule">
                <el-form-item label="开始时间" required>
                  <el-time-picker
                    :model-value="windowStartTime"
                    value-format="HH:mm"
                    format="HH:mm"
                    placeholder="选择开始时间"
                    :disabled="readonly"
                    style="width: 100%"
                    @update:model-value="updateWindowStartTime"
                  />
                </el-form-item>
                <el-form-item label="结束时间" required>
                  <el-time-picker
                    :model-value="windowEndTime"
                    value-format="HH:mm"
                    format="HH:mm"
                    placeholder="选择结束时间"
                    :disabled="readonly"
                    style="width: 100%"
                    @update:model-value="updateWindowEndTime"
                  />
                </el-form-item>
              </div>
              <div class="form-help window-help">结束日超过下月实际天数时，按下月最后一天处理。</div>
            </template>

            <div v-if="availabilityMode !== 'monthly_window'" class="config-grid config-grid--schedule">
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
            <el-form-item v-else label="周期生效日期">
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
          </template>
        </el-form>
      </section>

      <section class="config-section">
        <div class="config-section__heading">
          <div>
            <span class="config-section__index">03</span>
            <h3>发起次数限制</h3>
          </div>
          <el-tag size="small" type="info">{{ startLimitSummary }}</el-tag>
        </div>

        <el-form label-position="top" @submit.prevent>
          <el-form-item label="限制方式" required>
            <el-radio-group :model-value="startLimitMode" class="limit-mode" :disabled="readonly" @change="updateStartLimitMode">
              <el-radio-button value="unlimited">不限次数</el-radio-button>
              <el-radio-button value="limited">限制次数</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <template v-if="startLimitMode === 'limited'">
            <div class="config-grid">
              <el-form-item label="统计周期" required>
                <el-select :model-value="startLimitPeriod" :disabled="readonly" @change="updateStartLimitPeriod">
                  <el-option v-for="option in startLimitPeriodOptions" :key="option.value" :label="option.label" :value="option.value" />
                </el-select>
              </el-form-item>
              <el-form-item label="每人最多发起" required>
                <el-input-number
                  :model-value="startLimitMaxCount"
                  :min="1"
                  :max="10000"
                  :step="1"
                  controls-position="right"
                  :disabled="readonly"
                  @change="updateStartLimitMaxCount"
                />
              </el-form-item>
            </div>
            <div class="form-help">草稿不计次数；提交成功后计入。管理员代发时，次数归属所选员工。</div>
          </template>
        </el-form>
      </section>

      <section class="config-section config-section--identity">
        <div class="config-section__heading">
          <div>
            <span class="config-section__index">04</span>
            <h3>单据标识</h3>
          </div>
          <el-tag size="small" type="info">{{ businessPeriodSummary }}</el-tag>
        </div>

        <el-form label-position="top" @submit.prevent>
          <el-form-item label="单据标题命名规则" required>
            <el-input
              :model-value="titleTemplate"
              :disabled="readonly"
              maxlength="500"
              placeholder="{{starterName}}提交的{{workflowName}}"
              @input="updateTitleTemplate"
            />
            <div class="template-variables">
              <el-button
                v-for="variable in titleVariables"
                :key="variable.value"
                size="small"
                plain
                :disabled="readonly || (variable.value === '{{businessPeriod}}' && !businessPeriodEnabled)"
                @click="appendTitleVariable(variable.value)"
              >
                {{ variable.label }}
              </el-button>
              <el-dropdown v-if="titleFieldVariables.length" trigger="click" :disabled="readonly" @command="appendTitleVariable">
                <el-button size="small" plain :disabled="readonly">表单字段</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-for="field in titleFieldVariables" :key="field.value" :command="field.value">
                      {{ field.label }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </el-form-item>

          <el-form-item label="业务期间">
            <el-switch
              :model-value="businessPeriodEnabled"
              :disabled="readonly"
              active-text="启用"
              inactive-text="不启用"
              @change="updateBusinessPeriodEnabled"
            />
          </el-form-item>

          <template v-if="businessPeriodEnabled">
            <div class="config-grid">
              <el-form-item label="期间粒度" required>
                <el-select :model-value="businessPeriodGranularity" :disabled="readonly" @change="updateBusinessPeriodGranularity">
                  <el-option v-for="option in businessPeriodGranularityOptions" :key="option.value" :label="option.label" :value="option.value" />
                </el-select>
              </el-form-item>
              <el-form-item label="取值时间" required>
                <el-select :model-value="businessPeriodSource" :disabled="readonly" @change="updateBusinessPeriodSource">
                  <el-option v-for="option in businessPeriodSourceOptions" :key="option.value" :label="option.label" :value="option.value" />
                </el-select>
              </el-form-item>
            </div>
            <div class="config-grid">
              <el-form-item v-if="businessPeriodSource === 'form_field'" label="日期字段" required>
                <el-select :model-value="businessPeriodField" :disabled="readonly" placeholder="请选择日期字段" @change="updateBusinessPeriodField">
                  <el-option v-for="field in businessPeriodFieldOptions" :key="field.value" :label="field.label" :value="field.value" />
                </el-select>
              </el-form-item>
              <el-form-item label="期间偏移">
                <el-input-number
                  :model-value="businessPeriodOffset"
                  :min="-120"
                  :max="120"
                  :step="1"
                  controls-position="right"
                  :disabled="readonly"
                  @change="updateBusinessPeriodOffset"
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
  WorkflowBusinessPeriodGranularity,
  WorkflowBusinessPeriodSource,
  WorkflowDraft,
  WorkflowInitiatorScope,
  WorkflowStartAvailabilityConfig,
  WorkflowStartAvailabilityMode,
  WorkflowStartLimitMode,
  WorkflowStartLimitPeriod,
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
const businessPeriodGranularityOptions: Array<{ label: string, value: WorkflowBusinessPeriodGranularity }> = [
  { label: '按日', value: 'day' },
  { label: '按周', value: 'week' },
  { label: '按月', value: 'month' },
  { label: '按季度', value: 'quarter' },
  { label: '按年', value: 'year' },
]
const titleVariables = [
  { label: '流程名称', value: '{{workflowName}}' },
  { label: '发起人', value: '{{starterName}}' },
  { label: '业务期间', value: '{{businessPeriod}}' },
]

const startNode = computed(() => props.draft.nodes.find(node => node.type === 'start'))
const scope = computed<WorkflowInitiatorScope>(() => startNode.value?.initiator?.scope === 'specified' ? 'specified' : 'all')
const userIds = computed(() => normalizeIDs(startNode.value?.initiator?.userIds || []))
const departmentIds = computed(() => normalizeIDs(startNode.value?.initiator?.departmentIds || []))
const excludedUserIds = computed(() => normalizeIDs(startNode.value?.initiator?.excludedUserIds || []))
const availabilityMode = computed<WorkflowStartAvailabilityMode>(() => {
  const mode = startNode.value?.availability?.mode
  return mode === 'fixed' || mode === 'weekly' || mode === 'monthly' || mode === 'monthly_window' ? mode : 'always'
})
const weekdays = computed(() => normalizeDays(startNode.value?.availability?.weekdays || [], 7))
const monthDays = computed(() => normalizeDays(startNode.value?.availability?.monthDays || [], 31))
const lastDayOfMonth = computed(() => Boolean(startNode.value?.availability?.lastDayOfMonth))
const windowStartDayFromEnd = computed(() => normalizeInteger(startNode.value?.availability?.windowStartDayFromEnd, 1, 28, 1))
const windowEndDay = computed(() => normalizeInteger(startNode.value?.availability?.windowEndDay, 1, 31, 5))
const windowStartTime = computed(() => normalizeClock(startNode.value?.availability?.windowStartTime, '09:00'))
const windowEndTime = computed(() => normalizeClock(startNode.value?.availability?.windowEndTime, '18:00'))
const startLimitMode = computed<WorkflowStartLimitMode>(() => startNode.value?.startLimit?.mode === 'limited' ? 'limited' : 'unlimited')
const startLimitPeriod = computed<WorkflowStartLimitPeriod>(() => {
  const period = startNode.value?.startLimit?.period
  if (period === 'total' || period === 'day' || period === 'week' || period === 'month') return period
  if (period === 'availability' && availabilityMode.value !== 'always') return period
  return availabilityMode.value === 'always' ? 'month' : 'availability'
})
const startLimitMaxCount = computed(() => {
  const count = Number(startNode.value?.startLimit?.maxCount)
  return Number.isInteger(count) && count >= 1 && count <= 10000 ? count : 1
})
const startLimitPeriodOptions = computed<Array<{ label: string, value: WorkflowStartLimitPeriod }>>(() => [
  { label: '累计', value: 'total' },
  { label: '每天', value: 'day' },
  { label: '每周', value: 'week' },
  { label: '每月', value: 'month' },
  ...(availabilityMode.value === 'always' ? [] : [{ label: '每个开放期', value: 'availability' as const }]),
])
const startLimitSummary = computed(() => {
  if (startLimitMode.value === 'unlimited') return '不限次数'
  const periodLabel = startLimitPeriodOptions.value.find(option => option.value === startLimitPeriod.value)?.label || '当前周期'
  return `${periodLabel} ${startLimitMaxCount.value} 次`
})
const titleTemplate = computed(() => props.draft.instanceIdentity?.titleTemplate || '{{starterName}}提交的{{workflowName}}')
const businessPeriodEnabled = computed(() => Boolean(props.draft.instanceIdentity?.businessPeriod?.enabled))
const businessPeriodGranularity = computed<WorkflowBusinessPeriodGranularity>(() => {
  const value = props.draft.instanceIdentity?.businessPeriod?.granularity
  return value === 'day' || value === 'week' || value === 'quarter' || value === 'year' ? value : 'month'
})
const businessPeriodSource = computed<WorkflowBusinessPeriodSource>(() => {
  const value = props.draft.instanceIdentity?.businessPeriod?.source
  if (value === 'form_field') return value
  if ((value === 'availability_window_start' || value === 'availability_window_end') && availabilityMode.value !== 'always') return value
  return 'submit_time'
})
const businessPeriodField = computed(() => props.draft.instanceIdentity?.businessPeriod?.field || '')
const businessPeriodOffset = computed(() => normalizeInteger(props.draft.instanceIdentity?.businessPeriod?.offset, -120, 120, 0))
const businessPeriodSummary = computed(() => {
  if (!businessPeriodEnabled.value) return '未设置业务期间'
  return businessPeriodGranularityOptions.find(option => option.value === businessPeriodGranularity.value)?.label || '已启用'
})
const businessPeriodSourceOptions = computed<Array<{ label: string, value: WorkflowBusinessPeriodSource }>>(() => [
  { label: '提交时间', value: 'submit_time' },
  ...(availabilityMode.value === 'always'
    ? []
    : [
        { label: '开放窗口开始时间', value: 'availability_window_start' as const },
        { label: '开放窗口结束时间', value: 'availability_window_end' as const },
      ]),
  { label: '表单日期字段', value: 'form_field' },
])
const flatFormFields = computed(() => flattenFormFields(props.draft.form))
const businessPeriodFieldOptions = computed(() => flatFormFields.value
  .filter(field => field.type === 'date' || field.type === 'datetime')
  .map(field => ({ label: field.label, value: field.key })))
const titleFieldVariables = computed(() => flatFormFields.value
  .filter(field => ['text', 'textarea', 'number', 'amount', 'phone', 'email', 'select', 'radio', 'date', 'datetime', 'time', 'boolean', 'calculation'].includes(field.type))
  .map(field => ({ label: field.label, value: `{{field.${field.key}}}` })))

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
    ? { scope: 'specified', userIds: userIds.value, departmentIds: departmentIds.value, excludedUserIds: excludedUserIds.value }
    : { scope: 'all', excludedUserIds: excludedUserIds.value }
  emit('change')
}

function updateUserIds(values: number[]) {
  updateInitiator(normalizeIDs(values), departmentIds.value, excludedUserIds.value)
}

function updateDepartmentIds(values: unknown) {
  updateInitiator(userIds.value, normalizeIDs(Array.isArray(values) ? values : []), excludedUserIds.value)
}

function updateExcludedUserIds(values: number[]) {
  updateInitiator(userIds.value, departmentIds.value, normalizeIDs(values))
}

function updateInitiator(nextUserIds: number[], nextDepartmentIds: number[], nextExcludedUserIds: number[]) {
  const node = startNode.value
  if (!node || props.readonly) return
  node.initiator = scope.value === 'specified'
    ? { scope: 'specified', userIds: nextUserIds, departmentIds: nextDepartmentIds, excludedUserIds: nextExcludedUserIds }
    : { scope: 'all', excludedUserIds: nextExcludedUserIds }
  emit('change')
}

function updateAvailabilityMode(value: string | number | boolean | undefined) {
  if (props.readonly) return
  const mode: WorkflowStartAvailabilityMode = value === 'fixed' || value === 'weekly' || value === 'monthly' || value === 'monthly_window' ? value : 'always'
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
  if (mode === 'monthly') {
    updateAvailability({ mode: 'monthly', timezone, monthDays: [1], dailyStartTime: '09:00', dailyEndTime: '18:00' })
    return
  }
  updateAvailability({ mode: 'monthly_window', timezone, windowStartDayFromEnd: 1, windowStartTime: '09:00', windowEndDay: 5, windowEndTime: '18:00' })
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

function updateWindowStartDayFromEnd(value: number | undefined) {
  updateAvailability({ ...baseAvailability('monthly_window'), windowStartDayFromEnd: normalizeInteger(value, 1, 28, 1) })
}

function updateWindowEndDay(value: number | undefined) {
  updateAvailability({ ...baseAvailability('monthly_window'), windowEndDay: normalizeInteger(value, 1, 31, 5) })
}

function updateWindowStartTime(value: unknown) {
  updateAvailability({ ...baseAvailability('monthly_window'), windowStartTime: normalizeClock(value, '09:00') })
}

function updateWindowEndTime(value: unknown) {
  updateAvailability({ ...baseAvailability('monthly_window'), windowEndTime: normalizeClock(value, '18:00') })
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
    windowStartDayFromEnd: mode === 'monthly_window' ? windowStartDayFromEnd.value : undefined,
    windowStartTime: mode === 'monthly_window' ? windowStartTime.value : undefined,
    windowEndDay: mode === 'monthly_window' ? windowEndDay.value : undefined,
    windowEndTime: mode === 'monthly_window' ? windowEndTime.value : undefined,
    startsAt: mode === 'fixed' ? current?.startsAt : undefined,
    endsAt: mode === 'fixed' ? current?.endsAt : undefined,
  }
}

function updateAvailability(config: WorkflowStartAvailabilityConfig) {
  const node = startNode.value
  if (!node || props.readonly) return
  const configuredBusinessPeriodSource = props.draft.instanceIdentity?.businessPeriod?.source
  node.availability = config
  if (config.mode === 'always' && node.startLimit?.mode === 'limited' && node.startLimit.period === 'availability') {
    node.startLimit = { ...node.startLimit, period: 'month' }
  }
  if (config.mode === 'always' && (configuredBusinessPeriodSource === 'availability_window_start' || configuredBusinessPeriodSource === 'availability_window_end')) {
    updateBusinessPeriod({ source: 'submit_time', field: undefined })
  }
  emit('change')
}

function updateTitleTemplate(value: string) {
  if (props.readonly) return
  props.draft.instanceIdentity = {
    ...props.draft.instanceIdentity,
    titleTemplate: value,
  }
  emit('change')
}

function appendTitleVariable(value: unknown) {
  if (props.readonly || typeof value !== 'string') return
  updateTitleTemplate(`${titleTemplate.value}${value}`)
}

function updateBusinessPeriodEnabled(value: string | number | boolean) {
  if (props.readonly) return
  if (!value) {
    props.draft.instanceIdentity = {
      titleTemplate: titleTemplate.value.split('{{businessPeriod}}').join('').replace(/\s+/g, ' ').trim(),
    }
    emit('change')
    return
  }
  props.draft.instanceIdentity = {
    titleTemplate: props.draft.instanceIdentity?.titleTemplate || '{{businessPeriod}} {{starterName}}提交的{{workflowName}}',
    businessPeriod: {
      enabled: true,
      granularity: businessPeriodGranularity.value,
      source: availabilityMode.value === 'always' ? 'submit_time' : 'availability_window_start',
      offset: 0,
    },
  }
  emit('change')
}

function updateBusinessPeriodGranularity(value: WorkflowBusinessPeriodGranularity) {
  updateBusinessPeriod({ granularity: value })
}

function updateBusinessPeriodSource(value: WorkflowBusinessPeriodSource) {
  updateBusinessPeriod({ source: value, field: value === 'form_field' ? businessPeriodFieldOptions.value[0]?.value : undefined })
}

function updateBusinessPeriodField(value: string) {
  updateBusinessPeriod({ field: value })
}

function updateBusinessPeriodOffset(value: number | undefined) {
  updateBusinessPeriod({ offset: normalizeInteger(value, -120, 120, 0) })
}

function updateBusinessPeriod(patch: Partial<NonNullable<NonNullable<WorkflowDraft['instanceIdentity']>['businessPeriod']>>) {
  if (props.readonly) return
  props.draft.instanceIdentity = {
    titleTemplate: titleTemplate.value,
    businessPeriod: {
      enabled: true,
      granularity: businessPeriodGranularity.value,
      source: businessPeriodSource.value,
      offset: businessPeriodOffset.value,
      ...patch,
    },
  }
  emit('change')
}

function updateStartLimitMode(value: string | number | boolean | undefined) {
  const node = startNode.value
  if (!node || props.readonly) return
  if (value !== 'limited') {
    node.startLimit = { mode: 'unlimited' }
  }
  else {
    node.startLimit = {
      mode: 'limited',
      period: availabilityMode.value === 'always' ? 'month' : 'availability',
      maxCount: 1,
    }
  }
  emit('change')
}

function updateStartLimitPeriod(value: string) {
  const node = startNode.value
  if (!node || props.readonly) return
  const period: WorkflowStartLimitPeriod = value === 'total' || value === 'day' || value === 'week' || value === 'availability' ? value : 'month'
  node.startLimit = { mode: 'limited', period, maxCount: startLimitMaxCount.value }
  emit('change')
}

function updateStartLimitMaxCount(value: number | undefined) {
  const node = startNode.value
  if (!node || props.readonly) return
  const normalized = Math.min(10000, Math.max(1, Math.trunc(Number(value) || 1)))
  node.startLimit = { mode: 'limited', period: startLimitPeriod.value, maxCount: normalized }
  emit('change')
}

function normalizeIDs(values: Array<number | string>) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value > 0)))
}

function normalizeDays(values: unknown[], maximum: number) {
  return Array.from(new Set(values.map(Number).filter(value => Number.isInteger(value) && value >= 1 && value <= maximum))).sort((left, right) => left - right)
}

function normalizeInteger(value: unknown, minimum: number, maximum: number, fallback: number) {
  const normalized = Math.trunc(Number(value))
  return Number.isInteger(normalized) && normalized >= minimum && normalized <= maximum ? normalized : fallback
}

function normalizeClock(value: unknown, fallback: string) {
  return typeof value === 'string' && /^([01]\d|2[0-3]):[0-5]\d$/.test(value) ? value : fallback
}

function flattenFormFields(fields: WorkflowDraft['form']): WorkflowDraft['form'] {
  return fields.flatMap(field => field.type === 'group' ? flattenFormFields(field.fields || []) : [field])
}
</script>

<style scoped>
.workflow-start-config { width: 100%; min-width: 0; overflow: auto; background: #f7f9fc; }
.workflow-start-config__content { display: flex; flex-wrap: wrap; align-items: flex-start; gap: 20px; width: calc(100% - 20px); margin: 10px; padding: 28px 0 48px; }
.config-section { box-sizing: border-box; flex: 1 1 380px; max-width: 520px; min-width: 0; padding: 24px 28px 28px; border: 1px solid #dfe5ec; border-radius: 8px; background: #fff; box-shadow: 0 2px 10px rgb(15 23 42 / 5%); }
.config-section--identity { max-width: 700px; }
.config-section__heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin: -2px 0 22px; padding-bottom: 16px; border-bottom: 1px solid #edf1f5; }
.config-section__heading > div { display: flex; align-items: center; gap: 10px; }
.config-section__heading h3 { margin: 0; color: #1f2937; font-size: 17px; font-weight: 650; }
.config-section__index { display: grid; width: 30px; height: 24px; place-items: center; border-radius: 4px; color: #0f766e; background: #dff5f1; font-size: 11px; font-weight: 700; }
.scope-mode, .limit-mode { display: flex; width: 100%; }
.availability-mode { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); width: 100%; }
.scope-mode :deep(.el-radio-button), .availability-mode :deep(.el-radio-button), .limit-mode :deep(.el-radio-button) { flex: 1; }
.scope-mode :deep(.el-radio-button__inner), .availability-mode :deep(.el-radio-button__inner), .limit-mode :deep(.el-radio-button__inner) { width: 100%; }
.availability-mode :deep(.el-radio-button__inner) { border: 1px solid var(--el-border-color); }
.availability-mode :deep(.availability-mode__wide) { grid-column: 1 / -1; }
.config-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 24px; }
.config-grid--single { grid-template-columns: minmax(0, 1fr); }
.config-grid--schedule { align-items: start; }
.template-variables { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 10px; }
.form-help { margin-top: 6px; color: #8492a6; font-size: 12px; line-height: 1.5; }
.weekday-options { display: flex; width: 100%; }
.weekday-options :deep(.el-checkbox-button) { flex: 1; }
.weekday-options :deep(.el-checkbox-button__inner) { width: 100%; padding-inline: 8px; }
.month-day-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 18px; width: 100%; }
.month-day-row .el-select { width: 100%; }
.window-anchor { display: flex; align-items: center; gap: 10px; width: 100%; color: #4b5563; white-space: nowrap; }
.window-anchor :deep(.el-input-number) { flex: 1 1 92px; width: auto; min-width: 92px; max-width: 110px; }
.window-help { margin: -8px 0 16px; }
.config-section :deep(.el-form-item:last-child) { margin-bottom: 0; }
.config-section :deep(.el-select), .config-section :deep(.el-input-number) { width: 100%; }
@media (max-width: 900px) {
  .workflow-start-config__content { gap: 14px; padding: 16px 0 28px; }
  .config-section { padding: 20px 16px 22px; }
  .config-grid { grid-template-columns: 1fr; gap: 0; }
}
</style>
