const pendingExports = new Set<string>()

function csvCell(value: unknown): string {
  if (value != null && !['string', 'number', 'boolean'].includes(typeof value)) {
    throw new TypeError('CSV 单元格必须为文本、数字或布尔值')
  }
  let text = String(value ?? '')
  // Quoting alone does not stop spreadsheet formulas. Also handle prefixes
  // hidden behind whitespace, control characters, BOM or zero-width marks.
  if (/^[\t\r\n]|^[\s\p{Cc}\p{Cf}]*[=+\-@]/u.test(text)) text = `'${text}`
  return `"${text.replace(/"/g, '""')}"`
}

export function buildCsv(headers: readonly string[], rows: readonly (readonly unknown[])[]): string {
  return '\uFEFF' + [headers, ...rows].map(row => row.map(csvCell).join(',')).join('\r\n')
}

// The only source of rows is the freshly completed audited request. Repeated
// clicks while that request is pending do not issue another request or download.
export async function exportCsvPage<T>(
  filename: string,
  headers: readonly string[],
  fetchRows: () => Promise<readonly T[]>,
  toRow: (row: T) => readonly unknown[],
): Promise<void> {
  if (pendingExports.has(filename)) return
  pendingExports.add(filename)
  try {
    const rows = await fetchRows()
    if (!Array.isArray(rows)) throw new Error('导出列表格式错误')
    if (rows.length > 1000) throw new Error('单次导出不得超过 1000 条')
    const csv = buildCsv(headers, rows.map(toRow))
    const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' }))
    const anchor = document.createElement('a')
    try {
      anchor.href = url
      anchor.download = filename
      document.body.appendChild(anchor)
      anchor.click()
    } finally {
      anchor.remove()
      // Let the browser start consuming the download before releasing its URL.
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    }
  } finally {
    pendingExports.delete(filename)
  }
}
