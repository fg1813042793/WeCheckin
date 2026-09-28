import type {
  WorkflowCalculationDisplay,
  WorkflowFormField,
  WorkflowFormOption,
  WorkflowFormCalculation,
} from './types'

type WorkflowFormData = Record<string, unknown>

type CalculationNode =
  | { kind: 'number'; value: number }
  | { kind: 'reference'; reference: string }
  | { kind: 'unary'; operator: '+' | '-'; argument: CalculationNode }
  | { kind: 'binary'; operator: '+' | '-' | '*' | '/'; left: CalculationNode; right: CalculationNode }
  | { kind: 'aggregate'; function: CalculationAggregate; argument: CalculationNode; detailKey?: string }

type CalculationAggregate = 'SUM' | 'AVG' | 'MIN' | 'MAX' | 'COUNT'

interface CalculationSchema {
  fields: Map<string, WorkflowFormField>
  details: Array<{ key: string; columns: Map<string, WorkflowFormField> }>
}

export interface WorkflowCalculationEvaluation {
  value?: number
  error: string
}

export interface WorkflowCalculationReference {
  token: string
  label: string
  detailKey?: string
}

export interface WorkflowUnavailableCalculationOptionField {
  key: string
  label: string
  reason: string
}

export function workflowCalculationResultText(value: number, calculation?: WorkflowFormCalculation): string {
  const precision = workflowCalculationPrecision(calculation)
  const numericText = value.toFixed(precision)
  const display = calculation?.resultDisplay
  if (!display || display.mode !== 'label') return numericText
  for (const rule of display.rules || []) {
    if (rule.value !== undefined && Math.abs(value - Number(rule.value)) > 1e-9) continue
    if (rule.min !== undefined && value < Number(rule.min)) continue
    if (rule.max !== undefined && value > Number(rule.max)) continue
    if (String(rule.label || '').trim()) return String(rule.label)
  }
  return display.fallback?.trim() || numericText
}

const aggregateNames = new Set<CalculationAggregate>(['SUM', 'AVG', 'MIN', 'MAX', 'COUNT'])

export function workflowCalculationDisplay(calculation?: WorkflowFormCalculation): WorkflowCalculationDisplay {
  return calculation?.display === 'label' ? 'label' : 'field'
}

export function workflowCalculationPrecision(calculation?: WorkflowFormCalculation): number {
  const precision = calculation?.precision == null ? Number.NaN : Number(calculation.precision)
  return Number.isInteger(precision) && precision >= 0 && precision <= 6 ? precision : 2
}

export function workflowCalculationReferences(fields: WorkflowFormField[], ownerKey = ''): WorkflowCalculationReference[] {
  const result: WorkflowCalculationReference[] = []
  for (const field of flattenFields(fields)) {
    if (!field.key || field.key === ownerKey)
      continue
    if (isNumericField(field)) {
      result.push({
        token: `[${field.key}]`,
        label: isNumericOptionField(field) ? `${field.label || field.key}（选项计算值）` : (field.label || field.key),
      })
      continue
    }
    if (field.type !== 'detail_list')
      continue
    for (const column of field.columns || []) {
      if (!column.key || !isNumericField(column))
        continue
      result.push({
        token: `[${field.key}.${column.key}]`,
        label: `${field.label || field.key}.${column.label || column.key}`,
        detailKey: field.key,
      })
    }
  }
  return result
}

export function workflowUnavailableCalculationOptionFields(
  fields: WorkflowFormField[],
  ownerKey = '',
): WorkflowUnavailableCalculationOptionField[] {
  return flattenFields(fields)
    .filter(field => field.key && field.key !== ownerKey && (field.type === 'select' || field.type === 'radio'))
    .filter(field => !isNumericOptionField(field))
    .map((field) => {
      let reason = '存在未配置数字计算值的选项'
      if (!field.options?.length) {
        reason = field.optionSource?.type === 'api' ? '接口选项尚未同步' : '没有可用选项'
      }
      return {
        key: field.key,
        label: field.label || field.key,
        reason,
      }
    })
}

export function validateWorkflowCalculation(field: WorkflowFormField, fields: WorkflowFormField[]): string {
  if (field.type !== 'calculation')
    return ''
  const expression = String(field.calculation?.expression || '').trim()
  if (!expression)
    return '请输入计算公式'
  if (expression.length > 1000)
    return '计算公式不能超过1000个字符'
  if (field.calculation?.display && !['label', 'field'].includes(field.calculation.display))
    return '计算结果展示方式无效'
  const precision = Number(field.calculation?.precision ?? 2)
  if (!Number.isInteger(precision) || precision < 0 || precision > 6)
    return '小数位数必须在0到6之间'
  const resultDisplay = field.calculation?.resultDisplay
  if (resultDisplay && resultDisplay.mode && !['number', 'label'].includes(resultDisplay.mode))
    return '计算结果显示方式无效'
  for (const rule of resultDisplay?.rules || []) {
    if (!String(rule.label || '').trim()) return '计算结果标签不能为空'
    if (rule.value === undefined && rule.min === undefined && rule.max === undefined) return '计算结果标签必须配置匹配值'
    if (rule.value !== undefined && !Number.isFinite(Number(rule.value))) return '计算结果匹配值无效'
    if (rule.min !== undefined && !Number.isFinite(Number(rule.min))) return '计算结果最小值无效'
    if (rule.max !== undefined && !Number.isFinite(Number(rule.max))) return '计算结果最大值无效'
    if (rule.min !== undefined && rule.max !== undefined && Number(rule.min) > Number(rule.max)) return '计算结果标签区间无效'
  }
  try {
    const node = new CalculationParser(expression).parse()
    validateNode(node, buildSchema(fields), false)
    return ''
  }
  catch (error) {
    return calculationErrorMessage(error)
  }
}

export function evaluateWorkflowCalculation(field: WorkflowFormField, values: WorkflowFormData, fields: WorkflowFormField[] = []): WorkflowCalculationEvaluation {
  if (field.type !== 'calculation')
    return { error: '字段不是计算组件' }
  const expression = String(field.calculation?.expression || '').trim()
  if (!expression)
    return { error: '请输入计算公式' }
  try {
    const node = new CalculationParser(expression).parse()
    inferAggregateDetails(node, values)
    return { value: roundCalculationValue(evaluateNode(node, values, '', undefined, buildSchema(fields)), field.calculation), error: '' }
  }
  catch (error) {
    return { error: calculationErrorMessage(error) }
  }
}

export function calculateWorkflowFormData(fields: WorkflowFormField[], values: WorkflowFormData): WorkflowFormData {
  const result: WorkflowFormData = { ...values }
  const schema = buildSchema(fields)
  for (const field of flattenFields(fields)) {
    if (field.type !== 'calculation')
      continue
    try {
      const expression = String(field.calculation?.expression || '').trim()
      const node = new CalculationParser(expression).parse()
      validateNode(node, schema, false)
      result[field.key] = roundCalculationValue(evaluateNode(node, result, '', undefined, schema), field.calculation)
    }
    catch {
      delete result[field.key]
    }
  }
  return result
}

function roundCalculationValue(value: number, calculation?: WorkflowFormCalculation): number {
  const factor = 10 ** workflowCalculationPrecision(calculation)
  const absolute = Math.abs(value)
  const rounded = Math.round((absolute + Number.EPSILON * Math.max(1, absolute)) * factor) / factor
  const result = value < 0 ? -rounded : rounded
  if (!Number.isFinite(result))
    throw new Error('计算结果超出有效数字范围')
  return Object.is(result, -0) ? 0 : result
}

function flattenFields(fields: WorkflowFormField[]): WorkflowFormField[] {
  const result: WorkflowFormField[] = []
  for (const field of fields || []) {
    if (field.type === 'group')
      result.push(...flattenFields(field.fields || []))
    else
      result.push(field)
  }
  return result
}

function buildSchema(fields: WorkflowFormField[]): CalculationSchema {
  const schema: CalculationSchema = { fields: new Map(), details: [] }
  for (const field of flattenFields(fields)) {
    schema.fields.set(field.key, field)
    if (field.type === 'detail_list') {
      schema.details.push({
        key: field.key,
        columns: new Map((field.columns || []).map(column => [column.key, column])),
      })
    }
  }
  return schema
}

function isNumericField(field: WorkflowFormField): boolean {
  return field.type === 'number' || field.type === 'amount' || isNumericOptionField(field)
}

export function isNumericOptionField(field: Pick<WorkflowFormField, 'type' | 'options'>): boolean {
  if (field.type !== 'select' && field.type !== 'radio') return false
  if (!field.options?.length) return false
  return numericWorkflowOptions(field.options)
}

function numericWorkflowOptions(options: WorkflowFormOption[]): boolean {
  return options.every(option => {
    const valueIsNumeric = option.calculationValue !== undefined
      ? isNumericValue(option.calculationValue)
      : isNumericValue(option.value)
    return valueIsNumeric && (!option.children?.length || numericWorkflowOptions(option.children))
  })
}

function isNumericValue(value: unknown): boolean {
  if (typeof value === 'number') return Number.isFinite(value)
  if (typeof value !== 'string' || value.trim() === '') return false
  return Number.isFinite(Number(value))
}

function optionCalculationValue(field: WorkflowFormField, value: unknown): number {
  if (value === undefined || value === null || value === '') return 0
  const option = findWorkflowOption(field.options || [], value)
  if (!option) throw new Error(`选项值[${String(value)}]未配置有效计算值`)
  const number = option.calculationValue !== undefined ? Number(option.calculationValue) : Number(option.value)
  if (!Number.isFinite(number)) throw new Error(`选项值[${String(value)}]未配置有效计算值`)
  return number
}

function findWorkflowOption(options: WorkflowFormOption[], value: unknown): WorkflowFormOption | undefined {
  for (const option of options) {
    if (String(option.value) === String(value)) return option
    const nested = findWorkflowOption(option.children || [], value)
    if (nested) return nested
  }
  return undefined
}

function validateNode(node: CalculationNode, schema: CalculationSchema, inAggregate: boolean, aggregateDetail = { key: '' }): void {
  if (node.kind === 'number')
    return
  if (node.kind === 'reference') {
    if (inAggregate) {
      const detail = resolveDetailReference(schema, node.reference)
      if (detail) {
        if (!isNumericField(detail.column))
          throw new Error(`明细字段[${node.reference}]不是数字或金额`)
        if (aggregateDetail.key && aggregateDetail.key !== detail.key)
          throw new Error('一次明细聚合只能引用同一个明细列表')
        aggregateDetail.key = detail.key
        return
      }
    }
    const field = schema.fields.get(node.reference)
    if (!field) {
      if (resolveDetailReference(schema, node.reference))
        throw new Error(`明细字段[${node.reference}]必须放在SUM、AVG、MIN、MAX或COUNT中`)
      throw new Error(`字段[${node.reference}]不存在`)
    }
    if (!isNumericField(field))
      throw new Error(`字段[${node.reference}]不是数字或金额`)
    return
  }
  if (node.kind === 'unary') {
    validateNode(node.argument, schema, inAggregate, aggregateDetail)
    return
  }
  if (node.kind === 'binary') {
    validateNode(node.left, schema, inAggregate, aggregateDetail)
    validateNode(node.right, schema, inAggregate, aggregateDetail)
    return
  }
  if (inAggregate)
    throw new Error('明细聚合函数不能嵌套')
  const detail = { key: '' }
  validateNode(node.argument, schema, true, detail)
  if (!detail.key)
    throw new Error(`${node.function}函数必须引用明细列表中的数字列`)
  node.detailKey = detail.key
}

function resolveDetailReference(schema: CalculationSchema, reference: string) {
  for (const detail of schema.details) {
    const prefix = `${detail.key}.`
    if (!reference.startsWith(prefix))
      continue
    const column = detail.columns.get(reference.slice(prefix.length))
    if (column)
      return { key: detail.key, column }
  }
  return undefined
}

function inferAggregateDetails(node: CalculationNode, values: WorkflowFormData): void {
  if (node.kind === 'unary') {
    inferAggregateDetails(node.argument, values)
    return
  }
  if (node.kind === 'binary') {
    inferAggregateDetails(node.left, values)
    inferAggregateDetails(node.right, values)
    return
  }
  if (node.kind !== 'aggregate')
    return
  const detailKeys = new Set<string>()
  for (const reference of collectReferences(node.argument)) {
    const segments = reference.split('.')
    for (let index = segments.length - 1; index > 0; index -= 1) {
      const candidate = segments.slice(0, index).join('.')
      if (Array.isArray(values[candidate])) {
        detailKeys.add(candidate)
        break
      }
    }
  }
  if (detailKeys.size !== 1)
    throw new Error(`${node.function}函数必须且只能引用一个明细列表`)
  node.detailKey = [...detailKeys][0]
  inferAggregateDetails(node.argument, values)
}

function collectReferences(node: CalculationNode): string[] {
  if (node.kind === 'reference')
    return [node.reference]
  if (node.kind === 'unary' || node.kind === 'aggregate')
    return collectReferences(node.argument)
  if (node.kind === 'binary')
    return [...collectReferences(node.left), ...collectReferences(node.right)]
  return []
}

function evaluateNode(node: CalculationNode, data: WorkflowFormData, detailKey = '', row?: Record<string, unknown>, schema?: CalculationSchema): number {
  if (node.kind === 'number')
    return node.value
  if (node.kind === 'reference') {
    const prefix = detailKey ? `${detailKey}.` : ''
    const value = prefix && node.reference.startsWith(prefix)
      ? row?.[node.reference.slice(prefix.length)]
      : data[node.reference]
    const field = schema?.fields.get(node.reference)
    if (field && isNumericOptionField(field)) return optionCalculationValue(field, value)
    const detail = schema ? resolveDetailReference(schema, node.reference) : undefined
    if (detail && isNumericOptionField(detail.column)) return optionCalculationValue(detail.column, value)
    return calculationNumber(value)
  }
  if (node.kind === 'unary') {
    const value = evaluateNode(node.argument, data, detailKey, row, schema)
    return node.operator === '-' ? -value : value
  }
  if (node.kind === 'binary') {
    const left = evaluateNode(node.left, data, detailKey, row, schema)
    const right = evaluateNode(node.right, data, detailKey, row, schema)
    if (node.operator === '+')
      return left + right
    if (node.operator === '-')
      return left - right
    if (node.operator === '*')
      return left * right
    if (right === 0)
      throw new Error('不能除以0')
    return left / right
  }
  const rows = Array.isArray(data[node.detailKey || ''])
    ? (data[node.detailKey || ''] as unknown[]).filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object' && !Array.isArray(item))
    : []
  if (node.function === 'COUNT')
    return rows.length
  const aggregateValues = rows.map(item => evaluateNode(node.argument, data, node.detailKey, item, schema))
  if (aggregateValues.length === 0)
    return 0
  if (node.function === 'SUM')
    return aggregateValues.reduce((sum, value) => sum + value, 0)
  if (node.function === 'AVG')
    return aggregateValues.reduce((sum, value) => sum + value, 0) / aggregateValues.length
  if (node.function === 'MIN')
    return Math.min(...aggregateValues)
  return Math.max(...aggregateValues)
}

function calculationNumber(value: unknown): number {
  if (value === undefined || value === null || value === '')
    return 0
  const number = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(number))
    throw new Error(`字段值${String(value)}不是有效数字`)
  return number
}

function calculationErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : '计算公式无效'
}

class CalculationParser {
  private position = 0

  constructor(private readonly expression: string) {}

  parse(): CalculationNode {
    const node = this.parseExpression()
    this.skipSpaces()
    if (this.position !== this.expression.length)
      throw new Error(`位置${this.position + 1}存在无法识别的内容`)
    return node
  }

  private parseExpression(): CalculationNode {
    let left = this.parseTerm()
    while (true) {
      this.skipSpaces()
      const operator = this.peek()
      if (operator !== '+' && operator !== '-')
        return left
      this.position += 1
      left = { kind: 'binary', operator, left, right: this.parseTerm() }
    }
  }

  private parseTerm(): CalculationNode {
    let left = this.parseUnary()
    while (true) {
      this.skipSpaces()
      const operator = this.peek()
      if (operator !== '*' && operator !== '/')
        return left
      this.position += 1
      left = { kind: 'binary', operator, left, right: this.parseUnary() }
    }
  }

  private parseUnary(): CalculationNode {
    this.skipSpaces()
    const operator = this.peek()
    if (operator !== '+' && operator !== '-')
      return this.parsePrimary()
    this.position += 1
    return { kind: 'unary', operator, argument: this.parseUnary() }
  }

  private parsePrimary(): CalculationNode {
    this.skipSpaces()
    const current = this.peek()
    if (!current)
      throw new Error('计算公式不完整')
    if (current === '(') {
      this.position += 1
      const node = this.parseExpression()
      this.skipSpaces()
      if (this.peek() !== ')')
        throw new Error(`位置${this.position + 1}缺少右括号`)
      this.position += 1
      return node
    }
    if (current === '[')
      return this.parseReference()
    if (/\d|\./.test(current))
      return this.parseNumber()
    if (/[a-z]/i.test(current))
      return this.parseFunction()
    throw new Error(`位置${this.position + 1}的字符无法识别`)
  }

  private parseReference(): CalculationNode {
    const end = this.expression.indexOf(']', this.position + 1)
    if (end < 0)
      throw new Error(`位置${this.position + 1}缺少右方括号`)
    const reference = this.expression.slice(this.position + 1, end).trim()
    if (!reference)
      throw new Error('字段引用不能为空')
    this.position = end + 1
    return { kind: 'reference', reference }
  }

  private parseNumber(): CalculationNode {
    const start = this.position
    let dotSeen = false
    while (this.position < this.expression.length) {
      const current = this.peek()
      if (current === '.' && !dotSeen) {
        dotSeen = true
        this.position += 1
        continue
      }
      if (!/\d/.test(current))
        break
      this.position += 1
    }
    const text = this.expression.slice(start, this.position)
    const value = Number(text)
    if (!text || !Number.isFinite(value))
      throw new Error(`${text || '.'}不是有效数字`)
    return { kind: 'number', value }
  }

  private parseFunction(): CalculationNode {
    const start = this.position
    while (/[a-z]/i.test(this.peek()))
      this.position += 1
    const name = this.expression.slice(start, this.position).toUpperCase() as CalculationAggregate
    if (!aggregateNames.has(name))
      throw new Error(`不支持函数${name}`)
    this.skipSpaces()
    if (this.peek() !== '(')
      throw new Error(`函数${name}后缺少左括号`)
    this.position += 1
    const argument = this.parseExpression()
    this.skipSpaces()
    if (this.peek() !== ')')
      throw new Error(`函数${name}缺少右括号`)
    this.position += 1
    return { kind: 'aggregate', function: name, argument }
  }

  private skipSpaces(): void {
    while (/\s/.test(this.peek()))
      this.position += 1
  }

  private peek(): string {
    return this.expression[this.position] || ''
  }
}
