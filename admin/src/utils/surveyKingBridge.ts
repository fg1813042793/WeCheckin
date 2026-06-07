/** SurveyKing ↔ WeCheckin 双向类型/结构转换 */

// ============ 1. 类型名称映射 ============
export const SK_TO_WC: Record<string, string> = {
  Radio: 'radio', Checkbox: 'checkbox', Select: 'select',
  Cascader: 'cascade', Upload: 'file',
  FillBlank: 'input', Textarea: 'textarea', MultipleBlank: 'multiInput',
  HorzBlank: 'hInput', Barcode: 'scanCode', Signature: 'signature',
  Score: 'rating', Nps: 'nps',
  MatrixRadio: 'matrixRadio', MatrixCheckbox: 'matrixCheckbox',
  MatrixFillBlank: 'matrixFillBlank', MatrixAuto: 'matrixAuto',
  QuestionSet: 'questionSet', Remark: 'description', Pagination: 'pagination',
  SplitLine: 'divider', User: 'user', Dept: 'dept', RichText: 'richText',
  Address: 'location',
}
export const WC_TO_SK: Record<string, string> = {}
for (const [sk, wc] of Object.entries(SK_TO_WC)) WC_TO_SK[wc] = sk

export function toWcType(skType: string): string {
  return SK_TO_WC[skType] || skType
}
export function toSkType(wcType: string): string {
  return WC_TO_SK[wcType] || wcType
}

/** 判断一段 JSON 是否 SurveyKing 格式（根据根级 type 字段的 PascalCase 特征） */
export function isSurveyKingFormat(questions: any[]): boolean {
  if (!Array.isArray(questions) || !questions.length) return false
  const first = questions[0]
  if (typeof first.type !== 'string') return false
  return /^[A-Z]/.test(first.type) && SK_TO_WC[first.type] !== undefined
}

// ============ 2. 结构转换 ============

/** 递归生成短 id */
function shortId(): string {
  return Math.random().toString(36).slice(2, 6)
}

/** SurveyKing → WeCheckin */
export function importFromSurveyKing(skQuestions: any[]): any[] {
  return skQuestions.map(sk => convertOneSK(sk))
}

function convertOneSK(sk: any): any {
  const wcType = toWcType(sk.type || '')
  const wc: any = {
    id: sk.id || shortId(),
    title: sk.title || '',
    type: wcType,
    required: !!sk.attribute?.required,
    readOnly: false,
    dataType: '',
    placeholder: sk.attribute?.placeholder || '',
  }
  if (sk.description) wc.description = sk.description
  if (sk.attribute?.examAnswerMode) wc.examAnswerMode = sk.attribute.examAnswerMode

  wc.props = {}

  // Pagination
  if (wcType === 'pagination') {
    if (sk.attribute?.currentPage != null) wc.props.currentPage = sk.attribute.currentPage
    if (sk.attribute?.totalPage != null) wc.props.totalPage = sk.attribute.totalPage
  }

  // Remark / description
  if (wcType === 'description' && sk.attribute?.replaceTextRule) {
    wc.props.replaceTextRule = sk.attribute.replaceTextRule
  }

  // Matrix: rows from sk.row
  if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(wcType) && Array.isArray(sk.row)) {
    wc.props.rows = sk.row.map((r: any) => ({
      title: r.title || '',
      id: r.id || shortId(),
      width: r.attribute?.width,
    }))
  }

  // MatrixAuto & matrixFillBlank: columns from sk.children
  if (wcType === 'matrixAuto' && Array.isArray(sk.children)) {
    wc.props.columns = sk.children.map((c: any) => ({
      label: c.title || '',
      id: c.id || shortId(),
      width: c.attribute?.width,
      type: (c.attribute?.dataType === 'number' ? 'number'
             : c.attribute?.dataType === 'date' ? 'date'
             : c.attribute?.dataType === 'textarea' ? 'textarea'
             : 'input'),
    }))
    if (sk.attribute?.minRows != null) wc.props.minRows = sk.attribute.minRows
    if (sk.attribute?.maxRows != null) wc.props.maxRows = sk.attribute.maxRows
  }
  if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(wcType) && Array.isArray(sk.children)) {
    wc.props.columns = sk.children.map((c: any) => ({
      title: c.title || '',
      id: c.id || shortId(),
      width: c.attribute?.width,
    }))
  }

  // Choice types: options from sk.children
  if (['radio', 'checkbox', 'select', 'picker', 'cascade', 'judge'].includes(wcType) && Array.isArray(sk.children)) {
    wc.props.options = sk.children.map((c: any, i: number) => {
      const letter = i < 26 ? String.fromCharCode(65 + i) : `opt${i + 1}`
      return {
        label: c.title || '',
        value: c.value || letter,
        quota: c.attribute?.quota,
      }
    })
  }

  // Input types: children → options
  if (wcType === 'input' && Array.isArray(sk.children)) {
    const child0 = sk.children[0] || {}
    wc.dataType = child0.attribute?.dataType || ''
    wc.props.options = sk.children.map((c: any, i: number) => ({
      label: c.title || '',
      value: c.value || (i < 26 ? String.fromCharCode(65 + i) : `opt${i + 1}`),
      required: !!c.attribute?.required,
      readOnly: !!c.attribute?.readOnly,
      calculate: c.attribute?.calculate || '',
      dataType: c.attribute?.dataType || '',
      decimalPlaces: c.attribute?.decimalPlaces,
      minLength: c.attribute?.minLength,
      maxLength: c.attribute?.maxLength,
      suffix: c.attribute?.suffix || '',
      unique: !!c.attribute?.unique,
      placeholder: c.attribute?.placeholder || '',
    }))
  }

  // Textarea / Signature / Barcode / Upload: one child, no options
  if (['textarea', 'signature', 'scanCode', 'file'].includes(wcType) && Array.isArray(sk.children)) {
    wc.props.options = sk.children.map((c: any, i: number) => ({
      label: c.title || `字段${i + 1}`,
      value: c.value || (i < 26 ? String.fromCharCode(65 + i) : `opt${i + 1}`),
      required: !!c.attribute?.required,
      readOnly: !!c.attribute?.readOnly,
      calculate: c.attribute?.calculate || '',
      dataType: c.attribute?.dataType || '',
    }))
  }

  // Score / Nps: one child
  if (['rating', 'nps'].includes(wcType)) {
    wc.props.maxRating = sk.attribute?.maxRating || (wcType === 'nps' ? 10 : 5)
    wc.props.icon = sk.attribute?.icon || 'star'
  }

  // MultipleBlank → multiInput
  if (wcType === 'multiInput' && Array.isArray(sk.children)) {
    wc.props.fields = sk.children.map((c: any) => ({
      label: c.title || '',
      placeholder: c.attribute?.placeholder || '',
      dataType: c.attribute?.dataType || '',
      required: !!c.attribute?.required,
    }))
  }

  // HorzBlank → hInput
  if (wcType === 'hInput') {
    if (sk.attribute?.content) wc.props.content = sk.attribute.content
    if (Array.isArray(sk.children)) {
      wc.props.fields = sk.children.map((c: any) => ({
        label: c.id || '',
        placeholder: c.attribute?.placeholder || '',
        dataType: c.attribute?.dataType || '',
      }))
    }
  }

  // Cascader → cascade: children → options or dataSource → options
  if (wcType === 'cascade') {
    if (Array.isArray(sk.dataSource)) {
      wc.props.options = sk.dataSource
    } else if (Array.isArray(sk.children)) {
      wc.props.options = sk.children.map((c: any) => ({
        label: c.title || '',
        value: c.id || '',
      }))
    }
  }

  // Address → location
  if (wcType === 'location' && Array.isArray(sk.children)) {
    wc.props.options = sk.children.map((c: any) => ({
      label: c.title || '',
      value: c.id || '',
    }))
  }

  // MatrixAuto: from attribute minRows/maxRows
  if (wcType === 'matrixAuto') {
    wc.props.minRows = sk.attribute?.minRows ?? 0
    wc.props.maxRows = sk.attribute?.maxRows ?? 0
  }

  return wc
}

/** WeCheckin → SurveyKing */
export function exportToSurveyKing(wcQuestions: any[]): any[] {
  return wcQuestions.map(wc => convertOneWC(wc))
}

function convertOneWC(wc: any): any {
  const skType = toSkType(wc.type || wc.type)
  const sk: any = {
    id: wc.id || shortId(),
    type: skType,
    attribute: { required: !!wc.required },
  }
  if (wc.title) sk.title = wc.title
  if (wc.description) sk.description = wc.description

  // Pagination
  if (wc.type === 'pagination') {
    if (wc.props?.currentPage != null) sk.attribute.currentPage = wc.props.currentPage
    if (wc.props?.totalPage != null) sk.attribute.totalPage = wc.props.totalPage
  }

  // Remark / description
  if (wc.type === 'description' && wc.props?.replaceTextRule) {
    sk.attribute.replaceTextRule = wc.props.replaceTextRule
  }

  // Matrix: rows
  if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(wc.type) && Array.isArray(wc.props?.rows)) {
    sk.row = wc.props.rows.map((r: any) => ({
      title: typeof r === 'string' ? r : (r.title || ''),
      id: typeof r === 'string' ? shortId() : (r.id || shortId()),
      attribute: { width: typeof r === 'string' ? 150 : (r.width || 150) },
    }))
  }

  // Matrix: columns
  if (wc.type === 'matrixAuto' && Array.isArray(wc.props?.columns)) {
    sk.children = wc.props.columns.map((c: any) => ({
      title: c.label || '',
      id: c.id || shortId(),
      attribute: {
        width: c.width || 150,
        dataType: c.type === 'number' ? 'number' : c.type === 'date' ? 'date' : 'text',
      },
    }))
  }
  if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(wc.type) && Array.isArray(wc.props?.columns)) {
    sk.children = wc.props.columns.map((c: any) => ({
      title: typeof c === 'string' ? c : (c.title || ''),
      id: typeof c === 'string' ? shortId() : (c.id || shortId()),
      attribute: { width: typeof c === 'string' ? 150 : (c.width || 150) },
    }))
  }

  // Choice types: options → children
  if (['radio', 'checkbox', 'select', 'picker', 'cascade', 'judge'].includes(wc.type) && Array.isArray(wc.props?.options)) {
    sk.children = wc.props.options.map((o: any) => ({
      title: o.label || '',
      id: o.value || shortId(),
    }))
  }

  // Input → FillBlank
  if (wc.type === 'input' && Array.isArray(wc.props?.options)) {
    sk.children = wc.props.options.map((o: any) => ({
      id: o.value || shortId(),
      title: o.label || '',
      attribute: {
        dataType: o.dataType || wc.dataType || '',
        required: !!o.required,
        readOnly: !!o.readOnly,
        calculate: o.calculate || '',
        decimalPlaces: o.decimalPlaces,
        minLength: o.minLength,
        maxLength: o.maxLength,
        suffix: o.suffix || '',
        unique: !!o.unique,
        placeholder: o.placeholder || '',
      },
    }))
  } else if (wc.type === 'input') {
    // no options → create one child
    sk.children = [{
      id: shortId(),
      attribute: { dataType: wc.dataType || '' },
    }]
  }

  // Textarea / Signature / Barcode / Upload
  if (['textarea', 'signature', 'scanCode', 'file'].includes(wc.type)) {
    sk.children = Array.isArray(wc.props?.options) && wc.props.options.length
      ? wc.props.options.map((o: any) => ({
          id: o.value || shortId(),
          title: o.label || '',
          attribute: {
            required: !!o.required,
            readOnly: !!o.readOnly,
            calculate: o.calculate || '',
            dataType: o.dataType || '',
          },
        }))
      : [{ id: shortId() }]
  }

  // Score → Score, Nps → Nps
  if (['rating', 'nps'].includes(wc.type)) {
    sk.children = [{ id: shortId() }]
    sk.attribute.maxRating = wc.props?.maxRating || (wc.type === 'nps' ? 10 : 5)
    sk.attribute.icon = wc.props?.icon || 'star'
  }

  // MultipleBlank
  if (wc.type === 'multiInput' && Array.isArray(wc.props?.fields)) {
    sk.children = wc.props.fields.map((f: any) => ({
      title: f.label || '',
      id: shortId(),
      attribute: {
        placeholder: f.placeholder || '',
        dataType: f.dataType || '',
        required: !!f.required,
      },
    }))
  }

  // HorzBlank
  if (wc.type === 'hInput') {
    sk.attribute.content = wc.props?.content || ''
    sk.children = Array.isArray(wc.props?.fields)
      ? wc.props.fields.map((f: any) => ({
          id: f.label || shortId(),
          attribute: {
            placeholder: f.placeholder || '',
            dataType: f.dataType || '',
          },
        }))
      : []
  }

  // Cascader
  if (wc.type === 'cascade' && Array.isArray(wc.props?.options)) {
    sk.dataSource = wc.props.options
  }

  // MatrixAuto: minRows / maxRows
  if (wc.type === 'matrixAuto') {
    if (wc.props?.minRows != null) sk.attribute.minRows = wc.props.minRows
    if (wc.props?.maxRows != null) sk.attribute.maxRows = wc.props.maxRows
  }

  // Signature: examAnswerMode
  if (wc.type === 'signature' && wc.examAnswerMode) {
    sk.attribute.examAnswerMode = wc.examAnswerMode
  }
  if (wc.type === 'input' && wc.examAnswerMode) {
    sk.attribute.examAnswerMode = wc.examAnswerMode
  }

  return sk
}

/** 一键转换 questions 数组（自动检测格式并转为 WeCheckin 内部格式） */
export function normalizeQuestions(questions: any[]): any[] {
  if (!Array.isArray(questions)) return []
  if (questions.length && isSurveyKingFormat(questions)) {
    return importFromSurveyKing(questions)
  }
  return questions
}
