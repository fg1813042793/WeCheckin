import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch, type Ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { onBeforeRouteLeave, type RouteLocationNormalizedLoaded, type Router } from 'vue-router'
import { adminApi } from '../../../api'
import { showRequestError } from '../../../utils/request'
import { normalizeQuestions } from '../../../utils/surveyKingBridge'
import type { SurveyCompletionAction, SurveyDesignerQuestion, SurveyLogicRule } from '../survey-designer-types'
import type { SurveyDesignerIssue } from '../survey-designer-validation'
import { useSurveyDesignerHistory } from './useSurveyDesignerHistory'

const DEFAULT_RESULT_CONFIG = {
  statType: 'value',
  viewMode: 'aggregate',
  scope: ['all'],
  showChart: true,
  showDetail: true,
  exportField: 'value',
}

export function createSurveyDesignerForm() {
  return reactive<any>({
    id: 0, title: '', description: '', category: '', tags: '', mode: 'survey',
    visibility: 0, allowMultiBool: 0, anonymousBool: 0, showResultBool: 0,
    startDate: null, endDate: null, maxResponse: 0, statusBool: 1, deptIds: '',
    questionNumber: true, progressBar: false, autoSave: false,
    onePageOneQuestion: false, answerSheetVisible: true, copyEnabled: true,
    password: '', triggerType: 'onBlur', loginRequired: false,
    redirectUrl: '', endContent: '',
    examRankingEnabled: false, exerciseMode: false, randomOrder: false,
    minSubmitMinutes: 0, maxSubmitMinutes: 0,
    backgroundImages: [], headerImages: [],
    defaultAnswer: false, defaultLang: 'zh-CN',
    deviceLimit: 0, ipLimit: 0, userLimit: 0,
    publicQuery: false, showAnswerAnalysis: false,
    createBy: 0, collaborators: '', timeLimit: 0,
    fillTemplate: 'sf1',
    resultConfig: { ...DEFAULT_RESULT_CONFIG },
  })
}

interface SurveyDesignerDocumentOptions {
  form: Record<string, any>
  questions: Ref<SurveyDesignerQuestion[]>
  selected: Ref<SurveyDesignerQuestion | null>
  logicRules: Ref<SurveyLogicRule[]>
  completionAction: Ref<SurveyCompletionAction>
  issues: Ref<SurveyDesignerIssue[]>
  route: RouteLocationNormalizedLoaded
  router: Router
  locateIssue: (issue: SurveyDesignerIssue) => void
  syncQuestionIds: () => void
  reloadResources: () => void
}

export function useSurveyDesignerDocument(options: SurveyDesignerDocumentOptions) {
  const saving = ref(false)
  const designerLoading = ref(true)
  const designerLoadError = ref(false)
  const saveFailed = ref(false)
  const savedSnapshot = ref('')
  let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
  let dataLoaded = false
  let savePending = false

  const hasInvalidTimeRange = computed(() => {
    const start = Number(options.form.startDate || 0)
    const end = Number(options.form.endDate || 0)
    return start > 0 && end > 0 && end < start
  })
  const designerSnapshot = computed(() => JSON.stringify(buildSurveyPayload(
    JSON.stringify({ version: '2.0', questions: options.questions.value }),
    JSON.stringify(buildSettingsPayload()),
  )))
  const historySnapshot = computed(() => JSON.stringify({
    questions: options.questions.value,
    logicRules: options.logicRules.value,
  }))
  const hasUnsavedChanges = computed(() => dataLoaded && designerSnapshot.value !== savedSnapshot.value)
  const saveStatusLabel = computed(() => saving.value
    ? '保存中'
    : saveFailed.value
      ? '保存失败'
      : hasUnsavedChanges.value
        ? '未保存'
        : options.form.id ? '已保存' : '尚未保存')
  const saveStatusClass = computed(() => saveFailed.value
    ? 'is-error'
    : hasUnsavedChanges.value ? 'is-dirty' : saving.value ? 'is-saving' : 'is-saved')
  const saveStatusTagType = computed<'success' | 'warning' | 'danger' | 'info'>(() => saveFailed.value
    ? 'danger'
    : hasUnsavedChanges.value ? 'warning' : options.form.id ? 'success' : 'info')

  const designerHistory = useSurveyDesignerHistory({
    capture: () => historySnapshot.value,
    restore: snapshot => {
      const state = JSON.parse(snapshot) as { questions: SurveyDesignerQuestion[]; logicRules: SurveyLogicRule[] }
      const selectedId = options.selected.value?.id
      options.questions.value.splice(0, options.questions.value.length, ...normalizeQuestions(state.questions))
      options.logicRules.value = state.logicRules
      options.syncQuestionIds()
      options.selected.value = options.questions.value.find(question => question.id === selectedId) || null
    },
  })
  const { canUndo, canRedo } = designerHistory

  function syncCompletionActionFromForm() {
    options.completionAction.value = String(options.form.redirectUrl || '').trim()
      ? 'redirect'
      : String(options.form.endContent || '').trim() ? 'content' : 'default'
  }

  watch(options.completionAction, (value, oldValue) => {
    if (value === 'default') {
      options.form.redirectUrl = ''
      options.form.endContent = ''
    } else if (value === 'content' && oldValue !== 'content') {
      options.form.redirectUrl = ''
    } else if (value === 'redirect' && oldValue !== 'redirect') {
      options.form.endContent = ''
    }
  })
  watch(historySnapshot, snapshot => designerHistory.record(snapshot))
  watch(designerSnapshot, scheduleAutoSave)

  function buildSettingsPayload() {
    const form = options.form
    return {
      questionNumber: form.questionNumber, progressBar: form.progressBar, autoSave: form.autoSave,
      password: form.password, loginRequired: form.loginRequired,
      onePageOneQuestion: form.onePageOneQuestion, answerSheetVisible: form.answerSheetVisible,
      copyEnabled: form.copyEnabled, triggerType: form.triggerType,
      redirectUrl: form.redirectUrl, endContent: form.endContent,
      examRankingEnabled: form.examRankingEnabled, exerciseMode: form.exerciseMode,
      randomOrder: form.randomOrder, minSubmitMinutes: form.minSubmitMinutes,
      maxSubmitMinutes: form.maxSubmitMinutes, backgroundImages: form.backgroundImages,
      headerImages: form.headerImages, defaultAnswer: form.defaultAnswer,
      defaultLang: form.defaultLang, deviceLimit: form.deviceLimit, ipLimit: form.ipLimit,
      userLimit: form.userLimit, publicQuery: form.publicQuery,
      showAnswerAnalysis: form.showAnswerAnalysis, collaborators: form.collaborators,
      timeLimit: form.timeLimit, fillTemplate: form.fillTemplate,
      resultConfig: form.resultConfig, logicRules: JSON.stringify(options.logicRules.value),
    }
  }

  function buildSurveyPayload(schema: string, settings: string) {
    const form = options.form
    return {
      title: form.title, description: form.description, category: form.category, tags: form.tags,
      visibility: form.visibility, allowMulti: form.allowMultiBool, anonymous: form.anonymousBool,
      showResult: form.showResultBool, startTime: form.startDate || 0, endTime: form.endDate || 0,
      maxResponse: form.maxResponse, status: form.statusBool, schema, deptIds: form.deptIds,
      mode: form.mode, settings,
    }
  }

  function normalizeCompletionSettings() {
    options.form.title = String(options.form.title || '').trim()
    if (options.completionAction.value === 'default') {
      options.form.redirectUrl = ''
      options.form.endContent = ''
    } else if (options.completionAction.value === 'content') {
      options.form.redirectUrl = ''
    } else {
      options.form.redirectUrl = String(options.form.redirectUrl || '').trim()
      options.form.endContent = ''
    }
  }

  function validateSurveySettings(showMessage = true) {
    const blockingCodes = new Set([
      'survey_title_required', 'survey_department_required',
      'survey_time_range_invalid', 'survey_redirect_url_invalid',
    ])
    const issue = options.issues.value.find(item => blockingCodes.has(item.code))
    if (issue && showMessage) {
      ElMessage.warning(issue.message)
      options.locateIssue(issue)
    }
    return !issue
  }

  function scheduleAutoSave() {
    if (!options.form.id || !dataLoaded) return
    saveFailed.value = false
    clearAutoSaveTimer()
    autoSaveTimer = setTimeout(() => { void save(false) }, 800)
  }

  async function saveLogicRules(showMessage = true) {
    if (!options.form.id) {
      if (showMessage) ElMessage.warning('请先保存问卷')
      return
    }
    if (!validateSurveySettings(showMessage)) return
    normalizeCompletionSettings()
    saving.value = true
    try {
      await adminApi.surveyEdit({
        id: options.form.id,
        ...buildSurveyPayload(
          JSON.stringify({ version: '2.0', questions: options.questions.value }),
          JSON.stringify(buildSettingsPayload()),
        ),
      })
      if (showMessage) ElMessage.success('已保存')
    } catch (error) {
      if (showMessage) showRequestError(error, '保存失败')
    } finally {
      saving.value = false
    }
  }

  async function load() {
    const id = Number(options.route.query.id || 0)
    designerLoading.value = true
    designerLoadError.value = false
    if (!id) {
      resetSnapshots()
      designerLoading.value = false
      return
    }
    try {
      const response: any = await adminApi.surveyDetail(id)
      const survey = response.data.survey
      const base: Record<string, any> = {
        id: survey.id, title: survey.title, description: survey.description,
        category: survey.category, tags: survey.tags, visibility: survey.visibility,
        allowMultiBool: survey.allowMulti, anonymousBool: survey.anonymous,
        showResultBool: survey.showResult, startDate: survey.startTime || null,
        endDate: survey.endTime || null, maxResponse: survey.maxResponse,
        statusBool: survey.status, deptIds: survey.deptIds, mode: survey.mode || 'survey',
        createBy: survey.createBy || 0,
      }
      Object.assign(base, parseSettings(survey.settings))
      Object.assign(options.form, base)
      syncCompletionActionFromForm()
      options.reloadResources()
      const rawSchema = response.data?.schema || ''
      if (rawSchema) {
        try {
          const schema = JSON.parse(rawSchema)
          if (schema.questions) {
            options.questions.value = normalizeQuestions(schema.questions)
            options.syncQuestionIds()
          }
        } catch {}
      }
      await nextTick()
      resetSnapshots()
      saveFailed.value = false
    } catch (error) {
      designerLoadError.value = true
      showRequestError(error, '加载失败')
    } finally {
      designerLoading.value = false
    }
  }

  function retryLoad() {
    void load()
  }

  async function save(showMessage: boolean | Event = true) {
    const shouldShowMessage = typeof showMessage === 'boolean' ? showMessage : true
    if (saving.value) {
      savePending = true
      return false
    }
    if (!validateSurveySettings(shouldShowMessage)) return false
    normalizeCompletionSettings()
    clearAutoSaveTimer()
    const snapshotToSave = designerSnapshot.value
    saveFailed.value = false
    saving.value = true
    try {
      const payload: any = buildSurveyPayload(
        JSON.stringify({ version: '2.0', questions: options.questions.value }),
        JSON.stringify(buildSettingsPayload()),
      )
      if (options.form.id) {
        payload.id = options.form.id
        await adminApi.surveyEdit(payload)
        if (shouldShowMessage) ElMessage.success('已更新')
      } else {
        const response: any = await adminApi.surveyInsert(payload)
        options.form.id = response.data?.id || response.id
        if (shouldShowMessage) ElMessage.success('已创建')
        void options.router.replace({ query: { id: String(options.form.id) } })
      }
      savedSnapshot.value = snapshotToSave
      return true
    } catch (error) {
      saveFailed.value = true
      if (shouldShowMessage) showRequestError(error, '保存失败')
      return false
    } finally {
      const shouldRetry = savePending
      savePending = false
      saving.value = false
      if (shouldRetry && options.form.id && hasUnsavedChanges.value) {
        queueMicrotask(() => { void save(false) })
      }
    }
  }

  function undoDesignerChange() {
    designerHistory.undo()
  }

  function redoDesignerChange() {
    designerHistory.redo()
  }

  function showShortcutHelp() {
    void ElMessageBox.alert(
      'Ctrl/Cmd + S：保存问卷\nCtrl/Cmd + Z：撤销\nCtrl/Cmd + Shift + Z 或 Ctrl/Cmd + Y：重做',
      '快捷键',
      { confirmButtonText: '知道了' },
    )
  }

  function handleDesignerKeydown(event: KeyboardEvent) {
    if (!event.metaKey && !event.ctrlKey) return
    const key = event.key.toLowerCase()
    if (key === 's') {
      event.preventDefault()
      void save()
      return
    }
    if (isTextEditingTarget(event.target)) return
    if (key === 'z') {
      event.preventDefault()
      event.shiftKey ? redoDesignerChange() : undoDesignerChange()
    } else if (key === 'y') {
      event.preventDefault()
      redoDesignerChange()
    }
  }

  function handleBeforeUnload(event: BeforeUnloadEvent) {
    if (!hasUnsavedChanges.value) return
    event.preventDefault()
    event.returnValue = ''
  }

  onBeforeRouteLeave(async () => {
    if (!hasUnsavedChanges.value) return true
    try {
      await ElMessageBox.confirm(
        '当前问卷还有未保存的修改，离开后将丢失这些内容。',
        '确认离开',
        { confirmButtonText: '离开', cancelButtonText: '继续编辑', type: 'warning' },
      )
      return true
    } catch {
      return false
    }
  })

  onMounted(() => {
    window.addEventListener('keydown', handleDesignerKeydown)
    window.addEventListener('beforeunload', handleBeforeUnload)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleDesignerKeydown)
    window.removeEventListener('beforeunload', handleBeforeUnload)
    clearAutoSaveTimer()
    designerHistory.dispose()
  })

  return {
    saving,
    designerLoading,
    designerLoadError,
    hasInvalidTimeRange,
    saveStatusLabel,
    saveStatusClass,
    saveStatusTagType,
    canUndo,
    canRedo,
    load,
    retryLoad,
    save,
    saveLogicRules,
    undoDesignerChange,
    redoDesignerChange,
    showShortcutHelp,
  }

  function resetSnapshots() {
    dataLoaded = true
    savedSnapshot.value = designerSnapshot.value
    designerHistory.reset(historySnapshot.value)
  }

  function clearAutoSaveTimer() {
    if (autoSaveTimer) clearTimeout(autoSaveTimer)
    autoSaveTimer = null
  }

  function parseSettings(rawSettings: string) {
    if (!rawSettings) return defaultSettings()
    try {
      const settings = JSON.parse(rawSettings)
      try {
        options.logicRules.value = JSON.parse(settings.logicRules || '[]')
      } catch {
        options.logicRules.value = []
      }
      return {
        questionNumber: settings.questionNumber ?? true,
        progressBar: settings.progressBar ?? false,
        autoSave: settings.autoSave ?? false,
        password: settings.password || '',
        loginRequired: settings.loginRequired ?? false,
        onePageOneQuestion: settings.onePageOneQuestion ?? false,
        answerSheetVisible: settings.answerSheetVisible ?? true,
        copyEnabled: settings.copyEnabled ?? true,
        triggerType: settings.triggerType || 'onBlur',
        transcriptVisible: settings.transcriptVisible ?? true,
        rankVisible: settings.rankVisible ?? false,
        redirectUrl: settings.redirectUrl || '',
        endContent: settings.endContent || '',
        examRankingEnabled: settings.examRankingEnabled ?? false,
        exerciseMode: settings.exerciseMode ?? false,
        randomOrder: settings.randomOrder ?? false,
        minSubmitMinutes: settings.minSubmitMinutes || 0,
        maxSubmitMinutes: settings.maxSubmitMinutes || 0,
        backgroundImages: settings.backgroundImages || [],
        headerImages: settings.headerImages || [],
        defaultAnswer: settings.defaultAnswer ?? false,
        defaultLang: settings.defaultLang || 'zh-CN',
        deviceLimit: settings.deviceLimit || 0,
        ipLimit: settings.ipLimit || 0,
        userLimit: settings.userLimit || 0,
        publicQuery: settings.publicQuery ?? false,
        showAnswerAnalysis: settings.showAnswerAnalysis ?? false,
        collaborators: settings.collaborators || '',
        timeLimit: settings.timeLimit || 0,
        fillTemplate: settings.fillTemplate || 'sf1',
        resultConfig: { ...DEFAULT_RESULT_CONFIG, ...(settings.resultConfig || {}) },
      }
    } catch {
      options.logicRules.value = []
      return defaultSettings()
    }
  }
}

function defaultSettings() {
  return {
    questionNumber: true, progressBar: false, autoSave: false, password: '',
    loginRequired: false, onePageOneQuestion: false, answerSheetVisible: true,
    copyEnabled: true, triggerType: 'onBlur', transcriptVisible: true,
    rankVisible: false, redirectUrl: '', endContent: '', examRankingEnabled: false,
    exerciseMode: false, randomOrder: false, minSubmitMinutes: 0, maxSubmitMinutes: 0,
    backgroundImages: [], headerImages: [], defaultAnswer: false, defaultLang: 'zh-CN',
    deviceLimit: 0, ipLimit: 0, userLimit: 0, publicQuery: false,
    showAnswerAnalysis: false, collaborators: '', timeLimit: 0, fillTemplate: 'sf1',
    resultConfig: { ...DEFAULT_RESULT_CONFIG },
  }
}

function isTextEditingTarget(target: EventTarget | null) {
  return target instanceof HTMLElement && Boolean(target.closest('input, textarea, [contenteditable="true"]'))
}
