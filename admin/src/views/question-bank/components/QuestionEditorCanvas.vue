<template>
  <main class="question-editor-canvas">
    <header class="question-editor-canvas__toolbar">
      <div class="question-editor-canvas__modes" role="tablist" aria-label="题目编辑模式">
        <button v-for="item in modes" :key="item.value" type="button" :class="{ active: editorModeModel === item.value }" @click="editorModeModel = item.value">
          {{ item.label }}
        </button>
      </div>
      <span v-if="active && form.id" class="question-editor-canvas__id">题目 ID {{ form.id }}</span>
    </header>

    <div v-if="!active" class="question-editor-canvas__empty">
      <el-empty description="从左侧选择题目，或新增一道题目" :image-size="80" />
      <el-button v-if="canAdd" type="primary" :icon="Plus" @click="emit('add-question')">新增题目</el-button>
    </div>

    <div v-else-if="editorModeModel === 'json'" class="question-editor-canvas__json">
      <el-alert title="高级 JSON 会直接影响题目结构，请确认格式后保存" type="warning" :closable="false" show-icon />
      <el-input v-model="form.schema" type="textarea" :rows="24" :readonly="readonly" placeholder="题目 schema JSON" />
      <div class="question-editor-canvas__json-actions">
        <el-button :disabled="readonly" @click="formatSchema">格式化 JSON</el-button>
        <el-button type="primary" plain :disabled="readonly" @click="applyJsonToVisual">同步到可视化</el-button>
      </div>
    </div>

    <div v-else-if="editorModeModel === 'preview'" class="question-editor-canvas__preview">
      <div class="question-editor-canvas__paper">
        <QuestionPreview :q="visualQuestion" />
      </div>
    </div>

    <div v-else class="question-editor-canvas__scroll">
      <div class="question-editor-canvas__paper question-editor-canvas__paper--edit" :class="{ readonly }">
        <div class="question-editor-canvas__label">题目标题</div>
        <div class="bank-rich-editor bank-rich-title">
          <QuillEditor
            v-model:content="visualQuestion.title"
            content-type="html"
            :read-only="readonly"
            :options="richTitleEditorOptions"
            @update:content="syncPlainTitleFromVisual"
            @ready="(quill: any) => bindCompactRichEditor(quill, 'title')"
          />
        </div>

        <section v-if="usesOptions(visualQuestion.type)" class="question-editor-canvas__section">
          <div class="question-editor-canvas__section-title">选项</div>
          <div class="question-editor-canvas__rows">
            <div v-for="(opt, idx) in visualQuestion.props.options" :key="idx" class="question-editor-canvas__option-row">
              <span class="question-editor-canvas__index">{{ idx + 1 }}</span>
              <div class="bank-rich-editor bank-rich-option">
                <QuillEditor
                  v-model:content="opt.label"
                  content-type="html"
                  :read-only="readonly"
                  :options="richOptionEditorOptions"
                  @ready="(quill: any) => bindCompactRichEditor(quill, 'option', idx)"
                />
              </div>
              <el-input v-model="opt.value" :readonly="readonly" placeholder="选项值" />
              <el-button :icon="Delete" circle plain type="danger" :disabled="readonly" title="删除选项" @click="removeVisualOption(idx)" />
            </div>
          </div>
          <el-button size="small" :icon="Plus" :disabled="readonly" @click="addVisualOption">添加选项</el-button>
        </section>

        <section v-if="usesFields(visualQuestion.type)" class="question-editor-canvas__section">
          <div class="question-editor-canvas__section-title">字段</div>
          <div class="question-editor-canvas__rows">
            <div v-for="(field, idx) in visualQuestion.props.fields" :key="idx" class="question-editor-canvas__field-row">
              <span class="question-editor-canvas__index">{{ idx + 1 }}</span>
              <el-input v-model="field.label" :readonly="readonly" placeholder="字段名称" />
              <el-input v-model="field.placeholder" :readonly="readonly" placeholder="提示语" />
              <el-button :icon="Delete" circle plain type="danger" :disabled="readonly" title="删除字段" @click="removeVisualField(idx)" />
            </div>
          </div>
          <el-button size="small" :icon="Plus" :disabled="readonly" @click="addVisualField">添加字段</el-button>
        </section>

        <section v-if="usesMatrix(visualQuestion.type)" class="question-editor-canvas__section">
          <div class="question-editor-canvas__section-title">矩阵</div>
          <div class="question-editor-canvas__matrix">
            <div>
              <strong>行</strong>
              <div v-for="(row, idx) in visualQuestion.props.rows" :key="idx" class="question-editor-canvas__matrix-row">
                <el-input v-model="row.title" :readonly="readonly" placeholder="行标题" />
                <el-button :icon="Delete" circle plain type="danger" :disabled="readonly" @click="removeMatrixRow(idx)" />
              </div>
              <el-button size="small" :icon="Plus" :disabled="readonly" @click="addMatrixRow">添加行</el-button>
            </div>
            <div>
              <strong>列</strong>
              <div v-for="(column, idx) in visualQuestion.props.columns" :key="idx" class="question-editor-canvas__matrix-row">
                <el-input v-model="column.title" :readonly="readonly" placeholder="列标题" />
                <el-button :icon="Delete" circle plain type="danger" :disabled="readonly" @click="removeMatrixColumn(idx)" />
              </div>
              <el-button size="small" :icon="Plus" :disabled="readonly" @click="addMatrixColumn">添加列</el-button>
            </div>
          </div>
        </section>

        <section class="question-editor-canvas__section question-editor-canvas__section--preview">
          <div class="question-editor-canvas__section-title">实时预览</div>
          <QuestionPreview :q="visualQuestion" />
        </section>
      </div>
    </div>

    <QuestionRichEditorDialog :dialog="richEditDialog" :title="richEditTitle" :options="richFullEditorOptions" :confirm="confirmRichFullEdit" />
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Delete, Plus } from '@element-plus/icons-vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import QuestionPreview from '../../survey/formkit/QuestionPreview.vue'
import type { FormKitQuestion } from '../../formkit/shared/types'
import QuestionRichEditorDialog from './QuestionRichEditorDialog.vue'

type EditorMode = 'visual' | 'preview' | 'json'
type RichEditTarget = 'title' | 'option'

const props = defineProps<{
  active: boolean
  canAdd: boolean
  readonly: boolean
  form: Record<string, any>
  visualQuestion: FormKitQuestion
  editorMode: EditorMode
  richTitleEditorOptions: Record<string, any>
  richOptionEditorOptions: Record<string, any>
  richFullEditorOptions: Record<string, any>
  richEditDialog: { visible: boolean; target: RichEditTarget; optionIndex: number; content: string }
  richEditTitle: string
  bindCompactRichEditor: (quill: any, target: RichEditTarget, optionIndex?: number) => void
  syncPlainTitleFromVisual: () => void
  usesOptions: (type: string) => boolean
  usesFields: (type: string) => boolean
  usesMatrix: (type: string) => boolean
  addVisualOption: () => void
  removeVisualOption: (index: number) => void
  addVisualField: () => void
  removeVisualField: (index: number) => void
  addMatrixRow: () => void
  removeMatrixRow: (index: number) => void
  addMatrixColumn: () => void
  removeMatrixColumn: (index: number) => void
  formatSchema: () => void
  applyJsonToVisual: () => void
  confirmRichFullEdit: () => void
}>()

const emit = defineEmits<{
  (event: 'update:editorMode', value: EditorMode): void
  (event: 'add-question'): void
}>()

const modes: Array<{ value: EditorMode; label: string }> = [
  { value: 'visual', label: '编辑' }, { value: 'preview', label: '预览' }, { value: 'json', label: 'JSON' },
]

const editorModeModel = computed({
  get: () => props.editorMode,
  set: (value: EditorMode) => emit('update:editorMode', value),
})
</script>

<style scoped>
.question-editor-canvas { min-width:0; height:100%; display:flex; flex-direction:column; background:#f5f7fa; }
.question-editor-canvas__toolbar { min-height:48px; display:flex; align-items:center; justify-content:space-between; gap:12px; padding:0 14px; border-bottom:1px solid #e4e7ed; background:#fff; }
.question-editor-canvas__modes { display:flex; gap:3px; padding:3px; border-radius:6px; background:#f2f4f7; }
.question-editor-canvas__modes button { min-width:58px; height:28px; border:0; border-radius:4px; background:transparent; cursor:pointer; color:#667085; }
.question-editor-canvas__modes button.active { background:#fff; color:#2563eb; font-weight:600; box-shadow:0 1px 3px rgba(31,41,55,.12); }
.question-editor-canvas__id { color:#98a2b3; font-size:12px; }
.question-editor-canvas__scroll,.question-editor-canvas__preview,.question-editor-canvas__json { flex:1; min-height:0; overflow:auto; padding:22px 28px; }
.question-editor-canvas__paper { max-width:760px; margin:0 auto; border:1px solid #dfe3e8; border-radius:7px; background:#fff; }
.question-editor-canvas__paper--edit { padding:20px; }
.question-editor-canvas__paper--edit.readonly { opacity:.86; }
.question-editor-canvas__empty { flex:1; display:flex; flex-direction:column; align-items:center; justify-content:center; gap:8px; }
.question-editor-canvas__label,.question-editor-canvas__section-title { margin-bottom:8px; color:#475467; font-size:12px; font-weight:650; }
.question-editor-canvas__section { margin-top:18px; padding-top:18px; border-top:1px solid #eef0f3; }
.question-editor-canvas__rows { display:flex; flex-direction:column; gap:8px; margin-bottom:10px; }
.question-editor-canvas__option-row,.question-editor-canvas__field-row { display:grid; grid-template-columns:28px minmax(0,1fr) 120px 34px; gap:8px; align-items:start; }
.question-editor-canvas__field-row { grid-template-columns:28px minmax(0,1fr) minmax(0,1fr) 34px; }
.question-editor-canvas__index { display:flex; align-items:center; justify-content:center; width:24px; height:24px; margin-top:6px; border-radius:50%; background:#eff6ff; color:#2563eb; font-size:12px; }
.question-editor-canvas__matrix { display:grid; grid-template-columns:1fr 1fr; gap:16px; }
.question-editor-canvas__matrix strong { display:block; margin-bottom:8px; color:#667085; font-size:12px; }
.question-editor-canvas__matrix-row { display:grid; grid-template-columns:minmax(0,1fr) 34px; gap:8px; margin-bottom:8px; }
.question-editor-canvas__section--preview { padding:16px; border:1px solid #e4e7ed; border-radius:7px; background:#f8fafc; }
.question-editor-canvas__json { display:flex; flex-direction:column; gap:12px; }
.question-editor-canvas__json :deep(.el-textarea__inner) { font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size:12px; line-height:1.6; }
.question-editor-canvas__json-actions { display:flex; justify-content:flex-end; gap:8px; }
.bank-rich-editor { min-width:0; overflow:hidden; border:1px solid #dcdfe6; border-radius:7px; background:#fff; }
.bank-rich-editor:focus-within { border-color:#409eff; box-shadow:0 0 0 2px rgba(64,158,255,.1); }
.bank-rich-editor :deep(.ql-toolbar) { display:flex; flex-wrap:wrap; max-height:0; overflow:hidden; padding:0 8px; border:0; opacity:0; transition:max-height .18s,padding .18s,opacity .18s; }
.bank-rich-editor:focus-within :deep(.ql-toolbar) { max-height:74px; padding:6px 8px; border-bottom:1px solid #eef0f3; opacity:1; }
.bank-rich-editor :deep(.ql-container) { border:0; font-family:inherit; }
.bank-rich-title :deep(.ql-editor) { min-height:54px; padding:12px; font-size:16px; font-weight:600; }
.bank-rich-option :deep(.ql-editor) { min-height:34px; padding:7px 9px; font-size:13px; }
.bank-rich-editor :deep(.ql-editor.ql-blank:has(p:not(:empty))::before) { display:none; }
@media(max-width:1100px){ .question-editor-canvas__scroll,.question-editor-canvas__preview,.question-editor-canvas__json{padding:16px;} }
</style>
