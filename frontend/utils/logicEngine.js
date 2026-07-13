function getQuestionValueByIdx(questions, idx, answers) {
  const q = questions[idx]
  if (!q) return undefined
  return answers[q.id]
}

function evalCondition(cond, questions, answers) {
  const val = getQuestionValueByIdx(questions, cond.questionIdx, answers)
  const op = cond.operator || 'eq'

  if (op === 'filled') return val !== undefined && val !== null && val !== '' && !(Array.isArray(val) && val.length === 0)
  if (op === 'empty') return val === undefined || val === null || val === '' || (Array.isArray(val) && val.length === 0)

  if (cond.optionIdx !== undefined) {
    if (Array.isArray(val)) return val.includes(cond.optionIdx) || val.includes(String(cond.optionIdx))
    return String(val) === String(cond.optionIdx)
  }

  const cmp = cond.compareValue ?? ''
  switch (op) {
    case 'eq': return String(val) === String(cmp)
    case 'neq': return String(val) !== String(cmp)
    case 'gt': return Number(val) > Number(cmp)
    case 'lt': return Number(val) < Number(cmp)
    case 'gte': return Number(val) >= Number(cmp)
    case 'lte': return Number(val) <= Number(cmp)
    case 'contains': return String(val).includes(String(cmp))
    case 'notContains': return !String(val).includes(String(cmp))
    default: return String(val) === String(cmp)
  }
}

function evalConditions(rule, questions, answers) {
  if (rule.conditionType === 'none' || !rule.conditions || !rule.conditions.length) return true
  const results = rule.conditions.map(c => evalCondition(c, questions, answers))
  if (rule.conditionType === 'or') return results.some(Boolean)
  return results.every(Boolean)
}

export function evaluateFrontendRules(rules, questions, answers) {
  const visible = new Set(questions.map(q => q.id))
  const required = new Set()

  for (const rule of rules) {
    if (!evalConditions(rule, questions, answers)) continue
    const targetId = rule.targetQuestionIdx !== undefined ? questions[rule.targetQuestionIdx]?.id : undefined
    switch (rule.action) {
      case 'show':
        if (targetId) visible.add(targetId)
        break
      case 'hide':
        if (targetId) visible.delete(targetId)
        break
      case 'required':
        if (targetId) required.add(targetId)
        break
    }
  }

  const hiddenIds = new Set()
  for (const q of questions) {
    if (!visible.has(q.id)) hiddenIds.add(q.id)
  }
  return { hiddenIds, requiredIds: required }
}
