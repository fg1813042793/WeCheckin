import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { publicPortalApi } from '@/api/publicPortal'
import { hasPerm } from '@/utils/permission'
import type { PublicPortalAdminItem, PublicPortalItemInput, PublicPortalSettings } from '@/types/publicPortal'
import { canEditItem, canRemoveItem, canToggleItem, errorText, itemInputError, itemWritePayload, PORTAL_EDIT_PERMISSION, publicActivitiesUrl, settingsError, settingsWritePayload } from './portalModel'

export function usePublicPortalAdmin() {
  const settings = ref<PublicPortalSettings | null>(null)
  const items = ref<PublicPortalAdminItem[]>([])
  const loading = ref(false)
  const busy = ref(false)
  const loadError = ref('')
  const actionError = ref('')
  const dialogVisible = ref(false)
  const editing = ref<PublicPortalAdminItem | null>(null)
  const publicUrl = computed(() => publicActivitiesUrl())
  const sortedItems = computed(() => [...items.value].sort((a, b) => a.sort - b.sort || a.id - b.id))

  const canEdit = () => hasPerm(PORTAL_EDIT_PERMISSION)
  const mayEditItem = (item: PublicPortalAdminItem) => canEditItem(canEdit(), item)
  const mayRemoveItem = (item: PublicPortalAdminItem) => canRemoveItem(canEdit(), item)
  const mayToggleItem = (item: PublicPortalAdminItem) => canToggleItem(canEdit(), item)
  const writable = () => canEdit() && !loading.value && !busy.value && !loadError.value && Boolean(settings.value)

  async function load(reloadSettings = true): Promise<boolean> {
    if (loading.value) return false
    loading.value = true
    loadError.value = ''
    try {
      const { data } = await publicPortalApi.get()
      if (reloadSettings || !settings.value) settings.value = { ...data.settings }
      items.value = data.items
      return true
    } catch (error) {
      loadError.value = errorText(error, '加载公共入口配置失败，请刷新重试')
      return false
    } finally { loading.value = false }
  }

  async function saveSettings(input: PublicPortalSettings) {
    if (!writable()) return
    actionError.value = settingsError(input)
    if (actionError.value) return
    busy.value = true
    try {
      await publicPortalApi.saveSettings(settingsWritePayload(input))
      if (await load()) ElMessage.success('页面设置已保存')
      else actionError.value = '设置已保存，但读取最新配置失败；请刷新后继续操作'
    } catch (error) { actionError.value = errorText(error) } finally { busy.value = false }
  }

  function openCreate() {
    if (!writable()) return
    if (items.value.length >= 200) { actionError.value = '最多展示 200 项，请先移除不需要的内容'; return }
    editing.value = null
    actionError.value = ''
    dialogVisible.value = true
  }

  function openEdit(item: PublicPortalAdminItem) {
    if (!writable() || !mayEditItem(item)) return
    editing.value = { ...item }
    actionError.value = ''
    dialogVisible.value = true
  }

  async function saveItem(input: PublicPortalItemInput) {
    if (!writable() || (editing.value && !mayEditItem(editing.value))) return
    actionError.value = itemInputError(input)
    if (actionError.value) return
    busy.value = true
    try {
      if (editing.value) await publicPortalApi.update(editing.value.id, { ...input, version: editing.value.version })
      else await publicPortalApi.create(input)
      dialogVisible.value = false
      if (await load(false)) ElMessage.success('展示内容已保存')
      else actionError.value = '内容已保存，但读取最新列表失败；请刷新后继续操作'
    } catch (error) { actionError.value = errorText(error) } finally { busy.value = false }
  }

  async function changeItem(item: PublicPortalAdminItem, change: { enabled?: boolean; sort?: number }) {
    if (!writable() || !mayEditItem(item)) return
    if (change.enabled !== undefined && (!mayToggleItem(item) || (change.enabled && !item.available))) return
    if (change.sort !== undefined && !item.available) return
    const payload = { ...itemWritePayload(item), ...change }
    actionError.value = itemInputError(payload)
    if (actionError.value) return
    busy.value = true
    try {
      await publicPortalApi.update(item.id, payload)
      if (await load(false)) ElMessage.success('展示内容已更新')
      else actionError.value = '更新已保存，但读取最新列表失败；请刷新后继续操作'
    } catch (error) { actionError.value = errorText(error) } finally { busy.value = false }
  }

  async function removeItem(item: PublicPortalAdminItem) {
    if (!writable() || !mayRemoveItem(item)) return
    busy.value = true
    actionError.value = ''
    try {
      await ElMessageBox.confirm(`从公共入口移除“${item.title || item.displayTitle || item.managementTitle || `#${item.sourceId}`}”？原问卷或考试不会删除。`, '移除展示内容', { type: 'warning', confirmButtonText: '移除', cancelButtonText: '取消' })
      if (!canEdit() || !mayRemoveItem(item)) return
      await publicPortalApi.remove(item.id, item.version)
      if (await load(false)) ElMessage.success('已移除展示内容')
      else actionError.value = '移除已完成，但读取最新列表失败；请刷新后继续操作'
    } catch (error) {
      if (error !== 'cancel' && error !== 'close') actionError.value = errorText(error)
    } finally { busy.value = false }
  }

  onMounted(() => { void load() })
  return { settings, items, sortedItems, loading, busy, loadError, actionError, dialogVisible, editing, publicUrl, canEdit, mayEditItem, mayRemoveItem, mayToggleItem, load, saveSettings, openCreate, openEdit, saveItem, changeItem, removeItem }
}
