import { h } from 'vue'

export function readInputValue(event) {
  return event.detail?.value ?? event.target.value
}

export function estimateTextareaRows(value, minRows = 3, maxRows = 18, charsPerRow = 21) {
  const text = String(value || '')
  if (!text) return minRows
  const rows = text.split(/\r\n|\n|\r/).reduce((total, line) => {
    const visualLength = Array.from(line).reduce((sum, char) => {
      if (/\s/.test(char)) return sum + 0.35
      return sum + (char.charCodeAt(0) <= 255 ? 0.55 : 1)
    }, 0)
    return total + Math.max(1, Math.ceil(visualLength / charsPerRow))
  }, 0)
  return Math.max(minRows, Math.min(maxRows, rows))
}

export function textareaAutoHeightStyle(value, minRows = 3, maxRows = 18, charsPerRow = 21) {
  const rows = estimateTextareaRows(value, minRows, maxRows, charsPerRow)
  return {
    minHeight: `${rows * 22 + 26}px`,
    height: `${rows * 22 + 26}px`
  }
}

export function createObjective() {
  return {
    target: '',
    weight: 0,
    completion: '',
    result: ''
  }
}

export function createNextObjective() {
  return {
    target: '',
    weight: 0
  }
}

export function addCurrentObjective(review) {
  if (!Array.isArray(review.objectives)) review.objectives = []
  review.objectives.push(createObjective())
}

export function addNextObjective(review) {
  if (!Array.isArray(review.nextObjectives)) review.nextObjectives = []
  review.nextObjectives.push(createNextObjective())
}

export function removeCurrentObjective(review, index) {
  if (!Array.isArray(review.objectives)) return
  review.objectives.splice(index, 1)
}

export function removeNextObjective(review, index) {
  if (!Array.isArray(review.nextObjectives)) return
  review.nextObjectives.splice(index, 1)
}

export function confirmObjectiveDelete(index, label = '目标') {
  const content = `确认删除${label} ${index + 1}？删除后需要保存才会生效。`
  return new Promise((resolve) => {
    if (typeof uni !== 'undefined' && uni.showModal) {
      uni.showModal({
        title: `删除${label}`,
        content,
        confirmText: '删除',
        confirmColor: '#e34d59',
        success: (res) => resolve(Boolean(res.confirm)),
        fail: () => resolve(false)
      })
      return
    }
    if (typeof window !== 'undefined' && window.confirm) {
      resolve(window.confirm(content))
      return
    }
    resolve(false)
  })
}

export async function confirmRemoveCurrentObjective(review, index, onRemoved) {
  const confirmed = await confirmObjectiveDelete(index)
  if (!confirmed) return
  removeCurrentObjective(review, index)
  if (onRemoved) onRemoved()
}

export async function confirmRemoveNextObjective(review, index, onRemoved) {
  const confirmed = await confirmObjectiveDelete(index, '下月目标')
  if (!confirmed) return
  removeNextObjective(review, index)
  if (onRemoved) onRemoved()
}

export function hasManagerEvaluation(review) {
  if (!review) return false
  if (String(review.managerGrade || '').trim()) return true
  if (String(review.managerComment || '').trim()) return true
  return (review.values || []).some((item) => String(item.manager ?? '').trim())
}

export function hasHrbpEvaluation(review) {
  if (!review) return false
  if (String(review.hrbpGrade || '').trim()) return true
  if (String(review.hrbpComment || '').trim()) return true
  return (review.values || []).some((item) => String(item.hrbp ?? '').trim())
}

export function managerReviewTitleMeta(ctx, review) {
  const name = ctx.userName(review?.managerId)
  return name && name !== '无' ? `上级：${name}` : '上级未设置'
}

export function hrbpReviewTitleMeta(ctx, review) {
  const name = ctx.userName(review?.hrbpReviewerId || review?.hrbpId)
  return name && name !== '无' ? `HRBP：${name}` : 'HRBP未设置'
}

export function reviewGradeBadge(label, grade) {
  const value = String(grade || '').trim()
  if (!value) return null
  return h('text', { class: 'review-grade-badge' }, `${label}：${value}`)
}

export function valueRubricItems(tpl, item) {
  const source = Array.isArray(tpl?.rubric) && tpl.rubric.length > 0 ? tpl.rubric : item?.rubric
  return (Array.isArray(source) ? source : []).filter((rubric) =>
    String(rubric?.label || '').trim() ||
    String(rubric?.score ?? '').trim() ||
    String(rubric?.description || '').trim()
  )
}

export function valueRubricScoreText(score) {
  const value = Number(score)
  if (!Number.isFinite(value)) return '-'
  return Number.isInteger(value) ? String(value) : String(value.toFixed(1)).replace(/\.0$/, '')
}

export function renderValueRubricList(rubrics) {
  if (!rubrics.length) return null
  return h('view', { class: 'value-score-guide-list' }, rubrics.map((rubric) => {
    const label = String(rubric?.label || '').trim() || '未命名'
    const description = String(rubric?.description || '').trim()
    return h('view', { class: 'value-score-guide-item' }, [
      h('text', { class: 'value-score-guide-score' }, `${valueRubricScoreText(rubric?.score)}分`),
      h('view', { class: 'value-score-guide-copy' }, [
        h('text', { class: 'value-score-guide-name' }, label),
        description ? h('text', { class: 'value-score-guide-desc' }, description) : null
      ])
    ])
  }))
}
