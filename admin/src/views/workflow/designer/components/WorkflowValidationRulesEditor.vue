<template>
  <div class="validation-rules-editor" :class="{ 'validation-rules-editor--compact': compact }">
    <div class="validation-rules-heading">
      <span>{{ compact ? '高级校验' : '高级规则' }}</span>
      <el-button v-if="!readonly && ruleTypeOptions.length" link type="primary" icon="Plus" @click="addValidationRule">添加规则</el-button>
    </div>
    <section v-for="(rule, index) in rules" :key="rule.id" class="validation-rule-item">
      <header>
        <strong>规则 {{ index + 1 }}</strong>
        <el-button circle text size="small" type="danger" icon="Delete" title="删除规则" :disabled="readonly" @click="removeValidationRule(index)" />
      </header>
      <el-form-item label="规则类型">
        <el-select :model-value="rule.type" :disabled="readonly" @change="updateRuleType(rule, $event)">
          <el-option v-for="item in ruleTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>

      <div v-if="rule.type === 'min_length' || rule.type === 'max_length'" class="validation-rule-grid">
        <el-form-item :label="rule.type === 'min_length' ? '最小长度' : '最大长度'">
          <el-input-number
            v-if="rule.type === 'min_length'"
            v-model="rule.min"
            :min="0"
            :max="100000"
            :disabled="readonly"
            controls-position="right"
            @change="emitChange"
          />
          <el-input-number
            v-else
            v-model="rule.max"
            :min="0"
            :max="100000"
            :disabled="readonly"
            controls-position="right"
            @change="emitChange"
          />
        </el-form-item>
      </div>

      <template v-else-if="rule.type === 'pattern'">
        <el-form-item label="格式模板">
          <el-select :model-value="patternPreset(rule)" :disabled="readonly" @change="applyPatternPreset(rule, $event)">
            <el-option v-for="item in patternPresets" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="patternPreset(rule) === 'custom'" label="正则表达式">
          <el-input v-model="rule.pattern" maxlength="500" :disabled="readonly" @input="emitChange" />
        </el-form-item>
      </template>

      <div v-else-if="rule.type === 'number_range'" class="validation-rule-grid">
        <el-form-item label="最小值"><el-input-number v-model="rule.min" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
        <el-form-item label="最大值"><el-input-number v-model="rule.max" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
      </div>

      <el-form-item v-else-if="rule.type === 'decimal_places'" label="最大小数位">
        <el-input-number v-model="rule.precision" :min="0" :max="10" :disabled="readonly" controls-position="right" @change="emitChange" />
      </el-form-item>

      <div v-else-if="rule.type === 'selection_count'" class="validation-rule-grid">
        <el-form-item label="最少数量"><el-input-number v-model="rule.min" :min="0" :max="1000" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
        <el-form-item label="最多数量"><el-input-number v-model="rule.max" :min="0" :max="1000" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
      </div>

      <div v-else-if="rule.type === 'compare_field'" class="validation-rule-grid">
        <el-form-item label="目标字段">
          <el-select v-model="rule.field" :disabled="readonly" placeholder="选择同表单字段" @change="handleCompareFieldChange(rule)">
            <el-option v-for="item in compareFields" :key="item.key" :label="`${item.label}（${item.key}）`" :value="item.key" />
          </el-select>
        </el-form-item>
        <el-form-item label="比较关系">
          <el-select v-model="rule.operator" :disabled="readonly" @change="emitChange">
            <el-option v-for="item in fieldCompareOperators" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <p class="validation-rule-relation">
          <strong>{{ field.label || field.key }}</strong>
          <span>{{ compareOperatorLabel(rule.operator) }}</span>
          <strong>{{ compareFieldLabel(rule) }}</strong>
        </p>
      </div>

      <div v-else-if="rule.type === 'column_sum'" class="validation-rule-grid">
        <el-form-item label="合计列">
          <el-select v-model="rule.column" :disabled="readonly" @change="emitChange">
            <el-option v-for="item in summableColumns" :key="item.key" :label="item.label" :value="item.key" />
          </el-select>
        </el-form-item>
        <el-form-item label="比较关系">
          <el-select v-model="rule.operator" :disabled="readonly" @change="emitChange">
            <el-option v-for="item in compareOperators" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标值">
          <el-input-number v-model="rule.value" :disabled="readonly" controls-position="right" @change="emitChange" />
        </el-form-item>
      </div>

      <template v-else-if="rule.type === 'conditional_required'">
        <div class="validation-rule-grid">
          <el-form-item label="条件字段">
            <el-select v-model="rule.when!.field" :disabled="readonly" @change="handleConditionFieldChange(rule)">
              <el-option v-for="item in conditionFields" :key="item.key" :label="item.label" :value="item.key" />
            </el-select>
          </el-form-item>
          <el-form-item label="条件关系">
            <el-select v-model="rule.when!.operator" :disabled="readonly" @change="emitChange">
              <el-option v-for="item in conditionOperators" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
        </div>
        <el-form-item v-if="conditionNeedsValue(rule)" label="条件值">
          <el-select
            v-if="conditionField(rule)?.type === 'boolean'"
            v-model="rule.when!.value"
            :disabled="readonly"
            @change="emitChange"
          >
            <el-option label="开启" :value="true" />
            <el-option label="关闭" :value="false" />
          </el-select>
          <el-select
            v-else-if="conditionFieldOptions(rule).length"
            v-model="rule.when!.value"
            :disabled="readonly"
            clearable
            @change="emitChange"
          >
            <el-option v-for="item in conditionFieldOptions(rule)" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-else v-model="rule.when!.value" :disabled="readonly" @input="emitChange" />
        </el-form-item>
      </template>

      <el-form-item label="错误提示">
        <el-input v-model="rule.message" maxlength="200" clearable :disabled="readonly" placeholder="使用默认提示" @input="emitChange" />
      </el-form-item>
    </section>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type {
  WorkflowFormField,
  WorkflowValidationOperator,
  WorkflowValidationRule,
  WorkflowValidationRuleType,
} from '../../types'
import { workflowDataFields } from '../../formLayout'
import { flattenWorkflowOptions, normalizeWorkflowOptions } from '../../runtimeForm'
import { workflowCompareFieldCompatible, workflowCompareOperators } from '../../workflowValidationRules'

const props = withDefaults(defineProps<{
  field: WorkflowFormField
  fields: WorkflowFormField[]
  readonly?: boolean
  compact?: boolean
}>(), {
  readonly: false,
  compact: false,
})

const emit = defineEmits<{ change: [] }>()

const patternPresets = [
  { label: '身份证号码', value: 'id_card', pattern: '(^[0-9]{15}$)|(^[0-9]{17}[0-9Xx]$)' },
  { label: 'HTTP/HTTPS 地址', value: 'url', pattern: '^https?://[^\\s]+$' },
  { label: '字母和数字', value: 'alphanumeric', pattern: '^[A-Za-z0-9]+$' },
  { label: '自定义', value: 'custom', pattern: '' },
]

const compareOperators: Array<{ label: string; value: WorkflowValidationOperator }> = [
  { label: '等于', value: 'eq' }, { label: '不等于', value: 'ne' },
  { label: '大于', value: 'gt' }, { label: '大于等于', value: 'gte' },
  { label: '小于', value: 'lt' }, { label: '小于等于', value: 'lte' },
]

const conditionOperators: Array<{ label: string; value: WorkflowValidationOperator }> = [
  ...compareOperators,
  { label: '为空', value: 'empty' },
  { label: '不为空', value: 'not_empty' },
]

const rules = computed(() => {
  props.field.rules ||= []
  return props.field.rules
})

const conditionFields = computed(() => workflowDataFields(props.fields).filter(item => item.key !== props.field.key))
const compareFields = computed(() => conditionFields.value.filter(item => workflowCompareFieldCompatible(props.field.type, item.type)))
const fieldCompareOperators = computed(() => {
  const allowed = new Set(workflowCompareOperators(props.field.type))
  return compareOperators.filter(item => allowed.has(item.value))
})
const summableColumns = computed(() => (props.field.columns || []).filter(item => item.type === 'number' || item.type === 'amount'))
const ruleTypeOptions = computed(() => {
  const result: Array<{ label: string; value: WorkflowValidationRuleType }> = []
  if (['text', 'textarea', 'phone', 'email'].includes(props.field.type)) {
    result.push({ label: '最小长度', value: 'min_length' }, { label: '最大长度', value: 'max_length' }, { label: '格式匹配', value: 'pattern' })
  }
  if (props.field.type === 'number' || props.field.type === 'amount') {
    result.push({ label: '数值范围', value: 'number_range' }, { label: '小数位数', value: 'decimal_places' })
  }
  if (['multi_select', 'checkbox', 'attachment', 'user_multi', 'department_multi'].includes(props.field.type)) {
    result.push({ label: '选择数量', value: 'selection_count' })
  }
  if (props.field.type === 'detail_list' && summableColumns.value.length) {
    result.push({ label: '列合计', value: 'column_sum' })
  }
  if (compareFields.value.length) result.push({ label: '字段比较', value: 'compare_field' })
  if (conditionFields.value.length) result.push({ label: '条件必填', value: 'conditional_required' })
  return result
})

function addValidationRule() {
  const type = ruleTypeOptions.value[0]?.value
  if (!type) return
  rules.value.push(buildRule(type))
  emitChange()
}

function removeValidationRule(index: number) {
  if (props.readonly) return
  rules.value.splice(index, 1)
  emitChange()
}

function updateRuleType(rule: WorkflowValidationRule, value: string | number | boolean | undefined) {
  const type = String(value || '') as WorkflowValidationRuleType
  const replacement = buildRule(type, rule.id)
  for (const key of Object.keys(rule) as Array<keyof WorkflowValidationRule>) delete rule[key]
  Object.assign(rule, replacement)
  emitChange()
}

function buildRule(type: WorkflowValidationRuleType, existingID?: string): WorkflowValidationRule {
  const rule: WorkflowValidationRule = { id: existingID || nextRuleID(), type, message: '' }
  if (type === 'min_length') rule.min = 1
  if (type === 'max_length') rule.max = Math.max(1, Number(props.field.maxLength || 200))
  if (type === 'pattern') rule.pattern = patternPresets[2].pattern
  if (type === 'number_range') {
    rule.min = props.field.min ?? 0
    rule.max = props.field.max ?? 100
  }
  if (type === 'decimal_places') rule.precision = props.field.type === 'amount' ? 2 : 0
  if (type === 'selection_count') {
    rule.min = 1
    rule.max = 10
  }
  if (type === 'compare_field') {
    rule.field = ''
    rule.operator = workflowCompareOperators(props.field.type).includes('gte') ? 'gte' : 'eq'
  }
  if (type === 'column_sum') {
    rule.column = summableColumns.value[0]?.key || ''
    rule.operator = 'eq'
    rule.value = 100
  }
  if (type === 'conditional_required') {
    const conditionField = conditionFields.value[0]
    rule.when = { field: conditionField?.key || '', operator: 'eq', value: defaultConditionValue(conditionField) }
  }
  return rule
}

function nextRuleID() {
  let index = rules.value.length + 1
  let id = `rule_${Date.now()}_${index}`
  while (rules.value.some(item => item.id === id)) {
    index += 1
    id = `rule_${Date.now()}_${index}`
  }
  return id
}

function patternPreset(rule: WorkflowValidationRule) {
  return patternPresets.find(item => item.value !== 'custom' && item.pattern === rule.pattern)?.value || 'custom'
}

function applyPatternPreset(rule: WorkflowValidationRule, value: string | number | boolean | undefined) {
  const preset = patternPresets.find(item => item.value === value)
  if (preset && preset.value !== 'custom') rule.pattern = preset.pattern
  else rule.pattern = ''
  emitChange()
}

function conditionField(rule: WorkflowValidationRule) {
  return conditionFields.value.find(item => item.key === rule.when?.field)
}

function handleCompareFieldChange(rule: WorkflowValidationRule) {
  const allowed = workflowCompareOperators(props.field.type)
  if (!rule.operator || !allowed.includes(rule.operator)) {
    rule.operator = allowed.includes('gte') ? 'gte' : 'eq'
  }
  emitChange()
}

function compareFieldLabel(rule: WorkflowValidationRule) {
  const target = compareFields.value.find(item => item.key === rule.field)
  return target?.label || '请选择目标字段'
}

function compareOperatorLabel(operator?: WorkflowValidationOperator) {
  return compareOperators.find(item => item.value === operator)?.label || '比较'
}

function conditionFieldOptions(rule: WorkflowValidationRule) {
  const field = conditionField(rule)
  return field ? flattenWorkflowOptions(normalizeWorkflowOptions(field.options || [], field.optionSource)) : []
}

function handleConditionFieldChange(rule: WorkflowValidationRule) {
  if (!rule.when) return
  rule.when.value = defaultConditionValue(conditionField(rule))
  emitChange()
}

function defaultConditionValue(field?: WorkflowFormField) {
  if (!field) return ''
  if (field.type === 'boolean') return true
  const option = flattenWorkflowOptions(normalizeWorkflowOptions(field.options || [], field.optionSource))[0]
  return option?.value || ''
}

function conditionNeedsValue(rule: WorkflowValidationRule) {
  return rule.when?.operator !== 'empty' && rule.when?.operator !== 'not_empty'
}

function emitChange() {
  emit('change')
}
</script>

<style scoped>
.validation-rules-editor { min-width: 0; }
.validation-rules-heading { display: flex; align-items: center; justify-content: space-between; min-height: 32px; margin-bottom: 8px; color: #475569; font-size: 12px; font-weight: 600; }
.validation-rule-item { margin-top: 10px; padding: 10px; border: 1px solid #e3e8ef; border-radius: 6px; background: #fbfdff; }
.validation-rule-item > header { display: flex; align-items: center; justify-content: space-between; min-height: 28px; margin-bottom: 8px; color: #64748b; font-size: 11px; }
.validation-rule-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.validation-rule-relation { display: flex; grid-column: 1 / -1; align-items: center; min-width: 0; margin: -2px 0 8px; padding: 8px 10px; border-radius: 4px; gap: 6px; color: #64748b; background: #f1f5f9; font-size: 11px; line-height: 1.5; }
.validation-rule-relation strong { min-width: 0; overflow-wrap: anywhere; color: #334155; }
.validation-rules-editor :deep(.el-form-item) { margin-bottom: 10px; }
.validation-rules-editor--compact { margin-top: 8px; padding-top: 8px; border-top: 1px dashed #e2e8f0; }
@media (max-width: 1200px) { .validation-rule-grid { grid-template-columns: minmax(0, 1fr); } }
</style>
