<script setup lang="ts">
import type {
  WorkflowFieldAccessMap,
  WorkflowFieldActionsMap,
  WorkflowFormData,
  WorkflowFormField,
} from '@/types/workflow'
import { computed, ref } from 'vue'
import {
  calculateWorkflowFormData,
  evaluateWorkflowCalculation,
  workflowCalculationDisplay,
  workflowCalculationPrecision,
} from '../workflow-calculation'
import {
  cloneWorkflowValue,
  createWorkflowDetailRow,
  emptyWorkflowFieldValue,
  validateWorkflowFormData,
  visibleWorkflowFormFields,
  workflowDetailRowKey,
  workflowFieldAccessMap,
} from '../workflow-form'
import WorkflowFieldControl from './WorkflowFieldControl.vue'

defineOptions({ name: 'WorkflowRuntimeForm' })

const props = withDefaults(defineProps<{
  fields: WorkflowFormField[]
  modelValue: WorkflowFormData
  fieldAccess?: WorkflowFieldAccessMap
  fieldActions?: WorkflowFieldActionsMap
  readonly?: boolean
  readonlyAppearance?: 'disabled' | 'plain'
  embedded?: boolean
  showErrors?: boolean
  calculationFields?: WorkflowFormField[]
}>(), {
  readonly: false,
  readonlyAppearance: 'disabled',
  embedded: false,
  showErrors: false,
  fieldAccess: () => ({}),
  fieldActions: () => ({}),
})

const emit = defineEmits<{
  'update:modelValue': [value: WorkflowFormData]
  'change': [value: WorkflowFormData]
}>()

const localShowErrors = ref(false)
const helpVisible = ref(false)
const activeHelp = ref({ title: '字段说明', content: '' })
const accessMap = computed(() => {
  const defaults = workflowFieldAccessMap(props.fields, [], props.readonly ? 'read' : 'write')
  return { ...defaults, ...props.fieldAccess }
})
const visibleFields = computed(() => visibleWorkflowFormFields(props.fields, accessMap.value))
const validationErrors = computed(() => validateWorkflowFormData(props.fields, props.modelValue, accessMap.value))
const errorsVisible = computed(() => props.showErrors || localShowErrors.value)

function fieldReadonly(field: WorkflowFormField) {
  return field.type === 'calculation' || props.readonly || accessMap.value[field.key] !== 'write'
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

function fieldSpanStyle(field: WorkflowFormField) {
  const span = [6, 8, 12, 24].includes(Number(field.span || 24)) ? Number(field.span || 24) : 24
  return { width: `${span / 24 * 100}%` }
}

function fieldError(field: WorkflowFormField) {
  return errorsVisible.value ? validationErrors.value[field.key] || '' : ''
}

function updateField(field: WorkflowFormField, value: unknown) {
  const next = calculateWorkflowFormData(props.calculationFields || props.fields, {
    ...props.modelValue,
    [field.key]: cloneWorkflowValue(value),
  })
  emit('update:modelValue', next)
  emit('change', next)
}

function detailRows(field: WorkflowFormField): Array<Record<string, unknown>> {
  const value = props.modelValue[field.key]
  return Array.isArray(value)
    ? value.filter(item => item && typeof item === 'object' && !Array.isArray(item)) as Array<Record<string, unknown>>
    : []
}

function detailRowId(field: WorkflowFormField, row: Record<string, unknown>, index: number) {
  return String(row[workflowDetailRowKey(field)] || '') || `${field.key}_${index}`
}

function canAddRow(field: WorkflowFormField) {
  if (fieldReadonly(field))
    return false
  const actions = props.fieldActions[field.key] || []
  const maxRows = Number(field.maxRows || 0)
  return actions.includes('add') && (!maxRows || detailRows(field).length < maxRows)
}

function canDeleteRow(field: WorkflowFormField) {
  if (fieldReadonly(field))
    return false
  return (props.fieldActions[field.key] || []).includes('delete')
}

function addDetailRow(field: WorkflowFormField) {
  if (!canAddRow(field))
    return
  updateField(field, [...detailRows(field), createWorkflowDetailRow(field)])
}

function removeDetailRow(field: WorkflowFormField, index: number) {
  if (!canDeleteRow(field) || detailRows(field).length <= Number(field.minRows || 0))
    return
  updateField(field, detailRows(field).filter((_, rowIndex) => rowIndex !== index))
}

function updateDetailCell(field: WorkflowFormField, rowIndex: number, column: WorkflowFormField, value: unknown) {
  const rows = detailRows(field).map(row => ({ ...row }))
  if (!rows[rowIndex])
    return
  rows[rowIndex][column.key] = value ?? emptyWorkflowFieldValue(column)
  updateField(field, rows)
}

function openHelp(field: WorkflowFormField) {
  if (!field.help)
    return
  activeHelp.value = {
    title: field.help.title || field.label || '字段说明',
    content: field.help.content || '',
  }
  helpVisible.value = true
}

function validate() {
  localShowErrors.value = true
  const errors = validateWorkflowFormData(props.fields, props.modelValue, accessMap.value)
  return { valid: Object.keys(errors).length === 0, errors }
}

defineExpose({ validate })
</script>

<template>
  <view
    class="workflow-form app-workflow-form app-pc-control-scope"
    :class="{
      'workflow-form--embedded': embedded,
      'workflow-form--plain-readonly': readonlyAppearance === 'plain',
      'workflow-form--fully-readonly': readonly && readonlyAppearance === 'plain',
    }"
  >
    <view v-if="visibleFields.length === 0" class="workflow-form__empty">
      <u-icon name="file-text" size="54" color="#c8c9cc" />
      <text>暂无流程表单</text>
    </view>
    <view v-else class="workflow-form__grid">
      <template v-for="field in visibleFields" :key="field.key">
        <view v-if="field.type === 'group'" class="workflow-form__group">
          <view class="workflow-form__group-header">
            <text>{{ field.label || field.key }}</text>
            <u-icon v-if="field.help" class="workflow-form__help-action" name="info-circle" size="28" color="#2979ff" @click="openHelp(field)" />
          </view>
          <WorkflowRuntimeForm
            :fields="field.fields || []"
            :model-value="modelValue"
            :field-access="fieldAccess"
            :field-actions="fieldActions"
            :readonly="readonly"
            :readonly-appearance="readonlyAppearance"
            :calculation-fields="calculationFields || fields"
            embedded
            :show-errors="errorsVisible"
            @update:model-value="emit('update:modelValue', $event)"
            @change="emit('change', $event)"
          />
        </view>
        <view v-else-if="field.type === 'label'" class="workflow-form__label" :style="fieldSpanStyle(field)">
          {{ field.label }}
        </view>
        <view
          v-else-if="field.type === 'calculation' && calculationDisplay(field) === 'label'"
          class="workflow-form__calculation-label"
          :style="fieldSpanStyle(field)"
        >
          <text>{{ field.label || field.key }}</text>
          <text class="workflow-form__calculation-value">
            {{ calculationText(field) }}
          </text>
        </view>
        <view v-else-if="field.type === 'description'" class="workflow-form__description" :style="fieldSpanStyle(field)">
          {{ field.content }}
        </view>
        <view v-else-if="field.type === 'button'" class="workflow-form__layout-button" :style="fieldSpanStyle(field)">
          <u-button size="small" plain @click="openHelp(field)">
            {{ field.label }}
          </u-button>
        </view>
        <view v-else class="workflow-form__field" :style="fieldSpanStyle(field)">
          <view class="workflow-form__field-label">
            <text>{{ field.label || field.key }}</text>
            <text v-if="field.required && !fieldReadonly(field)" class="workflow-form__required">
              *
            </text>
            <u-icon v-if="field.help" class="workflow-form__help-action" name="info-circle" size="25" color="#2979ff" @click="openHelp(field)" />
          </view>
          <view v-if="field.type === 'detail_list'" class="workflow-detail">
            <view class="workflow-detail__toolbar">
              <text>共 {{ detailRows(field).length }} 行</text>
            </view>
            <view v-if="detailRows(field).length === 0" class="workflow-detail__empty">
              暂无明细
            </view>
            <view
              v-for="(row, rowIndex) in detailRows(field)"
              v-else
              :key="detailRowId(field, row, rowIndex)"
              class="workflow-detail__row"
            >
              <view class="workflow-detail__row-header">
                <text>第 {{ rowIndex + 1 }} 行</text>
                <u-icon
                  v-if="canDeleteRow(field)"
                  name="trash"
                  size="30"
                  color="#fa3534"
                  @click="removeDetailRow(field, rowIndex)"
                />
              </view>
              <view class="workflow-detail__columns">
                <view
                  v-for="column in field.columns || []"
                  :key="column.key"
                  class="workflow-detail__column"
                  :style="fieldSpanStyle(column)"
                >
                  <view class="workflow-form__field-label">
                    <text>{{ column.label || column.key }}</text>
                    <text v-if="column.required && !fieldReadonly(field)" class="workflow-form__required">
                      *
                    </text>
                  </view>
                  <WorkflowFieldControl
                    :field="column"
                    :model-value="row[column.key]"
                    :readonly="fieldReadonly(field)"
                    :readonly-appearance="readonlyAppearance"
                    :textarea-default-min-rows="2"
                    :textarea-default-max-rows="6"
                    @update:model-value="updateDetailCell(field, rowIndex, column, $event)"
                  />
                </view>
              </view>
            </view>
            <view v-if="canAddRow(field)" class="workflow-detail__add">
              <u-button
                custom-class="workflow-detail__add-button"
                size="mini"
                type="primary"
                plain
                icon="plus"
                @click="addDetailRow(field)"
              >
                新增行
              </u-button>
            </view>
          </view>
          <view v-else-if="field.type === 'calculation'" class="workflow-form__calculation-field">
            {{ calculationText(field) }}
          </view>
          <WorkflowFieldControl
            v-else
            :field="field"
            :model-value="modelValue[field.key]"
            :readonly="fieldReadonly(field)"
            :readonly-appearance="readonlyAppearance"
            @update:model-value="updateField(field, $event)"
          />
          <text v-if="fieldError(field)" class="workflow-form__error">
            {{ fieldError(field) }}
          </text>
        </view>
      </template>
    </view>
    <u-modal
      v-model="helpVisible"
      :title="activeHelp.title"
      :content="activeHelp.content"
      :z-index="10120"
      custom-class="workflow-help-modal app-pc-control-scope"
      width="640rpx"
      confirm-text="知道了"
      :show-cancel-button="false"
      :mask-close-able="true"
      :content-style="{ whiteSpace: 'pre-wrap', textAlign: 'left' }"
    />
  </view>
</template>

<style lang="scss" scoped>
.workflow-form__grid,
.workflow-detail__columns {
  display: flex;
  flex-wrap: wrap;
  margin: -10rpx;
}

.workflow-form__field,
.workflow-detail__column,
.workflow-form__label,
.workflow-form__calculation-label,
.workflow-form__description,
.workflow-form__layout-button {
  min-width: 0;
  padding: 10rpx;
  box-sizing: border-box;
}

.workflow-form__field-label {
  min-height: 42rpx;
  margin-bottom: 10rpx;
  display: flex;
  align-items: center;
  gap: 8rpx;
  color: $u-main-color;
  font-size: 25rpx;
  font-weight: 600;
}

.workflow-form__required,
.workflow-form__error {
  color: $u-type-error;
}

.workflow-form__error {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  line-height: 1.4;
}

.workflow-form__group {
  width: 100%;
  margin: 12rpx 10rpx;
  padding: 24rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  background: #f8fafc;
  box-sizing: border-box;
}

.workflow-form__group-header {
  margin-bottom: 18rpx;
  display: flex;
  align-items: center;
  gap: 10rpx;
  color: $u-main-color;
  font-size: 28rpx;
  font-weight: 700;
}

.workflow-form__label {
  color: $u-main-color;
  font-size: 30rpx;
  font-weight: 700;
}

.workflow-form__calculation-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20rpx;
  border-left: 6rpx solid $u-type-primary;
  background: #f4f8ff;
  color: $u-content-color;
  font-size: 25rpx;
}

.workflow-form__calculation-value {
  color: $u-type-primary;
  font-size: 30rpx;
  font-weight: 700;
}

.workflow-form__calculation-field {
  min-height: 70rpx;
  padding: 0 20rpx;
  display: flex;
  align-items: center;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  background: #f5f7fa;
  color: $u-main-color;
  font-size: 26rpx;
  box-sizing: border-box;
}

.workflow-form__description {
  color: $u-content-color;
  font-size: 24rpx;
  line-height: 1.7;
}

.workflow-form__empty,
.workflow-detail__empty {
  min-height: 150rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  color: $u-tips-color;
  font-size: 24rpx;
}

.workflow-detail {
  padding: 18rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  background: #f8fafc;
}

.workflow-detail__toolbar,
.workflow-detail__row-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.workflow-detail__toolbar {
  margin-bottom: 16rpx;
  justify-content: flex-start;
  color: $u-content-color;
  font-size: 23rpx;
}

.workflow-detail__add {
  margin-top: 18rpx;
  display: flex;
  justify-content: center;
}

.workflow-detail__add-button,
:deep(.workflow-detail__add-button) {
  width: auto;
  min-width: 160rpx;
  margin: 0;
}

.workflow-detail__row {
  margin-top: 14rpx;
  padding: 18rpx;
  border: 1px solid $u-border-color;
  border-radius: 8rpx;
  background: #fff;
}

.workflow-detail__row-header {
  margin-bottom: 12rpx;
  color: $u-main-color;
  font-size: 24rpx;
  font-weight: 700;
}

@media screen and (min-width: 769px) {
  .workflow-form--fully-readonly {
    :deep(.workflow-form__help-action) {
      display: none !important;
    }
  }

  .workflow-form--plain-readonly {
    :deep(.u-input--disabled),
    :deep(.u-textarea--disabled),
    :deep(.workflow-picker--disabled) {
      background-color: #fff !important;
    }

    :deep(.u-input--disabled .u-input__input),
    :deep(.u-input--disabled .u-input__textarea),
    :deep(.u-textarea--disabled .u-textarea__field),
    :deep(.workflow-picker--disabled) {
      color: $u-main-color !important;
      -webkit-text-fill-color: $u-main-color !important;
    }

    :deep(.u-radio__label--disabled),
    :deep(.u-checkbox__label--disabled) {
      color: $u-content-color !important;
    }

    :deep(.u-radio__icon-wrap--disabled),
    :deep(.u-checkbox__icon-wrap--disabled) {
      background-color: #fff !important;
    }

    :deep(.u-radio__icon-wrap--disabled--checked),
    :deep(.u-checkbox__icon-wrap--disabled--checked) {
      border-color: $u-type-primary !important;
      background-color: $u-type-primary !important;
      color: var(--u-white-color) !important;
    }

    :deep(.u-switch--disabled) {
      opacity: 1;
    }

    .workflow-form__group,
    .workflow-detail {
      background: #fff;
    }
  }

  :deep(.workflow-help-modal .u-mode-center-box) {
    width: 680px !important;
    max-width: calc(100vw - 64px);
  }
}

@media screen and (max-width: 768px) {
  .workflow-form__field,
  .workflow-detail__column,
  .workflow-form__label,
  .workflow-form__calculation-label,
  .workflow-form__description,
  .workflow-form__layout-button {
    width: 100% !important;
  }

  .workflow-form__group {
    padding: 20rpx;
  }
}
</style>
