<script setup lang="ts">
import type { AppConfig, DingTalkUser } from '@/types/dingtalk-h5'
import type { AppNavItem } from '@/types/navigation'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { changePassword as changeProfilePassword, updateProfile as updateProfileInfo, uploadAvatar as uploadProfileAvatarFile } from '@/api/dingtalk-h5/profile'
import { getNotificationUnreadCount } from '@/api/notifications'
import AppNotificationPanel from '@/components/app-notification-panel/app-notification-panel.vue'
import { useAppContentStore, useAppShellStore, useDingtalkAuthStore } from '@/stores'
import { userAvatarInitial } from '@/utils/avatar'
import { departmentLeafFromEntity } from '@/utils/departments'

const props = defineProps<{
  activeKey: string
  appConfig: AppConfig
  appTitle: string
  loading: boolean
  navItems: AppNavItem[]
  pageTitle: string
  user: DingTalkUser
}>()

const emit = defineEmits<{
  logout: []
  navigate: [item: AppNavItem]
  refresh: []
}>()

interface ProfileCenterForm {
  loading: boolean
  uploading: boolean
  avatar: string
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

interface TabScrollTemplateRef {
  $el?: HTMLElement
}

const shell = useAppShellStore()
const appContent = useAppContentStore()
const auth = useDingtalkAuthStore()
const profileOpen = ref(false)
const topMenuOpenKey = ref('')
const mobileShell = ref(resolveMobileShell())
const profileCenterVisible = ref(false)
const notificationPanelVisible = ref(false)
const notificationUnreadCount = ref(0)
const tabClosePromptVisible = ref(false)
const tabCloseSaving = ref(false)
const tabScrollContainer = ref<HTMLElement | TabScrollTemplateRef | null>(null)
const tabsOverflow = ref(false)
const canScrollTabsLeft = ref(false)
const canScrollTabsRight = ref(false)
const pendingCloseTab = ref<AppNavItem | null>(null)
const pendingCloseCanSaveDraft = computed(() => Boolean(
  pendingCloseTab.value && appContent.canSaveTabDraft(pendingCloseTab.value.key),
))
const profileCenterForm = ref<ProfileCenterForm>(emptyProfileCenterForm())
const profileCenterResetDelay = 280
let profileCenterResetTimer: ReturnType<typeof setTimeout> | null = null
let notificationCountTimer: ReturnType<typeof setInterval> | null = null
let tabScrollSyncTimer: ReturnType<typeof setTimeout> | null = null
let tabResizeObserver: ResizeObserver | null = null
let notificationCountRequest = 0
const sidebarCollapsed = computed(() => shell.sidebarCollapsed)
const isTopMenuLayout = computed(() => !mobileShell.value && shell.menuLayout === 'top')
const appShellClass = computed(() => ({
  'app-shell--mobile': mobileShell.value,
  'app-shell--sidebar-collapsed': sidebarCollapsed.value && !isTopMenuLayout.value,
  'app-shell--top-menu': isTopMenuLayout.value,
}))

function readSystemInfo() {
  try {
    return uni.getSystemInfoSync() as unknown as Record<string, unknown>
  }
  catch {
    return {}
  }
}

function resolveMobileShell() {
  const info = readSystemInfo()
  const widthValue = Number(info.windowWidth || info.screenWidth || 0)
  const width = Number.isFinite(widthValue) ? widthValue : 0
  const deviceType = String(info.deviceType || '').toLowerCase()
  const platform = String(info.platform || '').toLowerCase()
  const userAgent = typeof navigator === 'undefined' ? '' : navigator.userAgent
  const phoneUserAgent = /Android|iPhone|iPod|Mobile/i.test(userAgent) && !/iPad/i.test(userAgent)

  return (width > 0 && width <= 768) || deviceType === 'phone' || phoneUserAgent || (['android', 'ios'].includes(platform) && width <= 1024)
}

function syncMobileShell() {
  mobileShell.value = resolveMobileShell()
  if (mobileShell.value) {
    closeMobileSubmenu()
    closeTopMenu()
  }
}

function flattenNav(items: AppNavItem[]): AppNavItem[] {
  return items.flatMap(item => [item, ...flattenNav(item.children || [])])
}

const flatNavItems = computed(() => flattenNav(props.navItems))
const dynamicTabs = computed(() => appContent.dynamicTabs)
const dynamicContentActive = computed(() => Boolean(appContent.dynamicTab(props.activeKey)))
const allTabItems = computed(() => [...flatNavItems.value, ...dynamicTabs.value])

const activeNavItem = computed(() => {
  return allTabItems.value.find(item => item.key === props.activeKey) || flatNavItems.value[0] || null
})

const routeTabs = computed(() => {
  const dashboard = flatNavItems.value.find(item => item.key === 'dashboard') || props.navItems[0] || null
  const active = activeNavItem.value
  const tabKeys = [
    dashboard?.key,
    ...shell.openedTabKeys,
    ...dynamicTabs.value.map(item => item.key),
    active?.key,
  ].filter(Boolean) as string[]
  const tabs = tabKeys
    .filter((key, index, list) => list.indexOf(key) === index)
    .map(key => allTabItems.value.find(item => item.key === key))
    .filter((item): item is AppNavItem => Boolean(item))

  if (dashboard && !tabs.some(item => item.key === dashboard.key)) {
    tabs.unshift(dashboard)
  }
  return tabs
})

const userMeta = computed(() => {
  return [props.user.department, props.user.position].filter(Boolean).join(' · ')
})

const sidebarDepartment = computed(() => {
  return departmentLeafFromEntity(props.user)
})

const userInitial = computed(() => {
  return userAvatarInitial(props.user.name, props.user.account, props.user.id)
})

const userAvatarUrl = computed(() => {
  return firstText(props.user.avatar, props.user.avatarUrl, props.user.pic, props.user.userPic)
})

const profileCenterAvatarPreview = computed(() => {
  return firstText(profileCenterForm.value.avatar)
})

const profileCenterDisplayName = computed(() => {
  return firstText(props.user.name, props.user.account, props.user.id, '当前用户')
})

const profileCenterInitial = computed(() => {
  return userAvatarInitial(profileCenterDisplayName.value)
})

const profileTriggerAvatarStyle = {
  width: '32px',
  minWidth: '32px',
  maxWidth: '32px',
  height: '32px',
  minHeight: '32px',
  maxHeight: '32px',
  flex: '0 0 32px',
  borderRadius: '6px',
  overflow: 'hidden',
}

const profileDropdownButtonStyle = {
  display: 'flex',
  width: '100%',
  minWidth: 'auto',
  height: '28px',
  minHeight: '28px',
  margin: '0',
  padding: '0 6px',
  lineHeight: '28px',
  fontSize: '12px',
  borderRadius: '4px',
  border: '0',
}

function emptyProfileCenterForm(): ProfileCenterForm {
  return {
    loading: false,
    uploading: false,
    avatar: '',
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === 'object')
}

function firstText(...values: Array<unknown>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text) {
      return text
    }
  }
  return ''
}

function isActive(item: AppNavItem | null) {
  if (!item) {
    return false
  }
  return props.activeKey === item.key
}

function isRootActive(item: AppNavItem) {
  return isActive(item) || (item.children || []).some(child => isActive(child))
}

function hasChildren(item: AppNavItem) {
  return (item.children || []).length > 0
}

function isExpanded(item: AppNavItem) {
  return hasChildren(item) && !sidebarCollapsed.value && shell.isRootExpanded(item.key)
}

function rootNavClass(item: AppNavItem) {
  return ['sidebar-nav__item', isRootActive(item) ? 'sidebar-nav__item--active' : ''].filter(Boolean).join(' ')
}

function childNavClass(item: AppNavItem) {
  return ['sidebar-nav__child', isActive(item) ? 'sidebar-nav__child--active' : ''].filter(Boolean).join(' ')
}

function topNavClass(item: AppNavItem) {
  return [
    'top-nav__item',
    isRootActive(item) ? 'top-nav__item--active' : '',
    topMenuOpenKey.value === item.key ? 'top-nav__item--open' : '',
  ].filter(Boolean).join(' ')
}

function topChildNavClass(item: AppNavItem) {
  return ['top-nav__child', isActive(item) ? 'top-nav__child--active' : ''].filter(Boolean).join(' ')
}

function tabClass(item: AppNavItem) {
  return ['app-shell__tab', isActive(item) ? 'app-shell__tab--active' : ''].filter(Boolean).join(' ')
}

function clearTabScrollSyncTimer() {
  if (tabScrollSyncTimer) {
    clearTimeout(tabScrollSyncTimer)
    tabScrollSyncTimer = null
  }
}

function tabScrollElement(): HTMLElement | null {
  const target = tabScrollContainer.value
  if (target) {
    if ('$el' in target && target.$el)
      return target.$el
    if ('scrollWidth' in target)
      return target as HTMLElement
  }
  // #ifdef H5
  if (typeof document !== 'undefined')
    return document.querySelector<HTMLElement>('.app-shell__tabs-inner')
  // #endif
  return null
}

function syncTabScrollState() {
  const container = tabScrollElement()
  if (!container) {
    tabsOverflow.value = false
    canScrollTabsLeft.value = false
    canScrollTabsRight.value = false
    return
  }
  const maxScrollLeft = Math.max(0, container.scrollWidth - container.clientWidth)
  tabsOverflow.value = maxScrollLeft > 1
  canScrollTabsLeft.value = container.scrollLeft > 1
  canScrollTabsRight.value = container.scrollLeft < maxScrollLeft - 1
}

function scheduleTabScrollStateSync() {
  clearTabScrollSyncTimer()
  tabScrollSyncTimer = setTimeout(() => {
    tabScrollSyncTimer = null
    syncTabScrollState()
  }, 260)
}

function scrollRouteTabs(direction: -1 | 1) {
  const container = tabScrollElement()
  if (!container)
    return
  const maxScrollLeft = Math.max(0, container.scrollWidth - container.clientWidth)
  const distance = Math.max(160, Math.round(container.clientWidth * 0.65))
  const target = Math.min(maxScrollLeft, Math.max(0, container.scrollLeft + direction * distance))
  container.scrollTo({
    left: target,
    behavior: 'auto',
  })
  scheduleTabScrollStateSync()
}

function ensureActiveTabVisible() {
  const container = tabScrollElement()
  if (!container)
    return
  const activeTab = Array.from(container.querySelectorAll<HTMLElement>('.app-shell__tab'))
    .find(tab => tab.dataset.tabKey === props.activeKey)
  if (!activeTab) {
    syncTabScrollState()
    return
  }
  const containerRect = container.getBoundingClientRect()
  const activeTabRect = activeTab.getBoundingClientRect()
  const edgePadding = 8
  let target = container.scrollLeft
  if (activeTabRect.left < containerRect.left + edgePadding)
    target += activeTabRect.left - containerRect.left - edgePadding
  else if (activeTabRect.right > containerRect.right - edgePadding)
    target += activeTabRect.right - containerRect.right + edgePadding
  target = Math.max(0, target)
  if (Math.abs(target - container.scrollLeft) > 1) {
    container.scrollTo({
      left: target,
      behavior: 'auto',
    })
    scheduleTabScrollStateSync()
    return
  }
  syncTabScrollState()
}

function queueActiveTabVisibilitySync() {
  void nextTick(() => {
    ensureActiveTabVisible()
    scheduleTabScrollStateSync()
  })
}

function observeTabScrollContainer() {
  if (typeof ResizeObserver === 'undefined')
    return
  tabResizeObserver?.disconnect()
  const container = tabScrollElement()
  if (!container)
    return
  tabResizeObserver = new ResizeObserver(() => {
    syncTabScrollState()
    ensureActiveTabVisible()
  })
  tabResizeObserver.observe(container)
}

function handleWindowResize() {
  syncMobileShell()
  queueActiveTabVisibilitySync()
}

function scrollTabsHorizontally(event: WheelEvent) {
  const container = event.currentTarget as HTMLElement | null
  if (!container || container.scrollWidth <= container.clientWidth)
    return
  const delta = Math.abs(event.deltaX) > Math.abs(event.deltaY) ? event.deltaX : event.deltaY
  if (!delta)
    return
  container.scrollLeft += delta
  syncTabScrollState()
  event.preventDefault()
}

function firstNavigableItem(item: AppNavItem): AppNavItem | null {
  if (item.path) {
    return item
  }
  for (const child of item.children || []) {
    const matched = firstNavigableItem(child)
    if (matched) {
      return matched
    }
  }
  return null
}

function navigateToItem(item: AppNavItem) {
  profileOpen.value = false
  closeTopMenu()
  const target = firstNavigableItem(item)
  if (!target) {
    return
  }
  rememberOpenedTab(target.key)
  emit('navigate', target)
  closeMobileSubmenu()
}

function navigateByKey(key: string) {
  const item = allTabItems.value.find(navItem => navItem.key === key)
  if (item) {
    navigateToItem(item)
  }
}

function rememberOpenedTab(key: string) {
  const dashboard = flatNavItems.value.find(item => item.key === 'dashboard') || props.navItems[0] || null
  const item = flatNavItems.value.find(navItem => navItem.key === key)
  if (!item || item.key === dashboard?.key) {
    return
  }
  shell.rememberOpenedTab(item.key, dashboard?.key)
}

function handleNavItemClick(item: AppNavItem) {
  if (hasChildren(item)) {
    if (sidebarCollapsed.value) {
      shell.setSidebarCollapsed(false)
      shell.expandRootOnly(item.key)
      navigateToItem(item)
      return
    }
    shell.toggleRootExpansion(item.key)
    return
  }
  navigateToItem(item)
}

function handleTopNavItemClick(item: AppNavItem) {
  profileOpen.value = false
  if (hasChildren(item)) {
    topMenuOpenKey.value = topMenuOpenKey.value === item.key ? '' : item.key
    return
  }
  navigateToItem(item)
}

function closeTopMenu() {
  topMenuOpenKey.value = ''
}

function closeMobileSubmenu() {
  if (!mobileShell.value || shell.expandedRootKeys.length === 0) {
    return
  }
  shell.expandRootOnly('')
}

function closeFloatingMenus() {
  closeMobileSubmenu()
  closeTopMenu()
  profileOpen.value = false
}

function handleProfileOutsideClick(event: MouseEvent) {
  if (!profileOpen.value)
    return
  const target = event.target as HTMLElement | null
  if (target?.closest('.profile-menu'))
    return
  profileOpen.value = false
}

function toggleSidebarCollapsed() {
  shell.toggleSidebarCollapsed()
  profileOpen.value = false
}

function toggleMenuLayout() {
  shell.toggleMenuLayout()
  closeTopMenu()
  profileOpen.value = false
}

async function loadNotificationUnreadCount() {
  const request = ++notificationCountRequest
  try {
    const response = await getNotificationUnreadCount()
    if (request !== notificationCountRequest)
      return
    notificationUnreadCount.value = Math.max(0, Number(response?.data?.count || 0))
  }
  catch {
    // Keep the last known count when a background refresh fails.
  }
}

function openNotificationPanel() {
  closeTopMenu()
  profileOpen.value = false
  notificationPanelVisible.value = true
  void loadNotificationUnreadCount()
}

function handleNotificationUnreadChange(count: number) {
  notificationCountRequest += 1
  notificationUnreadCount.value = Math.max(0, Number(count || 0))
}

function notificationBadgeText() {
  return notificationUnreadCount.value > 99 ? '99+' : String(notificationUnreadCount.value)
}

function toggleProfileMenu() {
  closeTopMenu()
  profileOpen.value = !profileOpen.value
}

function currentProfileAvatar() {
  return firstText(props.user.avatar, props.user.avatarUrl, props.user.pic, props.user.userPic)
}

function clearProfileCenterResetTimer() {
  if (!profileCenterResetTimer) {
    return
  }
  clearTimeout(profileCenterResetTimer)
  profileCenterResetTimer = null
}

function resetProfileCenterForm() {
  profileCenterForm.value = emptyProfileCenterForm()
}

function scheduleProfileCenterReset() {
  clearProfileCenterResetTimer()
  profileCenterResetTimer = setTimeout(() => {
    profileCenterResetTimer = null
    if (!profileCenterVisible.value) {
      resetProfileCenterForm()
    }
  }, profileCenterResetDelay)
}

function resetProfileCenter() {
  profileCenterVisible.value = false
  scheduleProfileCenterReset()
}

function openProfileCenter() {
  closeTopMenu()
  profileOpen.value = false
  clearProfileCenterResetTimer()
  profileCenterForm.value = {
    ...emptyProfileCenterForm(),
    avatar: currentProfileAvatar(),
  }
  profileCenterVisible.value = true
}

function closeProfileCenter() {
  if (profileCenterForm.value.loading || profileCenterForm.value.uploading) {
    return
  }
  resetProfileCenter()
}

function handleProfileCenterClosed() {
  scheduleProfileCenterReset()
}

function showProfileToast(title: string) {
  uni.showToast({ title, icon: 'none' })
}

function profileErrorMessage(error: unknown, fallback: string) {
  if (!isRecord(error)) {
    return fallback
  }
  const nested = isRecord(error.data) ? error.data : {}
  return firstText(nested.msg, nested.message, error.msg, error.message, fallback)
}

function avatarFromUploadPayload(payload: unknown) {
  const data = isRecord(payload) && isRecord(payload.data) ? payload.data : {}
  return firstText(data.url, data.avatar, data.path)
}

function profileUserFromPayload(payload: unknown) {
  if (!isRecord(payload)) {
    return {}
  }
  const data = isRecord(payload.data) ? payload.data : {}
  if (isRecord(data.user)) {
    return data.user as Partial<DingTalkUser>
  }
  return data as Partial<DingTalkUser>
}

function mergedProfileUser(user: Partial<DingTalkUser>, avatar: string) {
  const nextId = firstText(user.id, user.account, props.user.id)
  const nextAvatar = firstText(user.avatar, user.avatarUrl, user.pic, user.userPic, avatar)
  return {
    ...props.user,
    ...user,
    id: nextId,
    account: firstText(user.account, nextId),
    avatar: nextAvatar,
  }
}

function chooseProfileAvatar() {
  if (profileCenterForm.value.loading || profileCenterForm.value.uploading) {
    return
  }
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      const tempFilePaths = Array.isArray(res.tempFilePaths) ? res.tempFilePaths : []
      const filePath = firstText(tempFilePaths[0])
      if (!filePath) {
        showProfileToast('请选择头像图片')
        return
      }
      uploadProfileAvatar(filePath)
    },
    fail: (error) => {
      if (!String(error.errMsg || '').toLowerCase().includes('cancel')) {
        showProfileToast('选择头像失败')
      }
    },
  })
}

async function uploadProfileAvatar(filePath: string) {
  if (!filePath || profileCenterForm.value.uploading) {
    return
  }
  profileCenterForm.value.uploading = true
  try {
    const res = await uploadProfileAvatarFile(filePath)
    const avatar = avatarFromUploadPayload(res)
    if (!avatar) {
      showProfileToast('头像上传失败')
      return
    }
    profileCenterForm.value.avatar = avatar
    showProfileToast('头像已上传，保存后生效')
  }
  catch (error) {
    showProfileToast(profileErrorMessage(error, '头像上传失败'))
  }
  finally {
    profileCenterForm.value.uploading = false
  }
}

function clearProfileAvatar() {
  if (profileCenterForm.value.loading || profileCenterForm.value.uploading) {
    return
  }
  profileCenterForm.value.avatar = ''
}

async function submitProfileCenter() {
  if (profileCenterForm.value.loading) {
    return
  }
  if (profileCenterForm.value.uploading) {
    showProfileToast('头像上传中，请稍后保存')
    return
  }

  const avatar = profileCenterForm.value.avatar.trim()
  const currentPassword = profileCenterForm.value.currentPassword.trim()
  const newPassword = profileCenterForm.value.newPassword.trim()
  const confirmPassword = profileCenterForm.value.confirmPassword.trim()
  const avatarChanged = avatar !== currentProfileAvatar()
  const passwordChanging = Boolean(newPassword || confirmPassword)

  if (passwordChanging && !currentPassword) {
    showProfileToast('修改密码时请填写当前密码')
    return
  }
  if (passwordChanging && newPassword.length < 6) {
    showProfileToast('新密码至少 6 位')
    return
  }
  if (passwordChanging && newPassword !== confirmPassword) {
    showProfileToast('两次输入的新密码不一致')
    return
  }
  if (!avatarChanged && !passwordChanging) {
    showProfileToast('没有需要保存的修改')
    return
  }

  profileCenterForm.value.loading = true
  try {
    if (avatarChanged) {
      const res = await updateProfileInfo({ avatar })
      auth.updateCurrentUser(mergedProfileUser(profileUserFromPayload(res), avatar))
    }
    if (passwordChanging) {
      await changeProfilePassword({
        currentPassword,
        oldPassword: currentPassword,
        newPassword,
        confirmPassword,
      })
    }
    resetProfileCenter()
    showProfileToast('个人信息已保存')
  }
  catch (error) {
    showProfileToast(profileErrorMessage(error, '保存失败'))
  }
  finally {
    profileCenterForm.value.loading = false
  }
}

function completeTabClose(item: AppNavItem) {
  if (item.key === 'dashboard') {
    return
  }
  const tabs = routeTabs.value
  const closingIndex = tabs.findIndex(tab => tab.key === item.key)
  const nextItem = tabs[Math.max(0, closingIndex - 1)] || tabs.find(tab => tab.key === 'dashboard')
  if (appContent.dynamicTab(item.key))
    appContent.removeDynamicTab(item.key)
  else
    shell.removeOpenedTab(item.key)
  if (item.key === props.activeKey) {
    if (nextItem)
      navigateToItem(nextItem)
    else
      appContent.switchContent('dashboard')
  }
}

function closeTab(item: AppNavItem) {
  if (item.key === 'dashboard')
    return
  if (!appContent.hasUnsavedTabChanges(item.key)) {
    completeTabClose(item)
    return
  }
  pendingCloseTab.value = item
  tabClosePromptVisible.value = true
}

function continueEditingTab() {
  tabClosePromptVisible.value = false
  pendingCloseTab.value = null
}

function discardAndCloseTab() {
  const item = pendingCloseTab.value
  tabClosePromptVisible.value = false
  pendingCloseTab.value = null
  if (item)
    completeTabClose(item)
}

async function saveAndCloseTab() {
  const item = pendingCloseTab.value
  if (!item || tabCloseSaving.value)
    return
  tabCloseSaving.value = true
  try {
    if (!await appContent.saveTabDraft(item.key))
      return
    tabClosePromptVisible.value = false
    pendingCloseTab.value = null
    completeTabClose(item)
  }
  finally {
    tabCloseSaving.value = false
  }
}

function handleLogout() {
  profileOpen.value = false
  emit('logout')
}

watch(() => props.activeKey, (key) => {
  rememberOpenedTab(key)
  queueActiveTabVisibilitySync()
  if (mobileShell.value) {
    shell.expandRootOnly('')
    return
  }
  const root = props.navItems.find(item => item.key === key || (item.children || []).some(child => child.key === key))
  if (root?.children?.length) {
    shell.expandRootOnly(root.key)
  }
}, { immediate: true })

watch(flatNavItems, () => {
  rememberOpenedTab(props.activeKey)
})

watch(() => routeTabs.value.map(item => item.key).join('|'), () => {
  queueActiveTabVisibilitySync()
})

watch(tabsOverflow, () => {
  queueActiveTabVisibilitySync()
})

watch(() => appContent.closeRequestTick, () => {
  const item = allTabItems.value.find(tab => tab.key === appContent.closeRequestKey)
  if (item)
    closeTab(item)
})

watch(() => appContent.refreshTick, () => {
  void loadNotificationUnreadCount()
})

onMounted(() => {
  syncMobileShell()
  queueActiveTabVisibilitySync()
  void nextTick(() => observeTabScrollContainer())
  void loadNotificationUnreadCount()
  notificationCountTimer = setInterval(() => void loadNotificationUnreadCount(), 60000)
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', handleWindowResize)
    window.addEventListener('click', handleProfileOutsideClick, true)
  }
})

onBeforeUnmount(() => {
  clearProfileCenterResetTimer()
  clearTabScrollSyncTimer()
  tabResizeObserver?.disconnect()
  tabResizeObserver = null
  if (notificationCountTimer) {
    clearInterval(notificationCountTimer)
    notificationCountTimer = null
  }
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', handleWindowResize)
    window.removeEventListener('click', handleProfileOutsideClick, true)
  }
})
</script>

<template>
  <view class="app-shell" :class="appShellClass" @click="closeFloatingMenus">
    <view class="app-shell__topbar">
      <view class="app-shell__brand">
        <image v-if="appConfig.logoUrl" class="brand__logo-image" :src="appConfig.logoUrl" mode="aspectFill" />
        <view v-else class="brand__logo-text">
          {{ appConfig.logoText || 'OKR' }}
        </view>
        <view class="brand__copy">
          <text class="brand__title">
            {{ appTitle }}
          </text>
        </view>
      </view>

      <view v-if="isTopMenuLayout" class="app-shell__top-nav" @click.stop>
        <view v-for="item in navItems" :key="item.key" class="top-nav__group">
          <u-button
            :custom-class="topNavClass(item)"
            @click.stop="handleTopNavItemClick(item)"
          >
            <u-icon class="nav-root-icon" :name="item.icon || 'grid'" size="17px" :color="isRootActive(item) ? '#1677ff' : '#86909c'" />
            <text class="top-nav__label">
              {{ item.label }}
            </text>
            <u-icon
              v-if="item.children.length"
              name="arrow-down"
              size="14px"
              :color="isRootActive(item) ? '#1677ff' : '#86909c'"
              class="top-nav__chevron"
              :class="{ 'top-nav__chevron--open': topMenuOpenKey === item.key }"
            />
          </u-button>

          <view v-if="topMenuOpenKey === item.key && item.children.length" class="top-nav__dropdown">
            <u-button
              v-for="child in item.children"
              :key="child.key"
              :custom-class="topChildNavClass(child)"
              @click.stop="navigateByKey(child.key)"
            >
              {{ child.label }}
            </u-button>
          </view>
        </view>
      </view>

      <view class="app-shell__top-actions">
        <view class="notification-trigger-wrap">
          <u-button custom-class="top-action-btn" title="站内信" @click.stop="openNotificationPanel">
            <u-icon name="bell" size="18px" color="#5f6b7a" />
          </u-button>
          <text v-if="notificationUnreadCount > 0" class="notification-badge">
            {{ notificationBadgeText() }}
          </text>
        </view>
        <u-button
          v-if="!mobileShell"
          custom-class="top-action-btn"
          :title="isTopMenuLayout ? '切换为左侧菜单' : '切换为顶部菜单'"
          @click.stop="toggleMenuLayout"
        >
          <u-icon :name="isTopMenuLayout ? 'list' : 'grid'" size="18px" color="#5f6b7a" />
        </u-button>
        <u-button custom-class="top-action-btn" title="刷新" @click.stop="emit('refresh')">
          <u-icon name="reload" size="18px" color="#5f6b7a" />
        </u-button>
        <view class="profile-menu">
          <u-button custom-class="profile-trigger" title="用户菜单" @click.stop="toggleProfileMenu">
            <u-avatar
              custom-class="profile-trigger__avatar"
              :custom-style="profileTriggerAvatarStyle"
              :src="userAvatarUrl"
              :text="userAvatarUrl ? '' : userInitial"
              size="32"
              mode="square"
              bg-color="transparent"
            />
          </u-button>
          <view v-if="profileOpen" class="profile-dropdown profile-dropdown--actions-only">
            <u-button custom-class="profile-dropdown__action" :custom-style="profileDropdownButtonStyle" @click.stop="openProfileCenter">
              <text class="profile-dropdown__action-text">
                个人信息
              </text>
            </u-button>
            <u-button custom-class="profile-dropdown__action profile-dropdown__action--danger" :custom-style="profileDropdownButtonStyle" @click.stop="handleLogout">
              <text class="profile-dropdown__action-text">
                退出登录
              </text>
            </u-button>
          </view>
        </view>
      </view>
    </view>

    <view class="app-shell__sidebar">
      <view class="sidebar-caption">
        <text class="sidebar-caption__title">
          功能菜单
        </text>
        <text class="sidebar-caption__subtitle">
          {{ sidebarDepartment || pageTitle }}
        </text>
      </view>

      <view class="sidebar-nav">
        <view v-for="item in navItems" :key="item.key" class="sidebar-nav__group">
          <u-button
            :custom-class="rootNavClass(item)"
            @click.stop="handleNavItemClick(item)"
          >
            <u-icon class="nav-root-icon" :name="item.icon || 'grid'" size="17px" :color="isRootActive(item) ? '#1677ff' : '#86909c'" />
            <text class="sidebar-nav__label">
              {{ item.label }}
            </text>
            <u-icon
              v-if="item.children.length"
              name="arrow-right"
              size="14px"
              :color="isRootActive(item) ? '#1677ff' : '#86909c'"
              class="sidebar-nav__chevron"
              :class="{ 'sidebar-nav__chevron--open': isExpanded(item) }"
            />
          </u-button>

          <view v-if="isExpanded(item)" class="sidebar-nav__children">
            <u-button
              v-for="child in item.children"
              :key="child.key"
              :custom-class="childNavClass(child)"
              @click.stop="navigateByKey(child.key)"
            >
              {{ child.label }}
            </u-button>
          </view>
        </view>
      </view>

      <view class="sidebar-bottom">
        <u-button custom-class="sidebar-collapse-btn" @click="toggleSidebarCollapsed">
          <text class="collapse-icon">
            {{ sidebarCollapsed ? '›' : '‹' }}
          </text>
          <text>{{ sidebarCollapsed ? '展开菜单' : '收起菜单' }}</text>
        </u-button>
      </view>
    </view>

    <view class="app-shell__main">
      <view class="app-shell__tabs" :class="{ 'app-shell__tabs--overflow': tabsOverflow }">
        <view
          v-show="tabsOverflow"
          class="app-shell__tab-scroll-control app-shell__tab-scroll-control--left"
          title="向左滚动页签"
          @click.stop="scrollRouteTabs(-1)"
        >
          <u-button
            custom-class="app-shell__tab-scroll-button"
            :disabled="!canScrollTabsLeft"
            @click.stop="scrollRouteTabs(-1)"
          >
            <u-icon name="arrow-left" size="14px" :color="canScrollTabsLeft ? '#4e5969' : '#c9cdd4'" />
          </u-button>
        </view>
        <view
          ref="tabScrollContainer"
          class="app-shell__tabs-inner"
          @scroll="syncTabScrollState"
          @wheel="scrollTabsHorizontally"
        >
          <view
            v-for="item in routeTabs"
            :key="item.key"
            :class="tabClass(item)"
            :data-tab-key="item.key"
            @click="navigateByKey(item.key)"
          >
            <text>{{ item.label }}</text>
            <u-button v-if="item.key !== 'dashboard'" custom-class="app-shell__tab-close" @click.stop="closeTab(item)">
              ×
            </u-button>
          </view>
        </view>
        <view
          v-show="tabsOverflow"
          class="app-shell__tab-scroll-control app-shell__tab-scroll-control--right"
          title="向右滚动页签"
          @click.stop="scrollRouteTabs(1)"
        >
          <u-button
            custom-class="app-shell__tab-scroll-button"
            :disabled="!canScrollTabsRight"
            @click.stop="scrollRouteTabs(1)"
          >
            <u-icon name="arrow-right" size="14px" :color="canScrollTabsRight ? '#4e5969' : '#c9cdd4'" />
          </u-button>
        </view>
      </view>

      <view class="app-shell__content app-pc-control-scope" :class="{ 'app-shell__content--dynamic': dynamicContentActive }">
        <slot />
      </view>
    </view>

    <AppNotificationPanel
      v-model="notificationPanelVisible"
      :mobile="mobileShell"
      :unread-count="notificationUnreadCount"
      @unread-change="handleNotificationUnreadChange"
    />

    <u-popup
      v-model="tabClosePromptVisible"
      mode="center"
      custom-class="workflow-tab-close-prompt app-pc-control-scope"
      border-radius="8"
      :mask-close-able="!tabCloseSaving"
    >
      <view class="tab-close-prompt">
        <view class="tab-close-prompt__header">
          <text class="tab-close-prompt__title">
            {{ pendingCloseCanSaveDraft ? '表单尚未保存' : '修改尚未提交' }}
          </text>
          <text class="tab-close-prompt__desc">
            {{ pendingCloseCanSaveDraft ? '关闭前是否保存当前填写内容？' : '关闭后将放弃当前修改，是否继续？' }}
          </text>
        </view>
        <view class="tab-close-prompt__actions">
          <u-button plain :disabled="tabCloseSaving" @click="continueEditingTab">
            继续填写
          </u-button>
          <u-button plain type="error" :disabled="tabCloseSaving" @click="discardAndCloseTab">
            不保存
          </u-button>
          <u-button v-if="pendingCloseCanSaveDraft" type="primary" :loading="tabCloseSaving" @click="saveAndCloseTab">
            保存草稿
          </u-button>
        </view>
      </view>
    </u-popup>

    <u-popup
      v-model="profileCenterVisible"
      mode="center"
      custom-class="app-pc-control-scope"
      border-radius="16"
      :mask-close-able="!profileCenterForm.loading && !profileCenterForm.uploading"
      @close="handleProfileCenterClosed"
    >
      <view class="profile-center">
        <view class="profile-center__header">
          <view class="profile-center__title-block">
            <text class="profile-center__title">
              个人信息
            </text>
            <text class="profile-center__desc">
              维护当前头像和登录密码。
            </text>
          </view>
          <u-button custom-class="profile-center__close app-icon-button" :disabled="profileCenterForm.loading || profileCenterForm.uploading" @click="closeProfileCenter">
            <u-icon name="close" size="18" color="#4e5969" />
          </u-button>
        </view>

        <view class="profile-center__body">
          <view class="profile-center__avatar-row">
            <view class="profile-center__avatar-preview" :class="{ 'profile-center__avatar-preview--uploading': profileCenterForm.uploading }" @click="chooseProfileAvatar">
              <u-avatar
                :src="profileCenterAvatarPreview"
                :text="profileCenterAvatarPreview ? '' : profileCenterInitial"
                size="116"
                mode="square"
                bg-color="#1677ff"
              />
              <text class="profile-center__avatar-mask">
                {{ profileCenterForm.uploading ? '上传中' : '更换' }}
              </text>
            </view>
            <view class="profile-center__avatar-meta">
              <text class="profile-center__avatar-name">
                {{ profileCenterDisplayName }}
              </text>
              <text class="profile-center__avatar-desc">
                {{ userMeta || '暂无部门岗位' }}
              </text>
              <view class="profile-center__avatar-actions">
                <u-button
                  custom-class="profile-center__mini-btn"
                  :disabled="profileCenterForm.loading || profileCenterForm.uploading"
                  :loading="profileCenterForm.uploading"
                  @click="chooseProfileAvatar"
                >
                  上传头像
                </u-button>
                <u-button
                  v-if="profileCenterForm.avatar"
                  custom-class="profile-center__mini-btn"
                  :disabled="profileCenterForm.loading || profileCenterForm.uploading"
                  @click="clearProfileAvatar"
                >
                  移除头像
                </u-button>
              </view>
              <text v-if="profileCenterForm.avatar" class="profile-center__avatar-tip">
                保存后生效
              </text>
            </view>
          </view>

          <view class="profile-center__form">
            <view class="profile-center__field">
              <text class="profile-center__label">
                当前密码
              </text>
              <u-input v-model="profileCenterForm.currentPassword" custom-class="profile-center__input" type="password" :border="true" placeholder="修改密码时必填" clearable />
            </view>
            <view class="profile-center__grid">
              <view class="profile-center__field">
                <text class="profile-center__label">
                  新密码
                </text>
                <u-input v-model="profileCenterForm.newPassword" custom-class="profile-center__input" type="password" :border="true" placeholder="不修改可留空" clearable />
              </view>
              <view class="profile-center__field">
                <text class="profile-center__label">
                  确认新密码
                </text>
                <u-input v-model="profileCenterForm.confirmPassword" custom-class="profile-center__input" type="password" :border="true" placeholder="再次输入新密码" clearable />
              </view>
            </view>
          </view>
        </view>

        <view class="profile-center__actions">
          <u-button custom-class="profile-center__cancel" :disabled="profileCenterForm.loading" @click="closeProfileCenter">
            取消
          </u-button>
          <u-button custom-class="profile-center__save" :loading="profileCenterForm.loading" @click="submitProfileCenter">
            保存
          </u-button>
        </view>
      </view>
    </u-popup>
  </view>
</template>

<style lang="scss" scoped>
@mixin mobile-shell-layout {
  height: auto;
  min-height: 100vh;
  display: block;
  overflow: visible;

  .app-shell__topbar {
    position: relative;
    z-index: 720;
    height: 54px;
    padding: 0 14px;
    overflow: visible;
  }

  .app-shell__top-nav {
    display: none;
  }

  .app-shell__sidebar {
    position: relative;
    z-index: 520;
    height: auto;
    overflow: visible;
    padding: 10px 12px;
    border-right: 0;
    border-bottom: 1px solid #e5e6eb;
  }

  .sidebar-caption,
  .sidebar-bottom {
    display: none;
  }

  .sidebar-nav {
    position: relative;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    overflow: visible;
  }

  .sidebar-nav__group {
    position: relative;
    flex: 0 0 auto;
  }

  .sidebar-nav__item,
  :deep(.sidebar-nav__item) {
    width: auto;
    padding: 0 12px;
    justify-content: flex-start;
  }

  .sidebar-nav__children {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    z-index: 560;
    min-width: 148px;
    margin: 0;
    padding: 6px;
    border: 1px solid #e5eaf3;
    border-radius: 6px;
    display: grid;
    gap: 2px;
    background: #fff;
    box-shadow: 0 12px 30px rgba(31, 35, 41, 0.14);
  }

  .sidebar-nav__child,
  :deep(.sidebar-nav__child) {
    width: 100%;
    min-width: 126px;
  }

  .app-shell__main {
    overflow: visible;
  }

  .app-shell__tabs {
    display: none;
  }

  .app-shell__content {
    padding: 16px 12px 28px;
    overflow: visible;
  }

  .profile-dropdown {
    top: 40px;
    right: 0;
    z-index: 760;
    width: min(124px, calc(100vw - 32px));
    max-width: calc(100vw - 32px);
    padding: 6px;
  }

  &.app-shell--sidebar-collapsed {
    .sidebar-nav__label,
    .sidebar-nav__chevron {
      display: inline-flex;
    }

    .sidebar-nav__children {
      display: grid;
    }

    :deep(.sidebar-nav__label),
    :deep(.sidebar-nav__chevron) {
      display: inline-flex;
    }

    :deep(.sidebar-nav__children) {
      display: grid;
    }
  }
}

.app-shell {
  height: 100vh;
  min-height: 0;
  display: grid;
  grid-template-columns: 196px minmax(0, 1fr);
  grid-template-rows: 56px minmax(0, 1fr);
  grid-template-areas:
    "topbar topbar"
    "sidebar main";
  overflow: hidden;
  background: #fff;
}

.app-shell--sidebar-collapsed {
  grid-template-columns: 64px minmax(0, 1fr);
}

.app-shell--top-menu {
  grid-template-columns: minmax(0, 1fr);
  grid-template-areas:
    "topbar"
    "main";
}

.app-shell--top-menu .app-shell__sidebar {
  display: none;
}

.app-shell__topbar {
  grid-area: topbar;
  position: relative;
  z-index: 30;
  height: 56px;
  padding: 0 20px;
  border-bottom: 1px solid #e5e6eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  background: #fff;
}

.app-shell__brand {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand__logo-image,
.brand__logo-text {
  width: 30px;
  height: 30px;
  border-radius: 6px;
  overflow: hidden;
  flex: 0 0 auto;
}

.brand__logo-text {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1677ff;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
}

.brand__copy {
  min-width: 0;
}

.brand__title {
  display: block;
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 56px;
  white-space: nowrap;
}

.app-shell__top-nav {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 4px;
  overflow: visible;
}

.top-nav__group {
  position: relative;
  flex: 0 0 auto;
}

.top-nav__item,
:deep(.top-nav__item) {
  height: 34px;
  min-height: 34px;
  max-width: 168px;
  padding: 0 12px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: transparent;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  line-height: 34px;
  white-space: nowrap;
}

.top-nav__item--active,
:deep(.top-nav__item--active),
.top-nav__item--open,
:deep(.top-nav__item--open) {
  background: #eaf3ff;
  color: #1677ff;
}

.nav-root-icon {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.top-nav__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.top-nav__chevron {
  transition: transform 0.16s ease;

  &--open {
    transform: rotate(180deg);
  }
}

.top-nav__dropdown {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  z-index: 80;
  width: 168px;
  padding: 6px;
  border: 1px solid #e5eaf3;
  border-radius: 6px;
  display: grid;
  gap: 2px;
  background: #fff;
  box-shadow: 0 14px 34px rgba(31, 35, 41, 0.14);
}

.top-nav__child,
:deep(.top-nav__child) {
  width: 100%;
  height: 32px;
  min-height: 32px;
  padding: 0 10px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  background: transparent;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  line-height: 32px;
  text-align: left;
}

.top-nav__child--active,
:deep(.top-nav__child--active) {
  background: #eaf3ff;
  color: #1677ff;
}

.app-shell__top-actions {
  position: relative;
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.notification-trigger-wrap {
  position: relative;
  flex: 0 0 32px;
}

.notification-badge {
  position: absolute;
  top: -6px;
  right: -8px;
  z-index: 2;
  min-width: 17px;
  height: 17px;
  padding: 0 4px;
  border: 2px solid #fff;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f53f3f;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  line-height: 13px;
}

.top-action-btn,
.profile-trigger,
.profile-dropdown__action,
.profile-center__close,
.profile-center__mini-btn,
.profile-center__cancel,
.profile-center__save,
.top-nav__item,
.top-nav__child,
.sidebar-nav__item,
.sidebar-nav__child,
.sidebar-collapse-btn,
.app-shell__tab-close,
:deep(.top-action-btn),
:deep(.profile-trigger),
:deep(.profile-dropdown__action),
:deep(.profile-center__close),
:deep(.profile-center__mini-btn),
:deep(.profile-center__cancel),
:deep(.profile-center__save),
:deep(.top-nav__item),
:deep(.top-nav__child),
:deep(.sidebar-nav__item),
:deep(.sidebar-nav__child),
:deep(.sidebar-collapse-btn),
:deep(.app-shell__tab-close) {
  margin: 0;
  border: 0;
  appearance: none;
  box-shadow: none;
}

.top-action-btn::after,
.profile-trigger::after,
.profile-dropdown__action::after,
.profile-center__close::after,
.profile-center__mini-btn::after,
.profile-center__cancel::after,
.profile-center__save::after,
.top-nav__item::after,
.top-nav__child::after,
.sidebar-nav__item::after,
.sidebar-nav__child::after,
.sidebar-collapse-btn::after,
.app-shell__tab-close::after,
:deep(.top-action-btn)::after,
:deep(.profile-trigger)::after,
:deep(.profile-dropdown__action)::after,
:deep(.profile-center__close)::after,
:deep(.profile-center__mini-btn)::after,
:deep(.profile-center__cancel)::after,
:deep(.profile-center__save)::after,
:deep(.top-nav__item)::after,
:deep(.top-nav__child)::after,
:deep(.sidebar-nav__item)::after,
:deep(.sidebar-nav__child)::after,
:deep(.sidebar-collapse-btn)::after,
:deep(.app-shell__tab-close)::after {
  display: none;
}

.top-action-btn,
:deep(.top-action-btn) {
  width: 32px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
}

.profile-menu {
  position: relative;
}

.profile-trigger,
:deep(.profile-trigger) {
  width: 32px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1677ff;
  color: #fff;
  font-size: 13px;
  font-weight: 800;
}

.profile-trigger__avatar,
:deep(.profile-trigger__avatar) {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  color: #fff !important;
  overflow: hidden;
}

:deep(.profile-trigger__avatar .u-line-1) {
  color: #fff !important;
  font-size: 13px !important;
  line-height: 1;
  font-weight: 800;
}

.profile-dropdown {
  position: absolute;
  top: 42px;
  right: 0;
  z-index: 40;
  width: 136px;
  padding: 6px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  background: #fff;
  box-shadow: 0 12px 32px rgba(31, 35, 41, 0.12);

  &__action {
    width: 100%;
    height: 28px;
    min-height: 28px;
    margin-top: 0;
    padding: 0 6px;
    border: 0;
    border-radius: 4px;
    background: transparent;
    color: #1677ff;
    font-size: 12px;
    font-weight: 600;
    line-height: 28px;
  }

  &__action:hover {
    background: #f2f7ff;
  }

  &__action--danger {
    color: #c42a2a;
  }

  &__action--danger:hover {
    background: #fff1f0;
  }

  &__action-text {
    display: block;
    width: 100%;
    font-size: 12px;
    line-height: 28px;
    text-align: center;
  }
}

.profile-dropdown--actions-only {
  width: min(136px, calc(100vw - 32px));
  padding: 6px;
  gap: 2px;
}

:deep(.profile-dropdown__action) {
  width: 100%;
  height: 28px;
  min-height: 28px;
  margin-top: 0;
  padding: 0 6px;
  border: 0;
  border-radius: 4px;
  background: transparent;
  color: #1677ff;
  font-size: 12px;
  font-weight: 600;
  line-height: 28px;
}

:deep(.profile-dropdown__action:hover) {
  background: #f2f7ff;
}

:deep(.profile-dropdown__action .u-button__text) {
  width: 100%;
  font-size: 12px;
  line-height: 28px;
}

:deep(.profile-dropdown__action--danger) {
  color: #c42a2a;
}

:deep(.profile-dropdown__action--danger:hover) {
  background: #fff1f0;
}

.profile-center {
  width: min(520px, calc(100vw - 32px));
  max-height: min(720px, calc(100vh - 48px));
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  overflow: hidden;
  background: #fff;
}

.profile-center__header {
  padding: 18px 20px 14px;
  border-bottom: 1px solid #f2f3f5;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.profile-center__title-block {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.profile-center__title,
.profile-center__desc,
.profile-center__avatar-name,
.profile-center__avatar-desc,
.profile-center__avatar-tip,
.profile-center__label {
  display: block;
}

.profile-center__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
  line-height: 1.35;
}

.profile-center__desc,
.profile-center__avatar-desc {
  color: #86909c;
  font-size: 13px;
  line-height: 1.5;
}

.profile-center__close,
:deep(.profile-center__close) {
  width: 34px;
  height: 34px;
  min-height: 34px;
  padding: 0;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f7f8fa;
}

.profile-center__body {
  min-height: 0;
  padding: 18px 20px;
  display: grid;
  gap: 18px;
  overflow-y: auto;
}

.profile-center__avatar-row {
  padding: 12px;
  border: 1px solid #e5eaf3;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: #f8fbff;
}

.profile-center__avatar-preview {
  position: relative;
  width: 58px;
  height: 58px;
  border-radius: 14px;
  flex: 0 0 58px;
  overflow: hidden;
  background: #eef3fb;
  cursor: pointer;
}

.profile-center__avatar-mask {
  position: absolute;
  inset: auto 0 0;
  padding: 4px 0;
  background: rgba(15, 23, 42, 0.62);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  text-align: center;
}

.profile-center__avatar-preview--uploading .profile-center__avatar-mask {
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.profile-center__avatar-meta {
  min-width: 0;
  flex: 1 1 auto;
}

.profile-center__avatar-name {
  color: #1f2329;
  font-size: 16px;
  font-weight: 800;
}

.profile-center__avatar-actions {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.profile-center__mini-btn,
:deep(.profile-center__mini-btn) {
  width: auto;
  height: 30px;
  min-height: 30px;
  padding: 0 12px;
  border: 1px solid #d7e7ff;
  border-radius: 4px;
  background: #f2f7ff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 700;
  line-height: 30px;
}

.profile-center__avatar-tip {
  margin-top: 8px;
  color: #86909c;
  font-size: 12px;
}

.profile-center__form {
  display: grid;
  gap: 14px;
}

.profile-center__field {
  min-width: 0;
  display: grid;
  gap: 8px;
}

.profile-center__label {
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
}

.profile-center__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.profile-center__input,
:deep(.profile-center__input) {
  min-height: 38px;
  font-size: 13px;
}

.profile-center__actions {
  padding: 14px 20px 18px;
  border-top: 1px solid #f2f3f5;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.profile-center__cancel,
.profile-center__save,
:deep(.profile-center__cancel),
:deep(.profile-center__save) {
  width: 92px;
  height: 38px;
  min-height: 38px;
  padding: 0 16px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 700;
  line-height: 38px;
}

.profile-center__cancel,
:deep(.profile-center__cancel) {
  border: 1px solid #cfd6e4;
  background: #fff;
  color: #4e5969;
}

.profile-center__save,
:deep(.profile-center__save) {
  background: #1677ff;
  color: #fff;
}

.app-shell__sidebar {
  grid-area: sidebar;
  min-height: 0;
  padding: 16px 8px;
  border-right: 1px solid #e5e6eb;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.sidebar-caption {
  min-height: 36px;
  margin: 0 0 8px;
  padding: 0 8px 8px;
  border-bottom: 1px solid #f2f3f5;
  display: grid;
  gap: 2px;
}

.sidebar-caption__title {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  line-height: 18px;
}

.sidebar-caption__subtitle {
  max-width: 160px;
  overflow: hidden;
  color: #86909c;
  font-size: 12px;
  line-height: 16px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-nav {
  min-height: 0;
  display: grid;
  gap: 2px;
  overflow-y: auto;
  scrollbar-width: thin;
}

.sidebar-nav__group {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.sidebar-nav__item,
:deep(.sidebar-nav__item) {
  width: 100%;
  height: 34px;
  min-height: 34px;
  padding: 0 10px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8px;
  background: transparent;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  line-height: 34px;
  text-align: left;
}

.sidebar-nav__item--active,
:deep(.sidebar-nav__item--active) {
  background: #eaf3ff;
  color: #1677ff;
}

.sidebar-nav__label {
  min-width: 0;
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-nav__chevron {
  transition: transform 0.16s ease;

  &--open {
    transform: rotate(90deg);
  }
}

.sidebar-nav__children {
  margin: 2px 0 6px 32px;
  display: grid;
  gap: 0;
}

.sidebar-nav__child,
:deep(.sidebar-nav__child) {
  width: 100%;
  height: 30px;
  min-height: 30px;
  padding: 0 10px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  background: transparent;
  color: #6b7785;
  font-size: 12px;
  font-weight: 600;
  line-height: 30px;
  text-align: left;
}

.sidebar-nav__child--active,
:deep(.sidebar-nav__child--active) {
  background: #eaf3ff;
  color: #1677ff;
  font-weight: 700;
}

.sidebar-bottom {
  margin-top: auto;
  padding-top: 16px;
}

.sidebar-collapse-btn,
:deep(.sidebar-collapse-btn) {
  width: 100%;
  height: 32px;
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid #e5e6eb;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: #fff;
  color: #4e5969;
  font-size: 13px;
  font-weight: 700;
  line-height: 32px;
}

.collapse-icon {
  font-size: 18px;
  font-weight: 800;
}

.app-shell--sidebar-collapsed {
  .app-shell__sidebar {
    padding-inline: 8px;
  }

  .sidebar-caption {
    display: none;
  }

  .sidebar-nav {
    padding-top: 4px;
  }

  .sidebar-caption__title,
  .sidebar-caption__subtitle,
  .sidebar-nav__label,
  .sidebar-nav__chevron,
  .sidebar-nav__children,
  .sidebar-collapse-btn text:not(.collapse-icon) {
    display: none;
  }

  .sidebar-nav__item {
    justify-content: center;
    padding: 0;
  }

  .sidebar-collapse-btn {
    padding: 0;
  }

  :deep(.sidebar-caption__title),
  :deep(.sidebar-caption__subtitle),
  :deep(.sidebar-nav__label),
  :deep(.sidebar-nav__chevron),
  :deep(.sidebar-nav__children),
  :deep(.sidebar-collapse-btn text:not(.collapse-icon)) {
    display: none;
  }

  :deep(.sidebar-nav__item) {
    justify-content: center;
    padding: 0;
  }

  :deep(.sidebar-collapse-btn) {
    padding: 0;
  }
}

.app-shell__main {
  grid-area: main;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.app-shell__tabs {
  position: relative;
  flex: 0 0 auto;
  height: 42px;
  padding: 0 12px;
  border-top: 1px solid #edf1f7;
  box-shadow: inset 0 -1px 0 #e5eaf3;
  display: flex;
  align-items: center;
  gap: 6px;
  background: #f7f9fc;
  box-sizing: border-box;
}

.app-shell__tabs--overflow {
  padding: 0 46px;
}

.app-shell__tab-scroll-control {
  position: absolute;
  top: 50%;
  z-index: 2;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  transform: translateY(-50%);
}

.app-shell__tab-scroll-control--left {
  left: 12px;
}

.app-shell__tab-scroll-control--right {
  right: 12px;
}

.app-shell__tab-scroll-button,
:deep(.app-shell__tab-scroll-button) {
  width: 100%;
  height: 100%;
  min-height: 28px;
  margin: 0;
  padding: 0;
  border: 1px solid #dfe5ee;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  box-sizing: border-box;
  line-height: 1;
  vertical-align: top;
}

.app-shell__tab-scroll-button::after,
:deep(.app-shell__tab-scroll-button)::after {
  display: none;
}

.app-shell__tabs-inner {
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  overflow-x: auto;
  overflow-y: hidden;
  overscroll-behavior-inline: contain;
  scrollbar-width: none;
  touch-action: pan-x;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar {
    display: none;
  }
}

.app-shell__tab {
  position: relative;
  flex: 0 0 auto;
  height: 34px;
  min-height: 34px;
  padding: 0 10px 0 16px;
  border: 1px solid transparent;
  border-radius: 6px 6px 0 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: transparent;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  line-height: 34px;
  white-space: nowrap;
  box-sizing: border-box;

  &--active {
    border-color: #cfe0ff;
    border-bottom-color: #fff;
    background: #fff;
    color: #1677ff;
  }
}

.app-shell__tab + .app-shell__tab::before {
  content: "";
  position: absolute;
  top: 10px;
  left: -1px;
  width: 1px;
  height: 14px;
  border-radius: 999px;
  background: #dfe5ef;
}

.app-shell__tab--active::before,
.app-shell__tab--active + .app-shell__tab::before {
  opacity: 0;
}

.app-shell__tab-close,
:deep(.app-shell__tab-close) {
  width: 18px;
  height: 18px;
  min-height: 18px;
  padding: 0;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  color: #86909c;
  font-size: 13px;
  font-weight: 700;
  line-height: 18px;
}

.app-shell__content {
  min-height: 0;
  flex: 1 1 auto;
  padding: 26px 24px 40px;
  overflow-x: hidden;
  overflow-y: auto;
}

.app-shell__content--dynamic {
  padding-bottom: 0;
}

.tab-close-prompt {
  width: min(420px, calc(100vw - 32px));
  padding: 22px;
  background: #fff;
  box-sizing: border-box;
}

.tab-close-prompt__header {
  display: grid;
  gap: 8px;
}

.tab-close-prompt__title {
  color: #1f2329;
  font-size: 17px;
  font-weight: 700;
}

.tab-close-prompt__desc {
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.tab-close-prompt__actions {
  margin-top: 22px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.tab-close-prompt__actions :deep(.u-btn) {
  width: 100%;
  min-width: 0;
  margin: 0;
}

@media (max-width: 520px) {
  .tab-close-prompt__actions {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .profile-center {
    width: calc(100vw - 28px);
    max-height: calc(100vh - 28px);
  }

  .profile-center__header,
  .profile-center__body,
  .profile-center__actions {
    padding-left: 16px;
    padding-right: 16px;
  }

  .profile-center__avatar-row {
    align-items: flex-start;
  }

  .profile-center__grid {
    grid-template-columns: 1fr;
  }

  .profile-center__actions {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .profile-center__cancel,
  .profile-center__save,
  :deep(.profile-center__cancel),
  :deep(.profile-center__save) {
    width: 100%;
  }
}

.app-shell--mobile {
  @include mobile-shell-layout;
}

@media (max-width: 768px), (hover: none) and (pointer: coarse) {
  .app-shell {
    @include mobile-shell-layout;
  }
}
</style>
