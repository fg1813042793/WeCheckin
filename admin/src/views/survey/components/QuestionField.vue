<template>
  <div class="q-item" :data-qid="q.id" :class="{ 'q-item--layout': LAYOUT_TYPES.includes(q.type) }">
    <template v-if="LAYOUT_TYPES.includes(q.type)">
      <div class="q-type">
        <span v-if="q.type==='description'" class="q-type-label">描述</span>
        <span v-else-if="q.type==='divider'" class="q-type-label">分隔</span>
        <span v-else class="q-type-label">分页</span>
        <div v-if="q.type==='description'" class="q-content" v-html="q.description" />
        <el-divider v-if="q.type==='divider'" style="margin:4px 0" />
      </div>
    </template>
    <template v-else>
      <div class="q-title" v-html="processedTitleHtml" />
      <div v-if="q.description && q.showDescription !== false" class="q-desc">
        <span class="q-desc-label">说明：</span>
        <div v-html="q.description" />
      </div>
      <div class="q-input">
        <el-input v-if="['input','text'].includes(q.type)" v-model="localAnswers[q.id]" :placeholder="q.placeholder || '请输入'" />
        <div v-else-if="q.type==='multiInput'" class="q-field-stack">
          <el-input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" v-model="localAnswers[q.id][fi]" :placeholder="f.placeholder||'请输入'" />
        </div>
        <div v-else-if="q.type==='hInput'" class="q-field-row">
          <el-input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" v-model="localAnswers[q.id][fi]" :placeholder="f.placeholder||'请输入'" />
        </div>
        <el-input v-else-if="q.type==='textarea'" v-model="localAnswers[q.id]" type="textarea" :rows="3" :placeholder="q.placeholder || '请输入'" />
        <el-input-number v-else-if="q.type==='number'" v-model="localAnswers[q.id]" style="width:100%;--el-input-width:100%" />
        <el-radio-group v-else-if="q.type==='radio'" v-model="localAnswers[q.id]" class="q-options" :style="optionGrid(q)">
          <el-radio v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label" /></el-radio>
        </el-radio-group>
        <el-checkbox-group v-else-if="q.type==='checkbox'" v-model="localAnswers[q.id]" class="q-options" :style="optionGrid(q)">
          <el-checkbox v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label" /></el-checkbox>
        </el-checkbox-group>
        <el-select v-else-if="q.type==='select'" v-model="localAnswers[q.id]" placeholder="请选择" style="width:100%" clearable :teleported="false">
          <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value" :label="o.label" />
        </el-select>
        <el-select v-else-if="q.type==='picker'" v-model="localAnswers[q.id]" placeholder="请选择" style="width:100%" clearable :teleported="false">
          <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value" :label="o.label" />
        </el-select>
        <el-cascader v-else-if="q.type==='cascade'" v-model="localAnswers[q.id]" placeholder="请选择" style="width:100%" :options="q.props?.options||[]" clearable />
        <el-radio-group v-else-if="q.type==='judge'" v-model="localAnswers[q.id]" class="q-options">
          <el-radio value="true">对</el-radio>
          <el-radio value="false">错</el-radio>
        </el-radio-group>
        <div v-else-if="q.type==='rating'" class="q-rating"><el-rate v-model="localAnswers[q.id]" :max="q.props?.maxRating || 5" /></div>
        <div v-else-if="q.type==='nps'" class="q-nps">
          <div class="q-nps-labels"><span>0</span><span>10</span></div>
          <el-rate v-model="localAnswers[q.id]" :max="10" show-score score-template="{value}" />
        </div>
        <el-date-picker v-else-if="q.type==='date'" v-model="localAnswers[q.id]" type="date" placeholder="选择日期" style="width:100%" />
        <el-time-picker v-else-if="q.type==='time'" v-model="localAnswers[q.id]" placeholder="选择时间" style="width:100%" />
        <el-switch v-else-if="q.type==='switch'" v-model="localAnswers[q.id]" />
        <div v-else-if="q.type==='file'" class="q-file">
          <input :ref="el => { if(el) fileInputs[q.id]=el as HTMLInputElement }" type="file" :multiple="q.props?.multiple !== false" style="display:none" @change="onFileInput" />
          <el-button text @click="triggerFileInput(q.id)">
            <svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7z"/></svg>选择文件
          </el-button>
          <div v-if="(fileLists[q.id]||[]).length" class="q-file-list">
            <div v-for="(f, fi) in fileLists[q.id]" :key="fi" class="q-file-item">
              <span class="q-file-name">{{ f.name }}</span>
              <el-button text size="small" type="danger" @click="removeFile(q.id, fi)" style="padding:0 4px">×</el-button>
            </div>
          </div>
        </div>
        <div v-else-if="q.type==='location'" class="q-location">
          <div v-if="localAnswers[q.id]" class="q-location-result">{{ localAnswers[q.id] }}</div>
          <el-button v-else text @click="pickLocation(q.id)">
            <svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M512 64C367.2 64 248 183.2 248 328c0 163.2 233.6 524.8 252 551.2 3.2 4.8 8 7.2 12 7.2s8.8-2.4 12-7.2C542.4 852.8 776 491.2 776 328 776 183.2 656.8 64 512 64z m0 400c-39.2 0-72-32.8-72-72s32.8-72 72-72 72 32.8 72 72-32.8 72-72 72z"/></svg>选择位置
          </el-button>
        </div>
        <el-input v-else-if="['phone','name','studentId','employeeId','class'].includes(q.type)" placeholder="请输入" v-model="localAnswers[q.id]" />
        <el-input v-else-if="q.type==='email'" placeholder="邮箱地址" v-model="localAnswers[q.id]" />
        <el-input v-else-if="q.type==='idCard'" placeholder="身份证号" v-model="localAnswers[q.id]" />
        <el-input v-else-if="q.type==='password'" type="password" placeholder="密码" v-model="localAnswers[q.id]" />
        <el-cascader v-else-if="['user','dept'].includes(q.type)" v-model="localAnswers[q.id]" :placeholder="q.type==='user'?'选择成员':'选择部门'" style="width:100%" :options="buildUserDeptTree(q)" :props="{ multiple: !!q.multiple, emitPath: false }" clearable />
        <div v-else-if="q.type==='dateRange'" class="q-field-stack">
          <el-date-picker v-model="localAnswers[q.id][0]" type="date" placeholder="开始日期" style="width:100%" />
          <el-date-picker v-model="localAnswers[q.id][1]" type="date" placeholder="结束日期" style="width:100%" />
        </div>
        <div v-else-if="q.type==='matrixRadio'" class="q-matrix"><table><thead><tr><th class="q-mtx-cnr">行\列</th><th v-for="c in (q.props?.columns||[])" :key="c?.title||c?.label||c">{{ c?.title||c?.label||c }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[])" :key="r?.title||r"><td class="q-mtx-label">{{ r?.title||r }}</td><td v-for="c in (q.props?.columns||[])" :key="c?.title||c?.label||c"><el-radio-group :model-value="localAnswers[q.id][ri]" @update:model-value="(v: any) => { localAnswers[q.id][ri] = v; emitUpdate() }"><el-radio :value="c?.title||c?.label||c" /></el-radio-group></td></tr></tbody></table></div>
        <div v-else-if="q.type==='matrixCheckbox'" class="q-matrix"><table><thead><tr><th class="q-mtx-cnr">行\列</th><th v-for="c in (q.props?.columns||[])" :key="c?.title||c?.label||c">{{ c?.title||c?.label||c }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[])" :key="r?.title||r"><td class="q-mtx-label">{{ r?.title||r }}</td><td v-for="c in (q.props?.columns||[])" :key="c?.title||c?.label||c"><el-checkbox-group :model-value="localAnswers[q.id][ri]||[]" @update:model-value="(v: any) => { localAnswers[q.id][ri] = v; emitUpdate() }"><el-checkbox :value="c?.title||c?.label||c" /></el-checkbox-group></td></tr></tbody></table></div>
        <div v-else-if="q.type==='matrixFillBlank'" class="q-matrix"><table><thead><tr><th class="q-mtx-cnr">行\列</th><th v-for="c in (q.props?.columns||[])" :key="c?.title||c?.label||c">{{ c?.title||c?.label||c }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[])" :key="r?.title||r"><td class="q-mtx-label">{{ r?.title||r }}</td><td v-for="(c, ci) in (q.props?.columns||[])" :key="c?.title||c?.label||c"><el-input :model-value="localAnswers[q.id][ri]?.[ci]" @update:model-value="(v: any) => { if(!localAnswers[q.id][ri]) localAnswers[q.id][ri]={}; localAnswers[q.id][ri][ci]=v; emitUpdate() }" placeholder="填空" size="small" style="width:100%" /></td></tr></tbody></table></div>
        <div v-else-if="q.type==='matrixAuto'" class="q-matrix"><table><thead><tr><th class="q-mtx-cnr">#</th><th v-for="c in (q.props?.columns||[])" :key="c?.label||c?.id||c">{{ c?.label||c?.id||c }}</th><th style="width:40px"></th></tr></thead><tbody><tr v-for="(r, ri) in (localAnswers[q.id]||[])" :key="ri"><td class="q-mtx-label">{{ ri+1 }}</td><td v-for="(c, ci) in (q.props?.columns||[])" :key="c?.label||c?.id||c"><el-input v-model="localAnswers[q.id][ri][ci]" size="small" :placeholder="c?.label||'值'" style="width:100%" /></td><td><el-button text size="small" type="danger" @click="removeMatrixRow(q.id, ri)" style="padding:2px"><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg></el-button></td></tr></tbody></table>
          <div class="q-mtx-add"><el-button size="small" text @click="addMatrixRow(q.id)">+ 添加行</el-button></div>
        </div>
        <div v-else-if="q.type==='richText'" class="q-quill-wrap">
          <QuillEditor v-model:content="localAnswers[q.id]" content-type="html" :options="{ theme: 'snow', placeholder: q.placeholder || '输入富文本内容...', modules: { imageResize: {} } }" style="min-height:150px" />
        </div>
        <el-input v-else-if="q.type==='scanCode'" v-model="localAnswers[q.id]" :placeholder="q.placeholder || '扫码'" class="scan-code-input">
          <template #prefix>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
          </template>
          <template #suffix>
            <el-button text type="primary" size="small" class="scan-code-btn" @click.stop="openScanner(q.id)">扫码</el-button>
          </template>
        </el-input>
        <div v-else-if="q.type==='signature'" class="q-sig-wrap">
          <canvas :ref="el => { if(el) { sigCanvasMap[q.id]=el as HTMLCanvasElement; restoreSignature(q.id, el as HTMLCanvasElement) } }" class="q-sig-canvas" @mousedown="sigStart($event, q.id)" @mousemove="sigMove($event, q.id)" @mouseup="sigEnd" @mouseleave="sigEnd" @touchstart.stop.prevent="e => sigTouchStart(e, q.id)" @touchmove.stop.prevent="e => sigTouchMove(e, q.id)" @touchend.stop="sigEnd" />
          <div class="q-sig-actions"><el-button size="small" text @click="clearSignature(q.id)">清除</el-button></div>
        </div>
        <el-input v-else v-model="localAnswers[q.id]" placeholder="请输入" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import '../../../utils/quill-image-resize'

const props = defineProps<{
  q: any
  index: number
  answers: any
  settings: any
  requiredQuestionIds?: Set<string>
  score?: number | string
  fileLists: Record<string, File[]>
  fileInputs: Record<string, HTMLInputElement>
  sigCanvasMap: Record<string, HTMLCanvasElement>
}>()

const emit = defineEmits<{
  'update:answers': [val: any]
  'triggerFile': [qid: string]
  'removeFile': [qid: string, idx: number]
  'pickLocation': [qid: string]
  'openScanner': [qid: string]
  'sigStart': [e: MouseEvent | Touch, id: string]
  'sigMove': [e: MouseEvent | Touch, id: string]
  'sigEnd': []
  'sigTouchStart': [e: TouchEvent, id: string]
  'sigTouchMove': [e: TouchEvent, id: string]
  'clearSignature': [id: string]
  'addMatrixRow': [qid: string]
  'removeMatrixRow': [qid: string, ri: number]
}>()

const LAYOUT_TYPES = ['description', 'divider', 'pagination']

const processedTitleHtml = computed(() => {
  let html = ''
  if (props.settings.questionNumber !== false) {
    html += `<span class="q-num">${props.index + 1}.</span>`
  }
  if (props.score) {
    html += `<span class="q-score-inline">(${props.score}分)</span>`
  }
  if (isRequired.value) {
    html += '<span class="q-req">*</span>'
  }
  html += unwrapOuterP(props.q.title || '')
  return html
})

function unwrapOuterP(html: string): string {
  const trimmed = html.trim()
  const match = trimmed.match(/^<p([^>]*)>([\s\S]*)<\/p>$/)
  if (!match) return html
  const attrs = match[1]
  const content = match[2]
  if (/\bql-align-\w+\b/.test(attrs)) {
    return `<span${attrs} style="display:inline-block;width:100%;text-align:inherit">${content}</span>`
  }
  return content
}

const localAnswers = computed({
  get: () => props.answers,
  set: (val) => emit('update:answers', val)
})
const isRequired = computed(() => !!props.q.required || !!props.requiredQuestionIds?.has(props.q.id))

function optionGrid(q: any) {
  const cols = q.optionLayout
  if (!cols || cols <= 1) return {}
  return { display: 'grid', gridTemplateColumns: `repeat(${cols}, 1fr)`, gap: '4px' }
}

function buildUserDeptTree(q: any) {
  const opts = q.props?.options || []
  if (q.type === 'user') {
    const deptMap: Record<string, any> = {}
    opts.forEach((o: any) => {
      const deptId = o.deptId || ''
      if (!deptMap[deptId]) deptMap[deptId] = { value: '__d__' + deptId, label: o.deptName || deptId || '未分组', children: [] }
      deptMap[deptId].children.push({ value: o.value, label: o.label || '成员' })
    })
    return Object.values(deptMap)
  }
  const map: Record<string, any> = {}
  opts.forEach((o: any) => { map[o.value] = { ...o, children: [] } })
  const roots: any[] = []
  opts.forEach((o: any) => {
    if (o.parentId && map[o.parentId]) map[o.parentId].children.push(map[o.value])
    else roots.push(map[o.value])
  })
  return roots
}

function emitUpdate() {
  emit('update:answers', props.answers)
}

function triggerFileInput(qid: string) { emit('triggerFile', qid) }
function onFileInput(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  if (!props.fileLists[props.q.id]) props.fileLists[props.q.id] = []
  for (const f of input.files) props.fileLists[props.q.id].push(f)
  input.value = ''
  props.answers[props.q.id] = props.fileLists[props.q.id].map(f => f.name)
}
function removeFile(qid: string, idx: number) { emit('removeFile', qid, idx) }
function pickLocation(qid: string) { emit('pickLocation', qid) }
function openScanner(qid: string) { emit('openScanner', qid) }
function sigStart(e: MouseEvent | Touch, id: string) { emit('sigStart', e, id) }
function sigMove(e: MouseEvent | Touch, id: string) { emit('sigMove', e, id) }
function sigEnd() { emit('sigEnd') }
function sigTouchStart(e: TouchEvent, id: string) { emit('sigTouchStart', e, id) }
function sigTouchMove(e: TouchEvent, id: string) { emit('sigTouchMove', e, id) }
function clearSignature(id: string) { emit('clearSignature', id) }
function restoreSignature(id: string, canvas: HTMLCanvasElement) {
  const data = props.answers?.[id]
  if (!data || typeof data !== 'string') return
  const img = new Image()
  img.onload = () => { canvas.getContext('2d')?.drawImage(img, 0, 0) }
  img.src = data
}
function addMatrixRow(qid: string) { emit('addMatrixRow', qid) }
function removeMatrixRow(qid: string, ri: number) { emit('removeMatrixRow', qid, ri) }
</script>

<style scoped>
.q-item {
  padding: 24px 36px;
  border-bottom: 1px solid #f0f0f0;
}
.q-item:last-child { border-bottom: none; }
.q-item--layout { padding: 12px 36px; }

.q-title {
  font-size: 15px; color: #303133; margin-bottom: 14px;
  font-weight: 500; line-height: 1.5; word-break: break-word; white-space: pre-wrap;
}
.q-title :deep(.q-num) { color: #3873f6; font-weight: 600; margin-right: 4px; }
.q-title :deep(.q-score-inline) { font-size: 12px; color: #fa8c16; font-weight: 500; }
.q-title :deep(.q-req) { color: #f56c6c; margin-left: 2px; font-weight: 600; }
.q-title :deep(img) { max-width: 100%; height: auto; }
:deep(.ql-align-center) { text-align: center; }
:deep(.ql-align-right) { text-align: right; }
:deep(.ql-align-justify) { text-align: justify; }
:deep(.ql-code-block-container) { background: #f5f5f5; border-radius: 4px; padding: 10px 14px; margin: 6px 0; font-family: monospace; font-size: 13px; line-height: 1.5; overflow-x: auto; }
:deep(.ql-code-block) { white-space: pre; }
.q-desc { background: #f5f5f5; border-radius: 6px; padding: 12px 16px; margin-bottom: 14px; font-size: 14px; color: #333; line-height: 1.6; word-break: break-word; display: flex; align-items: flex-start; white-space: pre-wrap; }
.q-desc-label { font-weight: 500; color: #666; flex-shrink: 0; }

.q-type { font-size: 14px; color: #909399; }
.q-type-label {
  display: inline-block; padding: 0 8px; border-radius: 3px;
  font-size: 12px; background: #f0f2f5; color: #909399; margin-bottom: 8px;
}
.q-content { font-size: 14px; color: #606266; white-space: pre-wrap; word-break: break-word; margin-top: 8px; }

.q-input { width: 100%; min-height: 28px; }
.q-input .el-input,
.q-input .el-select,
.q-input .el-date-editor,
.q-input .el-cascader,
.q-input .el-input-number { width: 100%; }

.q-field-stack { display: flex; flex-direction: column; gap: 8px; width: 100%; }
.q-field-row { display: flex; flex-wrap: wrap; gap: 8px; width: 100%; }
.q-field-row .el-input { flex: 1; min-width: 120px; }

.q-options { display: flex; flex-direction: column; gap: 0; width: 100%; }
.q-options :deep(.el-radio), .q-options :deep(.el-checkbox) {
  display: inline-flex; align-items: center; margin-right: 0; height: auto;
  padding: 4px 0; width: 100%; white-space: normal;
}
.q-options :deep(.el-radio__label), .q-options :deep(.el-checkbox__label) {
  font-size: 14px; line-height: 1.5; text-align: left; padding-left: 4px; white-space: normal; word-break: break-word;
  display: inline-block; vertical-align: middle; overflow: hidden;
}
.q-options :deep(.el-radio__label) img,
.q-options :deep(.el-checkbox__label) img { max-width:100%; height:auto; display:block; }
.q-options :deep(.el-radio__label) p,
.q-options :deep(.el-checkbox__label) p,
.q-options :deep(.el-radio__label) div,
.q-options :deep(.el-checkbox__label) div { margin:0; }

.q-rating { padding: 4px 0; }
.q-nps { padding: 4px 0; }
.q-nps-labels { display: flex; justify-content: space-between; font-size: 12px; color: #909399; margin-bottom: 4px; }

.q-matrix {
  border: 1px solid #e0e0e0; border-radius: 6px; overflow-x: auto;
}
.q-matrix table { width: 100%; border-collapse: collapse; font-size: 13px; }
.q-matrix th, .q-matrix td { border: 1px solid #e8e8e8; padding: 8px 12px; text-align: center; }
.q-matrix th { background: #f5f6fa; font-weight: 500; color: #303133; }
.q-mtx-cnr { background: #f5f6fa; color: #909399; font-size: 12px; min-width: 60px; }
.q-mtx-label { background: #fafafa; font-weight: 500; text-align: center; white-space: nowrap; }
.q-mtx-add { padding: 8px 12px; border-top: 1px solid #e8e8e8; }

.q-file {
  border: 1px dashed #d9d9d9; border-radius: 6px; padding: 16px;
}
.q-file-list { margin-top: 8px; display: flex; flex-direction: column; gap: 4px; }
.q-file-item { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.q-file-name { color: #303133; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.q-location-result { font-size: 13px; color: #3873f6; }

.q-sig-wrap { width: 100%; }
.q-sig-canvas { border: 1px dashed #d9d9d9; border-radius: 6px; width: 100%; height: 120px; cursor: crosshair; }
.q-sig-actions { margin-top: 6px; }

.q-quill-wrap { border: 1px solid #e0e0e0; border-radius: 6px; overflow: hidden; }
.scan-code-input :deep(.el-input__wrapper) { width: 100%; }
.scan-code-btn { margin-right: -8px; height: 28px; }
</style>
