<template>
  <div class="formkit-designer">
    <div class="designer-header">
      <h2>表单设计器 (Formkit)</h2>
      <div class="header-actions">
        <el-button @click="loadFromJson">导入 JSON</el-button>
        <el-button @click="exportToJson" type="primary">导出 JSON</el-button>
        <el-button @click="validateSchema" :loading="validating">校验</el-button>
      </div>
    </div>

    <div class="designer-body">
      <!-- 左侧：题型面板 -->
      <div class="designer-left">
        <el-tabs v-model="activeCategory" class="type-tabs">
          <el-tab-pane
            v-for="cat in categories"
            :key="cat.name"
            :label="cat.label"
            :name="cat.name"
          >
            <div class="type-grid">
              <div
                v-for="t in typesByCategory[cat.name] || []"
                :key="t.type"
                class="type-card"
                @click="addQuestion(t)"
              >
                <div class="type-name">{{ t.displayName }}</div>
                <div class="type-key">{{ t.type }}</div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 中间：画布 -->
      <div class="designer-center">
        <div class="canvas-header">
          <span>题目列表 ({{ questions.length }})</span>
          <el-button
            v-if="questions.length > 0"
            size="small"
            type="danger"
            text
            @click="clearAll"
          >清空</el-button>
        </div>
        <div class="canvas-body">
          <el-empty v-if="questions.length === 0" description="点击左侧题型添加" />
          <draggable-list
            v-else
            :questions="questions"
            @update:questions="questions = $event"
            @select="selectQuestion"
            :selected-id="selectedId"
          />
        </div>
      </div>

      <!-- 右侧：属性面板 -->
      <div class="designer-right">
        <div v-if="selected" class="props-panel">
          <h3>{{ selected.title || selected.id || '题目属性' }}</h3>
          <el-form label-position="top" size="small">
            <el-form-item label="题目 ID (不可重复)">
              <el-input v-model="selected.id" :disabled="!!selected._existing" />
            </el-form-item>
            <el-form-item label="标题">
              <el-input v-model="selected.title" />
            </el-form-item>
            <el-form-item label="说明">
              <el-input v-model="selected.description" type="textarea" :rows="2" />
            </el-form-item>
            <el-form-item label="必填">
              <el-switch v-model="selected.required" />
            </el-form-item>
            <el-form-item label="占位提示">
              <el-input v-model="selected.placeholder" />
            </el-form-item>
            <el-form-item v-if="hasOptions(selected) && selected.props" label="选项">
              <options-editor :model-value="selected.props.options || []" @update:model-value="onOptionsUpdate" />
            </el-form-item>
            <el-form-item label="校验规则">
              <validate-editor :model-value="selected.validate || []" @update:model-value="onValidateUpdate" />
            </el-form-item>
            <el-form-item label="计算表达式">
              <calc-editor :model-value="selected.calcValue" :env="envFromAnswers" @update:model-value="onCalcUpdate" />
            </el-form-item>
            <el-form-item label="显示逻辑">
              <logic-editor :model-value="selected.logic || []" :all-questions="questions" @update:model-value="onLogicUpdate" />
            </el-form-item>
            <el-form-item>
              <el-button type="danger" text @click="removeSelected">删除此题</el-button>
            </el-form-item>
          </el-form>
        </div>
        <el-empty v-else description="选择一道题目以编辑属性" />
      </div>
    </div>

    <!-- 导出 / 导入对话框 -->
    <el-dialog v-model="exportDialogVisible" title="导出的 Schema JSON" width="700px">
      <el-input
        v-model="exportedJson"
        type="textarea"
        :rows="20"
        readonly
        style="font-family: monospace"
      />
      <template #footer>
        <el-button @click="copyJson">复制到剪贴板</el-button>
        <el-button type="primary" @click="exportDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="importDialogVisible" title="导入 Schema JSON" width="700px">
      <el-input
        v-model="importInput"
        type="textarea"
        :rows="15"
        placeholder='粘贴 schema JSON，例如 {"version":"2.0","questions":[...]}'
        style="font-family: monospace"
      />
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="validating" @click="doImport">校验并导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi } from '../../../api'
import DraggableList from './DraggableList.vue'
import OptionsEditor from './OptionsEditor.vue'
import ValidateEditor from './ValidateEditor.vue'
import CalcEditor from './CalcEditor.vue'
import LogicEditor from './LogicEditor.vue'

interface TypeMeta {
  type: string
  displayName: string
  category: string
  defaultProps: Record<string, any>
}

interface Question {
  id: string
  type: string
  title: string
  description?: string
  required: boolean
  placeholder?: string
  props?: any
  validate?: any[]
  logic?: any[]
  calcValue?: any
  _existing?: boolean
}

const types = ref<TypeMeta[]>([])
const activeCategory = ref('base')
const questions = ref<Question[]>([])
const selectedId = ref<string | null>(null)
const validating = ref(false)
const exportDialogVisible = ref(false)
const importDialogVisible = ref(false)
const exportedJson = ref('')
const importInput = ref('')

const categories = [
  { name: 'base', label: '基础' },
  { name: 'select', label: '选择' },
  { name: 'media', label: '媒体' },
  { name: 'layout', label: '布局' },
  { name: 'advanced', label: '高级' }
]

const typesByCategory = computed(() => {
  const m: Record<string, TypeMeta[]> = {}
  for (const t of types.value) {
    if (!m[t.category]) m[t.category] = []
    m[t.category].push(t)
  }
  return m
})

const selected = computed(() =>
  questions.value.find((q) => q.id === selectedId.value) || null
)

const envFromAnswers = computed(() => {
  const env: Record<string, any> = {}
  for (const q of questions.value) {
    env[q.id] = q.title || ''
  }
  return env
})

const hasOptions = (q: Question) => {
  return ['select', 'radio', 'checkbox', 'picker'].includes(q.type)
}

onMounted(async () => {
  try {
    const res = await adminApi.formkitTypes()
    types.value = res.data || []
  } catch (e) {
    ElMessage.error('加载题型列表失败')
  }
})

let idCounter = 0
const nextId = () => {
  idCounter++
  return `q${idCounter}`
}

const addQuestion = (t: TypeMeta) => {
  const id = nextId()
  const q: Question = {
    id,
    type: t.type,
    title: t.displayName,
    description: '',
    required: false,
    placeholder: (t.defaultProps && t.defaultProps.placeholder) || '',
    props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {},
    validate: [],
    logic: []
  }
  questions.value.push(q)
  selectedId.value = id
}

const selectQuestion = (id: string) => {
  selectedId.value = id
}

const onOptionsUpdate = (val: any[]) => {
  if (selected.value) {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props.options = val
  }
}

const onValidateUpdate = (val: any[]) => {
  if (selected.value) selected.value.validate = val
}

const onCalcUpdate = (val: any) => {
  if (selected.value) selected.value.calcValue = val
}

const onLogicUpdate = (val: any[]) => {
  if (selected.value) selected.value.logic = val
}

const removeSelected = () => {
  if (!selectedId.value) return
  questions.value = questions.value.filter((q) => q.id !== selectedId.value)
  selectedId.value = null
}

const clearAll = () => {
  questions.value = []
  selectedId.value = null
  idCounter = 0
}

const exportToJson = () => {
  const schema = {
    version: '2.0',
    questions: questions.value.map((q) => {
      const { _existing, ...rest } = q
      return rest
    })
  }
  const json = JSON.stringify(schema, null, 2)
  exportedJson.value = json
  // 写入 localStorage 供 SurveyDesigner 读取
  try { localStorage.setItem('pending_import_schema', json) } catch {}
  exportDialogVisible.value = true
}

const copyJson = async () => {
  try {
    await navigator.clipboard.writeText(exportedJson.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const loadFromJson = () => {
  importInput.value = ''
  importDialogVisible.value = true
}

const doImport = async () => {
  validating.value = true
  try {
    const res = await adminApi.formkitParseSchema(importInput.value)
    const parsed = res.data
    questions.value = (parsed.questions || []).map((q: Question) => ({
      ...q,
      _existing: true
    }))
    idCounter = questions.value.length
    importDialogVisible.value = false
    ElMessage.success('导入成功')
  } catch (e: any) {
    ElMessage.error(e?.msg || '解析失败')
  } finally {
    validating.value = false
  }
}

const validateSchema = async () => {
  validating.value = true
  try {
    const json = JSON.stringify({
      version: '2.0',
      questions: questions.value.map(({ _existing, ...rest }) => rest)
    })
    await adminApi.formkitParseSchema(json)
    ElMessage.success('Schema 校验通过')
  } catch (e: any) {
    ElMessage.error(e?.msg || '校验失败')
  } finally {
    validating.value = false
  }
}
</script>

<style scoped>
.formkit-designer {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 100px);
  background: #f5f7fa;
}
.designer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}
.designer-header h2 {
  margin: 0;
  font-size: 18px;
}
.header-actions {
  display: flex;
  gap: 8px;
}
.designer-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.designer-left {
  width: 280px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  overflow-y: auto;
}
.type-tabs :deep(.el-tabs__header) {
  margin: 0;
}
.type-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 12px;
}
.type-card {
  background: #f5f7fa;
  border: 1px dashed #c0c4cc;
  border-radius: 6px;
  padding: 12px 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}
.type-card:hover {
  background: #ecf5ff;
  border-color: #409eff;
  transform: translateY(-1px);
}
.type-name {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
}
.type-key {
  font-size: 11px;
  color: #909399;
  margin-top: 4px;
  font-family: monospace;
}
.designer-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
  overflow: hidden;
}
.canvas-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  font-size: 13px;
  color: #606266;
}
.canvas-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.designer-right {
  width: 360px;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  overflow-y: auto;
  padding: 16px;
}
.props-panel h3 {
  margin: 0 0 16px 0;
  font-size: 15px;
  color: #303133;
  border-bottom: 1px solid #ebeef5;
  padding-bottom: 8px;
}
</style>
