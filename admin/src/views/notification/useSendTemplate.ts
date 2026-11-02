import { computed, ref } from 'vue'
import { notificationTemplatesApi } from '@/api/notificationTemplates'
import type { DingTalkNotificationSendPayload } from '@/api/types'
import type { DingTalkSendTemplate } from '@/types/notificationTemplate'
import { templateError } from './templateHelpers'

export function useSendTemplate(form: DingTalkNotificationSendPayload, confirmReplace: () => Promise<boolean>) {
  const options = ref<DingTalkSendTemplate[]>([])
  const selected = ref<DingTalkSendTemplate | null>(null)
  const loading = ref(false)
  const loaded = ref(false)
  const switching = ref(false)
  const error = ref('')
  let epoch = 0
  let switchEpoch = 0
  let initial = { title: '', content: '' }
  const edited = computed(() => form.title !== initial.title || form.content !== initial.content)
  const selectionError = computed(() => {
    if (!selected.value) return ''
    const latest = options.value.find(item => item.id === selected.value?.id)
    if (!latest || !latest.enabled) return '所选模板已停用或删除，请重新选择模板。'
    if (latest.version !== selected.value.version) return '所选模板已有新版本，请重新选择并确认内容。'
    return ''
  })
  const canSend = computed(() => !switching.value && (!selected.value || (
    loaded.value && !loading.value && !error.value && !selectionError.value
  )))

  async function loadOptions() {
    if (loading.value) return
    const requestEpoch = ++epoch
    loading.value = true
    error.value = ''
    try {
      const collected = new Map<number, DingTalkSendTemplate>()
      // Fetch every page explicitly; fail visibly if the bounded snapshot is incomplete.
      for (let page = 1; ; page++) {
        const { data } = await notificationTemplatesApi.options({ page, pageSize: 100 })
        if (requestEpoch !== epoch) return
        for (const item of data.list) collected.set(item.id, item)
        if (collected.size >= data.total) break
        if (!data.list.length || page >= 100) throw { msg: '模板未完整加载。' }
      }
      options.value = [...collected.values()].filter(item => item.enabled)
      loaded.value = true
    } catch (cause) {
      if (requestEpoch === epoch) error.value = templateError(cause, '模板加载失败。')
    } finally { if (requestEpoch === epoch) loading.value = false }
  }

  async function selectTemplate(id: number) {
    if (switching.value || (id !== 0 && (loading.value || error.value || !loaded.value))) return
    const next = id === 0 ? null : options.value.find(item => item.id === id && item.enabled)
    if (next === undefined || (next?.id === selected.value?.id && next?.version === selected.value?.version)) return
    const selectionEpoch = epoch
    const currentSwitch = ++switchEpoch
    switching.value = true
    try {
      if (edited.value && !await confirmReplace()) return
      if (selectionEpoch !== epoch) return
      selected.value = next ? { ...next, target: { ...next.target } } : null
      initial = { title: next?.title || '', content: next?.content || '' }
      form.title = initial.title
      form.content = initial.content
    } finally { if (currentSwitch === switchEpoch) switching.value = false }
  }

  function buildPayload(): DingTalkNotificationSendPayload {
    if (!canSend.value) throw new Error(error.value || selectionError.value || '请等待模板加载完成')
    return {
      title: form.title.trim(), content: form.content, scope: form.scope,
      ...(form.scope === 'users' ? { userIds: form.userIds } : {}),
      ...(form.scope === 'departments' ? { departmentIds: form.departmentIds } : {}),
      ...(selected.value ? { templateId: selected.value.id, templateVersion: selected.value.version } : {}),
    }
  }

  function reset() {
    epoch++
    switchEpoch++
    options.value = []
    selected.value = null
    loading.value = false
    loaded.value = false
    switching.value = false
    error.value = ''
    initial = { title: '', content: '' }
  }
  return { options, selected, loading, loaded, switching, error, edited, selectionError, canSend, loadOptions, selectTemplate, buildPayload, reset }
}
