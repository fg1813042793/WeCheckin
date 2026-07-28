<template>
  <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑题目' : '新增题目'" width="900px" :close-on-click-modal="false" class="question-edit-dialog">
    <el-form label-position="top" class="question-edit-form">
      <el-form-item label="题目标题">
        <div class="bank-rich-editor bank-rich-title">
          <QuillEditor
            v-model:content="visualQuestion.title"
            content-type="html"
            :options="richTitleEditorOptions"
            @update:content="syncPlainTitleFromVisual"
            @ready="(quill: any) => bindCompactRichEditor(quill, 'title')"
          />
        </div>
      </el-form-item>
      <el-form-item label="题型">
        <el-select v-model="form.type" placeholder="选择题型" filterable style="width:100%" @change="handleTypeChange">
          <el-option v-for="item in typeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="分类">
        <el-select v-model="form.category" placeholder="选择或输入分类" filterable allow-create clearable style="width:100%">
          <el-option v-for="item in categories" :key="item" :label="item" :value="item" />
        </el-select>
      </el-form-item>
      <el-form-item label="标签">
        <el-input v-model="form.tags" placeholder="多个标签用逗号分隔" />
      </el-form-item>
      <el-form-item label="题目配置">
        <el-tabs v-model="editorModeModel" class="schema-tabs" @tab-change="handleEditorModeChange">
          <el-tab-pane label="可视化编辑" name="visual">
            <div class="visual-editor">
              <div class="visual-main">
                <div class="visual-section">
                  <div class="visual-section-title">基础设置</div>
                  <div class="visual-grid">
                    <div class="visual-field">
                      <span>必填</span>
                      <el-switch v-model="visualQuestion.required" />
                    </div>
                    <div class="visual-field">
                      <span>只读</span>
                      <el-switch v-model="visualQuestion.readOnly" />
                    </div>
                    <div class="visual-field wide">
                      <span>提示语</span>
                      <el-input v-model="visualQuestion.placeholder" placeholder="请输入提示语" />
                    </div>
                    <div class="visual-field wide">
                      <span>说明</span>
                      <el-input v-model="visualQuestion.description" type="textarea" :rows="2" placeholder="请输入题目说明" />
                    </div>
                  </div>
                </div>

                <div v-if="usesOptions(visualQuestion.type)" class="visual-section">
                  <div class="visual-section-title">选项设置</div>
                  <div class="visual-option-list">
                    <div v-for="(opt, idx) in visualQuestion.props.options" :key="idx" class="visual-option-row">
                      <span class="option-index">{{ idx + 1 }}</span>
                      <div class="bank-rich-editor bank-rich-option">
                        <QuillEditor
                          v-model:content="opt.label"
                          content-type="html"
                          :options="richOptionEditorOptions"
                          @ready="(quill: any) => bindCompactRichEditor(quill, 'option', idx)"
                        />
                      </div>
                      <el-input v-model="opt.value" placeholder="选项值" />
                      <el-button icon="Delete" circle plain type="danger" @click="removeVisualOption(idx)" />
                    </div>
                  </div>
                  <el-button size="small" @click="addVisualOption">+ 添加选项</el-button>
                </div>

                <div v-if="usesFields(visualQuestion.type)" class="visual-section">
                  <div class="visual-section-title">字段设置</div>
                  <div class="visual-option-list">
                    <div v-for="(field, idx) in visualQuestion.props.fields" :key="idx" class="visual-option-row">
                      <span class="option-index">{{ idx + 1 }}</span>
                      <el-input v-model="field.label" placeholder="字段名称" />
                      <el-input v-model="field.placeholder" placeholder="提示语" />
                      <el-button icon="Delete" circle plain type="danger" @click="removeVisualField(idx)" />
                    </div>
                  </div>
                  <el-button size="small" @click="addVisualField">+ 添加字段</el-button>
                </div>

                <div v-if="usesMatrix(visualQuestion.type)" class="visual-section">
                  <div class="visual-section-title">矩阵设置</div>
                  <div class="matrix-editor-grid">
                    <div>
                      <div class="matrix-subtitle">行</div>
                      <div class="visual-option-list">
                        <div v-for="(row, idx) in visualQuestion.props.rows" :key="idx" class="visual-option-row compact">
                          <span class="option-index">{{ idx + 1 }}</span>
                          <el-input v-model="row.title" placeholder="行标题" />
                          <el-button icon="Delete" circle plain type="danger" @click="removeMatrixRow(idx)" />
                        </div>
                      </div>
                      <el-button size="small" @click="addMatrixRow">+ 添加行</el-button>
                    </div>
                    <div>
                      <div class="matrix-subtitle">列</div>
                      <div class="visual-option-list">
                        <div v-for="(col, idx) in visualQuestion.props.columns" :key="idx" class="visual-option-row compact">
                          <span class="option-index">{{ idx + 1 }}</span>
                          <el-input v-model="col.title" placeholder="列标题" />
                          <el-button icon="Delete" circle plain type="danger" @click="removeMatrixColumn(idx)" />
                        </div>
                      </div>
                      <el-button size="small" @click="addMatrixColumn">+ 添加列</el-button>
                    </div>
                  </div>
                </div>

                <div v-if="visualQuestion.type === 'file'" class="visual-section">
                  <div class="visual-section-title">上传设置</div>
                  <div class="visual-grid">
                    <div class="visual-field wide">
                      <span>文件类型</span>
                      <el-select v-model="visualQuestion.fileTypes" multiple clearable style="width:100%">
                        <el-option label="图片" value="image" />
                        <el-option label="文档" value="document" />
                        <el-option label="视频" value="video" />
                        <el-option label="音频" value="audio" />
                      </el-select>
                    </div>
                    <div class="visual-field">
                      <span>大小 MB</span>
                      <el-input-number v-model="visualQuestion.maxFileSize" :min="1" :max="200" controls-position="right" style="width:100%" />
                    </div>
                    <div class="visual-field">
                      <span>数量</span>
                      <el-input-number v-model="visualQuestion.maxFileCount" :min="1" :max="20" controls-position="right" style="width:100%" />
                    </div>
                  </div>
                </div>

                <div v-if="visualQuestion.type === 'rating' || visualQuestion.type === 'nps'" class="visual-section">
                  <div class="visual-section-title">评分设置</div>
                  <div class="visual-grid">
                    <div class="visual-field">
                      <span>最大分值</span>
                      <el-input-number v-model="visualQuestion.props.maxRating" :min="1" :max="20" controls-position="right" style="width:100%" />
                    </div>
                  </div>
                </div>

                <div v-if="activeScope === 'exam'" class="visual-section">
                  <div class="visual-section-title">考试设置</div>
                  <div class="visual-grid">
                    <div class="visual-field">
                      <span>分值</span>
                      <el-input-number v-model="visualQuestion.examScore" :min="0" :step="1" controls-position="right" style="width:100%" />
                    </div>
                    <div class="visual-field wide">
                      <span>正确答案</span>
                      <el-input v-model="visualQuestion.examCorrectAnswer" placeholder="请输入正确答案" />
                    </div>
                    <div class="visual-field wide">
                      <span>解析</span>
                      <el-input v-model="visualQuestion.examAnalysis" type="textarea" :rows="2" placeholder="请输入答案解析" />
                    </div>
                  </div>
                </div>
              </div>

              <div class="visual-preview">
                <div class="visual-preview-title">
                  <question-icon :type="visualQuestion.type" class="bank-type-icon" />
                  <span v-html="visualQuestion.title || typeName(visualQuestion.type)"></span>
                </div>
                <div class="visual-preview-body">
                  <template v-if="usesOptions(visualQuestion.type)">
                    <div v-for="(opt, idx) in visualQuestion.props.options" :key="idx" class="preview-option">
                      <span class="preview-dot" />
                      <span v-html="opt.label || `选项${idx + 1}`"></span>
                    </div>
                  </template>
                  <template v-else-if="usesFields(visualQuestion.type)">
                    <div v-for="(field, idx) in visualQuestion.props.fields" :key="idx" class="preview-field">
                      <span>{{ field.label || `字段${idx + 1}` }}</span>
                      <el-input :placeholder="field.placeholder || '请输入'" disabled />
                    </div>
                  </template>
                  <template v-else-if="usesMatrix(visualQuestion.type)">
                    <div class="preview-matrix">
                      <div class="preview-matrix-row header">
                        <span>行/列</span>
                        <span v-for="(col, idx) in visualQuestion.props.columns" :key="idx">{{ col.title || `列${idx + 1}` }}</span>
                      </div>
                      <div v-for="(row, idx) in visualQuestion.props.rows" :key="idx" class="preview-matrix-row">
                        <span>{{ row.title || `行${idx + 1}` }}</span>
                        <span v-for="(_col, cidx) in visualQuestion.props.columns" :key="cidx">-</span>
                      </div>
                    </div>
                  </template>
                  <el-input v-else-if="visualQuestion.type === 'textarea'" type="textarea" :placeholder="visualQuestion.placeholder || '请输入'" disabled />
                  <el-input v-else :placeholder="visualQuestion.placeholder || previewPlaceholder(visualQuestion.type)" disabled />
                </div>
              </div>
            </div>
          </el-tab-pane>
          <el-tab-pane label="高级 JSON" name="json">
            <div class="schema-editor">
              <el-input v-model="form.schema" type="textarea" :rows="14" placeholder="题目 schema JSON，可从设计器上传后编辑" />
              <div class="schema-actions">
                <el-button size="small" @click="formatSchema">格式化 JSON</el-button>
                <el-button size="small" @click="applyJsonToVisual">同步到可视化</el-button>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible=false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>

  <QuestionRichEditorDialog
    :dialog="richEditDialog"
    :title="richEditTitle"
    :options="richFullEditorOptions"
    :confirm="confirmRichFullEdit"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import QuestionIcon from '../../survey/formkit/QuestionIcon.vue'
import QuestionRichEditorDialog from './QuestionRichEditorDialog.vue'
import type { BankScope } from '../utils/importExport'

type EditorMode = 'visual' | 'json'
type RichEditTarget = 'title' | 'option'

const props = defineProps<{
  visible: boolean
  isEdit: boolean
  form: any
  visualQuestion: any
  categories: string[]
  typeOptions: Array<{ value: string; label: string }>
  editorMode: EditorMode
  activeScope: BankScope
  saving: boolean
  richTitleEditorOptions: Record<string, any>
  richOptionEditorOptions: Record<string, any>
  richFullEditorOptions: Record<string, any>
  richEditDialog: {
    visible: boolean
    target: RichEditTarget
    optionIndex: number
    content: string
  }
  richEditTitle: string
  bindCompactRichEditor: (quill: any, target: RichEditTarget, optionIndex?: number) => void
  syncPlainTitleFromVisual: () => void
  handleTypeChange: (type: string) => void
  handleEditorModeChange: (name: string | number) => void
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
  typeName: (type: string) => string
  previewPlaceholder: (type: string) => string
  formatSchema: () => void
  applyJsonToVisual: () => void
  confirmRichFullEdit: () => void
  save: () => void
}>()

const emit = defineEmits<{
  (event: 'update:visible', value: boolean): void
  (event: 'update:editorMode', value: EditorMode): void
}>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value)
})

const editorModeModel = computed({
  get: () => props.editorMode,
  set: (value: EditorMode) => emit('update:editorMode', value)
})
</script>

<style scoped>
.schema-editor {
  width: 100%;
}

.schema-editor :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 12px;
  line-height: 1.6;
}

.schema-actions {
  margin-top: 8px;
}

.schema-tabs {
  width: 100%;
}

.bank-rich-editor {
  display: block;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  transition: border-color 0.18s, box-shadow 0.18s;
}

.bank-rich-editor:hover {
  border-color: #c0c4cc;
}

.bank-rich-editor:focus-within {
  border-color: var(--bank-accent);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}

.bank-rich-editor :deep(.ql-toolbar) {
  display: flex;
  flex-wrap: wrap;
  box-sizing: border-box;
  width: 100%;
  max-height: 0;
  overflow: hidden;
  padding: 0 8px;
  border: 0;
  border-bottom: 0 solid #eef2f7;
  background: #fff;
  opacity: 0;
  transition: max-height 0.18s ease, padding 0.18s ease, opacity 0.18s ease;
}

.bank-rich-editor:focus-within :deep(.ql-toolbar) {
  max-height: 74px;
  padding: 6px 8px;
  border-bottom-width: 1px;
  opacity: 1;
}

.bank-rich-editor :deep(.ql-toolbar .ql-formats) {
  margin-right: 6px;
}

.bank-rich-editor :deep(.ql-toolbar button),
.bank-rich-editor :deep(.ql-toolbar .ql-picker-label) {
  border-radius: 5px;
}

.bank-rich-editor :deep(.ql-toolbar .ql-bank-full) {
  width: 68px;
  padding: 0 6px;
  color: var(--bank-accent);
  font-size: 12px;
  font-weight: 600;
}

.bank-rich-editor :deep(.ql-toolbar button:hover),
.bank-rich-editor :deep(.ql-toolbar button.ql-active),
.bank-rich-editor :deep(.ql-toolbar .ql-picker-label:hover) {
  background: #f1f5f9;
}

.bank-rich-editor :deep(.ql-container) {
  display: block;
  box-sizing: border-box;
  width: 100%;
  height: auto;
  border: 0;
  border-radius: 8px;
  font-size: 13px;
  font-family: inherit;
}

.bank-rich-title :deep(.ql-editor) {
  min-height: 44px;
  padding: 10px 12px;
}

.bank-rich-option :deep(.ql-editor) {
  min-height: 38px;
  padding: 8px 10px;
}

.bank-rich-title :deep(.ql-editor p),
.bank-rich-option :deep(.ql-editor p) {
  margin: 0;
}

.bank-rich-title :deep(.ql-editor.ql-blank::before),
.bank-rich-option :deep(.ql-editor.ql-blank::before) {
  left: 12px;
  color: #a8b0bd;
  font-style: normal;
}


.question-edit-form :deep(.el-form-item__content) {
  min-width: 0;
}

.question-edit-form :deep(.el-form-item) {
  margin-bottom: 14px;
}

.question-edit-form :deep(.el-form-item__label) {
  margin-bottom: 6px;
  color: #475467;
  font-weight: 600;
  line-height: 1.3;
}

.visual-editor {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 14px;
  width: 100%;
}

.visual-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.visual-section,
.visual-preview {
  border: 1px solid var(--bank-border);
  border-radius: 8px;
  background: #fff;
}

.visual-section {
  padding: 12px;
}

.visual-section-title {
  margin-bottom: 10px;
  color: #344054;
  font-size: 13px;
  font-weight: 650;
}

.visual-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.visual-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.visual-field > span,
.matrix-subtitle {
  color: #667085;
  font-size: 12px;
}

.visual-field.wide {
  grid-column: 1 / -1;
}

.visual-option-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 8px;
}

.visual-option-row {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) minmax(0, 120px) 34px;
  gap: 8px;
  align-items: start;
}

.visual-option-row.compact {
  grid-template-columns: 28px minmax(0, 1fr) 34px;
}

.option-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-top: 7px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 12px;
}

.visual-option-row > .el-button {
  margin-top: 4px;
}

.matrix-editor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.matrix-subtitle {
  margin-bottom: 8px;
}

.visual-preview {
  align-self: start;
  overflow: hidden;
  background: #f8fafc;
}

.visual-preview-title {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid var(--bank-border);
  color: #1f2937;
  font-weight: 650;
}

.bank-type-icon {
  width: 16px;
  height: 16px;
  color: #94a3b8;
}

.visual-preview-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
}

.preview-option {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 28px;
  color: #344054;
  font-size: 13px;
}

.preview-dot {
  width: 12px;
  height: 12px;
  border: 1px solid #cbd5e1;
  border-radius: 999px;
  background: #fff;
}

.preview-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #344054;
  font-size: 12px;
}

.preview-matrix {
  overflow-x: auto;
  border: 1px solid var(--bank-border);
  border-radius: 6px;
  background: #fff;
}

.preview-matrix-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(72px, 1fr));
}

.preview-matrix-row > span {
  padding: 7px 8px;
  border-right: 1px solid var(--bank-border);
  border-bottom: 1px solid var(--bank-border);
  color: #475467;
  font-size: 12px;
}

.preview-matrix-row.header > span {
  background: #f1f5f9;
  color: #334155;
  font-weight: 650;
}

@media (max-width: 768px) {
  .visual-editor,
  .matrix-editor-grid,
  .visual-grid {
    grid-template-columns: 1fr;
  }

  .visual-option-row,
  .visual-option-row.compact {
    grid-template-columns: 24px minmax(0, 1fr) 34px;
  }

  .visual-option-row:not(.compact) :deep(.el-input:nth-of-type(2)) {
    grid-column: 2 / 3;
  }
}
</style>
