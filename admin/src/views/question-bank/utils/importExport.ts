export type BankScope = 'survey' | 'exam'

export type QuestionPayload = {
  title: string
  type: string
  category: string
  tags: string
  schema: string
}

type ExportPayloadParams = {
  scope: BankScope
  scopeName: string
  filters: {
    keyword: string
    category: string
    type: string
  }
  items: QuestionPayload[]
}

export function normalizeSchemaValue(value: unknown) {
  if (value === undefined || value === null || value === '') return ''
  if (typeof value === 'string') return value
  return JSON.stringify(value)
}

export function normalizeExportQuestion(row: any): QuestionPayload {
  return {
    title: String(row.title || ''),
    type: String(row.type || ''),
    category: String(row.category || ''),
    tags: String(row.tags || ''),
    schema: normalizeSchemaValue(row.schema)
  }
}

export function downloadJson(filename: string, data: unknown) {
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

export function exportFilename(scope: BankScope) {
  const stamp = new Date().toISOString().replace(/[:.]/g, '-')
  return `${scope === 'survey' ? 'survey' : 'exam'}-question-bank-${stamp}.json`
}

export function buildQuestionBankExportPayload(params: ExportPayloadParams) {
  return {
    version: '1.0',
    module: 'question-bank',
    scope: params.scope,
    scopeName: params.scopeName,
    exportedAt: new Date().toISOString(),
    filters: params.filters,
    total: params.items.length,
    items: params.items
  }
}

export function extractImportRows(raw: any) {
  if (Array.isArray(raw)) return raw
  const candidates = [raw?.items, raw?.list, raw?.questions, raw?.data?.items, raw?.data?.list]
  const rows = candidates.find(Array.isArray)
  if (!rows) throw new Error('未找到题目数组，请选择题库导出的 JSON 文件')
  return rows
}

export function normalizeImportQuestion(row: any, index: number, fallbackCategory = ''): QuestionPayload {
  const title = String(row?.title ?? row?.label ?? row?.question ?? row?.name ?? '').trim()
  const type = String(row?.type ?? row?.qType ?? row?.questionType ?? '').trim()
  if (!title || !type) throw new Error(`第 ${index + 1} 题缺少标题或题型`)
  return {
    title,
    type,
    category: String(row?.category || fallbackCategory || ''),
    tags: String(row?.tags || ''),
    schema: normalizeSchemaValue(row?.schema ?? row)
  }
}

export function parseImportQuestions(text: string, fallbackCategory = '') {
  let raw: any
  try {
    raw = JSON.parse(text)
  } catch {
    throw new Error('JSON 文件格式不正确')
  }
  return extractImportRows(raw).map((row: any, index: number) => normalizeImportQuestion(row, index, fallbackCategory))
}
