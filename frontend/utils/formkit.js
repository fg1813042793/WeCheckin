// 表单工具：老/新 schema/answer 兼容转换
// 供 H5 表单运行时使用

// 老格式 field
// { label: string, type: string, required?: bool, placeholder?: string, options?: [{label,value}] }
// 老格式 answer
// ["value1", "value2", ...]

// 新格式 schema (formkit)
// { version: "2.0", questions: [{ id, type, title, required, placeholder, props, validate, logic, calcValue }] }
// 新格式 answer
// { "q1": "value1", "q2": "value2" }

// 检测 schema 格式
export function isOldSchema(schemaStr) {
  if (!schemaStr) return true
  try {
    const s = typeof schemaStr === 'string' ? JSON.parse(schemaStr) : schemaStr
    if (Array.isArray(s)) return true
    if (s && s.version === '2.0' && Array.isArray(s.questions)) return false
    return true
  } catch (e) {
    return true
  }
}

// 把 schema 统一转换为 questions 数组（不区分老新）
// 每项规范化为 { id, type, title, required, placeholder, options, validate, logic, calcValue, _oldIndex }
export function normalizeSchema(schemaStr) {
  if (!schemaStr) return []
  let s
  try {
    s = typeof schemaStr === 'string' ? JSON.parse(schemaStr) : schemaStr
  } catch (e) {
    return []
  }
  if (isOldSchema(schemaStr)) {
    // 老格式
    return (s || []).map((f, i) => ({
      id: `q${i + 1}`,
      type: f.type || 'input',
      title: f.label || `字段${i + 1}`,
      required: !!f.required,
      placeholder: f.placeholder || '',
      options: f.options || [],
      _oldIndex: i
    }))
  }
  // 新格式
  return (s.questions || []).map((q) => ({
    id: q.id,
    type: q.type || 'input',
    title: q.title || q.id,
    description: q.description || '',
    required: !!q.required,
    placeholder: q.placeholder || '',
    options: (q.props && q.props.options) || [],
    validate: q.validate || [],
    logic: q.logic || [],
    calcValue: q.calcValue || null
  }))
}

// 初始化 answers（新格式用对象，老格式用数组）
export function initAnswers(questions, isOld) {
  if (isOld) {
    return questions.map(() => '')
  }
  const obj = {}
  for (const q of questions) {
    obj[q.id] = q.type === 'checkbox' ? [] : ''
  }
  return obj
}

// 从 answers 提取某个 question 的当前值
export function getAnswerValue(answers, q, isOld) {
  if (isOld) {
    return answers[q._oldIndex]
  }
  return answers[q.id]
}

// 设置某个 question 的值
export function setAnswerValue(answers, q, value, isOld) {
  if (isOld) {
    answers[q._oldIndex] = value
  } else {
    answers[q.id] = value
  }
}

// 序列化为可提交格式（数组或对象转 JSON）
// 老格式：保持数组 → JSON.stringify
// 新格式：保持对象 → JSON.stringify
export function serializeAnswers(answers) {
  if (Array.isArray(answers)) return JSON.stringify(answers)
  return JSON.stringify(answers)
}

// 校验必填
export function validateRequired(questions, answers, isOld) {
  for (const q of questions) {
    if (q.type === 'divider' || q.type === 'description') continue
    if (q.type === 'autopop') continue
    if (!q.required) continue
    const v = getAnswerValue(answers, q, isOld)
    if (v === '' || v === null || v === undefined) return q
    if (Array.isArray(v) && v.length === 0) return q
  }
  return null
}
