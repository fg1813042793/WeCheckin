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
    <el-form-item label="结果显示">
      <el-select :model-value="resultDisplayMode" :disabled="readonly" @change="updateResultDisplayMode">
        <el-option label="显示数字" value="number" />
        <el-option label="按结果显示标签" value="label" />
      </el-select>
    </el-form-item>
    <div v-if="resultDisplayMode === 'label'" class="result-label-settings">
      <div class="result-label-settings__header">
        <span>结果标签规则（按顺序匹配）</span>
        <el-button link type="primary" :disabled="readonly" @click="addResultRule">新增规则</el-button>
      </div>
      <div v-for="(rule, index) in resultDisplayRules" :key="index" class="result-label-rule">
        <el-input-number :model-value="rule.value" :disabled="readonly" :controls="false" placeholder="精确值" @change="updateResultRule(index, 'value', $event)" />
        <el-input-number :model-value="rule.min" :disabled="readonly" :controls="false" placeholder="最小值" @change="updateResultRule(index, 'min', $event)" />
        <el-input-number :model-value="rule.max" :disabled="readonly" :controls="false" placeholder="最大值" @change="updateResultRule(index, 'max', $event)" />
        <el-input :model-value="rule.label" :disabled="readonly" placeholder="显示标签" @input="updateResultRule(index, 'label', $event)" />
        <el-button circle text type="danger" :disabled="readonly" title="删除规则" @click="removeResultRule(index)">×</el-button>
      </div>
      <el-input :model-value="resultDisplayFallback" :disabled="readonly" placeholder="未匹配时显示，例如：未评级" @input="updateResultDisplayFallback" />
      <p class="result-label-settings__tip">按规则顺序匹配；可配置精确值或最小值、最大值区间，区间边界包含在内。</p>
    </div>
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
        <span>明细聚合（仅明细列表）</span>
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
      <div v-if="unavailableOptionFields.length" class="formula-tools__group formula-tools__group--warning">
        <span>暂不可计算的下拉字段</span>
        <div class="formula-tools__references">
          <el-button
            v-for="optionField in unavailableOptionFields"
            :key="optionField.key"
            link
            type="warning"
            :disabled="readonly"
            @click="emit('select-field', optionField.key)"
          >
            {{ optionField.label }}：{{ optionField.reason }}，去配置
          </el-button>
        </div>
      </div>
    </div>
    <p class="calculation-editor__tip">
      普通数字、金额和已同步计算值的下拉字段直接使用 [字段编码]；AVG、SUM 等聚合函数只用于明细列表，例如 AVG([items.score])。
    </p>
  </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue'
import type { WorkflowCalculationDisplay, WorkflowCalculationResultRule, WorkflowFormField } from '../../types'
import { validateWorkflowCalculation, workflowCalculationDisplay, workflowCalculationPrecision, workflowCalculationReferences, workflowUnavailableCalculationOptionFields } from '../../workflowCalculation'

const props = defineProps<{
  field: WorkflowFormField
  fields: WorkflowFormField[]
  readonly?: boolean
}>()
const emit = defineEmits<{
  change: []
  'select-field': [fieldKey: string]
}>()

const formulaInput = ref<{ textarea?: HTMLTextAreaElement }>()
const operators = [' + ', ' - ', ' * ', ' / ', '(', ')']
const aggregates = ['SUM', 'AVG', 'MIN', 'MAX', 'COUNT'] as const
const expression = computed(() => String(props.field.calculation?.expression || ''))
const display = computed(() => workflowCalculationDisplay(props.field.calculation))
const precision = computed(() => workflowCalculationPrecision(props.field.calculation))
const resultDisplayMode = computed(() => props.field.calculation?.resultDisplay?.mode === 'label' ? 'label' : 'number')
const resultDisplayRules = computed(() => props.field.calculation?.resultDisplay?.rules || [])
const resultDisplayFallback = computed(() => props.field.calculation?.resultDisplay?.fallback || '')
const references = computed(() => workflowCalculationReferences(props.fields, props.field.key))
const unavailableOptionFields = computed(() => workflowUnavailableCalculationOptionFields(props.fields, props.field.key))
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

function updateResultDisplayMode(value: string | number | boolean | undefined) {
  if (props.readonly || (value !== 'number' && value !== 'label')) return
  const current = props.field.calculation?.resultDisplay
  updateCalculation({
    resultDisplay: value === 'label'
      ? { mode: 'label', rules: current?.rules || [], fallback: current?.fallback || '' }
      : { mode: 'number' },
  })
}

function updateResultRule(index: number, key: keyof WorkflowCalculationResultRule, value: string | number | undefined) {
  if (props.readonly) return
  const rules = resultDisplayRules.value.map(rule => ({ ...rule }))
  if (!rules[index]) return
  if (key === 'label') rules[index].label = String(value || '')
  else if (typeof value === 'number' && Number.isFinite(value)) rules[index][key] = value
  else delete rules[index][key]
  updateCalculation({ resultDisplay: { mode: 'label', rules, fallback: resultDisplayFallback.value } })
}

function addResultRule() {
  if (props.readonly) return
  updateCalculation({
    resultDisplay: {
      mode: 'label',
      rules: [...resultDisplayRules.value, { min: undefined, max: undefined, label: '' }],
      fallback: resultDisplayFallback.value,
    },
  })
}

function removeResultRule(index: number) {
  if (props.readonly) return
  const rules = resultDisplayRules.value.filter((_, ruleIndex) => ruleIndex !== index)
  updateCalculation({ resultDisplay: { mode: 'label', rules, fallback: resultDisplayFallback.value } })
}

function updateResultDisplayFallback(value: string) {
  if (props.readonly) return
  updateCalculation({ resultDisplay: { mode: 'label', rules: resultDisplayRules.value, fallback: value } })
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
.formula-tools__group--warning { padding: 8px 10px; border: 1px solid #f4d7a1; border-radius: 6px; background: #fff8e8; }
.formula-tools__buttons, .formula-tools__references { display: flex; flex-wrap: wrap; gap: 6px; }
.formula-tools__buttons :deep(.el-button), .formula-tools__references :deep(.el-button) { margin-left: 0; }
.formula-tools__references :deep(.el-button) { min-height: 28px; padding: 3px 6px; white-space: normal; }
.calculation-editor__tip { margin: 12px 0 0; color: #8492a6; font-size: 11px; line-height: 1.6; }
.result-label-settings { margin: -2px 0 12px; padding: 10px; border: 1px solid #e5e7eb; border-radius: 6px; background: #f8fafc; }
.result-label-settings__header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; color: #64748b; font-size: 11px; font-weight: 600; }
.result-label-rule { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)) minmax(100px, 1.5fr) 28px; gap: 5px; margin-bottom: 6px; }
.result-label-rule :deep(.el-input-number), .result-label-rule :deep(.el-input) { width: 100%; }
.result-label-settings__tip { margin: 7px 0 0; color: #94a3b8; font-size: 11px; line-height: 1.5; }
</style>
