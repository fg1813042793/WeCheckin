import { ref } from 'vue'
import { notificationTemplatesApi } from '@/api/notificationTemplates'
import type { TemplateWorkflowOption } from '@/types/notificationTemplate'
import { templateError } from './templateHelpers'

export function useTemplateWorkflowOptions() {
  const options = ref<TemplateWorkflowOption[]>([])
  const selected = ref<TemplateWorkflowOption | null>(null)
  const loading = ref(false)
  const resolving = ref(false)
  const error = ref('')
  const selectedError = ref('')
  const page = ref(1)
  const total = ref(0)
  const keyword = ref('')
  const pageSize = 20
  let searchEpoch = 0
  let selectionEpoch = 0
  async function search(value = keyword.value, nextPage = 1) {
    const epoch = ++searchEpoch
    keyword.value = value
    loading.value = true
    error.value = ''
    try {
      const { data } = await notificationTemplatesApi.workflowOptions({ keyword: value.trim(), page: nextPage, pageSize })
      if (epoch !== searchEpoch) return
      options.value = data.list
      total.value = data.total
      page.value = data.page
    } catch (cause) {
      if (epoch === searchEpoch) error.value = templateError(cause, '已发布流程加载失败，请重试。')
    } finally { if (epoch === searchEpoch) loading.value = false }
  }
  async function resolveSelected(id?: number) {
    const epoch = ++selectionEpoch
    selected.value = null
    selectedError.value = ''
    resolving.value = Boolean(id)
    if (!id) return
    try {
      const { data } = await notificationTemplatesApi.workflowOptions({ id, page: 1, pageSize })
      if (epoch !== selectionEpoch) return
      selected.value = data.list.find(item => item.id === id) || null
      if (!selected.value) selectedError.value = '所选流程已不存在或未发布，请重新选择。'
    } catch (cause) {
      if (epoch === selectionEpoch) selectedError.value = templateError(cause, '所选流程校验失败，请重试。')
    } finally { if (epoch === selectionEpoch) resolving.value = false }
  }
  function dispose() { searchEpoch++; selectionEpoch++ }
  return { options, selected, loading, resolving, error, selectedError, page, total, keyword, pageSize, search, resolveSelected, dispose }
}
