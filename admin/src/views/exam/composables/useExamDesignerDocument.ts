import { computed, onBeforeUnmount, reactive, ref, watch, type Ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router'
import { adminApi } from '../../../api'
import { showRequestError } from '../../../utils/request'
import type { ExamCompletionAction, ExamDesignerQuestion } from '../exam-designer-types'

export function createExamDesignerForm() {
  return reactive<any>({
    id: 0, title: '', description: '', category: '', tags: '',
    visibility: 0, allowMultiBool: 0, anonymousBool: 0, showResultBool: 0,
    startDate: null, endDate: null, maxResponse: 0, statusBool: 1, deptIds: '',
    questionNumber: true, progressBar: false, autoSave: false,
    onePageOneQuestion: false, answerSheetVisible: true, answerVisible: true,
    copyEnabled: true, language: 'zh', password: '', triggerType: 'onBlur',
    loginRequired: false, transcriptVisible: true, rankVisible: false, showAnalysis: true,
    redirectUrl: '', endContent: '', examRankingEnabled: false, exerciseMode: false,
    randomOrder: false, minSubmitMinutes: 0, maxSubmitMinutes: 0,
    duration: 60, maxAttempts: 1, showScore: 1,
    deviceLimit: 0, ipLimit: 0, userLimit: 0,
    createBy: 0, collaborators: '', backgroundImages: [], headerImages: [],
    fillTemplate: 'ef',
  })
}

interface ExamDesignerDocumentOptions {
  form: Record<string, any>
  questions: Ref<ExamDesignerQuestion[]>
  selected: Ref<ExamDesignerQuestion | null>
  route: RouteLocationNormalizedLoaded
  router: Router
  onQuestionsLoaded: () => void
  reloadResources: () => void | Promise<void>
  reloadSettings: () => void | Promise<void>
}

export function useExamDesignerDocument(options: ExamDesignerDocumentOptions) {
  const saving = ref(false)
  const completionAction = ref<ExamCompletionAction>('default')
  let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
  let designerLoaded = false

  const designerAutoSaveSnapshot = computed(() => JSON.stringify({
    title: options.form.title, description: options.form.description,
    category: options.form.category, tags: options.form.tags,
    visibility: options.form.visibility, allowMultiBool: options.form.allowMultiBool,
    anonymousBool: options.form.anonymousBool, showResultBool: options.form.showResultBool,
    startDate: options.form.startDate, endDate: options.form.endDate,
    maxResponse: options.form.maxResponse, statusBool: options.form.statusBool,
    deptIds: options.form.deptIds, duration: options.form.duration,
    maxAttempts: options.form.maxAttempts, showScore: options.form.showScore,
    collaborators: options.form.collaborators, questionNumber: options.form.questionNumber,
    progressBar: options.form.progressBar, autoSave: options.form.autoSave,
    password: options.form.password, loginRequired: options.form.loginRequired,
    onePageOneQuestion: options.form.onePageOneQuestion,
    answerSheetVisible: options.form.answerSheetVisible, answerVisible: options.form.answerVisible,
    copyEnabled: options.form.copyEnabled, language: options.form.language,
    triggerType: options.form.triggerType, transcriptVisible: options.form.transcriptVisible,
    showAnalysis: options.form.showAnalysis, rankVisible: options.form.rankVisible,
    redirectUrl: options.form.redirectUrl, endContent: options.form.endContent,
    examRankingEnabled: options.form.examRankingEnabled, exerciseMode: options.form.exerciseMode,
    randomOrder: options.form.randomOrder, minSubmitMinutes: options.form.minSubmitMinutes,
    maxSubmitMinutes: options.form.maxSubmitMinutes, deviceLimit: options.form.deviceLimit,
    ipLimit: options.form.ipLimit, userLimit: options.form.userLimit,
    backgroundImages: options.form.backgroundImages, headerImages: options.form.headerImages,
    fillTemplate: options.form.fillTemplate,
  }))

  watch(completionAction, (value, oldValue) => {
    if (value === 'default') {
      options.form.redirectUrl = ''; options.form.endContent = ''
    } else if (value === 'content' && oldValue !== 'content') {
      options.form.redirectUrl = ''
    } else if (value === 'redirect' && oldValue !== 'redirect') {
      options.form.endContent = ''
    }
  })
  watch(designerAutoSaveSnapshot, scheduleAutoSave)

  function syncCompletionActionFromForm() {
    completionAction.value = String(options.form.redirectUrl || '').trim()
      ? 'redirect'
      : String(options.form.endContent || '').trim() ? 'content' : 'default'
  }

  function normalizeCompletionSettings() {
    if (completionAction.value === 'default') {
      options.form.redirectUrl = ''; options.form.endContent = ''
    } else if (completionAction.value === 'content') {
      options.form.redirectUrl = ''
    } else {
      options.form.redirectUrl = String(options.form.redirectUrl || '').trim()
      options.form.endContent = ''
    }
  }

  function isValidRedirectUrl(raw: string) {
    if (!raw || raw.startsWith('/')) return true
    try { const url = new URL(raw); return ['http:', 'https:'].includes(url.protocol) } catch { return false }
  }

  function buildSettings() {
    const form = options.form
    return {
      collaborators: form.collaborators, questionNumber: form.questionNumber,
      progressBar: form.progressBar, autoSave: form.autoSave, password: form.password,
      loginRequired: form.loginRequired, onePageOneQuestion: form.onePageOneQuestion,
      answerSheetVisible: form.answerSheetVisible, answerVisible: form.answerVisible,
      copyEnabled: form.copyEnabled, language: form.language, triggerType: form.triggerType,
      transcriptVisible: form.transcriptVisible, showAnalysis: form.showAnalysis,
      rankVisible: form.rankVisible, redirectUrl: form.redirectUrl, endContent: form.endContent,
      examRankingEnabled: form.examRankingEnabled, exerciseMode: form.exerciseMode,
      randomOrder: form.randomOrder, minSubmitMinutes: form.minSubmitMinutes,
      maxSubmitMinutes: form.maxSubmitMinutes, deviceLimit: form.deviceLimit,
      ipLimit: form.ipLimit, userLimit: form.userLimit,
      backgroundImages: form.backgroundImages, headerImages: form.headerImages,
      fillTemplate: form.fillTemplate,
    }
  }

  function buildPayload() {
    const form = options.form
    return {
      title: form.title, description: form.description, category: form.category, tags: form.tags,
      visibility: form.visibility, allowMulti: form.allowMultiBool,
      anonymous: form.anonymousBool, showResult: form.showResultBool,
      startTime: form.startDate || 0, endTime: form.endDate || 0,
      maxResponse: form.maxResponse, status: form.statusBool, mode: 'exam',
      deptIds: form.deptIds,
      schema: JSON.stringify({ version: '2.0', questions: options.questions.value, setting: {} }),
      settings: JSON.stringify(buildSettings()),
      duration: form.duration, maxAttempts: form.maxAttempts, showScore: form.showScore,
    }
  }

  function scheduleAutoSave() {
    if (!options.form.id || !designerLoaded) return
    clearAutoSaveTimer()
    autoSaveTimer = setTimeout(() => { void save(false) }, 500)
  }

  async function load() {
    const id = Number(options.route.query.id || 0)
    if (!id) { designerLoaded = true; return }
    try {
      const response: any = await adminApi.examDetail(id)
      const survey = response.data?.survey
      if (!survey) { ElMessage.error('加载失败'); return }
      Object.assign(options.form, {
        id: survey.id, title: survey.title, description: survey.description || '',
        category: survey.category || '', tags: survey.tags || '', visibility: survey.visibility ?? 0,
        allowMultiBool: survey.allowMulti ?? 0, anonymousBool: survey.anonymous ?? 0,
        showResultBool: survey.showResult ?? 0, startDate: survey.startTime || null,
        endDate: survey.endTime || null, maxResponse: survey.maxResponse ?? 0,
        statusBool: survey.status ?? 1, deptIds: survey.deptIds || '',
        duration: survey.duration ?? 60, maxAttempts: survey.maxAttempts ?? 1,
        showScore: survey.showScore ?? 1, createBy: survey.createBy || 0,
      })
      if (survey.settings) applySettings(survey.settings)
      syncCompletionActionFromForm()
      const rawSchema = response.data?.schema || ''
      if (rawSchema) {
        try {
          const schema = JSON.parse(rawSchema)
          if (schema.questions) {
            options.questions.value = schema.questions
            options.selected.value = options.questions.value[0] || null
            options.onQuestionsLoaded()
          }
        } catch {}
      }
      options.reloadResources()
      await options.reloadSettings()
    } catch (error) {
      showRequestError(error, '加载失败')
    } finally {
      designerLoaded = true
    }
  }

  async function save(showMessage: boolean | Event = true, successMessage = '已保存') {
    const shouldShowMessage = typeof showMessage === 'boolean' ? showMessage : true
    if (!options.form.title) { if (shouldShowMessage) ElMessage.warning('请填写标题'); return false }
    normalizeCompletionSettings()
    if (!isValidRedirectUrl(options.form.redirectUrl)) { if (shouldShowMessage) ElMessage.warning('请输入有效的跳转链接'); return false }
    clearAutoSaveTimer()
    saving.value = true
    try {
      const payload: any = buildPayload()
      if (options.form.id) payload.id = options.form.id
      const response: any = await adminApi.examSave(payload)
      if (!options.form.id) {
        options.form.id = response.id || response.data?.id
        void options.router.replace({ query: { id: String(options.form.id) } })
      }
      if (shouldShowMessage) ElMessage.success(successMessage)
      return true
    } catch (error) {
      if (shouldShowMessage) showRequestError(error, '保存失败')
      return false
    } finally {
      saving.value = false
    }
  }

  async function saveLogicRules() {
    if (!options.form.id) { ElMessage.warning('请先保存考试'); return }
    await save(true, '逻辑规则已保存')
  }

  function applySettings(raw: string) {
    try {
      const settings = JSON.parse(raw)
      Object.assign(options.form, {
        questionNumber: settings.questionNumber ?? true, progressBar: settings.progressBar ?? false,
        autoSave: settings.autoSave ?? false, password: settings.password || '',
        loginRequired: settings.loginRequired ?? false, onePageOneQuestion: settings.onePageOneQuestion ?? false,
        answerSheetVisible: settings.answerSheetVisible ?? true, answerVisible: settings.answerVisible ?? true,
        copyEnabled: settings.copyEnabled ?? true, language: settings.language || 'zh',
        triggerType: settings.triggerType || 'onBlur', transcriptVisible: settings.transcriptVisible ?? true,
        showAnalysis: settings.showAnalysis ?? true, rankVisible: settings.rankVisible ?? false,
        redirectUrl: settings.redirectUrl || '', endContent: settings.endContent || '',
        examRankingEnabled: settings.examRankingEnabled ?? false, exerciseMode: settings.exerciseMode ?? false,
        randomOrder: settings.randomOrder ?? false, minSubmitMinutes: settings.minSubmitMinutes || 0,
        maxSubmitMinutes: settings.maxSubmitMinutes || 0, deviceLimit: settings.deviceLimit || 0,
        ipLimit: settings.ipLimit || 0, userLimit: settings.userLimit || 0,
        backgroundImages: settings.backgroundImages || [], headerImages: settings.headerImages || [],
        collaborators: settings.collaborators || '', fillTemplate: settings.fillTemplate || 'ef',
      })
    } catch {}
  }

  function markLoaded() { designerLoaded = true }
  function clearAutoSaveTimer() { if (autoSaveTimer) clearTimeout(autoSaveTimer); autoSaveTimer = null }
  onBeforeUnmount(clearAutoSaveTimer)

  return { saving, completionAction, load, save, saveLogicRules, markLoaded }
}
