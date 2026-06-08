<template>
  <div class="question-preview" :class="{ editing, 'is-hidden': q.defaultHidden }">
    <!-- 题目标题 -->
    <div v-if="editing" class="edit-title-line">
      <el-tag size="small" :type="tagType(q.type)||undefined" style="flex-shrink:0">{{ typeName(q.type) }}</el-tag>
      <div
        ref="titleRef"
        class="title-editable"
        contenteditable
        placeholder="输入标题"
        @blur="onTitleBlur"
        @keydown.enter.prevent="onTitleEnter"
      >{{ q.title }}</div>
      <el-tag v-if="q.required" type="danger" size="small" effect="plain" style="flex-shrink:0">必填</el-tag>
      <span class="preview-actions">
        <el-tooltip content="逻辑设置" placement="top">
          <el-button text size="small" @click.stop="$emit('open-logic')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="复制" placement="top">
          <el-button text size="small" @click.stop="$emit('copy')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="删除" placement="top">
          <el-button text size="small" type="danger" @click.stop="$emit('remove')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="上传题库" placement="top">
          <el-button text size="small" @click.stop="$emit('upload-bank')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          </el-button>
        </el-tooltip>
      </span>
    </div>
    <div v-else class="preview-header">
      <el-tag size="small" :type="tagType(q.type)||undefined">{{ typeName(q.type) }}</el-tag>
      <span class="preview-title">{{ q.title }}</span>
      <el-tag v-if="q.required" type="danger" size="small" effect="plain">必填</el-tag>
    </div>

    <!-- 说明 -->
    <div v-if="editing && q.showDescription === true" class="edit-desc-line">
      <div
        ref="descRef"
        class="desc-editable"
        contenteditable
        placeholder="添加说明"
        @blur="onDescBlur"
      >{{ q.description }}</div>
    </div>
    <div v-else-if="q.description && q.showDescription === true" class="preview-desc">{{ q.description }}</div>

    <!-- 媒体 -->
    <div v-if="q.mediaUrl" class="preview-media" :style="{ textAlign: q.mediaAlign || 'center' }">
      <img v-if="q.mediaType==='image'" :src="q.mediaUrl" :style="{ maxWidth: q.mediaWidth ? q.mediaWidth+'px' : '100%', maxHeight:'300px' }" class="media-preview" />
      <video v-else-if="q.mediaType==='video'" :src="q.mediaUrl" controls :style="{ maxWidth: q.mediaWidth ? q.mediaWidth+'px' : '100%', maxHeight:'300px' }" class="media-preview" />
      <audio v-else-if="q.mediaType==='audio'" :src="q.mediaUrl" controls class="media-preview" style="width:100%;max-width:400px" />
    </div>

    <!-- 选项画布编辑 (选择题) -->
      <div v-if="editing && isChoiceType" class="edit-options-area" :style="optionGrid(q)">
      <div v-for="(o, i) in (q.props?.options||[])" :key="o.value || i" class="edit-option-row" @click.stop="emit('select-option', i)">
        <span class="option-icon">{{ fieldIcon }}</span>
        <div
          class="option-editable"
          contenteditable
          @blur="e => onOptionLabelBlur(i, e)"
          @keydown.enter.prevent="(e: any) => e.target.blur()"
        >{{ o.label }}</div>
        <el-button text size="small" type="danger" class="opt-del-btn" @click="removeOption(i)">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
        <el-tag v-if="o.value===examCorrectAnswer" size="small" type="success" effect="dark" style="margin-left:4px;">✓</el-tag>
      </div>
    </div>
    <el-button v-if="editing && isChoiceType" text size="small" type="primary" @click="addOption" class="add-option-btn">添加选项</el-button>

    <!-- 矩阵题行/列画布编辑 -->
    <div v-if="editing && isMatrixType" class="edit-options-area">
      <div class="matrix-edit-header" style="display:flex;gap:8px;margin-bottom:4px;font-size:12px;color:#999;padding:0 4px">
        <span>行</span>
      </div>
      <div v-for="(r, ri) in (q.props?.rows||[])" :key="ri" class="matrix-edit-row" @click.stop="emit('select-option', ri)" style="display:flex;align-items:center;gap:4px;margin-bottom:2px;border-radius:4px;padding:2px 4px;cursor:pointer">
        <span class="matrix-edit-order" style="font-size:11px;color:#999;width:16px;flex-shrink:0">{{ ri+1 }}.</span>
        <div class="option-editable" style="flex:0 0 auto;min-width:60px" contenteditable @blur="e => onMatrixRowLabelBlur(ri, e)" @keydown.enter.prevent="(e:any)=>e.target.blur()">{{ typeof r==='string'?r:r.title }}</div>
        <template v-if="q.type!=='matrixFillBlank'">
          <span class="matrix-col-label" v-for="col in (q.props?.columns||[])" :key="col.id||col.title||col" style="font-size:12px;color:#666;padding:2px 8px;margin:0 2px;background:#f5f5f5;border-radius:3px">{{ typeof col==='string'?col:(col.title||col.label||'') }}</span>
        </template>
        <template v-else>
          <el-input disabled size="small" placeholder="填空" style="width:100px;pointer-events:none;margin-left:8px" />
        </template>
        <el-button text size="small" type="danger" class="opt-del-btn" @click.stop="removeMatrixRow(ri)" style="margin-left:auto">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
      </div>
      <div class="matrix-edit-add" style="display:flex;align-items:center;gap:4px;margin-top:4px">
        <el-input v-model="newMatrixRowName" size="small" placeholder="新行" style="width:100px" @keyup.enter="addMatrixRow" />
        <el-button text size="small" type="primary" @click="addMatrixRow">+ 添加行</el-button>
      </div>
    </div>

    <!-- 单项填空/多行文本 → 直接渲染输入框 -->
    <div v-if="editing && isSingleInput" class="edit-options-area" style="padding:4px 8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none">
        <el-input v-if="q.type==='input'" size="small" :placeholder="(q.props?.options?.[0]?.placeholder)||'请输入'" disabled />
        <el-input v-else size="small" type="textarea" :placeholder="(q.props?.options?.[0]?.placeholder)||'请输入'" :rows="2" disabled />
      </div>
    </div>

    <!-- 评分 / NPS 输入（点击弹出评分设置） -->
    <div v-if="editing && (q.type==='rating'||q.type==='nps')" class="edit-options-area" style="padding:8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none;display:flex;align-items:center;gap:8px">
        <span v-if="q.type==='rating'" class="rate-icons-preview" v-for="i in (q.props?.maxRating || 5)" :key="i" v-html="rateIconSvg(q.props?.icon || 'star')"></span>
        <template v-else>
          <span style="font-size:11px;color:#909399">0</span>
          <el-rate :model-value="0" disabled :max="10" />
          <span style="font-size:11px;color:#909399">10</span>
        </template>
      </div>
    </div>

    <!-- 文件上传（画布中显示上传框） -->
    <div v-if="editing && q.type==='file'" class="edit-options-area" style="padding:8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div class="file-upload-placeholder" style="border:1px dashed #d9d9d9;border-radius:6px;padding:16px 12px;text-align:center;color:#999;pointer-events:none">
        <svg viewBox="0 0 1024 1024" width="24" height="24" fill="currentColor" style="display:block;margin:0 auto 6px"><path d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7z"/></svg>
        <span>上传文件</span>
        <div style="font-size:11px;margin-top:4px">点击设置上传参数</div>
      </div>
    </div>

    <!-- 其他填空题子字段预览 (signature/scanCode) -->
    <div v-if="editing && isInputSubFieldType && !isSingleInput && q.type!=='file'" class="edit-options-area">
      <div v-for="(o, i) in (q.props?.options||[])" :key="i" class="edit-option-row input-field-row" @click.stop="emit('select-option', i)">
        <span class="option-icon">▸</span>
        <span class="input-field-label">{{ o.label || '字段' }}</span>
        <el-input size="small" :placeholder="o.placeholder||'请输入'" disabled style="flex:1;max-width:200px;pointer-events:none" />
        <el-tag v-if="o.calculate" size="small" type="warning" effect="plain" style="margin:0 4px">公式</el-tag>
        <el-tag v-if="o.dataType" size="small" effect="plain" style="margin:0 4px">{{ o.dataType }}</el-tag>
        <el-button text size="small" type="danger" class="opt-del-btn" @click.stop="removeOption(i)">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
      </div>
    </div>
    <el-button v-if="editing && isInputSubFieldType && !isSingleInput && q.type!=='file'" text size="small" type="primary" @click="addOption" class="add-option-btn">添加字段</el-button>

    <!-- 子字段画布编辑 (multiInput / hInput) -->
    <div v-if="editing && hasFields" class="edit-options-area">
      <div v-for="(f, i) in (q.props?.fields||[])" :key="i" class="edit-option-row input-field-row" @click.stop="emit('select-option', i)">
        <span class="option-icon">⊞</span>
        <span class="input-field-label">{{ f.label || '字段' }}</span>
        <el-input size="small" :placeholder="f.placeholder||'请输入'" disabled style="flex:1;max-width:200px;pointer-events:none" />
        <el-tag v-if="f.dataType" size="small" effect="plain" style="margin:0 4px">{{ f.dataType }}</el-tag>
        <el-button text size="small" type="danger" class="opt-del-btn" @click.stop="removeField(i)">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
      </div>
    </div>
    <el-button v-if="editing && hasFields" text size="small" type="primary" @click="addField" class="add-option-btn">添加字段</el-button>

    <!-- 预览体 -->
    <div v-if="!editing" class="preview-body">
      <el-input v-if="['input','text'].includes(q.type)" :placeholder="q.placeholder||'请输入'" disabled />
      <el-input v-else-if="q.type==='multiInput'" placeholder="多项填空" disabled />
      <el-input v-else-if="q.type==='hInput'" placeholder="横向填空" disabled />
      <el-input v-else-if="q.type==='scanCode'" placeholder="扫码" disabled>
        <template #prefix><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg></template>
      </el-input>
      <div v-else-if="q.type==='signature'" class="preview-plain"><el-button text>签名</el-button></div>
      <el-input v-else-if="q.type==='textarea'" type="textarea" :placeholder="q.placeholder||'请输入'" :rows="2" disabled />
      <el-input-number v-else-if="q.type==='number'" :model-value="0" disabled style="width:100%;--el-input-width:100%" />
      <el-radio-group v-else-if="q.type==='radio'" :model-value="''" disabled class="preview-options preview-radio-group" :style="optionGrid(q)">
        <el-radio v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value">{{ o.label }}</el-radio>
      </el-radio-group>
      <el-checkbox-group v-else-if="q.type==='checkbox'" :model-value="[]" disabled class="preview-options preview-checkbox-group" :style="optionGrid(q)">
        <el-checkbox v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value">{{ o.label }}</el-checkbox>
      </el-checkbox-group>
      <el-select v-else-if="q.type==='select'" :model-value="''" disabled placeholder="请选择" style="width:100%">
        <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value" :label="o.label" />
      </el-select>
      <el-select v-else-if="q.type==='picker'" :model-value="''" disabled placeholder="请选择" style="width:100%">
        <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value" :label="o.label" />
      </el-select>
      <el-cascader v-else-if="q.type==='cascade'" :model-value="[]" disabled placeholder="请选择" style="width:100%" :options="q.props?.options||[]" />
      <el-radio-group v-else-if="q.type==='judge'" :model-value="''" disabled class="preview-options preview-radio-group">
        <el-radio value="true">对</el-radio>
        <el-radio value="false">错</el-radio>
      </el-radio-group>
      <span v-else-if="q.type==='rating'" class="preview-rate-icons" style="display:inline-flex;gap:4px">
        <span v-for="i in (q.props?.maxRating || 5)" :key="i" v-html="rateIconSvg(q.props?.icon || 'star')"></span>
      </span>
      <el-date-picker v-else-if="q.type==='date'" :model-value="''" disabled type="date" placeholder="选择日期" style="width:100%" />
      <el-time-picker v-else-if="q.type==='time'" :model-value="''" disabled placeholder="选择时间" style="width:100%" />
      <el-switch v-else-if="q.type==='switch'" disabled />
      <el-divider v-else-if="q.type==='divider'" style="margin:4px 0" />
      <div v-else-if="q.type==='description'" class="preview-plain">{{ q.description || '说明文字' }}</div>
      <div v-else-if="q.type==='file'" style="border:1px dashed #d9d9d9;border-radius:6px;padding:12px;text-align:center;color:#999;margin:4px 0">
        <svg viewBox="0 0 1024 1024" width="22" height="22" fill="currentColor" style="display:block;margin:0 auto 4px"><path d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7z"/></svg>
        <span style="font-size:13px">上传文件</span>
      </div>
      <div v-else-if="q.type==='location'" class="preview-plain"><el-button text><svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M512 64C367.2 64 248 183.2 248 328c0 163.2 233.6 524.8 252 551.2 3.2 4.8 8 7.2 12 7.2s8.8-2.4 12-7.2C542.4 852.8 776 491.2 776 328 776 183.2 656.8 64 512 64z m0 400c-39.2 0-72-32.8-72-72s32.8-72 72-72 72 32.8 72 72-32.8 72-72 72z"/></svg>选择位置</el-button></div>
      <div v-else-if="q.type==='signature'" class="preview-plain"><el-button text>签名</el-button></div>
      <el-input v-else-if="q.type==='phone'" placeholder="手机号" disabled />
      <el-input v-else-if="q.type==='email'" placeholder="邮箱地址" disabled />
      <el-input v-else-if="q.type==='idCard'" placeholder="身份证号" disabled />
      <el-input v-else-if="q.type==='password'" type="password" placeholder="密码" disabled />
      <el-date-picker v-else-if="q.type==='dateRange'" :model-value="['','']" disabled type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width:100%" />
      <div v-else-if="q.type==='matrixRadio'" class="preview-matrix">
        <div class="matrix-row" v-for="r in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><span class="matrix-label">{{ typeof r==='string'?r:r.title }}</span><el-radio-group disabled :model-value="''"><el-radio v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)" :value="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</el-radio></el-radio-group></div>
      </div>
      <div v-else-if="q.type==='matrixCheckbox'" class="preview-matrix">
        <div class="matrix-row" v-for="r in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><span class="matrix-label">{{ typeof r==='string'?r:r.title }}</span><el-checkbox-group disabled :model-value="[]"><el-checkbox v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)" :value="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</el-checkbox></el-checkbox-group></div>
      </div>
      <div v-else-if="q.type==='matrixFillBlank'" class="preview-matrix">
        <div class="matrix-row" v-for="r in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><span class="matrix-label">{{ typeof r==='string'?r:r.title }}</span><el-input disabled placeholder="填空" style="width:120px" /></div>
      </div>
      <div v-else-if="q.type==='matrixAuto'" class="preview-matrix">
        <el-button text size="small">+ 添加行</el-button>
        <div class="matrix-row" style="margin-top:4px" v-for="col in (q.props?.columns||[]).slice(0,1)" :key="col.label||col.id||col"><span class="matrix-label">{{ col.label||col }}</span><el-input disabled placeholder="值" style="width:120px" /></div>
      </div>
      <div v-else-if="q.type==='questionSet'" class="preview-plain">问题组（内部题）</div>
      <div v-else-if="q.type==='pagination'" class="preview-plain">—— 分页 ——</div>
      <div v-else-if="q.type==='user'" class="preview-plain"><el-button text>选择成员</el-button></div>
      <div v-else-if="q.type==='dept'" class="preview-plain"><el-button text>选择部门</el-button></div>
      <div v-else-if="q.type==='richText'" class="preview-plain"><div class="rich-text-placeholder">富文本内容</div></div>
      <div v-else-if="q.type==='autopop'" class="preview-plain"><el-input disabled placeholder="自动填充" /></div>
      <div v-else-if="q.type==='nps'" class="preview-nps">
        <div class="nps-labels"><span>0</span><span>10</span></div>
        <el-rate :model-value="0" disabled :max="10" show-score score-template="{value}" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{ q: any; editing?: boolean }>()
const emit = defineEmits<{
  'update:title': [v: string]
  'update:description': [v: string]
  remove: []
  'open-logic': []
  copy: []
  'upload-bank': []
  'select-option': [idx: number]
}>()

const titleRef = ref<HTMLElement>()
const descRef = ref<HTMLElement>()

const hasOptions = computed(() => ['select','radio','checkbox','picker','judge','cascade','matrixRadio','matrixCheckbox'].includes(props.q.type))
const isInputSubFieldType = computed(() => ['input','textarea','signature','scanCode','file'].includes(props.q.type))
const isSingleInput = computed(() => ['input','textarea'].includes(props.q.type))
const hasFields = computed(() => ['multiInput','hInput'].includes(props.q.type))
const isChoiceType = computed(() => ['select','radio','checkbox','picker','judge','cascade'].includes(props.q.type))
const isMatrixChoiceType = computed(() => ['matrixRadio','matrixCheckbox'].includes(props.q.type))
const isMatrixType = computed(() => ['matrixRadio','matrixCheckbox','matrixFillBlank'].includes(props.q.type))

const newMatrixRowName = ref('')
function ensureRows() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.rows)) props.q.props.rows = []
}
function onMatrixRowLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureRows()
  const existing = props.q.props.rows[idx] || {}
  props.q.props.rows[idx] = typeof existing === 'string' ? text : { ...existing, title: text }
}
function addMatrixRow() {
  ensureRows()
  const name = newMatrixRowName.value.trim() || `行${props.q.props.rows.length + 1}`
  props.q.props.rows.push({ title: name, id: Date.now() + '_' + Math.random().toString(36).slice(2,6), width: 150 })
  newMatrixRowName.value = ''
}
function removeMatrixRow(idx: number) {
  ensureRows()
  props.q.props.rows.splice(idx, 1)
}

const fieldIcon = computed(() => {
  if (props.q.type === 'checkbox') return '☐'
  if (props.q.type === 'judge') return '○'
  if (isChoiceType.value) return '○'
  return '▸'
})

const examCorrectAnswer = computed(() => props.q.examCorrectAnswer)

function onTitleBlur(e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (text) emit('update:title', text)
}
function onTitleEnter(e: KeyboardEvent) {
  (e.target as HTMLElement).blur()
}
function onDescBlur(e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  emit('update:description', text)
}
/** 直接修改父级 reactive 数据，绕过多层 emit 闭包问题 */
function ensureOpts() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.options)) props.q.props.options = []
}
function onOptionLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureOpts()
  const existing = props.q.props.options[idx] || {}
  props.q.props.options[idx] = { ...existing, label: text }
}
function removeOption(idx: number) {
  ensureOpts()
  props.q.props.options.splice(idx, 1)
}
function addOption() {
  ensureOpts()
  const n = props.q.props.options.length + 1
  const letter = n <= 26 ? String.fromCharCode(64 + n) : `opt${n}`
  props.q.props.options.push({ label: `选项${n}`, value: letter })
}

function ensureFields() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.fields)) props.q.props.fields = []
}
function onFieldLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureFields()
  const existing = props.q.props.fields[idx] || {}
  props.q.props.fields[idx] = { ...existing, label: text }
}
function addField() {
  ensureFields()
  props.q.props.fields.push({ label: '字段' + (props.q.props.fields.length + 1), placeholder: '' })
}
function removeField(idx: number) {
  ensureFields()
  props.q.props.fields.splice(idx, 1)
}

const typeName = (t: string) => {
  const map: Record<string, string> = {
    input: '单行文本', text: '文本', textarea: '多行', number: '数字',
    select: '下拉', radio: '单选', checkbox: '多选', picker: '选择器',
    cascade: '级联选择', judge: '判断',
    multiInput: '多项填空', hInput: '横向填空', scanCode: '扫码',
    signature: '签名', file: '上传文件',
    rating: '评分', nps: 'NPS评分',
    date: '日期', time: '时间', dateRange: '日期范围',
    phone: '手机', email: '邮箱', idCard: '身份证', password: '密码',
    switch: '开关', location: '地理位置',
    matrixRadio: '矩阵单选', matrixCheckbox: '矩阵多选',
    matrixFillBlank: '矩阵填空', matrixAuto: '表格自增',
    divider: '分割线', description: '说明',
    questionSet: '问题组', pagination: '分页',
    user: '成员', dept: '部门', richText: '富文本',
    autopop: '自动填充',
  }
  return map[t] || t
}
const tagType = (t: string) => {
  const map: Record<string, string> = {
    input: '', text: '', textarea: '', number: '', multiInput: '', hInput: '', scanCode: '',
    select: 'warning', radio: 'warning', checkbox: 'warning', picker: 'warning', judge: 'warning', cascade: 'warning',
    rating: 'success', nps: 'success',
    file: '', signature: '', location: '',
    phone: '', email: '', idCard: '', password: '', switch: '',
    matrixRadio: '', matrixCheckbox: '', matrixFillBlank: '', matrixAuto: '',
    divider: 'info', description: 'info',
    questionSet: 'info', pagination: 'info',
    user: '', dept: '', richText: '', autopop: '',
    date: '', time: '', dateRange: '',
  }
  return map[t] || ''
}
function rateIconSvg(icon: string) {
  const gray = '#c0c4cc'
  if (icon === 'heart') {
    return `<svg viewBox="0 0 24 24" width="18" height="18" fill="${gray}"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>`
  }
  if (icon === 'smiley') {
    return `<svg viewBox="0 0 24 24" width="18" height="18" fill="${gray}"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm5 8h-2v-2h2v2zm-6 0H9v-2h2v2zm1 6c-2.33 0-4.31-1.46-5.11-3.5h10.22c-.8 2.04-2.78 3.5-5.11 3.5z"/></svg>`
  }
  return `<svg viewBox="0 0 24 24" width="18" height="18" fill="${gray}"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg>`
}
function optionGrid(q: any) {
  const cols = q.optionLayout
  if (!cols || cols <= 1) return {}
  return { display: 'grid', gridTemplateColumns: `repeat(${cols}, 1fr)`, gap: '4px' }
}
</script>

<style scoped>
.question-preview {
  padding: 6px 0;
}
.question-preview.editing {
  padding: 8px;
  border: 1px dashed #fb454c;
  border-radius: 8px;
  background: #fff;
}
.question-preview.is-hidden {
  opacity: 0.4;
  filter: grayscale(0.6);
  position: relative;
}
.question-preview.is-hidden::after {
  content: '已隐藏';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  color: #909399;
  background: rgba(255,255,255,0.9);
  padding: 2px 10px;
  border-radius: 4px;
  z-index: 1;
  pointer-events: none;
}
.preview-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.preview-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.preview-actions {
  display: flex; flex-shrink: 0; margin-left: auto;
  background: #f5f5f5; border-radius: 5px; overflow: hidden; font-size: 0;
}
.preview-actions :deep(.el-button) { padding: 2px 3px; }
.preview-desc {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}
.preview-body {
  min-height: 28px;
}
.preview-media {
  margin-bottom: 8px;
}
.media-preview {
  border-radius: 6px;
  border: 1px solid #eee;
}
.preview-options {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}
.preview-radio-group,
.preview-checkbox-group {
  align-items: flex-start;
  text-align: left;
}
.preview-plain {
  font-size: 13px;
  color: #606266;
  padding: 4px 0;
}
.preview-matrix {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.matrix-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.matrix-label {
  font-size: 12px;
  color: #606266;
  min-width: 40px;
}
.preview-nps {
  padding: 4px 0;
}
.preview-rate-icons,
.rate-icons-preview {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.preview-rate-icons span,
.rate-icons-preview {
  line-height: 1;
}
.nps-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #909399;
  margin-bottom: 2px;
}
.rich-text-placeholder {
  min-height:60px; border:1px dashed #ddd; border-radius:4px;
  display:flex; align-items:center; justify-content:center;
  color:#ccc; font-size:13px;
}

/* ======= 编辑模式 ======= */
.edit-title-line {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}
.title-editable {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  padding: 4px 6px;
  border: 1px solid transparent;
  border-radius: 4px;
  outline: none;
  min-width: 0;
  line-height: 1.5;
}
.title-editable:focus {
  border-color: #fb454c;
  background: #fff;
}
.title-editable:empty::before {
  content: attr(placeholder);
  color: #ccc;
}
.edit-desc-line {
  margin-bottom: 6px;
}
.desc-editable {
  font-size: 12px;
  color: #909399;
  padding: 3px 6px;
  border: 1px solid transparent;
  border-radius: 4px;
  outline: none;
  min-height: 20px;
}
.desc-editable:focus {
  border-color: #ddd;
  background: #fff;
}
.desc-editable:empty::before {
  content: attr(placeholder);
  color: #ddd;
}
.edit-options-area {
  margin: 8px 0;
  padding: 8px;
  background: #fafafa;
  border-radius: 6px;
}
.edit-option-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  padding: 3px 4px;
  border-radius: 4px;
  transition: background 0.15s;
}
.edit-option-row:hover {
  background: #fff;
}
.option-icon {
  flex-shrink: 0;
  font-size: 14px;
  color: #666;
  width: 16px;
  text-align: center;
}
.option-editable {
  flex: 1;
  font-size: 13px;
  color: #303133;
  padding: 3px 6px;
  border: 1px solid transparent;
  border-radius: 4px;
  outline: none;
  min-width: 0;
  line-height: 1.5;
}
.option-editable:focus {
  border-color: #fb454c;
  background: #fff;
}
.opt-del-btn {
  opacity: 0.3;
  transition: opacity 0.15s;
}
.edit-option-row:hover .opt-del-btn {
  opacity: 1;
}
.add-option-btn {
  margin-top: 6px;
}
</style>
