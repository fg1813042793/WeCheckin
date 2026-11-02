import { ref } from 'vue'
import { notificationTemplatesApi } from '@/api/notificationTemplates'
import type { DingTalkSendTemplate, DingTalkSendTemplateInput } from '@/types/notificationTemplate'
import { newTemplateDraft, templatePayload, templateError, validateTemplate } from './templateHelpers'

export function useTemplateManager(canEdit: () => boolean, confirmDelete: (name: string) => Promise<boolean>) {
  const items = ref<DingTalkSendTemplate[]>([])
  const loading = ref(false)
  const error = ref('')
  const keyword = ref('')
  const page = ref(1)
  const pageSize = ref(20)
  const total = ref(0)
  const busy = ref(false)
  const mutationError = ref('')
  const editorVisible = ref(false)
  const editing = ref<DingTalkSendTemplate | null>(null)
  const draft = ref(newTemplateDraft())
  async function load() {
    if (loading.value) return
    loading.value = true
    error.value = ''
    try {
      const { data } = await notificationTemplatesApi.list({ keyword: keyword.value.trim(), page: page.value, pageSize: pageSize.value })
      items.value = data.list
      total.value = data.total
    }
    catch (cause) { error.value = templateError(cause, '模板列表加载失败，请重试。') }
    finally { loading.value = false }
  }
  function openEditor(row?: DingTalkSendTemplate, copy = false) {
    if (!canEdit() || busy.value || loading.value) return
    editing.value = row && !copy ? { ...row, target: { ...row.target } } : null
    draft.value = newTemplateDraft(row, copy)
    mutationError.value = ''
    editorVisible.value = true
  }
  async function save(value: DingTalkSendTemplateInput): Promise<boolean> {
    if (!canEdit() || busy.value || !editorVisible.value || Object.keys(validateTemplate(value)).length) return false
    busy.value = true
    mutationError.value = ''
    try {
      const payload = templatePayload(value)
      if (editing.value) await notificationTemplatesApi.update(editing.value.id, { ...payload, version: editing.value.version })
      else await notificationTemplatesApi.create(payload)
      editorVisible.value = false
      await load()
      return true
    } catch (cause) {
      mutationError.value = templateError(cause, '保存失败，已保留编辑内容，请重试。')
      return false
    } finally { busy.value = false }
  }
  async function toggle(row: DingTalkSendTemplate): Promise<boolean> {
    if (!canEdit() || busy.value || loading.value) return false
    busy.value = true
    mutationError.value = ''
    try {
      await notificationTemplatesApi.update(row.id, { ...templatePayload(row), enabled: !row.enabled, version: row.version })
      await load()
      return true
    } catch (cause) {
      mutationError.value = templateError(cause, '状态更新失败，请刷新列表后重试。')
      return false
    } finally { busy.value = false }
  }
  async function remove(row: DingTalkSendTemplate): Promise<boolean> {
    if (!canEdit() || busy.value || loading.value) return false
    busy.value = true
    mutationError.value = ''
    try {
      if (!await confirmDelete(row.name) || !canEdit()) return false
      await notificationTemplatesApi.remove(row.id, row.version)
      if (items.value.length === 1 && page.value > 1) page.value--
      await load()
      return true
    } catch (cause) {
      mutationError.value = templateError(cause, '删除失败，请刷新列表后重试。')
      return false
    } finally { busy.value = false }
  }
  function search() { if (!loading.value && !busy.value) { page.value = 1; void load() } }
  return { items, loading, error, keyword, page, pageSize, total, busy, mutationError, editorVisible, editing, draft, load, search, openEditor, save, toggle, remove }
}
