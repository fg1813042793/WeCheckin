import { computed, watch, type Ref } from 'vue'
import type { ExamDesignerQuestion } from '../exam-designer-types'
import { questionScore } from '../examScore'

export function useExamDesignerScore(questions: Ref<ExamDesignerQuestion[]>, selected: Ref<ExamDesignerQuestion | null>) {
  const partialTotalScore = computed(() => {
    if (!selected.value || selected.value.examScoreMode === 'none') return 0
    const items = ['multiInput','hInput'].includes(selected.value.type) ? (selected.value.props?.fields||[]) : (selected.value.props?.options||[])
    return items.reduce((sum: number, o: any) => sum + (o.examCorrect ? (Number(o.examScore) || 0) : 0), 0)
  })
  const perOptionTotalScore = computed(() => {
    if (!selected.value || selected.value.examScoreMode === 'none') return 0
    const items = ['multiInput','hInput'].includes(selected.value.type) ? (selected.value.props?.fields||[]) : (selected.value.props?.options||[])
    return items.reduce((sum: number, o: any) => sum + (Number(o.examScore) || 0), 0)
  })

  const paperTotalScore = computed(() => questions.value.reduce((sum, q) => sum + questionScore(q), 0))

  watch([partialTotalScore, perOptionTotalScore], () => {
    if (!selected.value) return
    if (selected.value.examScoreMode === 'partialCorrect') selected.value.examScore = partialTotalScore.value
    else if (selected.value.examScoreMode === 'perOption') selected.value.examScore = perOptionTotalScore.value
  })

  watch(() => [selected.value, selected.value?.examScoreMode] as const, ([question, cur], [previousQuestion, prev]) => {
    if (!question) return
    if (cur === 'none' || (question === previousQuestion && prev && !cur)) {
      question.examScoreMode = 'none'
      question.examScore = 0
      question.examCorrectAnswer = undefined
      return
    }
    if (question !== previousQuestion || !prev) return
    if (prev === 'single' && cur !== 'manual') {
      question.examCorrectAnswer = undefined
    }
    if (prev === 'perOption' || prev === 'partialCorrect') {
      question.props?.options?.forEach((o: any) => { o.examScore = undefined })
      question.props?.fields?.forEach((f: any) => { f.examScore = undefined })
    }
    if (prev === 'allCorrect' || prev === 'partialCorrect') {
      question.props?.options?.forEach((o: any) => { o.examCorrect = undefined })
      question.props?.fields?.forEach((f: any) => { f.examCorrect = undefined })
    }
    if (cur === 'single' && ['multiInput','hInput'].includes(question.type)) {
      populateFieldsFromCorrectAnswer()
    }
  })

  function populateFieldsFromCorrectAnswer() {
    if (!selected.value || selected.value.examScoreMode === 'none' || !['multiInput','hInput'].includes(selected.value.type)) return
    const correctAnswer = selected.value.examCorrectAnswer
    if (!correctAnswer) return
    const fields = selected.value.props?.fields || []
    const parts = correctAnswer.split(',')
    fields.forEach((f: any, i: number) => {
      if (parts[i] !== undefined) f.examCorrectAnswer = parts[i]
    })
  }

  function deriveCorrectAnswerFromMode() {
    if (!selected.value) return
    if (selected.value.examScoreMode === 'manual' || selected.value.examScoreMode === 'none') return
    if (!selected.value.examScoreMode || selected.value.examScoreMode === 'single') {
      if (['multiInput','hInput'].includes(selected.value.type)) {
        const fields = selected.value.props?.fields || []
        const vals = fields.map((f: any) => f.examCorrectAnswer || '').filter(Boolean)
        selected.value.examCorrectAnswer = vals.length ? vals.join(',') : undefined
      } else {
        if (selected.value.examCorrectAnswer?.includes(',')) selected.value.examCorrectAnswer = undefined
      }
      return
    }
    const isInputType = ['multiInput','hInput'].includes(selected.value.type)
    const items = isInputType ? (selected.value.props?.fields || []) : (selected.value.props?.options || [])
    let correctVals: string[] = []
    if (selected.value.examScoreMode === 'allCorrect' || selected.value.examScoreMode === 'partialCorrect') {
      correctVals = items.filter((item: any) => item.examCorrect).map((item: any) => item.value || item.examCorrectAnswer || '')
    } else if (selected.value.examScoreMode === 'perOption') {
      correctVals = items.filter((item: any) => Number(item.examScore) > 0).map((item: any) => isInputType ? (item.examCorrectAnswer || '') : (item.value || ''))
    }
    selected.value.examCorrectAnswer = correctVals.length ? correctVals.join(',') : undefined
  }

  function syncFieldsCorrectAnswer() {
    if (!selected.value || selected.value.examScoreMode === 'none') return
    const fields = selected.value.props?.fields || []
    const vals = fields.map((f: any) => f.examCorrectAnswer || '').filter(Boolean)
    selected.value.examCorrectAnswer = vals.length ? vals.join(',') : undefined
  }
  watch(() => selected.value?.examScoreMode, deriveCorrectAnswerFromMode)
  watch(() => selected.value?.props?.options?.map((o: any) => o.examCorrect), deriveCorrectAnswerFromMode, { deep: true })
  watch(() => selected.value?.props?.options?.map((o: any) => o.examScore), deriveCorrectAnswerFromMode, { deep: true })
  watch(() => selected.value?.props?.fields?.map((f: any) => f.examCorrect), deriveCorrectAnswerFromMode, { deep: true })
  watch(() => selected.value?.props?.fields?.map((f: any) => f.examScore), deriveCorrectAnswerFromMode, { deep: true })
  watch(() => selected.value?.props?.fields?.map((f: any) => f.examCorrectAnswer), deriveCorrectAnswerFromMode, { deep: true })

  return { partialTotalScore, perOptionTotalScore, paperTotalScore, populateFieldsFromCorrectAnswer, syncFieldsCorrectAnswer }
}
