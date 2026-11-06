interface ExamNavigationQuestion {
  type?: unknown
  defaultHidden?: unknown
}

const LAYOUT_TYPES = new Set(['description', 'divider', 'pagination'])

function isRealQuestion(question: ExamNavigationQuestion | undefined): boolean {
  return Boolean(question && !question.defaultHidden && !LAYOUT_TYPES.has(String(question.type || '')))
}

export function examQuestionNavIndex(questions: ExamNavigationQuestion[], currentIndex: number): number {
  if (!Array.isArray(questions) || currentIndex < 0 || currentIndex >= questions.length) return -1
  const realQuestions = questions.filter(isRealQuestion)
  if (!realQuestions.length) return -1
  const currentQuestion = questions[currentIndex]
  const directIndex = realQuestions.indexOf(currentQuestion)
  if (directIndex >= 0) return directIndex
  const previousCount = questions.slice(0, currentIndex).filter(isRealQuestion).length
  return Math.min(Math.max(previousCount - 1, 0), realQuestions.length - 1)
}
