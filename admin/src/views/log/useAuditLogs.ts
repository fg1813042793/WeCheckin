import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { adminApi } from '@/api'
import { hasPerm } from '@/utils/permission'
import type { AuditLogCategory, AuditLogQuery, AuditLogRecord } from '@/types/auditLog'
import { auditCsv, buildAuditQuery, createRequestGate, emptyAuditFilters, EXPORT_LIMIT, exportSummary } from './auditLogHelpers'

export function useAuditLogs() {
  const canReadOperations = computed(() => hasPerm('admin:menu:log:list'))
  const canReadLogin = computed(() => hasPerm('admin:menu:log:security'))
  const category = ref<AuditLogCategory>(canReadOperations.value ? 'operation' : 'login')
  const canRead = computed(() => category.value === 'login' ? canReadLogin.value : canReadOperations.value)
  const canExport = computed(() => canRead.value && hasPerm(category.value === 'login' ? 'admin:menu:log:security-export' : 'admin:menu:log:export'))
  const filters = reactive(emptyAuditFilters())
  const rows = ref<AuditLogRecord[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const dirty = ref(false)
  const error = ref('')
  const filterError = ref('')
  const exporting = ref(false)
  const exportNotice = ref('')
  const detailOpen = ref(false)
  const detail = ref<AuditLogRecord | null>(null)
  const detailLoading = ref(false)
  const detailError = ref('')
  let detailId: AuditLogRecord['id'] | null = null
  let appliedQuery: AuditLogQuery | null = null
  let suspended = false
  const listGate = createRequestGate()
  const detailGate = createRequestGate()
  const exportGate = createRequestGate()

  function closeDetail() {
    detailGate.invalidate()
    detailOpen.value = false
    detailLoading.value = false
    detail.value = null
    detailId = null
  }

  function invalidate() {
    listGate.invalidate()
    exportGate.invalidate()
    closeDetail()
    loading.value = false
    exporting.value = false
    exportNotice.value = ''
    error.value = ''
    filterError.value = ''
    rows.value = []
    total.value = 0
    appliedQuery = null
  }

  async function load() {
    exportGate.invalidate()
    closeDetail()
    exporting.value = false
    exportNotice.value = ''
    error.value = ''
    filterError.value = ''
    appliedQuery = null
    if (!canRead.value) {
      rows.value = []
      total.value = 0
      return
    }
    let query: AuditLogQuery
    try { query = buildAuditQuery(filters, page.value, pageSize.value) }
    catch (cause) {
      filterError.value = cause instanceof Error ? cause.message : '请检查筛选条件'
      return
    }
    const ticket = listGate.next()
    const target = category.value
    loading.value = true
    dirty.value = false
    try {
      const response = await (target === 'login' ? adminApi.loginLogList(query) : adminApi.logList(query))
      if (!listGate.isCurrent(ticket)) return
      const data = response.data
      if (!Array.isArray(data?.list) || !Number.isSafeInteger(data.total) || data.total < data.list.length) throw new Error('Invalid log list')
      rows.value = data.list
      total.value = data.total
      appliedQuery = query
    } catch {
      if (listGate.isCurrent(ticket)) {
        rows.value = []
        total.value = 0
        error.value = '日志加载失败，请点击重试。'
      }
    } finally {
      if (listGate.isCurrent(ticket)) loading.value = false
    }
  }

  function search() { page.value = 1; return load() }
  function reset() { Object.assign(filters, emptyAuditFilters()); return search() }
  function changePage(value: number) { if (page.value !== value) { page.value = value; void load() } }
  function changeSize(value: number) { if (pageSize.value !== value) { pageSize.value = value; void search() } }

  async function showDetail(row: Pick<AuditLogRecord, 'id'>) {
    if (!canRead.value) return
    const ticket = detailGate.next()
    const target = category.value
    detailId = row.id
    detail.value = null
    detailError.value = ''
    detailLoading.value = true
    detailOpen.value = true
    try {
      const response = await (target === 'login' ? adminApi.loginLogDetail(row.id) : adminApi.logDetail(row.id))
      if (!detailGate.isCurrent(ticket)) return
      if (!response.data || response.data.id == null) throw new Error('Invalid log detail')
      detail.value = response.data
    } catch {
      if (detailGate.isCurrent(ticket)) detailError.value = '日志详情加载失败，请重试。'
    } finally {
      if (detailGate.isCurrent(ticket)) detailLoading.value = false
    }
  }

  function retryDetail() { if (detailId !== null) return showDetail({ id: detailId }) }

  async function exportData() {
    if (!canExport.value || exporting.value || loading.value || dirty.value || !appliedQuery) return
    const ticket = exportGate.next()
    const target = category.value
    const query = { ...appliedQuery, page: 1, pageSize: EXPORT_LIMIT }
    exporting.value = true
    exportNotice.value = ''
    try {
      const response = await (target === 'login' ? adminApi.loginLogExport(query) : adminApi.logExport(query))
      if (!exportGate.isCurrent(ticket) || !canExport.value) return
      const summary = exportSummary(response.data)
      if (!summary.count) { exportNotice.value = '当前筛选条件下没有可导出的日志。'; return }
      if (summary.truncated) {
        try {
          await ElMessageBox.confirm(`共匹配 ${summary.total} 条日志，本次仅返回前 ${summary.count} 条（最多 ${EXPORT_LIMIT} 条），其余 ${summary.total - summary.count} 条不会导出。可取消后缩小时间或其他筛选范围。`, '导出数量受限', { confirmButtonText: `导出这 ${summary.count} 条`, cancelButtonText: '取消并调整筛选', type: 'warning' })
        } catch {
          if (exportGate.isCurrent(ticket)) exportNotice.value = '已取消导出，请缩小筛选范围后重试。'
          return
        }
      }
      if (!exportGate.isCurrent(ticket) || !canExport.value) return
      const csv = auditCsv(response.data.list)
      const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' }))
      const link = document.createElement('a')
      link.href = url
      link.download = `${target === 'login' ? '登录' : '操作'}日志-${new Date().toISOString().slice(0, 10)}.csv`
      document.body.append(link)
      link.click()
      link.remove()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
      exportNotice.value = `已导出 ${summary.count} / ${summary.total} 条${summary.truncated ? '，结果已截断，请缩小筛选范围获取其余记录' : ''}。`
    } catch {
      if (exportGate.isCurrent(ticket)) exportNotice.value = '导出失败，请重试或缩小筛选范围。'
    } finally {
      if (exportGate.isCurrent(ticket)) exporting.value = false
    }
  }

  watch(filters, () => { invalidate(); dirty.value = true; page.value = 1 }, { deep: true, flush: 'sync' })
  watch(category, () => { invalidate(); page.value = 1; void load() }, { flush: 'sync' })
  watch(detailOpen, open => { if (!open) closeDetail() }, { flush: 'sync' })
  onMounted(load)
  onDeactivated(() => { suspended = true; invalidate() })
  onActivated(() => { if (suspended) { suspended = false; void load() } })
  onBeforeUnmount(invalidate)

  return { category, canReadOperations, canReadLogin, canRead, canExport, filters, rows, total, page, pageSize,
    loading, dirty, error, filterError, exporting, exportNotice, detailOpen, detail, detailLoading, detailError,
    load, search, reset, changePage, changeSize, showDetail, retryDetail, exportData }
}
