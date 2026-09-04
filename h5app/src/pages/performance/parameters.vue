<script setup lang="ts">
import type { PerformanceTemplate } from '@/types/dingtalk-h5'
import { computed, ref, watch } from 'vue'
import { getTemplate, saveTemplate } from '@/api/dingtalk-h5'
import { useDingtalkAuthStore } from '@/stores'
import { useAppContentStore } from '@/stores/appContent'
import PerformanceResizableTextarea from './components/PerformanceResizableTextarea.vue'

interface TemplateObjective {
  id?: string
  target?: string
  weight?: number | string
}

interface TemplateGrade {
  label?: string
  grade: string
  coefficient?: number | string
}

interface TemplateRubric {
  label?: string
  score?: number | string
  description?: string
}

interface TemplateValue {
  id?: string
  name?: string
  definition?: string
  rubric?: TemplateRubric[]
}

interface ParameterTemplate extends PerformanceTemplate {
  objectiveDefaults: TemplateObjective[]
  nextObjectiveDefaults: TemplateObjective[]
  gradeLevels: TemplateGrade[]
  values: TemplateValue[]
}

type ParameterTabKey = 'objectiveDefaults' | 'nextObjectiveDefaults' | 'gradeLevels' | 'values'

interface ParameterTab {
  name: string
  key: ParameterTabKey
}

function resolveMobilePage() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    const deviceType = String(info.deviceType || '').toLowerCase()
    const platform = String(info.platform || '').toLowerCase()
    return (width > 0 && width <= 768) || deviceType === 'phone' || (['android', 'ios'].includes(platform) && width <= 1024)
  }
  catch {
    return false
  }
}

const auth = useDingtalkAuthStore()
const appContent = useAppContentStore()
const template = ref<PerformanceTemplate | null>(null)
const loading = ref(false)
const isMobilePage = ref(resolveMobilePage())
const templateEditing = ref(false)
const templateSaving = ref(false)
const templateDraft = ref<ParameterTemplate>(emptyTemplate())
const listTitle = '绩效参数'
const listDesc = '目标模板、价值观标尺和绩效工资系数'
const parameterTabs: ParameterTab[] = [
  { name: '默认目标', key: 'objectiveDefaults' },
  { name: '下月目标', key: 'nextObjectiveDefaults' },
  { name: '绩效工资系数', key: 'gradeLevels' },
  { name: '价值观', key: 'values' },
]
const activeParameterTab = ref(0)
const activeParameterKey = computed<ParameterTabKey>(() => {
  return parameterTabs[activeParameterTab.value]?.key || 'objectiveDefaults'
})
const parameterTemplate = computed(() => {
  return templateEditing.value ? templateDraft.value : cloneTemplate(template.value)
})

function emptyTemplate(): ParameterTemplate {
  return {
    objectiveDefaults: [],
    nextObjectiveDefaults: [],
    gradeLevels: [],
    values: [],
  }
}

function cloneTemplate(template: PerformanceTemplate | null | undefined): ParameterTemplate {
  const source = template || {}
  return JSON.parse(JSON.stringify({
    objectiveDefaults: Array.isArray(source.objectiveDefaults) ? source.objectiveDefaults : [],
    nextObjectiveDefaults: Array.isArray(source.nextObjectiveDefaults) ? source.nextObjectiveDefaults : [],
    gradeLevels: Array.isArray(source.gradeLevels) ? source.gradeLevels : [],
    values: Array.isArray(source.values) ? source.values : [],
  }))
}

async function loadParametersPageData() {
  loading.value = true
  try {
    const res = await getTemplate()
    template.value = res.data || null
  }
  finally {
    loading.value = false
  }
  if (!templateEditing.value) {
    templateDraft.value = cloneTemplate(template.value)
  }
}

function startTemplateEdit() {
  templateDraft.value = cloneTemplate(template.value)
  templateEditing.value = true
}

function cancelTemplateEdit() {
  templateEditing.value = false
  templateDraft.value = cloneTemplate(template.value)
}

async function submitTemplate() {
  templateSaving.value = true
  try {
    const res = await saveTemplate(templateDraft.value)
    template.value = res.data || templateDraft.value
    templateEditing.value = false
    uni.showToast({
      title: '已保存',
      icon: 'success',
    })
  }
  finally {
    templateSaving.value = false
  }
}

function addObjective(field: 'objectiveDefaults' | 'nextObjectiveDefaults', prefix: string) {
  templateDraft.value[field].push({ id: `${prefix}-${Date.now()}`, target: '', weight: 0 })
}

function addGrade() {
  templateDraft.value.gradeLevels.push({ label: '', grade: '', coefficient: 1 })
}

function addValue() {
  templateDraft.value.values.push({
    id: `value-${Date.now()}`,
    name: '',
    definition: '',
    rubric: [
      { label: '卓越', score: 50, description: '持续超出要求，对团队或业务产生明显正向影响' },
      { label: '优秀', score: 40, description: '高质量完成要求，表现稳定且有主动贡献' },
      { label: '良好', score: 30, description: '符合岗位要求，能够稳定完成相关表现' },
    ],
  })
}

function addRubric(item: TemplateValue) {
  if (!Array.isArray(item.rubric)) {
    item.rubric = []
  }
  item.rubric.push({ label: '', score: 0, description: '' })
}

function removeItem<T>(items: T[], index: number) {
  items.splice(index, 1)
}

function handleParameterTabChange(index: number | string) {
  const nextIndex = Number(index)
  if (!Number.isFinite(nextIndex)) {
    activeParameterTab.value = 0
    return
  }
  activeParameterTab.value = Math.min(Math.max(nextIndex, 0), parameterTabs.length - 1)
}

watch(
  () => [appContent.currentKey, appContent.refreshTick],
  () => {
    if (appContent.currentKey === 'performance:template') {
      void loadParametersPageData()
    }
  },
  { immediate: true },
)
</script>

<template>
  <view class="performance-page">
    <view class="page-head">
      <view v-if="!isMobilePage" class="page-head__copy">
        <text class="page-title">
          {{ listTitle }}
        </text>
        <text class="page-desc">
          {{ listDesc }}
        </text>
      </view>
      <view class="head-actions">
        <u-button
          v-if="auth.hasButtonPermission('dingtalk_h5:button:template:edit') && !templateEditing"
          custom-class="dt-btn dt-btn-primary page-action-btn parameter-edit-btn"
          @click="startTemplateEdit"
        >
          编辑
        </u-button>
        <u-button v-if="templateEditing" custom-class="dt-btn dt-btn-light page-action-btn" @click="cancelTemplateEdit">
          取消
        </u-button>
        <u-button v-if="templateEditing" custom-class="dt-btn dt-btn-primary page-action-btn" :disabled="templateSaving" :loading="templateSaving" @click="submitTemplate">
          保存
        </u-button>
      </view>
    </view>

    <view v-if="loading" class="panel performance-loading-panel">
      <u-loading mode="circle" />
      <text>加载中...</text>
    </view>

    <template v-else>
      <view class="panel template-tabs-panel" :class="{ editing: templateEditing }">
        <view class="template-tabs-head">
          <u-tabs
            custom-class="template-tabs"
            :list="parameterTabs"
            :current="activeParameterTab"
            :is-scroll="true"
            active-color="#1677ff"
            inactive-color="#5f6b7a"
            :height="64"
            :font-size="26"
            :bar-width="36"
            :bar-height="4"
            @change="handleParameterTabChange"
          />
        </view>

        <view class="template-tab-body">
          <view v-if="activeParameterKey === 'objectiveDefaults'" class="template-section">
            <view class="panel-head">
              <text class="panel-title">
                默认目标
              </text>
              <u-button v-if="templateEditing" custom-class="dt-btn dt-btn-light small" @click="addObjective('objectiveDefaults', 'objective')">
                添加
              </u-button>
            </view>
            <view class="template-list">
              <view v-for="(item, index) in parameterTemplate.objectiveDefaults" :key="item.id || `objective-${index}`" class="template-row">
                <template v-if="!templateEditing">
                  <text class="template-row-main">
                    {{ item.target || '-' }}
                  </text>
                  <text class="template-weight-badge">
                    {{ item.weight || 0 }}%
                  </text>
                </template>
                <template v-else>
                  <PerformanceResizableTextarea v-model="item.target" :height="84" placeholder="目标描述" />
                  <view class="template-inline-fields">
                    <u-input v-model="item.weight" custom-class="template-editor-input" type="number" :border="true" placeholder="权重%" />
                    <u-button custom-class="dt-btn dt-btn-danger-light small" @click="removeItem(parameterTemplate.objectiveDefaults, index)">
                      删除
                    </u-button>
                  </view>
                </template>
              </view>
              <view v-if="parameterTemplate.objectiveDefaults.length === 0" class="template-empty">
                暂无模板项
              </view>
            </view>
          </view>

          <view v-if="activeParameterKey === 'nextObjectiveDefaults'" class="template-section">
            <view class="panel-head">
              <text class="panel-title">
                下月目标
              </text>
              <u-button v-if="templateEditing" custom-class="dt-btn dt-btn-light small" @click="addObjective('nextObjectiveDefaults', 'next')">
                添加
              </u-button>
            </view>
            <view class="template-list">
              <view v-for="(item, index) in parameterTemplate.nextObjectiveDefaults" :key="item.id || `next-${index}`" class="template-row">
                <template v-if="!templateEditing">
                  <text class="template-row-main">
                    {{ item.target || '-' }}
                  </text>
                  <text class="template-weight-badge">
                    {{ item.weight || 0 }}%
                  </text>
                </template>
                <template v-else>
                  <PerformanceResizableTextarea v-model="item.target" :height="84" placeholder="目标描述" />
                  <view class="template-inline-fields">
                    <u-input v-model="item.weight" custom-class="template-editor-input" type="number" :border="true" placeholder="权重%" />
                    <u-button custom-class="dt-btn dt-btn-danger-light small" @click="removeItem(parameterTemplate.nextObjectiveDefaults, index)">
                      删除
                    </u-button>
                  </view>
                </template>
              </view>
              <view v-if="parameterTemplate.nextObjectiveDefaults.length === 0" class="template-empty">
                暂无模板项
              </view>
            </view>
          </view>

          <view v-if="activeParameterKey === 'gradeLevels'" class="template-section">
            <view class="panel-head">
              <text class="panel-title">
                绩效工资系数
              </text>
              <u-button v-if="templateEditing" custom-class="dt-btn dt-btn-light small" @click="addGrade">
                添加
              </u-button>
            </view>
            <view class="template-list">
              <view v-for="(item, index) in parameterTemplate.gradeLevels" :key="`${item.grade || 'grade'}-${index}`" class="template-row">
                <template v-if="!templateEditing">
                  <text class="template-row-main">
                    {{ item.label || '-' }} · {{ item.grade || '-' }}
                  </text>
                  <text class="template-weight-badge">
                    {{ item.coefficient || '-' }}
                  </text>
                </template>
                <template v-else>
                  <view class="template-inline-fields">
                    <u-input v-model="item.label" custom-class="template-editor-input" :border="true" placeholder="等级标签" />
                    <u-input v-model="item.grade" custom-class="template-editor-input" :border="true" placeholder="档位" />
                    <u-input v-model="item.coefficient" custom-class="template-editor-input" type="number" :border="true" placeholder="系数" />
                    <u-button custom-class="dt-btn dt-btn-danger-light small" @click="removeItem(parameterTemplate.gradeLevels, index)">
                      删除
                    </u-button>
                  </view>
                </template>
              </view>
              <view v-if="parameterTemplate.gradeLevels.length === 0" class="template-empty">
                暂无模板项
              </view>
            </view>
          </view>

          <view v-if="activeParameterKey === 'values'" class="template-section">
            <view class="panel-head">
              <text class="panel-title">
                价值观
              </text>
              <u-button v-if="templateEditing" custom-class="dt-btn dt-btn-light small" @click="addValue">
                添加
              </u-button>
            </view>
            <view class="template-list">
              <view v-for="(item, index) in parameterTemplate.values" :key="item.id || `value-${index}`" class="template-row value">
                <template v-if="!templateEditing">
                  <text class="template-row-main">
                    {{ item.name || '-' }}
                  </text>
                  <text class="template-row-desc">
                    {{ item.definition || '-' }}
                  </text>
                  <view v-if="item.rubric?.length" class="template-rubric-preview">
                    <text v-for="(rubric, rubricIndex) in item.rubric" :key="`${rubric.label || 'rubric'}-${rubricIndex}`" class="template-rubric-preview-item">
                      {{ rubric.label || '-' }} {{ rubric.score || 0 }}分
                    </text>
                  </view>
                </template>
                <template v-else>
                  <view class="template-inline-fields template-value-head">
                    <u-input v-model="item.name" custom-class="template-editor-input" :border="true" placeholder="价值观名称" />
                    <u-button custom-class="dt-btn dt-btn-danger-light small" @click="removeItem(parameterTemplate.values, index)">
                      删除
                    </u-button>
                  </view>
                  <PerformanceResizableTextarea v-model="item.definition" :height="84" placeholder="价值观定义" />
                  <view class="template-rubric-list">
                    <view v-for="(rubric, rubricIndex) in item.rubric || []" :key="`${rubric.label || 'rubric'}-${rubricIndex}`" class="template-rubric-row">
                      <view class="template-rubric-fields">
                        <u-input v-model="rubric.label" custom-class="template-editor-input" :border="true" placeholder="评分名称" />
                        <u-input v-model="rubric.score" custom-class="template-editor-input" type="number" :border="true" placeholder="分值" />
                        <u-button custom-class="dt-btn dt-btn-danger-light small" @click="removeItem(item.rubric || [], rubricIndex)">
                          删除
                        </u-button>
                      </view>
                      <PerformanceResizableTextarea v-model="rubric.description" :height="84" placeholder="评分说明" />
                    </view>
                    <u-button custom-class="dt-btn dt-btn-light small" @click="addRubric(item)">
                      添加评分
                    </u-button>
                  </view>
                </template>
              </view>
              <view v-if="parameterTemplate.values.length === 0" class="template-empty">
                暂无模板项
              </view>
            </view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<style lang="scss" scoped src="./components/performance-page.scss"></style>
