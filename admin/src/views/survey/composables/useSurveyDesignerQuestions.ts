import { computed, nextTick, ref, type Ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { SurveyDesignerQuestion, SurveyLogicRule } from '../survey-designer-types'

interface QuestionStateOptions {
  logicRules: Ref<SurveyLogicRule[]>
  panelMode: Ref<'edit' | 'json' | 'preview'>
  showSurveySettings: Ref<boolean>
}

export function useSurveyDesignerQuestions(options: QuestionStateOptions) {
  const questions = ref<SurveyDesignerQuestion[]>([])
  const selected = ref<SurveyDesignerQuestion | null>(null)
  const selectedOptIdx = ref(-1)
  let idCounter = 0

  const selectedQuestionContext = computed(() => {
    if (!selected.value) return ''
    const index = questions.value.findIndex(question => question.id === selected.value?.id)
    const title = stripMarkup(String(selected.value.title || '')).split('\n')[0].replace(/\s+/g, ' ').trim() || '未命名'
    return `${index >= 0 ? `第 ${index + 1} 项` : '当前项'} · ${title}`
  })
  const answerableQuestionCount = computed(() => questions.value.filter(question => !['divider', 'description', 'pagination', 'questionSet'].includes(question.type)).length)

  function syncIdCounterFromQuestions(list: SurveyDesignerQuestion[] = questions.value) {
    idCounter = list.reduce((max, question) => {
      const match = String(question.id || '').match(/^q(\d+)$/)
      return match ? Math.max(max, Number(match[1])) : max
    }, 0)
  }

  function genId() {
    let id = ''
    do {
      id = `q${++idCounter}`
    } while (questions.value.some(question => question.id === id))
    return id
  }

  function addQuestion(type: any) {
    const question: SurveyDesignerQuestion = {
      id: genId(), type: type.type, title: type.displayName, required: false, readOnly: false,
      dataType: '', placeholder: '', mediaType: '', mediaUrl: '', mediaWidth: '', mediaAlign: 'center',
      props: type.defaultProps ? JSON.parse(JSON.stringify(type.defaultProps)) : {},
    }
    initializeQuestion(question)
    questions.value.push(question)
    selected.value = question
  }

  function initializeQuestion(question: SurveyDesignerQuestion) {
    const type = question.type
    if (!question.props) question.props = {}
    if (type === 'judge') question.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
    if (['select', 'radio', 'checkbox', 'picker', 'cascade'].includes(type)) {
      question.props.options = ['A', 'B', 'C', 'D'].map(letter => ({ label: `选项${letter}`, value: letter }))
    }
    if (!['description', 'questionSet', 'pagination', 'divider'].includes(type)) question.showDescription = true
    if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(type)) {
      question.props.rows = [1, 2, 3].map(number => ({ title: `行${number}`, id: genId() }))
      question.props.columns = ['A', 'B'].map(letter => ({ title: `列${letter}`, id: genId(), width: 150 }))
    }
    if (type === 'matrixAuto') {
      question.props.columns = [{ label: '姓名', id: genId(), width: 150, type: 'input' }, { label: '数值', id: genId(), width: 150, type: 'number' }]
      question.props.minRows = 0; question.props.maxRows = 0
    }
    if (type === 'questionSet') question.props.children = []
    if (type === 'number') {
      question.props.min = undefined; question.props.max = undefined; question.props.decimalPlaces = 0; question.props.calculateFormula = ''
    }
    if (type === 'rating' || type === 'nps') {
      question.props.maxRating = type === 'nps' ? 10 : 5; question.props.icon = 'star'
    }
    if (type === 'multiInput' || type === 'hInput') question.props.fields = [{ label: '填空1', placeholder: '' }, { label: '填空2', placeholder: '' }]
    if (['input', 'textarea'].includes(type)) {
      question.props.options = [{ label: '', value: 'A', calculate: '', dataType: '', placeholder: '' }]
      question.props.calculateFormula = ''
    }
    if (type === 'user' || type === 'dept') {
      question.props.options = type === 'user'
        ? [{ label: '成员A', value: 'memberA', deptName: '', deptId: '', parentDeptId: '' }, { label: '成员B', value: 'memberB', deptName: '', deptId: '', parentDeptId: '' }]
        : [{ label: '部门A', value: 'deptA', parentId: '' }, { label: '部门B', value: 'deptB', parentId: '' }]
      question.multiple = false
    }
    if (['date', 'time', 'dateRange'].includes(type)) question.props.calculateFormula = ''
    if (type === 'description') question.props.replaceTextRule = ''
    if (type === 'pagination') { question.props.currentPage = 1; question.props.totalPage = 1 }
    if (type === 'signature') question.examAnswerMode = ''
    if (type === 'file') {
      question.fileTypes = ['image']; question.fileExtensions = ''; question.maxFileSize = 10; question.maxFileCount = 1
    }
  }

  function scrollQuestionIntoView(id: string) {
    nextTick(() => {
      const cards = Array.from(document.querySelectorAll<HTMLElement>('[data-survey-question-id]'))
      cards.find(element => element.dataset.surveyQuestionId === id)?.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'nearest' })
    })
  }

  function selectQuestion(id: string, shouldScroll = false) {
    selected.value = questions.value.find(question => question.id === id) || null
    selectedOptIdx.value = -1
    options.showSurveySettings.value = false
    if (shouldScroll && selected.value) {
      options.panelMode.value = 'edit'
      scrollQuestionIntoView(id)
    }
  }

  function selectOption(questionId: string, optionIndex: number) {
    selected.value = questions.value.find(question => question.id === questionId) || null
    selectedOptIdx.value = optionIndex
    options.showSurveySettings.value = false
    if (!selected.value || !['user', 'dept'].includes(selected.value.type)) return
    if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
    while (selected.value.props.options.length <= optionIndex) {
      selected.value.props.options.push(selected.value.type === 'user'
        ? { label: '', value: '', deptName: '', deptId: '', parentDeptId: '' }
        : { label: '', value: '', parentId: '' })
    }
  }

  function onQuestionsUpdate(nextQuestions: SurveyDesignerQuestion[]) {
    const selectedId = selected.value?.id
    replaceQuestions(nextQuestions)
    selected.value = selectedId ? questions.value.find(question => question.id === selectedId) || null : null
  }

  async function removeSelected() {
    if (!selected.value) return
    try {
      await ElMessageBox.confirm('确定删除此题？', '确认', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' })
    } catch { return }
    const selectedId = selected.value.id
    replaceQuestions(questions.value.filter(question => question.id !== selectedId))
    selected.value = questions.value[questions.value.length - 1] || null
    selectedOptIdx.value = -1
  }

  async function clearAll() {
    if (!questions.value.length) return
    try {
      await ElMessageBox.confirm(`确定清空全部 ${questions.value.length} 题？`, '确认', { confirmButtonText: '清空', cancelButtonText: '取消', type: 'warning' })
    } catch { return }
    questions.value = []
    options.logicRules.value = []
    selected.value = null
    selectedOptIdx.value = -1
    idCounter = 0
  }

  function replaceQuestions(nextQuestions: SurveyDesignerQuestion[]) {
    const oldQuestions = [...questions.value]
    questions.value.splice(0, questions.value.length, ...nextQuestions)
    syncLogicRulesAfterQuestionChange(oldQuestions, questions.value)
  }

  function removeQuestionById(id: string) {
    replaceQuestions(questions.value.filter(question => question.id !== id))
    if (selected.value?.id === id) {
      selected.value = null
      selectedOptIdx.value = -1
    }
  }

  function syncLogicRulesAfterQuestionChange(oldQuestions: SurveyDesignerQuestion[], nextQuestions: SurveyDesignerQuestion[]) {
    if (!options.logicRules.value.length) return
    const nextIndexById = new Map(nextQuestions.map((question, index) => [question.id, index]))
    const remappedRules = options.logicRules.value.map(rule => ({
      ...rule,
      conditions: (rule.conditions || []).map(condition => ({ ...condition, questionIdx: remapQuestionIndex(condition.questionIdx, oldQuestions, nextIndexById) })),
      targetQuestionIdx: remapQuestionIndex(rule.targetQuestionIdx, oldQuestions, nextIndexById),
      branchFromIdx: remapQuestionIndex(rule.branchFromIdx, oldQuestions, nextIndexById),
      branchToIdx: rule.branchToEnd ? undefined : remapQuestionIndex(rule.branchToIdx, oldQuestions, nextIndexById),
    }))
    const nextRules = remappedRules.filter(isLogicRuleUsable)
    const removed = remappedRules.length - nextRules.length
    options.logicRules.value = nextRules
    if (removed > 0) ElMessage.warning(`已移除 ${removed} 条关联失效题目的逻辑规则`)
  }

  function deselectQuestion() { selected.value = null; options.showSurveySettings.value = false }
  function openSurveySettings() { selected.value = null; options.showSurveySettings.value = true }

  return {
    questions, selected, selectedOptIdx, selectedQuestionContext, answerableQuestionCount,
    genId, syncIdCounterFromQuestions, addQuestion, selectQuestion, selectOption,
    onQuestionsUpdate, removeSelected, clearAll, replaceQuestions, removeQuestionById,
    deselectQuestion, openSurveySettings,
  }
}

function stripMarkup(value: string) { return value.replace(/<[^>]*>/g, '') }
function remapQuestionIndex(index: number | undefined, oldQuestions: SurveyDesignerQuestion[], nextIndexById: Map<string, number>) {
  if (index === undefined) return undefined
  const oldId = oldQuestions[index]?.id
  return oldId ? nextIndexById.get(oldId) : undefined
}
function isLogicRuleUsable(rule: SurveyLogicRule) {
  if (rule.conditionType !== 'none' && (rule.conditions || []).some(condition => condition.questionIdx === undefined)) return false
  if (['show', 'hide', 'required', 'check', 'assignment', 'validate', 'replace'].includes(rule.action) && rule.targetQuestionIdx === undefined) return false
  if (['branch', 'end'].includes(rule.action) && rule.branchFromIdx === undefined) return false
  return rule.action !== 'branch' || rule.branchToEnd || rule.branchToIdx !== undefined
}
