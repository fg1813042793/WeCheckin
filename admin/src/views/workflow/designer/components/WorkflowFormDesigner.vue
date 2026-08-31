<template>
  <div class="form-designer">
    <aside class="field-palette">
      <div class="panel-heading">
        <div>
          <strong>字段组件</strong>
          <span>{{ fieldTypes.length }} 个组件</span>
        </div>
      </div>
      <div class="palette-content">
        <section v-for="group in fieldGroups" :key="group.label" class="palette-group">
          <h3>{{ group.label }}</h3>
          <div class="palette-grid">
            <button
              v-for="item in group.items"
              :key="item.type"
              type="button"
              :disabled="readonly"
              @click="addField(item.type)"
            >
              <span class="palette-type-icon"><el-icon><component :is="item.icon" /></el-icon></span>
              <span>{{ item.label }}</span>
              <el-icon class="palette-add-icon"><Plus /></el-icon>
            </button>
          </div>
        </section>
      </div>
    </aside>

    <main class="form-canvas">
      <div class="canvas-heading">
        <div>
          <strong>流程表单</strong>
          <span>共 {{ fields.length }} 个字段</span>
        </div>
        <el-button v-if="fields.length && !readonly" size="small" icon="Plus" @click="addField('text')">添加字段</el-button>
      </div>

      <div class="canvas-stage">
        <section class="form-sheet">
          <el-empty v-if="fields.length === 0" :image-size="88" description="从左侧选择字段开始设计表单" />
          <div v-else class="field-list">
            <article
              v-for="(field, index) in fields"
              :key="field.key"
              class="field-item"
              :class="{
                active: selectedField === field,
                dragging: dragIndex === index,
                'drop-before': dropIndex === index,
                'field-item--compact': fieldSpan(field) <= 8,
              }"
              :style="{ gridColumn: `span ${fieldSpan(field)}` }"
              @click="selectField(field)"
              @dragover.prevent
              @dragenter="handleDragEnter(index)"
              @drop.prevent="handleDrop(index, $event)"
            >
              <div class="field-item__content">
                <div class="field-item__heading">
                  <div class="field-item__main">
                    <button
                      v-if="!readonly"
                      type="button"
                      class="field-drag-handle"
                      draggable="true"
                      title="拖动调整字段顺序"
                      @click.stop
                      @dragstart="handleDragStart(index, $event)"
                      @dragend="handleDragEnd"
                    >
                      <el-icon><Rank /></el-icon>
                    </button>
                    <span class="field-type-icon"><el-icon><component :is="fieldTypeMeta(field.type).icon" /></el-icon></span>
                    <div>
                      <strong>{{ field.label || '未命名字段' }}<i v-if="field.required">*</i></strong>
                      <p>{{ fieldTypeMeta(field.type).label }} · {{ field.key }} · {{ fieldSpanLabel(field) }}</p>
                    </div>
                  </div>
                  <div v-if="!readonly" class="field-actions" @click.stop>
                    <el-button circle size="small" icon="ArrowUp" title="上移" :disabled="index === 0" @click="moveField(index, -1)" />
                    <el-button circle size="small" icon="ArrowDown" title="下移" :disabled="index === fields.length - 1" @click="moveField(index, 1)" />
                    <el-button circle size="small" type="danger" plain icon="Delete" title="删除" @click="removeField(index)" />
                  </div>
                </div>

                <div class="field-preview" @click.stop="selectField(field)">
                  <el-input
                    v-if="field.type === 'text'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || '请输入内容'"
                    disabled
                  />
                  <el-input
                    v-else-if="field.type === 'textarea'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || '请输入内容'"
                    type="textarea"
                    :rows="2"
                    resize="none"
                    disabled
                  />
                  <el-input-number
                    v-else-if="field.type === 'number' || field.type === 'amount'"
                    :model-value="numberDefault(field.default)"
                    :placeholder="field.placeholder || (field.type === 'amount' ? '请输入金额' : '请输入数字')"
                    :precision="field.type === 'amount' ? 2 : undefined"
                    controls-position="right"
                    disabled
                  />
                  <el-input
                    v-else-if="field.type === 'phone' || field.type === 'email'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || (field.type === 'phone' ? '请输入手机号' : '请输入邮箱')"
                    disabled
                  >
                    <template #suffix><el-icon><component :is="field.type === 'phone' ? 'Cellphone' : 'Message'" /></el-icon></template>
                  </el-input>
                  <el-select
                    v-else-if="field.type === 'select' || field.type === 'multi_select'"
                    :model-value="selectDefault(field)"
                    :multiple="field.type === 'multi_select'"
                    :placeholder="field.placeholder || '请选择'"
                    disabled
                  >
                    <el-option v-for="option in field.options" :key="option.value" :label="option.label" :value="option.value" />
                  </el-select>
                  <el-radio-group v-else-if="field.type === 'radio'" :model-value="stringDefault(field.default)" disabled>
                    <el-radio v-for="option in field.options" :key="option.value" :value="option.value">{{ option.label }}</el-radio>
                  </el-radio-group>
                  <el-checkbox-group v-else-if="field.type === 'checkbox'" :model-value="arrayDefault(field.default)" disabled>
                    <el-checkbox v-for="option in field.options" :key="option.value" :value="option.value">{{ option.label }}</el-checkbox>
                  </el-checkbox-group>
                  <el-input
                    v-else-if="field.type === 'date' || field.type === 'datetime'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || (field.type === 'date' ? '请选择日期' : '请选择日期时间')"
                    disabled
                  >
                    <template #suffix><el-icon><component :is="field.type === 'date' ? 'Calendar' : 'Clock'" /></el-icon></template>
                  </el-input>
                  <el-time-picker
                    v-else-if="field.type === 'time'"
                    :model-value="null"
                    :placeholder="field.placeholder || '请选择时间'"
                    disabled
                  />
                  <el-date-picker
                    v-else-if="field.type === 'date_range'"
                    :model-value="[]"
                    type="daterange"
                    range-separator="至"
                    start-placeholder="开始日期"
                    end-placeholder="结束日期"
                    disabled
                  />
                  <el-input
                    v-else-if="['user', 'user_multi', 'department', 'department_multi'].includes(field.type)"
                    :placeholder="field.placeholder || organizationPlaceholder(field.type)"
                    disabled
                  >
                    <template #suffix><el-icon><component :is="field.type.startsWith('user') ? 'User' : 'OfficeBuilding'" /></el-icon></template>
                  </el-input>
                  <el-button v-else-if="field.type === 'attachment'" icon="Upload" disabled>选择附件</el-button>
                  <el-switch v-else :model-value="Boolean(field.default)" disabled />
                </div>
              </div>
            </article>
            <div
              v-if="dragIndex >= 0"
              class="field-drop-zone field-drop-zone--tail"
              :class="{ active: dropIndex === fields.length }"
              @dragover.prevent
              @dragenter="handleDragEnter(fields.length)"
              @drop.prevent="handleDrop(fields.length, $event)"
            >
              拖到此处置底
            </div>
          </div>
        </section>
      </div>
    </main>

    <aside class="property-panel">
      <div class="panel-heading">
        <div>
          <strong>字段属性</strong>
          <span>{{ selectedField ? fieldTypeMeta(selectedField.type).label : '请选择字段' }}</span>
        </div>
      </div>
      <el-empty v-if="!selectedField" :image-size="72" description="选择中间字段后配置属性" />
      <el-form v-else label-position="top" class="property-form">
        <section class="property-section">
          <h3>基础信息</h3>
          <el-form-item label="字段名称" required>
            <el-input v-model="selectedField.label" maxlength="60" :disabled="readonly" @input="emitChange" />
          </el-form-item>
          <el-form-item label="字段编码" required>
            <el-input v-model="selectedField.key" maxlength="100" :disabled="readonly" @input="emitChange">
              <template #append><el-tooltip content="以字母开头，可使用字母、数字、点、下划线和中划线"><el-icon><QuestionFilled /></el-icon></el-tooltip></template>
            </el-input>
          </el-form-item>
          <el-form-item label="提示文字">
            <el-input v-model="selectedField.placeholder" maxlength="120" :disabled="readonly" @input="emitChange" />
          </el-form-item>
        </section>

        <section class="property-section layout-setting">
          <h3>布局设置</h3>
          <el-form-item label="字段宽度">
            <el-radio-group :model-value="fieldSpan(selectedField)" :disabled="readonly" @change="updateFieldSpan">
              <el-radio-button v-for="item in fieldSpanOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </el-radio-button>
            </el-radio-group>
          </el-form-item>
          <p>PC 端按栅格同行排列，移动端自动切换为整行。</p>
        </section>

        <section class="property-section">
          <h3>校验规则</h3>
          <div class="required-setting">
            <span>必填字段</span>
            <el-switch v-model="selectedField.required" :disabled="readonly" @change="emitChange" />
          </div>
          <el-form-item v-if="['text', 'textarea', 'phone', 'email'].includes(selectedField.type)" label="最大长度">
            <el-input-number
              v-model="selectedField.maxLength"
              :min="0"
              :max="100000"
              :disabled="readonly"
              controls-position="right"
              @change="emitChange"
            />
          </el-form-item>
          <div class="number-range">
            <el-form-item v-if="selectedField.type === 'number' || selectedField.type === 'amount'" label="最小值"><el-input-number v-model="selectedField.min" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
            <el-form-item v-if="selectedField.type === 'number' || selectedField.type === 'amount'" label="最大值"><el-input-number v-model="selectedField.max" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
          </div>
        </section>

        <section v-if="supportsDefault(selectedField)" class="property-section">
          <h3>默认值</h3>
          <el-form-item>
            <el-switch
              v-if="selectedField.type === 'boolean'"
              :model-value="Boolean(selectedField.default)"
              :disabled="readonly"
              @change="updateDefault"
            />
            <el-input-number
              v-else-if="selectedField.type === 'number' || selectedField.type === 'amount'"
              :model-value="numberDefault(selectedField.default)"
              :precision="selectedField.type === 'amount' ? 2 : undefined"
              :disabled="readonly"
              controls-position="right"
              @change="updateDefault"
            />
            <el-select
              v-else-if="isOptionField(selectedField)"
              :model-value="selectedField.default"
              :multiple="selectedField.type === 'multi_select' || selectedField.type === 'checkbox'"
              clearable
              :disabled="readonly"
              style="width: 100%"
              @change="updateDefault"
            >
              <el-option v-for="option in selectedField.options" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
            <el-input
              v-else
              :model-value="stringDefault(selectedField.default)"
              clearable
              :disabled="readonly"
              @input="updateDefault"
            />
          </el-form-item>
        </section>

        <section v-if="isOptionField(selectedField)" class="property-section option-editor">
          <div class="option-editor__heading">
            <h3>选项配置</h3>
            <el-button v-if="!readonly" link type="primary" icon="Plus" @click="addOption">新增选项</el-button>
          </div>
          <div v-for="(option, index) in selectedField.options" :key="index" class="option-row">
            <span class="option-index">{{ index + 1 }}</span>
            <div class="option-row__inputs">
              <el-input v-model="option.label" placeholder="选项名称" :disabled="readonly" @input="emitChange" />
              <el-input v-model="option.value" placeholder="选项值" :disabled="readonly" @input="emitChange" />
            </div>
            <el-button circle size="small" type="danger" plain icon="Delete" :disabled="readonly || (selectedField.options?.length || 0) <= 1" @click="removeOption(index)" />
          </div>
        </section>
      </el-form>
    </aside>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'
import type { WorkflowFormField, WorkflowFormFieldSpan, WorkflowFormFieldType } from '../../types'

const props = defineProps<{ fields: WorkflowFormField[]; readonly?: boolean }>()
const emit = defineEmits<{ change: [] }>()

const fieldTypes: Array<{ type: WorkflowFormFieldType; label: string; icon: string }> = [
  { type: 'text', label: '单行文本', icon: 'EditPen' },
  { type: 'textarea', label: '多行文本', icon: 'Document' },
  { type: 'number', label: '数字', icon: 'Odometer' },
  { type: 'amount', label: '金额', icon: 'Money' },
  { type: 'phone', label: '手机号', icon: 'Cellphone' },
  { type: 'email', label: '邮箱', icon: 'Message' },
  { type: 'boolean', label: '开关', icon: 'Switch' },
  { type: 'select', label: '单选下拉', icon: 'Select' },
  { type: 'multi_select', label: '多选下拉', icon: 'Finished' },
  { type: 'radio', label: '单选框', icon: 'CircleCheck' },
  { type: 'checkbox', label: '复选框', icon: 'Checked' },
  { type: 'date', label: '日期', icon: 'Calendar' },
  { type: 'datetime', label: '日期时间', icon: 'Clock' },
  { type: 'time', label: '时间', icon: 'Timer' },
  { type: 'date_range', label: '日期区间', icon: 'Calendar' },
  { type: 'user', label: '人员', icon: 'User' },
  { type: 'user_multi', label: '多人', icon: 'UserFilled' },
  { type: 'department', label: '部门', icon: 'OfficeBuilding' },
  { type: 'department_multi', label: '多部门', icon: 'CopyDocument' },
  { type: 'attachment', label: '附件', icon: 'Paperclip' },
]
const fieldGroups = [
  { label: '基础字段', items: fieldTypes.filter(item => ['text', 'textarea', 'number', 'amount', 'phone', 'email', 'boolean'].includes(item.type)) },
  { label: '选择与时间', items: fieldTypes.filter(item => ['select', 'multi_select', 'radio', 'checkbox', 'date', 'datetime', 'time', 'date_range'].includes(item.type)) },
  { label: '组织与附件', items: fieldTypes.filter(item => ['user', 'user_multi', 'department', 'department_multi', 'attachment'].includes(item.type)) },
]
const fieldSpanOptions: Array<{ label: string; value: WorkflowFormFieldSpan }> = [
  { label: '1/4', value: 6 },
  { label: '1/3', value: 8 },
  { label: '1/2', value: 12 },
  { label: '整行', value: 24 },
]

const selectedField = ref<WorkflowFormField | null>(null)
const dragIndex = ref(-1)
const dropIndex = ref(-1)

watch(() => props.fields, (fields) => {
  if (!fields.length) selectedField.value = null
  else if (!selectedField.value || !fields.includes(selectedField.value)) selectedField.value = fields[0]
}, { immediate: true, deep: false })

function fieldTypeMeta(type: WorkflowFormFieldType) {
  return fieldTypes.find(item => item.type === type) || fieldTypes[0]
}

function nextFieldKey() {
  let index = props.fields.length + 1
  while (props.fields.some(item => item.key === `field_${index}`)) index += 1
  return `field_${index}`
}

function buildField(type: WorkflowFormFieldType): WorkflowFormField {
  const meta = fieldTypeMeta(type)
  const field: WorkflowFormField = {
    key: nextFieldKey(),
    label: meta.label,
    type,
    required: false,
    placeholder: '',
    span: defaultFieldSpan(type),
  }
  if (type === 'text' || type === 'textarea' || type === 'phone' || type === 'email') {
    field.maxLength = type === 'textarea' ? 2000 : type === 'email' ? 254 : type === 'phone' ? 20 : 200
  }
  if (type === 'number' || type === 'amount') {
    field.min = 0
    field.max = type === 'amount' ? 1000000 : 100
  }
  if (['select', 'multi_select', 'radio', 'checkbox'].includes(type)) {
    field.options = [
      { label: '选项一', value: 'option_1' },
      { label: '选项二', value: 'option_2' },
    ]
  }
  return field
}

function defaultFieldSpan(type: WorkflowFormFieldType): WorkflowFormFieldSpan {
  return ['textarea', 'radio', 'checkbox', 'attachment', 'date_range'].includes(type) ? 24 : 12
}

function fieldSpan(field: WorkflowFormField): WorkflowFormFieldSpan {
  return fieldSpanOptions.some(item => item.value === field.span) ? field.span as WorkflowFormFieldSpan : 24
}

function fieldSpanLabel(field: WorkflowFormField) {
  return fieldSpanOptions.find(item => item.value === fieldSpan(field))?.label || '整行'
}

function updateFieldSpan(value: string | number | boolean | undefined) {
  if (!selectedField.value) return
  const span = Number(value)
  if (!fieldSpanOptions.some(item => item.value === span)) return
  selectedField.value.span = span as WorkflowFormFieldSpan
  emitChange()
}

function addField(type: WorkflowFormFieldType) {
  if (props.readonly) return
  const field = buildField(type)
  props.fields.push(field)
  selectedField.value = field
  emitChange()
}

function selectField(field: WorkflowFormField) {
  selectedField.value = field
}

function moveField(index: number, offset: number) {
  const target = index + offset
  if (target < 0 || target >= props.fields.length) return
  const [field] = props.fields.splice(index, 1)
  props.fields.splice(target, 0, field)
  emitChange()
}

function handleDragStart(index: number, event: DragEvent) {
  if (props.readonly) return
  dragIndex.value = index
  dropIndex.value = -1
  selectedField.value = props.fields[index] || null
  if (!event.dataTransfer) return
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', props.fields[index]?.key || String(index))
  const card = (event.currentTarget as HTMLElement | null)?.closest('.field-item') as HTMLElement | null
  if (!card) return
  const rect = card.getBoundingClientRect()
  event.dataTransfer.setDragImage(card, Math.min(event.clientX - rect.left, rect.width), Math.min(event.clientY - rect.top, rect.height))
}

function handleDragEnter(index: number) {
  const source = dragIndex.value
  if (source < 0 || index === source || (index > source && index === source + 1)) {
    dropIndex.value = -1
    return
  }
  dropIndex.value = index
}

function handleDrop(index: number, event: DragEvent) {
  event.preventDefault()
  const source = dragIndex.value
  if (source < 0 || source >= props.fields.length) {
    handleDragEnd()
    return
  }
  const target = index > source ? index - 1 : index
  if (target === source || target < 0 || target >= props.fields.length) {
    handleDragEnd()
    return
  }
  const [field] = props.fields.splice(source, 1)
  props.fields.splice(target, 0, field)
  selectedField.value = field
  handleDragEnd()
  emitChange()
}

function handleDragEnd() {
  dragIndex.value = -1
  dropIndex.value = -1
}

function removeField(index: number) {
  if (props.readonly) return
  const [removed] = props.fields.splice(index, 1)
  if (removed === selectedField.value) selectedField.value = props.fields[Math.min(index, props.fields.length - 1)] || null
  emitChange()
}

function isOptionField(field: WorkflowFormField) {
  return ['select', 'multi_select', 'radio', 'checkbox'].includes(field.type)
}

function addOption() {
  if (!selectedField.value || !isOptionField(selectedField.value)) return
  selectedField.value.options ||= []
  const index = selectedField.value.options.length + 1
  selectedField.value.options.push({ label: `选项${index}`, value: `option_${index}` })
  emitChange()
}

function removeOption(index: number) {
  if (!selectedField.value?.options || selectedField.value.options.length <= 1) return
  selectedField.value.options.splice(index, 1)
  emitChange()
}

function supportsDefault(field: WorkflowFormField) {
  return !['attachment', 'user', 'user_multi', 'department', 'department_multi', 'date_range'].includes(field.type)
}

function stringDefault(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function numberDefault(value: unknown) {
  return typeof value === 'number' ? value : undefined
}

function arrayDefault(value: unknown) {
  return Array.isArray(value) ? value.filter(item => typeof item === 'string') : []
}

function selectDefault(field: WorkflowFormField) {
  if (field.type === 'multi_select') return Array.isArray(field.default) ? field.default.filter(value => typeof value === 'string') : []
  return typeof field.default === 'string' ? field.default : undefined
}

function organizationPlaceholder(type: WorkflowFormFieldType) {
  if (type === 'user_multi') return '请选择人员，可多选'
  if (type === 'department_multi') return '请选择部门，可多选'
  return type === 'user' ? '请选择人员' : '请选择部门'
}

function updateDefault(value: unknown) {
  if (!selectedField.value) return
  if (value === '' || value === undefined || value === null || (Array.isArray(value) && value.length === 0)) delete selectedField.value.default
  else selectedField.value.default = value
  emitChange()
}

function emitChange() {
  emit('change')
}
</script>

<style scoped>
.form-designer { display: grid; grid-template-columns: 320px minmax(480px, 1fr) 360px; width: 100%; min-width: 0; min-height: 0; height: 100%; overflow: hidden; background: #f2f5f8; }
.field-palette, .property-panel { display: flex; min-width: 0; min-height: 0; overflow: hidden; flex-direction: column; background: #fff; }
.field-palette { border-right: 1px solid #dfe6ee; }
.property-panel { border-left: 1px solid #dfe6ee; }
.panel-heading, .canvas-heading { display: flex; align-items: center; flex: 0 0 auto; justify-content: space-between; gap: 12px; min-height: 58px; padding: 0 16px; border-bottom: 1px solid #e7ebf0; background: #fff; }
.panel-heading > div { display: flex; align-items: baseline; justify-content: space-between; width: 100%; gap: 8px; }
.panel-heading strong, .canvas-heading strong { color: #1f2937; font-size: 14px; }
.panel-heading span, .canvas-heading span { color: #94a3b8; font-size: 11px; }
.canvas-heading > div { display: flex; align-items: baseline; gap: 8px; }
.palette-content { min-height: 0; overflow-y: auto; flex: 1; padding: 14px 12px 24px; }
.palette-group + .palette-group { margin-top: 20px; }
.palette-group h3 { margin: 0 0 9px 2px; color: #64748b; font-size: 11px; font-weight: 600; }
.palette-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.palette-grid button { position: relative; display: flex; align-items: center; gap: 7px; min-width: 0; height: 40px; padding: 0 8px; border: 1px solid #e2e8f0; border-radius: 6px; color: #475569; background: #fff; font: inherit; font-size: 12px; cursor: pointer; transition: border-color .15s, background-color .15s, color .15s; }
.palette-grid button:hover:not(:disabled) { border-color: #9fc2f7; color: #1677ff; background: #f6f9ff; }
.palette-grid button:disabled { cursor: not-allowed; opacity: .55; }
.palette-type-icon { display: grid; flex: 0 0 auto; place-items: center; width: 24px; height: 24px; border-radius: 5px; color: #1677ff; background: #edf5ff; }
.palette-add-icon { position: absolute; right: 7px; opacity: 0; color: #1677ff; transition: opacity .15s; }
.palette-grid button:hover .palette-add-icon { opacity: 1; }
.form-canvas { display: flex; min-width: 0; min-height: 0; overflow: hidden; flex-direction: column; background: #f2f5f8; }
.canvas-stage { min-height: 0; overflow-y: auto; flex: 1; padding: 18px 22px 32px; }
.form-sheet { width: min(920px, 100%); min-height: calc(100% - 2px); margin: 0 auto; padding: 16px; border: 1px solid #dfe6ee; border-radius: 8px; background: #fff; box-shadow: 0 8px 24px rgb(15 23 42 / 4%); }
.form-sheet > .el-empty { min-height: 360px; }
.field-list { display: grid; grid-template-columns: repeat(24, minmax(0, 1fr)); align-items: start; gap: 10px; }
.field-item { position: relative; display: block; min-height: 112px; padding: 14px; border: 1px solid #e3e8ef; border-radius: 7px; background: #fff; cursor: pointer; transition: border-color .15s, box-shadow .15s, background-color .15s; }
.field-item:last-child { margin-bottom: 0; }
.field-item:hover { border-color: #b8cff1; background: #fcfdff; }
.field-item.active { border-color: #74a9f7; box-shadow: 0 0 0 2px rgb(22 119 255 / 7%); }
.field-item.active::before { position: absolute; top: 14px; bottom: 14px; left: -1px; width: 3px; border-radius: 0 2px 2px 0; background: #1677ff; content: ''; }
.field-item.dragging { opacity: .45; border-style: dashed; box-shadow: none; }
.field-item.drop-before::after { position: absolute; z-index: 2; top: -7px; right: 8px; left: 8px; height: 2px; border-radius: 2px; background: #1677ff; box-shadow: 0 0 0 3px rgb(22 119 255 / 9%); content: ''; }
.field-item__heading { display: flex; align-items: center; justify-content: space-between; min-width: 0; gap: 14px; }
.field-item__main { display: flex; align-items: center; min-width: 0; gap: 11px; }
.field-item__main > div { min-width: 0; }
.field-drag-handle { display: grid; flex: 0 0 auto; place-items: center; width: 24px; height: 34px; padding: 0; border: 0; border-radius: 4px; color: #a7b2c1; background: transparent; cursor: grab; }
.field-drag-handle:hover { color: #1677ff; background: #edf5ff; }
.field-drag-handle:active { cursor: grabbing; }
.field-type-icon { display: grid; flex: 0 0 auto; place-items: center; width: 34px; height: 34px; border-radius: 6px; color: #1677ff; background: #eaf3ff; }
.field-item strong { color: #273548; font-size: 13px; }
.field-item strong i { margin-left: 3px; color: #ef4444; font-style: normal; }
.field-item p { overflow: hidden; margin: 4px 0 0; color: #94a3b8; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.field-actions { display: flex; flex: 0 0 auto; gap: 5px; opacity: 0; transition: opacity .15s; }
.field-item:hover .field-actions, .field-item.active .field-actions, .field-actions:focus-within { opacity: 1; }
.field-item--compact .field-item__main { padding-right: 24px; gap: 6px; }
.field-item--compact .field-type-icon { width: 30px; height: 30px; }
.field-item--compact .field-actions { position: absolute; z-index: 3; top: 8px; right: 8px; padding: 3px; border: 1px solid #e5eaf0; border-radius: 6px; background: #fff; box-shadow: 0 5px 14px rgb(15 23 42 / 10%); }
.field-item--compact .field-preview { margin-left: 0; }
.field-preview { max-width: 720px; margin: 12px 0 0 45px; pointer-events: none; }
.field-preview :deep(.el-select), .field-preview :deep(.el-input-number), .field-preview :deep(.el-date-editor) { width: 100%; }
.field-preview :deep(.el-radio-group), .field-preview :deep(.el-checkbox-group) { display: flex; flex-wrap: wrap; gap: 8px 18px; }
.field-preview :deep(.el-radio), .field-preview :deep(.el-checkbox) { margin-right: 0; }
.field-preview :deep(.is-disabled .el-input__wrapper), .field-preview :deep(.el-textarea.is-disabled .el-textarea__inner) { background: #f8fafc; box-shadow: 0 0 0 1px #e6ebf1 inset; }
.field-drop-zone--tail { display: grid; grid-column: 1 / -1; place-items: center; height: 42px; border: 1px dashed #cbd5e1; border-radius: 6px; color: #94a3b8; background: #f8fafc; font-size: 11px; }
.field-drop-zone--tail.active { border-color: #1677ff; color: #1677ff; background: #f2f7ff; }
.property-panel > .panel-heading { position: sticky; top: 0; z-index: 2; }
.property-form { min-height: 0; overflow-y: auto; flex: 1; }
.property-form :deep(.el-form-item) { margin-bottom: 14px; }
.property-form :deep(.el-form-item__label) { margin-bottom: 5px; color: #475569; font-size: 12px; line-height: 20px; }
.property-form :deep(.el-input-number) { width: 100%; }
.property-section { padding: 16px; border-bottom: 1px solid #edf0f4; }
.property-section:last-child { border-bottom: 0; }
.property-section > h3, .option-editor__heading h3 { margin: 0 0 14px; color: #273548; font-size: 12px; font-weight: 650; }
.layout-setting :deep(.el-radio-group) { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); width: 100%; }
.layout-setting :deep(.el-radio-button), .layout-setting :deep(.el-radio-button__inner) { width: 100%; }
.layout-setting :deep(.el-radio-button__inner) { padding-right: 8px; padding-left: 8px; }
.layout-setting > p { margin: -3px 0 0; color: #94a3b8; font-size: 11px; line-height: 1.6; }
.required-setting { display: flex; align-items: center; justify-content: space-between; min-height: 34px; margin-bottom: 12px; color: #475569; font-size: 12px; }
.number-range { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.option-editor__heading { display: flex; align-items: center; justify-content: space-between; min-height: 32px; }
.option-editor__heading h3 { margin-bottom: 0; }
.option-row { display: grid; grid-template-columns: 22px minmax(0, 1fr) 28px; align-items: center; gap: 7px; margin-top: 10px; }
.option-index { display: grid; place-items: center; width: 22px; height: 22px; border-radius: 50%; color: #64748b; background: #f1f5f9; font-size: 10px; }
.option-row__inputs { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
@media (max-width: 1380px) {
  .form-designer { grid-template-columns: 200px minmax(420px, 1fr) 320px; }
  .canvas-stage { padding-right: 16px; padding-left: 16px; }
}
@media (max-width: 1120px) {
  .form-designer { grid-template-columns: 176px minmax(360px, 1fr) 286px; }
  .palette-grid { grid-template-columns: 1fr; }
  .option-row__inputs { grid-template-columns: 1fr; }
  .field-preview { margin-left: 0; }
}
@media (max-width: 760px) {
  .field-item { grid-column: 1 / -1 !important; }
}
</style>
