<template>
  <div class="workflow-runtime-form" :class="{ 'workflow-runtime-form--embedded': embedded }">
    <el-empty v-if="visibleFields.length === 0" :image-size="72" :description="emptyText" />
    <component :is="embedded ? 'div' : ElForm" v-else label-position="top" class="runtime-form-grid" @submit.prevent>
      <template v-for="field in visibleFields" :key="field.key">
        <section v-if="field.type === 'group'" class="runtime-form-group">
          <header class="runtime-form-group__header">
            <strong>{{ field.label || field.key }}</strong>
            <el-button v-if="field.help" link type="primary" @click="openFieldHelp(field)">
              {{ helpButtonText(field) }}
            </el-button>
          </header>
          <WorkflowRuntimeForm
            :fields="field.fields || []"
            :model-value="modelValue"
            :field-access="props.fieldAccess"
            :field-actions="props.fieldActions"
            :readonly="readonly"
            :user-name-map="props.userNameMap"
            :calculation-fields="calculationFields || fields"
            embedded
            :show-errors="errorsVisible"
            @update:model-value="emit('update:modelValue', $event)"
            @change="emit('change', $event)"
          />
        </section>
        <h3
          v-else-if="field.type === 'label'"
          class="runtime-form-label"
          :style="{ gridColumn: `span ${fieldSpan(field)}` }"
        >{{ field.label }}</h3>
        <p
          v-else-if="field.type === 'description'"
          class="runtime-form-description"
          :style="{ gridColumn: `span ${fieldSpan(field)}` }"
        >{{ field.content }}</p>
        <div
          v-else-if="field.type === 'button'"
          class="runtime-form-button"
          :style="{ gridColumn: `span ${fieldSpan(field)}` }"
        >
          <el-button @click="openFieldHelp(field)">{{ field.label }}</el-button>
        </div>
        <div
          v-else-if="field.type === 'calculation' && calculationDisplay(field) === 'label'"
          class="runtime-form-calculation-label"
          :style="{ gridColumn: `span ${fieldSpan(field)}` }"
        >
          <span>{{ field.label || field.key }}</span>
          <strong>{{ calculationText(field) }}</strong>
        </div>
        <el-form-item
          v-else
          :required="fieldIsRequired(field) && !fieldReadonly(field)"
          :error="fieldError(field)"
          :style="{ gridColumn: `span ${fieldSpan(field)}` }"
        >
        <template #label>
          <span class="runtime-field-label">
            <span>{{ field.label || field.key }}</span>
            <el-button v-if="field.help" link type="primary" @click.stop="openFieldHelp(field)">
              {{ helpButtonText(field) }}
            </el-button>
          </span>
        </template>
        <el-input
          v-if="field.type === 'calculation'"
          :model-value="calculationText(field)"
          readonly
        />
        <el-input
          v-else-if="field.type === 'user' && fieldReadonly(field)"
          :model-value="userDisplayName(stringValue(field))"
          readonly
        />
        <el-input
          v-else-if="['text', 'phone', 'email', 'user', 'department'].includes(field.type)"
          :model-value="stringValue(field)"
          :placeholder="field.placeholder || placeholderFor(field)"
          :maxlength="field.maxLength || undefined"
          :readonly="fieldReadonly(field)"
          clearable
          @update:model-value="updateField(field, $event)"
        />
        <el-input
          v-else-if="field.type === 'textarea'"
          :model-value="stringValue(field)"
          :placeholder="field.placeholder || '请输入内容'"
          :maxlength="field.maxLength || undefined"
          :readonly="fieldReadonly(field)"
          type="textarea"
          :autosize="workflowTextareaAutosize(field, 3, 8)"
          :show-word-limit="Boolean(field.maxLength)"
          resize="none"
          @update:model-value="updateField(field, $event)"
        />
        <el-input-number
          v-else-if="field.type === 'number' || field.type === 'amount'"
          :model-value="numberValue(field)"
          :placeholder="field.placeholder || (field.type === 'amount' ? '请输入金额' : '请输入数字')"
          :precision="field.type === 'amount' ? 2 : undefined"
          :min="field.min"
          :max="field.max"
          :disabled="fieldReadonly(field)"
          controls-position="right"
          @update:model-value="updateField(field, $event)"
        />
        <el-tree-select
          v-else-if="isDropdownField(field) && shouldUseTreeSelect(field)"
          :model-value="field.type === 'multi_select' ? arrayValue(field) : stringValue(field)"
          :data="fieldOptions(field)"
          :multiple="field.type === 'multi_select'"
          :props="optionTreeProps"
          node-key="value"
          check-strictly
          filterable
          clearable
          :loading="fieldOptionsLoading(field)"
          :placeholder="field.placeholder || '请选择'"
          :disabled="fieldReadonly(field)"
          style="width: 100%"
          @update:model-value="updateField(field, $event)"
        />
        <el-select
          v-else-if="isDropdownField(field)"
          :model-value="field.type === 'multi_select' ? arrayValue(field) : stringValue(field)"
          :multiple="field.type === 'multi_select'"
          :placeholder="field.placeholder || '请选择'"
          :disabled="fieldReadonly(field)"
          clearable
          style="width: 100%"
          @update:model-value="updateField(field, $event)"
        >
          <el-option v-for="option in fieldOptions(field)" :key="option.value" :label="option.label" :value="option.value" />
        </el-select>
        <el-radio-group
          v-else-if="field.type === 'radio'"
          :model-value="stringValue(field)"
          :disabled="fieldReadonly(field)"
          @update:model-value="updateField(field, $event)"
        >
          <el-radio v-for="option in flatFieldOptions(field)" :key="option.value" :value="option.value">{{ option.label }}</el-radio>
        </el-radio-group>
        <el-checkbox-group
          v-else-if="field.type === 'checkbox'"
          :model-value="arrayValue(field)"
          :disabled="fieldReadonly(field)"
          @update:model-value="updateField(field, $event)"
        >
          <el-checkbox v-for="option in flatFieldOptions(field)" :key="option.value" :value="option.value">{{ option.label }}</el-checkbox>
        </el-checkbox-group>
        <el-switch
          v-else-if="field.type === 'boolean'"
          :model-value="booleanValue(field)"
          :disabled="fieldReadonly(field)"
          @update:model-value="updateField(field, $event)"
        />
        <el-date-picker
          v-else-if="field.type === 'date'"
          :model-value="stringValue(field)"
          type="date"
          value-format="YYYY-MM-DD"
          :placeholder="field.placeholder || '请选择日期'"
          :disabled="fieldReadonly(field)"
          style="width: 100%"
          @update:model-value="updateField(field, $event)"
        />
        <el-date-picker
          v-else-if="field.type === 'datetime'"
          :model-value="stringValue(field)"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          :placeholder="field.placeholder || '请选择日期时间'"
          :disabled="fieldReadonly(field)"
          style="width: 100%"
          @update:model-value="updateField(field, $event)"
        />
        <el-time-picker
          v-else-if="field.type === 'time'"
          :model-value="stringValue(field)"
          value-format="HH:mm:ss"
          :placeholder="field.placeholder || '请选择时间'"
          :disabled="fieldReadonly(field)"
          style="width: 100%"
          @update:model-value="updateField(field, $event)"
        />
        <el-date-picker
          v-else-if="field.type === 'date_range'"
          :model-value="arrayValue(field)"
          type="daterange"
          value-format="YYYY-MM-DD"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          :disabled="fieldReadonly(field)"
          style="width: 100%"
          @update:model-value="updateField(field, $event || [])"
        />
        <div v-else-if="field.type === 'detail_list'" class="detail-list-field">
          <div class="detail-list-toolbar">
            <span class="detail-list-count">{{ detailRows(field).length }} 条明细</span>
            <el-button
              v-if="canAddDetailRow(field)"
              size="small"
              type="primary"
              plain
              icon="Plus"
              @click="addDetailRow(field)"
            >新增行</el-button>
          </div>
          <div v-if="detailRows(field).length > 0" class="detail-list-rows">
            <section v-for="(row, index) in detailRows(field)" :key="detailRowKey(field, row, index)" class="detail-list-row">
              <header class="detail-list-row__head">
                <span>第 {{ index + 1 }} 行</span>
                <el-button
                  v-if="canDeleteDetailRow(field)"
                  circle
                  text
                  size="small"
                  type="danger"
                  icon="Delete"
                  title="删除行"
                  :disabled="detailRows(field).length <= detailMinRows(field)"
                  @click="removeDetailRow(field, index)"
                />
              </header>
              <div class="detail-list-columns">
                <label
                  v-for="column in detailColumns(field)"
                  :key="column.key"
                  class="detail-list-column"
                  :style="{ gridColumn: `span ${detailColumnSpan(column)}` }"
                >
                  <span class="detail-list-column__label">
                    {{ column.label || column.key }}
                    <i v-if="column.required">*</i>
                  </span>
                  <el-input
                    v-if="column.type === 'user' && fieldReadonly(field)"
                    :model-value="userDisplayName(detailStringValue(row, column))"
                    readonly
                  />
                  <el-input
                    v-else-if="['text', 'phone', 'email', 'user', 'department'].includes(column.type)"
                    :model-value="detailStringValue(row, column)"
                    :placeholder="column.placeholder || placeholderFor(column)"
                    :maxlength="column.maxLength || undefined"
                    :readonly="fieldReadonly(field)"
                    clearable
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-input
                    v-else-if="column.type === 'textarea'"
                    :model-value="detailStringValue(row, column)"
                    :placeholder="column.placeholder || '请输入内容'"
                    :maxlength="column.maxLength || undefined"
                    :readonly="fieldReadonly(field)"
                    type="textarea"
                    :autosize="workflowTextareaAutosize(column, 2, 6)"
                    :show-word-limit="Boolean(column.maxLength)"
                    resize="none"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-input-number
                    v-else-if="column.type === 'number' || column.type === 'amount'"
                    :model-value="detailNumberValue(row, column)"
                    :precision="column.type === 'amount' ? 2 : undefined"
                    :min="column.min"
                    :max="column.max"
                    :disabled="fieldReadonly(field)"
                    controls-position="right"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-tree-select
                    v-else-if="isDropdownField(column) && shouldUseTreeSelect(column)"
                    :model-value="column.type === 'multi_select' ? detailArrayValue(row, column) : detailStringValue(row, column)"
                    :data="fieldOptions(column)"
                    :multiple="column.type === 'multi_select'"
                    :props="optionTreeProps"
                    node-key="value"
                    check-strictly
                    filterable
                    clearable
                    :loading="fieldOptionsLoading(column)"
                    :placeholder="column.placeholder || '请选择'"
                    :disabled="fieldReadonly(field)"
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-select
                    v-else-if="isDropdownField(column)"
                    :model-value="column.type === 'multi_select' ? detailArrayValue(row, column) : detailStringValue(row, column)"
                    :multiple="column.type === 'multi_select'"
                    :placeholder="column.placeholder || '请选择'"
                    :disabled="fieldReadonly(field)"
                    clearable
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  >
                    <el-option v-for="option in fieldOptions(column)" :key="option.value" :label="option.label" :value="option.value" />
                  </el-select>
                  <el-radio-group
                    v-else-if="column.type === 'radio'"
                    :model-value="detailStringValue(row, column)"
                    :disabled="fieldReadonly(field)"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  >
                    <el-radio v-for="option in flatFieldOptions(column)" :key="option.value" :value="option.value">{{ option.label }}</el-radio>
                  </el-radio-group>
                  <el-checkbox-group
                    v-else-if="column.type === 'checkbox'"
                    :model-value="detailArrayValue(row, column)"
                    :disabled="fieldReadonly(field)"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  >
                    <el-checkbox v-for="option in flatFieldOptions(column)" :key="option.value" :value="option.value">{{ option.label }}</el-checkbox>
                  </el-checkbox-group>
                  <el-switch
                    v-else-if="column.type === 'boolean'"
                    :model-value="detailBooleanValue(row, column)"
                    :disabled="fieldReadonly(field)"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-date-picker
                    v-else-if="column.type === 'date'"
                    :model-value="detailStringValue(row, column)"
                    type="date"
                    value-format="YYYY-MM-DD"
                    :placeholder="column.placeholder || '请选择日期'"
                    :disabled="fieldReadonly(field)"
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-date-picker
                    v-else-if="column.type === 'datetime'"
                    :model-value="detailStringValue(row, column)"
                    type="datetime"
                    value-format="YYYY-MM-DD HH:mm:ss"
                    :placeholder="column.placeholder || '请选择日期时间'"
                    :disabled="fieldReadonly(field)"
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-time-picker
                    v-else-if="column.type === 'time'"
                    :model-value="detailStringValue(row, column)"
                    value-format="HH:mm:ss"
                    :placeholder="column.placeholder || '请选择时间'"
                    :disabled="fieldReadonly(field)"
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-date-picker
                    v-else-if="column.type === 'date_range'"
                    :model-value="detailArrayValue(row, column)"
                    type="daterange"
                    value-format="YYYY-MM-DD"
                    range-separator="至"
                    start-placeholder="开始日期"
                    end-placeholder="结束日期"
                    :disabled="fieldReadonly(field)"
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event || [])"
                  />
                  <el-input
                    v-else-if="column.type === 'user_multi' && fieldReadonly(field)"
                    :model-value="userDisplayNames(detailArrayValue(row, column))"
                    readonly
                  />
                  <el-select
                    v-else-if="column.type === 'user_multi' || column.type === 'department_multi'"
                    :model-value="detailArrayValue(row, column)"
                    multiple
                    filterable
                    allow-create
                    default-first-option
                    :placeholder="column.placeholder || placeholderFor(column)"
                    :disabled="fieldReadonly(field)"
                    style="width: 100%"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                  <el-input
                    v-else-if="column.type === 'attachment'"
                    :model-value="detailAttachmentTextModel(row, column)"
                    :placeholder="column.placeholder || '每行填写一个附件地址或标识'"
                    :readonly="fieldReadonly(field)"
                    type="textarea"
                    :rows="2"
                    resize="vertical"
                    @update:model-value="updateDetailArrayTextCell(field, index, column, $event)"
                  />
                  <el-input
                    v-else
                    :model-value="detailStringValue(row, column)"
                    :readonly="fieldReadonly(field)"
                    @update:model-value="updateDetailCell(field, index, column, $event)"
                  />
                </label>
              </div>
            </section>
          </div>
          <el-empty v-else :image-size="48" description="暂无明细" />
        </div>
        <el-input
          v-else-if="field.type === 'user_multi' && fieldReadonly(field)"
          :model-value="userDisplayNames(arrayValue(field))"
          readonly
        />
        <el-select
          v-else-if="field.type === 'user_multi' || field.type === 'department_multi'"
          :model-value="arrayValue(field)"
          multiple
          filterable
          allow-create
          default-first-option
          :placeholder="field.placeholder || placeholderFor(field)"
          :disabled="fieldReadonly(field)"
          style="width: 100%"
          @update:model-value="updateField(field, $event)"
        />
        <el-input
          v-else-if="field.type === 'attachment'"
          :model-value="attachmentTextModel(field)"
          :placeholder="field.placeholder || '每行填写一个附件地址或标识'"
          :readonly="fieldReadonly(field)"
          type="textarea"
          :rows="3"
          resize="vertical"
          @update:model-value="updateArrayTextField(field, $event)"
        />
        </el-form-item>
      </template>
    </component>
    <el-dialog
      v-model="helpDialogVisible"
      :title="activeHelp?.title || '说明'"
      width="min(520px, calc(100vw - 32px))"
      append-to-body
    >
      <div class="runtime-help-content">{{ activeHelp?.content }}</div>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue'
import { ElForm, ElMessage } from 'element-plus'
import request from '../../../utils/request'
import type { WorkflowFieldAccess, WorkflowFormField, WorkflowFormHelp, WorkflowFormOption, WorkflowOptionSource } from '../types'
import { workflowDataFields } from '../formLayout'
import { calculateWorkflowFormData, evaluateWorkflowCalculation, workflowCalculationDisplay, workflowCalculationPrecision } from '../workflowCalculation'
import {
  createWorkflowDetailRow,
  emptyWorkflowFieldValue,
  flattenWorkflowOptions,
  hasWorkflowOptionChildren,
  initialWorkflowFormData,
  normalizeWorkflowFormValue,
  normalizeWorkflowAttachments,
  normalizeWorkflowOptions,
  optionPathValue,
  visibleWorkflowFormFields,
  validateWorkflowFormData,
  workflowFieldIsRequired,
  workflowTextareaAutosize,
  workflowDetailRowKey,
  type WorkflowFieldActionMap,
  type WorkflowFieldAccessMap,
  type WorkflowFormData,
} from '../runtimeForm'

const props = withDefaults(defineProps<{
  fields: WorkflowFormField[]
  modelValue: WorkflowFormData
  fieldAccess?: WorkflowFieldAccessMap
  fieldActions?: WorkflowFieldActionMap
  readonly?: boolean
  emptyText?: string
  embedded?: boolean
  showErrors?: boolean
  userNameMap?: Record<string, string>
  calculationFields?: WorkflowFormField[]
}>(), {
  readonly: false,
  emptyText: '暂无流程表单',
  embedded: false,
  showErrors: false,
  userNameMap: () => ({}),
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: WorkflowFormData): void
  (event: 'change', value: WorkflowFormData): void
}>()

const optionTreeProps = { label: 'label', value: 'value', children: 'children' }
const remoteOptions = reactive<Record<string, WorkflowFormOption[]>>({})
const remoteOptionLoading = reactive<Record<string, boolean>>({})
const helpDialogVisible = ref(false)
const activeHelp = ref<WorkflowFormHelp | null>(null)
const showValidationErrors = ref(false)
const touchedFields = reactive<Record<string, boolean>>({})

const accessMap = computed<WorkflowFieldAccessMap>(() => {
  const result: WorkflowFieldAccessMap = {}
  for (const field of workflowDataFields(props.fields || [])) {
    if (field?.key) result[field.key] = field.type === 'calculation' ? 'read' : props.fieldAccess?.[field.key] || 'write'
  }
  return result
})

const visibleFields = computed(() => visibleWorkflowFormFields(props.fields || [], accessMap.value))
const validationErrors = computed(() => validateWorkflowFormData(props.fields || [], props.modelValue || {}, accessMap.value))
const errorsVisible = computed(() => props.showErrors || showValidationErrors.value)

const remoteOptionSignature = computed(() => {
  return collectRemoteOptionFields()
    .map((field) => optionSourceCacheKey(field))
    .filter(Boolean)
    .sort()
    .join('|')
})

watch(remoteOptionSignature, () => {
  void loadRemoteOptionFields()
}, { immediate: true })

function fieldAccess(field: WorkflowFormField): WorkflowFieldAccess {
  return accessMap.value[field.key] || 'write'
}

function fieldReadonly(field: WorkflowFormField) {
  return field.type === 'calculation' || props.readonly || fieldAccess(field) !== 'write'
}

function calculationDisplay(field: WorkflowFormField) {
  return workflowCalculationDisplay(field.calculation)
}

function calculationText(field: WorkflowFormField) {
  const result = evaluateWorkflowCalculation(field, props.modelValue || {})
  if (result.error || result.value === undefined)
    return '-'
  return result.value.toFixed(workflowCalculationPrecision(field.calculation))
}

function fieldIsRequired(field: WorkflowFormField) {
  return workflowFieldIsRequired(field, props.modelValue || {})
}

function fieldError(field: WorkflowFormField) {
  if (!errorsVisible.value && !touchedFields[field.key]) return ''
  return validationErrors.value[field.key] || ''
}

function helpButtonText(field: WorkflowFormField) {
  return field.help?.buttonText?.trim() || '查看说明'
}

function openFieldHelp(field: WorkflowFormField) {
  if (!field.help) return
  activeHelp.value = field.help
  helpDialogVisible.value = true
}

function fieldSpan(field: WorkflowFormField) {
  const span = Number(field.span || 24)
  return [6, 8, 12, 24].includes(span) ? span : 24
}

function fieldValue(field: WorkflowFormField) {
  const values = props.modelValue || {}
  if (Object.prototype.hasOwnProperty.call(values, field.key)) return values[field.key]
  return initialWorkflowFormData([field])[field.key]
}

function stringValue(field: WorkflowFormField) {
  const value = fieldValue(field)
  return typeof value === 'string' ? value : ''
}

function userDisplayName(value: string) {
  const userID = value.trim()
  return props.userNameMap?.[userID]?.trim() || userID
}

function userDisplayNames(values: string[]) {
  return values.map(userDisplayName).filter(Boolean).join('、')
}

function numberValue(field: WorkflowFormField) {
  const value = fieldValue(field)
  return typeof value === 'number' ? value : undefined
}

function booleanValue(field: WorkflowFormField) {
  return Boolean(fieldValue(field))
}

function arrayValue(field: WorkflowFormField): string[] {
  const value = normalizeWorkflowFormValue(field, fieldValue(field))
  return Array.isArray(value) ? value.map((item) => String(item)) : []
}

function attachmentTextModel(field: WorkflowFormField) {
  return normalizeWorkflowAttachments(fieldValue(field)).map((attachment) => attachment.name || attachment.url).join('\n')
}

function detailColumns(field: WorkflowFormField) {
  return Array.isArray(field.columns) ? field.columns.filter((column) => column?.key) : []
}

function isDropdownField(field: WorkflowFormField) {
  return field.type === 'select' || field.type === 'multi_select'
}

function isOptionField(field: WorkflowFormField) {
  return isDropdownField(field) || field.type === 'radio' || field.type === 'checkbox'
}

function fieldOptions(field: WorkflowFormField): WorkflowFormOption[] {
  if (fieldUsesRemoteOptions(field)) {
    const cacheKey = optionSourceCacheKey(field)
    return remoteOptions[cacheKey] || normalizeWorkflowOptions(field.options || [], field.optionSource)
  }
  return normalizeWorkflowOptions(field.options || [], field.optionSource)
}

function flatFieldOptions(field: WorkflowFormField) {
  return flattenWorkflowOptions(fieldOptions(field))
}

function shouldUseTreeSelect(field: WorkflowFormField) {
  return fieldUsesRemoteOptions(field) || hasWorkflowOptionChildren(fieldOptions(field))
}

function fieldOptionsLoading(field: WorkflowFormField) {
  return Boolean(remoteOptionLoading[optionSourceCacheKey(field)])
}

function fieldUsesRemoteOptions(field: WorkflowFormField) {
  return field.optionSource?.type === 'api'
}

function collectRemoteOptionFields() {
  const result: WorkflowFormField[] = []
  for (const field of visibleFields.value) {
    if (isOptionField(field) && fieldUsesRemoteOptions(field)) result.push(field)
    if (field.type === 'detail_list') {
      for (const column of detailColumns(field)) {
        if (isOptionField(column) && fieldUsesRemoteOptions(column)) result.push(column)
      }
    }
  }
  return result
}

async function loadRemoteOptionFields() {
  await Promise.all(collectRemoteOptionFields().map((field) => loadRemoteOptions(field)))
}

async function loadRemoteOptions(field: WorkflowFormField) {
  const source = field.optionSource
  const cacheKey = optionSourceCacheKey(field)
  if (!source || !cacheKey || remoteOptionLoading[cacheKey]) return
  if (!validBackendOptionSourceURL(source.url)) {
    remoteOptions[cacheKey] = normalizeWorkflowOptions(field.options || [], source)
    return
  }
  remoteOptionLoading[cacheKey] = true
  try {
    const method = optionSourceMethod(source)
    const response = method === 'POST'
      ? await request.post<unknown>(source.url || '', {})
      : await request.get<unknown>(source.url || '')
    remoteOptions[cacheKey] = normalizeWorkflowOptions(optionSourceResponsePayload(response, source), source)
  } catch {
    remoteOptions[cacheKey] = normalizeWorkflowOptions(field.options || [], source)
  } finally {
    remoteOptionLoading[cacheKey] = false
  }
}

function optionSourceResponsePayload(response: unknown, source: WorkflowOptionSource) {
  const path = source.responsePath?.trim() || 'data'
  if (response && typeof response === 'object' && !Array.isArray(response)) {
    const fromResponse = optionPathValue(response as Record<string, unknown>, path)
    if (fromResponse !== undefined) return fromResponse
    const responseData = (response as { data?: unknown }).data
    if (responseData && typeof responseData === 'object' && !Array.isArray(responseData)) {
      const fromData = optionPathValue(responseData as Record<string, unknown>, path)
      if (fromData !== undefined) return fromData
    }
    if (responseData !== undefined) return responseData
  }
  return response
}

function optionSourceCacheKey(field: WorkflowFormField) {
  if (!field.optionSource || field.optionSource.type !== 'api') return ''
  const source = field.optionSource
  return [
    source.url || '',
    optionSourceMethod(source),
    source.responsePath || '',
    source.labelField || '',
    source.valueField || '',
    source.childrenField || '',
  ].join('\u0001')
}

function optionSourceMethod(source: WorkflowOptionSource) {
  return source.method === 'POST' ? 'POST' : 'GET'
}

function validBackendOptionSourceURL(rawURL?: string) {
  const optionURL = String(rawURL || '').trim()
  return Boolean(optionURL) && optionURL.startsWith('/api/') && !optionURL.startsWith('//') && !optionURL.includes('://') && !/[\s]/.test(optionURL)
}

function detailRows(field: WorkflowFormField): Record<string, unknown>[] {
  const value = normalizeWorkflowFormValue(field, fieldValue(field))
  return Array.isArray(value) ? value.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object' && !Array.isArray(item)) : []
}

function detailRowKey(field: WorkflowFormField, row: Record<string, unknown>, index: number) {
  const rowID = row[workflowDetailRowKey(field)]
  return typeof rowID === 'string' && rowID.trim() ? rowID : `${field.key}_${index}`
}

function detailMinRows(field: WorkflowFormField) {
  return Math.max(0, Number(field.minRows || 0))
}

function detailMaxRows(field: WorkflowFormField) {
  return Math.max(0, Number(field.maxRows || 0))
}

function detailColumnSpan(column: WorkflowFormField) {
  return fieldSpan(column)
}

function detailCellValue(row: Record<string, unknown>, column: WorkflowFormField) {
  if (Object.prototype.hasOwnProperty.call(row, column.key)) return row[column.key]
  return emptyWorkflowFieldValue(column)
}

function detailStringValue(row: Record<string, unknown>, column: WorkflowFormField) {
  const value = detailCellValue(row, column)
  return typeof value === 'string' ? value : ''
}

function detailNumberValue(row: Record<string, unknown>, column: WorkflowFormField) {
  const value = detailCellValue(row, column)
  return typeof value === 'number' ? value : undefined
}

function detailBooleanValue(row: Record<string, unknown>, column: WorkflowFormField) {
  return Boolean(detailCellValue(row, column))
}

function detailArrayValue(row: Record<string, unknown>, column: WorkflowFormField): string[] {
  const value = normalizeWorkflowFormValue(column, detailCellValue(row, column))
  return Array.isArray(value) ? value.map((item) => String(item)) : []
}

function detailAttachmentTextModel(row: Record<string, unknown>, column: WorkflowFormField) {
  return normalizeWorkflowAttachments(detailCellValue(row, column)).map((attachment) => attachment.name || attachment.url).join('\n')
}

function detailActions(field: WorkflowFormField) {
  return props.fieldActions?.[field.key] || { add: false, delete: false }
}

function canAddDetailRow(field: WorkflowFormField) {
  return !fieldReadonly(field) && detailActions(field).add
}

function canDeleteDetailRow(field: WorkflowFormField) {
  return !fieldReadonly(field) && detailActions(field).delete
}

function emitFormData(next: WorkflowFormData) {
  emit('update:modelValue', next)
  emit('change', next)
}

function updateField(field: WorkflowFormField, value: unknown) {
  touchedFields[field.key] = true
  const next = calculateWorkflowFormData(props.calculationFields || props.fields, {
    ...(props.modelValue || {}),
    [field.key]: normalizeWorkflowFormValue(field, value),
  })
  emitFormData(next)
}

function updateArrayTextField(field: WorkflowFormField, value: string | number) {
  updateField(field, String(value || ''))
}

function updateDetailCell(field: WorkflowFormField, rowIndex: number, column: WorkflowFormField, value: unknown) {
  const rows = detailRows(field)
  const nextRows = rows.map((row, index) => {
    if (index !== rowIndex) return { ...row }
    return { ...row, [column.key]: normalizeWorkflowFormValue(column, value) }
  })
  updateField(field, nextRows)
}

function updateDetailArrayTextCell(field: WorkflowFormField, rowIndex: number, column: WorkflowFormField, value: string | number) {
  updateDetailCell(field, rowIndex, column, String(value || ''))
}

function addDetailRow(field: WorkflowFormField) {
  const rows = detailRows(field)
  const maxRows = detailMaxRows(field)
  if (maxRows > 0 && rows.length >= maxRows) {
    ElMessage.warning(`${field.label || field.key}最多允许${maxRows}行`)
    return
  }
  updateField(field, [...rows, createWorkflowDetailRow(field)])
}

function removeDetailRow(field: WorkflowFormField, rowIndex: number) {
  const rows = detailRows(field)
  const minRows = detailMinRows(field)
  if (rows.length <= minRows) {
    ElMessage.warning(`${field.label || field.key}至少需要${minRows}行`)
    return
  }
  updateField(field, rows.filter((_, index) => index !== rowIndex))
}

function placeholderFor(field: WorkflowFormField) {
  if (field.type === 'phone') return '请输入手机号'
  if (field.type === 'email') return '请输入邮箱'
  if (field.type === 'user') return '请输入用户 ID'
  if (field.type === 'user_multi') return '输入后回车添加用户 ID'
  if (field.type === 'department') return '请输入部门 ID'
  if (field.type === 'department_multi') return '输入后回车添加部门 ID'
  return '请输入内容'
}

function validate() {
  showValidationErrors.value = true
  return Object.keys(validationErrors.value).length === 0
}

function resetValidation() {
  showValidationErrors.value = false
  for (const key of Object.keys(touchedFields)) delete touchedFields[key]
}

defineExpose({ validate, resetValidation })
</script>

<style scoped>
.workflow-runtime-form { width: 100%; }
.runtime-form-grid {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  column-gap: 16px;
  row-gap: 4px;
}
.runtime-form-grid :deep(.el-form-item) { margin-bottom: 16px; }
.runtime-form-grid :deep(.el-form-item__label) {
  color: #475569;
  font-size: 13px;
  line-height: 1.4;
}
.runtime-form-grid :deep(.el-input-number) { width: 100%; }
.runtime-form-group { grid-column: 1 / -1; padding: 14px 0 4px; border-top: 1px solid #dfe6ee; }
.runtime-form-group__header { display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 32px; margin-bottom: 12px; }
.runtime-form-group__header strong { color: #273548; font-size: 15px; }
.workflow-runtime-form--embedded { min-width: 0; }
.runtime-form-label { align-self: end; margin: 8px 0 10px; color: #273548; font-size: 15px; line-height: 1.5; }
.runtime-form-description { margin: 0 0 16px; color: #64748b; font-size: 13px; line-height: 1.8; white-space: pre-wrap; }
.runtime-form-button { display: flex; align-items: flex-start; margin-bottom: 16px; }
.runtime-form-calculation-label { min-width: 0; margin-bottom: 16px; padding: 10px 12px; border-left: 3px solid #1677ff; display: flex; align-items: center; justify-content: space-between; gap: 12px; box-sizing: border-box; color: #475569; background: #f6f9ff; font-size: 13px; }
.runtime-form-calculation-label strong { color: #1677ff; font-size: 16px; }
.runtime-field-label { display: inline-flex; align-items: center; gap: 6px; min-width: 0; }
.runtime-field-label :deep(.el-button) { height: auto; padding: 0; font-size: 12px; vertical-align: baseline; }
.runtime-help-content { color: #475569; font-size: 14px; line-height: 1.8; white-space: pre-wrap; overflow-wrap: anywhere; }
.detail-list-field { width: 100%; }
.detail-list-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}
.detail-list-count { color: #64748b; font-size: 13px; }
.detail-list-rows {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.detail-list-row {
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fbfdff;
}
.detail-list-row__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 28px;
  margin-bottom: 10px;
  color: #475569;
  font-size: 13px;
  font-weight: 600;
}
.detail-list-columns {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  gap: 12px;
}
.detail-list-column {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}
.detail-list-column__label {
  color: #64748b;
  font-size: 12px;
  line-height: 1.3;
}
.detail-list-column__label i {
  margin-left: 3px;
  color: #f56c6c;
  font-style: normal;
}
@media (max-width: 768px) {
  .runtime-form-grid { grid-template-columns: minmax(0, 1fr); }
  .runtime-form-grid :deep(.el-form-item),
  .runtime-form-group,
  .runtime-form-label,
  .runtime-form-description,
  .runtime-form-button,
  .runtime-form-calculation-label { grid-column: 1 / -1 !important; }
  .detail-list-columns { grid-template-columns: minmax(0, 1fr); }
  .detail-list-column { grid-column: 1 / -1 !important; }
}
</style>
