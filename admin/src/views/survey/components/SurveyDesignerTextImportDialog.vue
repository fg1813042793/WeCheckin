<template>
  <el-dialog v-model="visible" title="文本导入" width="600px" :close-on-click-modal="false" @closed="reset">
    <div class="import-toolbar">
      <el-radio-group v-model="mode" size="small">
        <el-radio-button value="paste">粘贴文本</el-radio-button>
        <el-radio-button value="file">上传文件</el-radio-button>
      </el-radio-group>
      <el-tag type="info" class="format-hint">每行一个题目，选项用 A. B. C. 开头</el-tag>
    </div>
    <el-input
      v-if="mode === 'paste'"
      v-model="text"
      type="textarea"
      :rows="12"
      placeholder="粘贴题目文本&#10;&#10;格式示例：&#10;1. 您最喜欢的颜色？&#10;A. 红色&#10;B. 蓝色&#10;C. 绿色&#10;&#10;2. 您的姓名&#10;&#10;3. [多选题] 您的爱好&#10;A. 阅读&#10;B. 运动&#10;C. 音乐"
    />
    <div v-else>
      <el-upload drag :auto-upload="false" :show-file-list="false" :on-change="onTextFileChange" accept=".txt,.docx,.md">
        <el-icon class="upload-icon"><DocumentAdd /></el-icon>
        <div class="upload-text">拖拽或点击选择 .txt / .docx 文件</div>
      </el-upload>
      <div v-if="text" class="parsed-lines">已解析 {{ text.split('\n').filter(line => line.trim()).length }} 行</div>
    </div>
    <div v-if="preview.length" class="import-preview">
      <div class="preview-count">预览（{{ preview.length }} 题）：</div>
      <div v-if="surveyTitle" class="preview-title">{{ surveyTitle }}</div>
      <div v-for="(question, index) in preview" :key="index" class="preview-row">
        <el-tag size="small">{{ typeName(question.type) }}</el-tag>
        <span>{{ stripSurveyMarkup(question.title) }}</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :disabled="!parsed.length" @click="confirmImport">导入 {{ parsed.length }} 题</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { DocumentAdd } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getSurveyQuestionTypeName as typeName } from '../survey-designer-question-types'
import { parseSurveyText, stripSurveyMarkup, type ParsedSurveyTextQuestion } from '../survey-designer-text-transfer'
import type { SurveyTextImportPayload } from '../survey-designer-types'

const emit = defineEmits<{ import: [payload: SurveyTextImportPayload] }>()
const visible = ref(false)
const mode = ref<'paste' | 'file'>('paste')
const text = ref('')
const parsed = ref<ParsedSurveyTextQuestion[]>([])
const preview = ref<ParsedSurveyTextQuestion[]>([])
const surveyTitle = ref('')

watch(text, value => {
  if (mode.value === 'paste' && value) parse(value)
  else if (!value) clearParsed()
})
watch(mode, clearParsed)

function parse(value: string) {
  const result = parseSurveyText(value)
  surveyTitle.value = result.title
  parsed.value = result.questions
  preview.value = result.questions.slice(0, 20)
}

function clearParsed() {
  parsed.value = []
  preview.value = []
  surveyTitle.value = ''
}

async function onTextFileChange(file: any) {
  const source = file.raw || file
  if (!source) return
  if (source.name.endsWith('.docx')) {
    try {
      const libraryName = 'mamm' + 'oth'
      const mammothModule = await import(/* @vite-ignore */ libraryName)
      const mammoth = (mammothModule.default || mammothModule) as any
      const result = await mammoth.extractRawText({ arrayBuffer: await source.arrayBuffer() })
      text.value = result.value
    } catch {
      ElMessage.error('Word 文件解析失败，请先安装 mammoth: npm install mammoth')
      return
    }
  } else {
    text.value = await source.text()
  }
  parse(text.value)
}

function confirmImport() {
  if (!parsed.value.length) return
  emit('import', { title: surveyTitle.value, questions: parsed.value })
  visible.value = false
}

function reset() {
  mode.value = 'paste'
  text.value = ''
  clearParsed()
}

function open() {
  reset()
  visible.value = true
}

defineExpose({ open })
</script>

<style scoped>
.import-toolbar { display:flex; align-items:center; gap:8px; margin-bottom:8px; }
.format-hint { font-size:11px; }
.upload-icon { margin-bottom:8px; font-size:40px; }
.upload-text { color:#666; font-size:13px; }
.parsed-lines { margin-top:8px; color:#999; font-size:12px; }
.import-preview { max-height:200px; overflow-y:auto; margin-top:12px; padding:8px; border:1px solid #e8e8e8; border-radius:6px; }
.preview-count { margin-bottom:4px; color:#666; font-size:12px; }
.preview-title { margin-bottom:4px; padding:2px 0 6px; border-bottom:1px dashed #eee; font-size:13px; font-weight:bold; }
.preview-row { display:flex; gap:6px; padding:2px 0; font-size:12px; }
.preview-row :deep(.el-tag) { flex-shrink:0; }
.preview-row span { overflow:hidden; color:#333; text-overflow:ellipsis; white-space:nowrap; }
</style>
