import { computed, nextTick, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '../../../api'
import { showRequestError } from '../../../utils/request'
import type { SurveyDesignerQuestion } from '../survey-designer-types'

interface SurveyResponseRecord {
  id: number
  nickname?: string
  submitTime?: number
  startTime?: number
  duration?: number
  ip?: string
  browser?: string
  deviceType?: string
  platformType?: string
  isAutoSubmit?: number
  answers: Record<string, unknown>
}

interface SurveyResponseColumn {
  key: string
  label: string
  width?: number
  minWidth?: number
  showOverflowTooltip?: boolean
  fixed?: 'left' | 'right'
  prop?: string
  qId?: string
  qType?: string
}

interface UseSurveyResponsesOptions {
  getSurveyId: () => number
  getQuestions: () => SurveyDesignerQuestion[]
}

export function useSurveyResponses(options: UseSurveyResponsesOptions) {
  const responseRows = ref<SurveyResponseRecord[]>([])
  const responseTotal = ref(0)
  const responsePage = ref(1)
  const responsePageSize = ref(20)
  const responseKeyword = ref('')
  const responseSource = ref<SurveyResponseRecord[]>([])
  const dataTableRef = ref<any>(null)
  const dragColKey = ref('')
  const selectedResponseIds = ref<number[]>([])
  const dataBaseOrder = ref(['nickname', 'startTime', 'submitTime', 'duration', 'ip', 'browser', 'deviceType', 'platformType', 'isAutoSubmit'])

  const dataColumnKeys = computed(() => {
    const keys = [...dataBaseOrder.value]
    options.getQuestions().forEach(question => {
      const answerKey = `answer-${question.id}`
      if (!keys.includes(answerKey)) keys.push(answerKey)
    })
    keys.push('actions')
    return keys
  })

  const dataColumns = computed<SurveyResponseColumn[]>(() => dataColumnKeys.value.flatMap(key => {
    if (key.startsWith('answer-')) {
      const qId = key.replace('answer-', '')
      const question = options.getQuestions().find(item => item.id === qId)
      return question
        ? [{ key, label: stripHtml(question.title), minWidth: 120, showOverflowTooltip: true, qId, qType: question.type }]
        : []
    }
    const definitions: Record<string, SurveyResponseColumn> = {
      nickname: { key: 'nickname', label: '提交人', width: 90 },
      startTime: { key: 'startTime', label: '开始时间', width: 150 },
      submitTime: { key: 'submitTime', label: '提交时间', width: 150 },
      duration: { key: 'duration', label: '答题时长', width: 80 },
      ip: { key: 'ip', label: 'IP地址', width: 120 },
      browser: { key: 'browser', label: '浏览器', width: 100 },
      deviceType: { key: 'deviceType', label: '设备类型', width: 90 },
      platformType: { key: 'platformType', label: '平台类型', width: 90 },
      isAutoSubmit: { key: 'isAutoSubmit', label: '自动提交', width: 80 },
      actions: { key: 'actions', label: '操作', width: 80, fixed: 'right' },
    }
    return definitions[key] ? [definitions[key]] : []
  }))

  function setupColumnDrag() {
    void nextTick(() => {
      const wrapper = dataTableRef.value?.$el?.querySelector('.el-table__header-wrapper')
      if (!wrapper) return
      const headers = wrapper.querySelectorAll('th:not(.el-table__cell--selection)')
      headers.forEach((header: HTMLElement, index: number) => {
        header.draggable = true
        header.addEventListener('dragstart', () => { dragColKey.value = dataColumnKeys.value[index] })
        header.addEventListener('dragover', event => {
          if (!dragColKey.value || dragColKey.value === dataColumnKeys.value[index]) return
          event.preventDefault()
          const order = [...dataBaseOrder.value]
          const from = order.indexOf(dragColKey.value)
          const to = dataColumnKeys.value.indexOf(dataColumnKeys.value[index])
          if (from >= 0) {
            const [key] = order.splice(from, 1)
            order.splice(Math.min(to, order.length), 0, key)
            dataBaseOrder.value = order
          }
        })
        header.addEventListener('dragend', () => { dragColKey.value = '' })
      })
    })
  }

  async function loadResponses(page = 1) {
    const surveyId = options.getSurveyId()
    if (!surveyId) return
    responsePage.value = page
    try {
      const keyword = responseKeyword.value.trim()
      const response: any = await adminApi.surveyResponseList({ surveyId, page, pageSize: responsePageSize.value, keyword })
      const list: any[] = response.data?.list || []
      responseTotal.value = response.data?.total || 0
      responseSource.value = list.map(item => {
        let answers: Record<string, unknown> = item.answers || {}
        if (typeof item.answers === 'string') {
          try { answers = JSON.parse(item.answers) } catch { answers = {} }
        }
        return {
          id: item.id,
          nickname: item.nickname,
          submitTime: item.submitTime,
          startTime: item.startTime,
          duration: item.duration,
          ip: item.ip,
          browser: item.browser || '-',
          deviceType: item.deviceType || '-',
          platformType: item.platformType || '-',
          isAutoSubmit: item.isAutoSubmit,
          answers,
        }
      })
      responseRows.value = responseSource.value
      setupColumnDrag()
    } catch {
      responseRows.value = []
    }
  }

  function onSearch() { void loadResponses(1) }
  function onPageChange(page: number) { void loadResponses(page) }
  function onPageSizeChange(size: number) {
    responsePageSize.value = size
    void loadResponses(1)
  }
  function onSelectionChange(rows: SurveyResponseRecord[]) {
    selectedResponseIds.value = rows.map(row => row.id)
  }

  async function deleteResponse(id: number) {
    try {
      await adminApi.surveyResponseDel({ id })
      responseSource.value = responseSource.value.filter(row => row.id !== id)
      responseTotal.value--
      responseRows.value = responseSource.value
    } catch (error) {
      showRequestError(error, '删除失败')
    }
  }

  function batchDeleteResponses() {
    if (!selectedResponseIds.value.length) {
      ElMessage.warning('请先选择数据')
      return
    }
    ElMessageBox.confirm(`确认删除 ${selectedResponseIds.value.length} 条数据?`, '提示', { type: 'warning' }).then(async () => {
      for (const id of selectedResponseIds.value) {
        try { await adminApi.surveyResponseDel({ id }) } catch {}
      }
      selectedResponseIds.value = []
      await loadResponses(responsePage.value)
      ElMessage.success('已删除')
    }).catch(() => {})
  }

  function refreshResponses() {
    void loadResponses(responsePage.value)
    ElMessage.success('已刷新')
  }

  return {
    responseRows,
    responseTotal,
    responsePage,
    responsePageSize,
    responseKeyword,
    dataTableRef,
    dataColumns,
    loadResponses,
    onSearch,
    onPageChange,
    onPageSizeChange,
    onSelectionChange,
    deleteResponse,
    batchDeleteResponses,
    refreshResponses,
  }
}

export function formatSurveyResponseTime(timestamp?: number) {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

function formatDate(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\//g, '-')
}

export function formatSurveyAnswer(value: unknown, questionType?: string) {
  if (value == null || value === '') return '-'
  if (questionType === 'date') return formatDate(String(value))
  if (questionType === 'time') {
    const date = new Date(String(value))
    return Number.isNaN(date.getTime())
      ? String(value)
      : date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })
  }
  if (questionType === 'dateRange') {
    return Array.isArray(value) ? value.map(item => formatDate(String(item))).join(' ~ ') : formatDate(String(value))
  }
  if (Array.isArray(value)) return value.join(', ')
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function stripHtml(html: string) {
  const container = document.createElement('div')
  container.innerHTML = html
  return container.textContent || container.innerText || ''
}
