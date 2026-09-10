<template>
  <div class="survey-sidebar-panel">
    <div class="survey-sidebar-panel-tabs">
      <button v-for="tab in tabs" :key="tab.name" class="survey-sidebar-panel-tabs-pane" :class="{ active: middleTab === tab.name }" :title="tab.label" @click="middleTab = tab.name">
        <component :is="tab.icon" />
        <span class="tab-label">{{ tab.label }}</span>
        <span v-if="tab.name === 'result' && form.resultConfig.viewMode === 'perResponse'" class="tab-badge">逐份</span>
      </button>
    </div>
    <div v-show="middleTab !== 'logic' && middleTab !== 'result'" class="survey-sidebar-panel-tabs-content">
      <el-tabs v-if="middleTab === 'item'" v-model="sideSubTab" class="side-sub-tabs">
        <el-tab-pane label="题型" name="types"><SurveyDesignerQuestionPalette :types="types" @add="emit('add', $event)" /></el-tab-pane>
        <el-tab-pane label="题库" name="bank">
          <div class="bank-panel">
            <div class="bank-search"><el-input v-model="bankKeyword" placeholder="搜索题目..." size="small" clearable @input="loadBank" /></div>
            <div class="bank-list">
              <template v-if="bankTree.length">
                <div v-for="category in bankTree" :key="category.key" class="bank-cat">
                  <div class="bank-cat-title" @click="toggleBankExpand(`cat:${category.expandKey}`)"><span class="bank-arrow" :class="{ expanded: category._expanded }" /><span class="bank-cat-name">{{ category.label }}</span><span class="bank-count">{{ category.count }} 题</span></div>
                  <div v-show="category._expanded" class="bank-cat-body">
                    <div v-for="group in category.children" :key="group.key" class="bank-type-group">
                      <div class="bank-type-title" @click="toggleBankExpand(`type:${category.expandKey}|${group.expandKey}`)"><span class="bank-arrow" :class="{ expanded: group._expanded }" /><QuestionIcon :type="group.label" class="bank-icon" /><span class="bank-type-name">{{ typeName(group.label) }}</span><span class="bank-count">{{ group.children.length }} 题</span></div>
                      <div v-show="group._expanded" class="bank-type-body">
                        <button v-for="question in group.children" :key="question.id" class="bank-item" @click="emit('addBank', question)"><span class="bank-item-main"><span class="bank-title" :title="bankQuestionTitle(question)">{{ bankQuestionTitle(question) }}</span><span class="bank-meta">#{{ question.id }}</span></span><span class="bank-type">{{ typeName(question.type) }}</span></button>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
              <el-empty v-if="!bankTree.length && !bankLoading" description="题库暂无题目" :image-size="40" />
              <div v-if="bankLoadingVisible" class="bank-loading">加载中...</div>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="大纲" name="outline"><SurveyDesignerOutline :title="form.title" :questions="questions" :selected-id="selectedId" :invalid-ids="invalidIds" :type-name="typeName" @select="emit('select', $event)" /></el-tab-pane>
      </el-tabs>

      <div v-if="middleTab === 'appearance'" class="appearance-panel">
        <el-tabs v-model="appearanceTab" class="appearance-tabs">
          <el-tab-pane v-for="resourceType in resourceTypes" :key="resourceType.name" :label="resourceType.label" :name="resourceType.name">
            <div class="appearance-grid">
              <div v-for="(item, index) in resourcesFor(resourceType.name)" :key="item.id || index" class="appearance-thumb" :class="{ active: isActiveImage(resourceType.field, item) }" @click="applyImage(resourceType.field, item)"><img :src="resourceUrl(item)" /><div class="appearance-overlay"><span>点击应用</span></div><button class="appearance-remove" @click.stop="removeResource(resourceType.field, item)">✕</button></div>
              <el-upload action="/api/v2/admin/survey-resources" :show-file-list="false" :on-success="handleUploadSuccess" :on-error="handleUploadError" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: resourceType.name }" :before-upload="checkSaved"><div class="appearance-add">+</div></el-upload>
            </div>
            <el-empty v-if="!resourcesFor(resourceType.name).length" :description="`暂无${resourceType.label}`" :image-size="30" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import QuestionIcon from '../formkit/QuestionIcon.vue'
import SurveyDesignerOutline from './SurveyDesignerOutline.vue'
import SurveyDesignerQuestionPalette from './SurveyDesignerQuestionPalette.vue'
import type { SurveyDesignerQuestion } from '../survey-designer-types'

const props = defineProps<{ form: Record<string, any>; types: any[]; questions: SurveyDesignerQuestion[]; selectedId: string | null; invalidIds: string[]; typeName: (type: string) => string }>()
const middleTab = defineModel<string>({ required: true })
const emit = defineEmits<{ add: [type: any]; addBank: [question: any]; select: [questionId: string]; save: [] }>()
const form = computed(() => props.form)
const { typeName } = props
const sideSubTab = ref('types')
const appearanceTab = ref('bg')
const bankKeyword = ref('')
const bankQuestions = ref<any[]>([])
const bankLoading = ref(false)
const bankLoadingVisible = ref(false)
const bankExpanded = ref<Record<string, boolean>>({})
const backgroundResources = ref<any[]>([])
const headerResources = ref<any[]>([])
let bankTimer: ReturnType<typeof setTimeout> | null = null
let loadingTimer: ReturnType<typeof setTimeout> | null = null
let bankLoadSequence = 0

const tabs = [
  { name: 'item', label: '题目', icon: makeIcon('M810.7 128H213.3A85.3 85.3 0 00128 213.3v597.4a85.3 85.3 0 0085.3 85.3h597.4a85.3 85.3 0 0085.3-85.3V213.3A85.3 85.3 0 00810.7 128zM768 682.7H256v-85.4h512v85.4zm0-170.7H256v-85.3h512V512zm0-170.7H256V256h512v85.3z') },
  { name: 'appearance', label: '外观', icon: makeIcon('M512 128a384 384 0 100 768 384 384 0 000-768zm0 682.7a298.7 298.7 0 110-597.4 298.7 298.7 0 010 597.4zm0-469.4a170.7 170.7 0 100 341.4 170.7 170.7 0 000-341.4z') },
  { name: 'logic', label: '逻辑', icon: makeIcon('M633 57a57 57 0 00-57-57H455a57 57 0 00-57 57v121H171a57 57 0 00-57 57v113a57 57 0 0057 57h121v370a57 57 0 0057 57h113a57 57 0 0057-57V462h114a57 57 0 0057-57V292a57 57 0 00-57-57V57z') },
  { name: 'result', label: '统计', icon: makeIcon('M128 128h768v85H213v683h-85V128zm171 171h85v426h-85V299zm170 128h86v298h-86V427zm171-43h85v341h-85V384z') },
]
const resourceTypes = [
  { name: 'bg', label: '背景图', field: 'backgroundImages' },
  { name: 'header', label: '页眉图', field: 'headerImages' },
] as const
const uploadHeaders = { Authorization: localStorage.getItem('admin_token') || '' }

const bankTree = computed(() => {
  const grouped: Record<string, Record<string, any[]>> = {}
  bankQuestions.value.forEach(question => {
    const category = question.category || ''
    const type = question.type || ''
    ;(grouped[category] ||= {})[type] ||= []
    grouped[category][type].push(question)
  })
  return Object.entries(grouped).map(([category, types]) => {
    const children = Object.entries(types).map(([type, items]) => ({ key: type || '__unknown_type__', expandKey: type, label: type, _expanded: bankExpanded.value[`type:${category}|${type}`] ?? false, children: items }))
    return { key: category || '__uncategorized__', expandKey: category, label: category || '未分类', count: children.reduce((sum, group) => sum + group.children.length, 0), _expanded: bankExpanded.value[`cat:${category}`] ?? false, children }
  })
})

function makeIcon(path: string) { return defineComponent(() => () => h('svg', { viewBox: '0 0 1024 1024', width: 20, height: 20, fill: 'currentColor' }, [h('path', { d: path })])) }
function toggleBankExpand(key: string) { bankExpanded.value[key] = !bankExpanded.value[key] }
function bankQuestionTitle(question: any) { return String(question?.title || '').replace(/<[^>]*>/g, '').split('\n')[0].replace(/\s+/g, ' ').trim() || '未命名' }
function loadBank() {
  if (bankTimer) clearTimeout(bankTimer)
  if (loadingTimer) clearTimeout(loadingTimer)
  bankLoadingVisible.value = false
  const sequence = ++bankLoadSequence
  bankTimer = setTimeout(async () => {
    bankLoading.value = true
    loadingTimer = setTimeout(() => { if (bankLoading.value && sequence === bankLoadSequence) bankLoadingVisible.value = true }, 450)
    try {
      const response: any = await adminApi.surveyQuestionBankList({ keyword: bankKeyword.value, page: 1, pageSize: 100 })
      if (sequence === bankLoadSequence) bankQuestions.value = response.data?.list || []
    } catch { if (sequence === bankLoadSequence) bankQuestions.value = [] }
    finally {
      if (sequence === bankLoadSequence) { bankLoading.value = false; bankLoadingVisible.value = false }
      if (loadingTimer) { clearTimeout(loadingTimer); loadingTimer = null }
    }
  }, 300)
}
function resourceUrl(item: any) { return `${item.domain || ''}${item.url}` }
function resourcesFor(type: 'bg' | 'header') { return type === 'bg' ? backgroundResources.value : headerResources.value }
function isActiveImage(field: string, item: any) {
  const current = form.value[field]?.[0]
  return current && typeof current === 'object' && current.id ? current.id === item.id : current === resourceUrl(item)
}
function applyImage(field: string, item: any) {
  if (isActiveImage(field, item)) return
  form.value[field] = [{ id: item.id, url: resourceUrl(item) }]
  emit('save')
}
async function removeResource(field: string, item: any) {
  if (item.id) { try { await adminApi.surveyResourceDelete({ id: item.id }) } catch {} }
  if (isActiveImage(field, item)) { form.value[field] = []; emit('save') }
  await loadResources()
}
function checkSaved() { if (!form.value.id) { ElMessage.warning('请先保存问卷'); return false }; return true }
function handleUploadSuccess(response: any) { if (!response.data?.url) return ElMessage.error('上传失败：响应数据异常'); ElMessage.success('已上传到资源库，点击图片应用'); loadResources() }
function handleUploadError() { ElMessage.error('上传失败') }
async function loadResources() {
  if (!form.value.id) return
  try {
    const response: any = await adminApi.surveyResourceList({ surveyId: form.value.id })
    const resources = Array.isArray(response.data) ? response.data : []
    backgroundResources.value = resources.filter((resource: any) => resource.type === 'bg')
    headerResources.value = resources.filter((resource: any) => resource.type === 'header')
  } catch {}
}

watch(middleTab, value => { if (value === 'appearance') loadResources() })
watch(sideSubTab, value => { if (value === 'bank') loadBank() })
onBeforeUnmount(() => { if (bankTimer) clearTimeout(bankTimer); if (loadingTimer) clearTimeout(loadingTimer) })
function showOutline() { middleTab.value = 'item'; sideSubTab.value = 'outline' }
defineExpose({ loadBank, loadResources, showOutline })
</script>

<style scoped>
.survey-sidebar-panel { position:relative; z-index:2; display:flex; flex:0 0 auto; user-select:none; box-shadow:2px 0 14px rgba(15,23,42,.04); }
.survey-sidebar-panel-tabs { display:flex; flex-direction:column; font-size:12px; background:#f7f8fb; border-right:1px solid var(--designer-border); padding:6px 0; }
.survey-sidebar-panel-tabs-pane { position:relative; display:flex; flex-direction:column; align-items:center; justify-content:center; width:44px; height:42px; color:#8a94a6; cursor:pointer; border:0; border-radius:8px; margin:2px 5px; background:transparent; transition:all .15s; }
.survey-sidebar-panel-tabs-pane:hover, .survey-sidebar-panel-tabs-pane.active { color:var(--designer-accent); background:#fff; box-shadow:0 6px 14px rgba(15,23,42,.08); }
.tab-badge { position:absolute; top:-2px; right:-6px; font-size:9px; background:var(--designer-accent); color:#fff; border-radius:8px; padding:1px 4px; line-height:1.4; }
.survey-sidebar-panel-tabs-content { display:flex; flex-direction:column; width:256px; height:100%; border-right:1px solid var(--designer-border); overflow-y:auto; background:#fff; }
.side-sub-tabs { display:flex; flex-direction:column; height:100%; }
.side-sub-tabs :deep(.el-tabs__header) { margin:0; padding:0 10px; border-bottom:1px solid var(--designer-border); }
.side-sub-tabs :deep(.el-tabs__nav), .side-sub-tabs :deep(.el-tabs__nav-scroll) { display:flex; width:100%; }
.side-sub-tabs :deep(.el-tabs__item) { flex:1; justify-content:center; padding:0; font-size:13px; height:36px; color:#667085; }
.side-sub-tabs :deep(.el-tabs__content) { flex:1; overflow-y:auto; }
.appearance-panel { padding:14px 16px 14px 18px; }
.appearance-panel :deep(.el-tabs__header) { margin:0 0 8px; }
.appearance-grid { display:grid; grid-template-columns:repeat(auto-fill,86px); gap:10px; padding-left:2px; }
.appearance-grid :deep(.el-upload), .appearance-add, .appearance-thumb { width:86px; }
.appearance-add, .appearance-thumb { height:86px; border-radius:8px; overflow:hidden; }
.appearance-add { display:flex; align-items:center; justify-content:center; border:2px dashed var(--designer-border); font-size:24px; color:#ccc; cursor:pointer; }
.appearance-thumb { position:relative; cursor:pointer; border:2px solid transparent; }
.appearance-thumb:hover, .appearance-thumb.active { border-color:var(--designer-accent); }
.appearance-thumb img { display:block; width:100%; height:100%; object-fit:cover; }
.appearance-overlay { position:absolute; inset:0; display:flex; align-items:center; justify-content:center; background:rgba(0,0,0,.4); opacity:0; color:#fff; font-size:11px; }
.appearance-thumb:hover .appearance-overlay { opacity:1; }
.appearance-remove { position:absolute; top:2px; right:2px; width:18px; height:18px; border:0; border-radius:50%; background:rgba(0,0,0,.5); color:#fff; cursor:pointer; }
.bank-panel { padding:10px 8px 12px; display:flex; flex-direction:column; height:100%; box-sizing:border-box; }
.bank-search { margin-bottom:10px; }
.bank-list { flex:1; min-height:0; overflow-y:auto; display:flex; flex-direction:column; gap:6px; }
.bank-cat { border:1px solid #eef2f7; border-radius:8px; overflow:hidden; }
.bank-cat-title, .bank-type-title { display:flex; align-items:center; gap:6px; cursor:pointer; }
.bank-cat-title { padding:7px 8px; font-size:13px; font-weight:650; background:#f8fafc; }
.bank-type-title { padding:5px 6px; font-size:12px; color:#475569; }
.bank-cat-name, .bank-type-name, .bank-item-main, .bank-title { flex:1; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bank-arrow::before { content:''; display:block; width:6px; height:6px; border-right:1.5px solid currentColor; border-bottom:1.5px solid currentColor; transform:rotate(-45deg); }
.bank-arrow.expanded::before { transform:rotate(45deg); }
.bank-count, .bank-type, .bank-meta { flex-shrink:0; color:#94a3b8; font-size:11px; }
.bank-cat-body { padding:4px 6px 6px; }
.bank-type-body { padding-left:12px; }
.bank-item { width:100%; display:flex; align-items:center; gap:6px; border:0; border-radius:6px; padding:6px; background:transparent; text-align:left; cursor:pointer; }
.bank-item:hover { background:#f8fafc; }
.bank-item-main { display:flex; align-items:center; gap:6px; }
.bank-title { font-size:13px; color:#1f2937; }
.bank-icon { width:20px; height:20px; }
.bank-loading { padding:16px; text-align:center; color:#98a2b3; font-size:12px; }
</style>
