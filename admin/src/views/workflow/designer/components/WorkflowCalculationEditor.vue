<template>
  <div class="calculation-editor">
    <el-form-item label="展示方式">
      <el-radio-group :model-value="display" :disabled="readonly" @change="updateDisplay">
        <el-radio-button value="label">标签结果</el-radio-button>
        <el-radio-button value="field">只读字段</el-radio-button>
      </el-radio-group>
    </el-form-item>
    <el-form-item label="小数位数">
      <el-input-number
        :model-value="precision"
        :min="0"
        :max="6"
        :disabled="readonly"
        controls-position="right"
        @change="updatePrecision"
      />
    </el-form-item>
    <el-form-item label="计算公式" required :error="formulaError">
      <el-input
        ref="formulaInput"
        :model-value="expression"
        type="textarea"
        :rows="4"
        maxlength="1000"
        show-word-limit
        resize="vertical"
        :disabled="readonly"
        placeholder="例如：[quantity] * [price] 或 SUM([items.quantity] * [items.price])"
        @input="updateExpression"
      />
    </el-form-item>

    <div class="formula-tools">
      <div class="formula-tools__group">
        <span>运算符</span>
        <div class="formula-tools__buttons">
          <el-button v-for="operator in operators" :key="operator" size="small" :disabled="readonly" @click="insertFormula(operator)">
            {{ operator }}
          </el-button>
        </div>
      </div>
      <div class="formula-tools__group">
        <span>明细聚合</span>
        <div class="formula-tools__buttons">
          <el-button v-for="aggregate in aggregates" :key="aggregate" size="small" :disabled="readonly" @click="insertAggregate(aggregate)">
            {{ aggregate }}
          </el-button>
        </div>
      </div>
      <div v-if="scalarReferences.length" class="formula-tools__group">
        <span>表单数值字段</span>
        <div class="formula-tools__references">
          <el-button v-for="reference in scalarReferences" :key="reference.token" link type="primary" :disabled="readonly" @click="insertFormula(reference.token)">
            {{ reference.label }}（{{ reference.token }}）
          </el-button>
        </div>
      </div>
      <div v-for="group in detailReferenceGroups" :key="group.key" class="formula-tools__group">
        <span>{{ group.label }}</span>
        <div class="formula-tools__references">
          <el-button v-for="reference in group.references" :key="reference.token" link type="primary" :disabled="readonly" @click="insertFormula(reference.token)">
            {{ reference.label }}（{{ reference.token }}）
          </el-button>
        </div>
      </div>
    </div>
    <p class="calculation-editor__tip">
      方括号内使用字段编码。明细列必须放入聚合函数；同一函数内可先做四则运算，例如 SUM([items.quantity] * [items.price])。
    </p>
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue'
import type { WorkflowCalculationDisplay, WorkflowFormField } from '../../types'
import { validateWorkflowCalculation, workflowCalculationDisplay, workflowCalculationPrecision, workflowCalculationReferences } from '../../workflowCalculation'

const props = defineProps<{
  field: WorkflowFormField
  fields: WorkflowFormField[]
  readonly?: boolean
}>()
const emit = defineEmits<{ change: [] }>()

const formulaInput = ref<{ textarea?: HTMLTextAreaElement }>()
const operators = [' + ', ' - ', ' * ', ' / ', '(', ')']
const aggregates = ['SUM', 'AVG', 'MIN', 'MAX', 'COUNT'] as const
const expression = computed(() => String(props.field.calculation?.expression || ''))
const display = computed(() => workflowCalculationDisplay(props.field.calculation))
const precision = computed(() => workflowCalculationPrecision(props.field.calculation))
const references = computed(() => workflowCalculationReferences(props.fields, props.field.key))
const scalarReferences = computed(() => references.value.filter(item => !item.detailKey))
const detailReferenceGroups = computed(() => {
  const groups = new Map<string, typeof references.value>()
  for (const reference of references.value) {
    if (!reference.detailKey)
      continue
    const group = groups.get(reference.detailKey) || []
    group.push(reference)
    groups.set(reference.detailKey, group)
  }
  return [...groups.entries()].map(([key, items]) => ({
    key,
    label: props.fields.flatMap(flattenFields).find(item => item.key === key)?.label || key,
    references: items,
  }))
})
const formulaError = computed(() => validateWorkflowCalculation(props.field, props.fields))

function flattenFields(field: WorkflowFormField): WorkflowFormField[] {
  return field.type === 'group' ? (field.fields || []).flatMap(flattenFields) : [field]
}

function updateCalculation(patch: Partial<NonNullable<WorkflowFormField['calculation']>>) {
  if (props.readonly)
    return
  props.field.calculation = {
    expression: '',
    display: 'field',
    precision: 2,
    ...props.field.calculation,
    ...patch,
  }
  emit('change')
}

function updateExpression(value: string) {
  updateCalculation({ expression: value })
}

function updateDisplay(value: string | number | boolean | undefined) {
  if (value === 'label' || value === 'field')
    updateCalculation({ display: value as WorkflowCalculationDisplay })
}

function updatePrecision(value: number | undefined) {
  if (typeof value === 'number')
    updateCalculation({ precision: Math.max(0, Math.min(6, Math.trunc(value))) })
}

function insertFormula(token: string) {
  if (props.readonly)
    return
  const textarea = formulaInput.value?.textarea
  const start = textarea?.selectionStart ?? expression.value.length
  const end = textarea?.selectionEnd ?? start
  const next = `${expression.value.slice(0, start)}${token}${expression.value.slice(end)}`
  updateCalculation({ expression: next })
  nextTick(() => {
    const position = start + token.length
    formulaInput.value?.textarea?.focus()
    formulaInput.value?.textarea?.setSelectionRange(position, position)
  })
}

function insertAggregate(name: typeof aggregates[number]) {
  if (props.readonly)
    return
  const textarea = formulaInput.value?.textarea
  const start = textarea?.selectionStart ?? expression.value.length
  const end = textarea?.selectionEnd ?? start
  const selected = expression.value.slice(start, end)
  const token = `${name}(${selected})`
  const next = `${expression.value.slice(0, start)}${token}${expression.value.slice(end)}`
  updateCalculation({ expression: next })
  nextTick(() => {
    const position = selected ? start + token.length : start + name.length + 1
    formulaInput.value?.textarea?.focus()
    formulaInput.value?.textarea?.setSelectionRange(position, position)
  })
}
</script>

<style scoped>
.calculation-editor { min-width: 0; }
.calculation-editor :deep(.el-radio-group) { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); width: 100%; }
.calculation-editor :deep(.el-radio-button__inner) { width: 100%; }
.calculation-editor :deep(.el-input-number) { width: 100%; }
.formula-tools { display: flex; flex-direction: column; gap: 12px; }
.formula-tools__group { display: flex; min-width: 0; flex-direction: column; gap: 7px; }
.formula-tools__group > span { color: #64748b; font-size: 11px; font-weight: 600; }
.formula-tools__buttons, .formula-tools__references { display: flex; flex-wrap: wrap; gap: 6px; }
.formula-tools__buttons :deep(.el-button), .formula-tools__references :deep(.el-button) { margin-left: 0; }
.formula-tools__references :deep(.el-button) { min-height: 28px; padding: 3px 6px; white-space: normal; }
.calculation-editor__tip { margin: 12px 0 0; color: #8492a6; font-size: 11px; line-height: 1.6; }
</style>
