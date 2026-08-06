import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { firstText } from '../../views/performance/common/helpers'

const MENU_LAYOUT_KEY = 'dingtalk-h5-menu-layout'

const navIconMap = {
  dashboard: [
    'M4 10.5 12 4l8 6.5',
    'M6 9.5V20h5v-6h2v6h5V9.5'
  ],
  mine: [
    'M5 12.5 9.2 17 19 7',
    'M4 5h16v14H4V5Z'
  ],
  history: [
    'M12 6v6l4 2',
    'M4 12a8 8 0 1 0 2.3-5.7',
    'M4 4v5h5'
  ],
  manager: [
    'M8 4h8l1 3H7l1-3Z',
    'M6 7h12v13H6V7Z',
    'M9 12h6',
    'M9 16h4'
  ],
  hrbp: [
    'M12 4l2.3 4.7 5.2.8-3.8 3.7.9 5.2L12 16l-4.6 2.4.9-5.2-3.8-3.7 5.2-.8L12 4Z'
  ],
  summary: [
    'M5 19V9',
    'M12 19V5',
    'M19 19v-7',
    'M4 19h16'
  ],
  org: [
    'M9 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z',
    'M3.5 20a5.5 5.5 0 0 1 11 0',
    'M17 10a2.5 2.5 0 1 0 0-5',
    'M15.5 14.5A4.8 4.8 0 0 1 21 20'
  ],
  template: [
    'M6 4h12v16H6V4Z',
    'M9 8h6',
    'M9 12h6',
    'M9 16h4'
  ],
  account: [
    'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Z',
    'M4 21a8 8 0 0 1 16 0',
    'M18 5v4',
    'M16 7h4'
  ],
  performance: [
    'M5 4h14v5H5V4Z',
    'M5 13h6v7H5v-7Z',
    'M15 13h4v7h-4v-7Z'
  ]
}

function readMenuLayout() {
  try {
    if (typeof uni === 'undefined') return 'side'
    return uni.getStorageSync(MENU_LAYOUT_KEY) === 'top' ? 'top' : 'side'
  } catch (error) {
    return 'side'
  }
}

function persistMenuLayout(layout) {
  try {
    if (typeof uni !== 'undefined') {
      uni.setStorageSync(MENU_LAYOUT_KEY, layout)
    }
  } catch (error) {
    // 本地偏好保存失败不影响页面切换。
  }
}

function performanceTabKey(key) {
  return String(key || '').replace('performance:', '')
}

function hasChildren(item) {
  return Array.isArray(item?.children) && item.children.length > 0
}

function navIcon(name) {
  const key = String(name || '').trim()
  const normalizedKey = key.startsWith('performance:') ? key.replace('performance:', '') : key
  return navIconMap[key] || navIconMap[normalizedKey] || navIconMap.dashboard
}

export function useAppShellNavigation(props, emit) {
  const sidebarCollapsed = ref(false)
  const profileMenuOpen = ref(false)
  const mobileSubmenuOpenKey = ref('')
  const topSubmenuOpenKey = ref('')
  const menuLayout = ref(readMenuLayout())
  const collapsedMenuKeys = ref(new Set())
  const desktopPageTabsInnerRef = ref(null)
  const desktopTabsOverflowing = ref(false)
  const desktopTabsCanScrollLeft = ref(false)
  const desktopTabsCanScrollRight = ref(false)
  let desktopPageTabsResizeObserver = null

  const activeNavItem = computed(() => {
    return props.navItems.find((item) => item.key === props.activeView) || props.navItems[0] || null
  })

  const mobileSubmenuOpenItem = computed(() => {
    return props.navItems.find((item) => item.key === mobileSubmenuOpenKey.value && hasChildren(item)) || null
  })

  const userAvatarUrl = computed(() => props.user.avatar || props.user.avatarUrl || props.user.pic || props.user.userPic || '')
  const appDisplayTitle = computed(() => firstText(props.appConfig.appTitle, props.appConfig.appName, props.appTitle, '钉钉H5应用'))
  const appLogoText = computed(() => firstText(props.appConfig.logoText, 'H5').slice(0, 4))
  const appLogoUrl = computed(() => firstText(props.appConfig.logoUrl, props.appConfig.logoURL))
  const userInitial = computed(() => {
    const name = String(props.user.name || props.user.id || 'U').trim()
    return (name.slice(0, 1) || 'U').toUpperCase()
  })

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setMenuLayout(layout) {
    menuLayout.value = layout === 'top' ? 'top' : 'side'
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    topSubmenuOpenKey.value = ''
    persistMenuLayout(menuLayout.value)
  }

  function toggleMenuLayout() {
    setMenuLayout(menuLayout.value === 'side' ? 'top' : 'side')
  }

  function toggleProfileMenu() {
    mobileSubmenuOpenKey.value = ''
    topSubmenuOpenKey.value = ''
    profileMenuOpen.value = !profileMenuOpen.value
  }

  function closeProfileMenu() {
    profileMenuOpen.value = false
  }

  function handleLogout() {
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    topSubmenuOpenKey.value = ''
    emit('logout')
  }

  function handleOpenProfile() {
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    topSubmenuOpenKey.value = ''
    emit('open-profile')
  }

  function closeTopSubmenu() {
    topSubmenuOpenKey.value = ''
  }

  function closeMobileSubmenu() {
    mobileSubmenuOpenKey.value = ''
  }

  function handleMobileNavClick(item) {
    profileMenuOpen.value = false
    if (hasChildren(item)) {
      mobileSubmenuOpenKey.value = mobileSubmenuOpenKey.value === item.key ? '' : item.key
      return
    }
    closeMobileSubmenu()
    emit('switch-view', item.key)
  }

  function handleMobileChildClick(item) {
    closeMobileSubmenu()
    emit('switch-view', item.key)
  }

  function handleTopNavClick(item) {
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    if (hasChildren(item)) {
      topSubmenuOpenKey.value = topSubmenuOpenKey.value === item.key ? '' : item.key
      return
    }
    closeTopSubmenu()
    emit('switch-view', item.key)
  }

  function handleTopChildClick(item) {
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    closeTopSubmenu()
    emit('switch-view', item.key)
  }

  function isMenuExpanded(item) {
    if (!hasChildren(item) || props.activeView !== item.key) return false
    return !collapsedMenuKeys.value.has(item.key)
  }

  function expandMenu(key) {
    if (!collapsedMenuKeys.value.has(key)) return
    const next = new Set(collapsedMenuKeys.value)
    next.delete(key)
    collapsedMenuKeys.value = next
  }

  function toggleMenuCollapse(key) {
    const next = new Set(collapsedMenuKeys.value)
    if (next.has(key)) {
      next.delete(key)
    } else {
      next.add(key)
    }
    collapsedMenuKeys.value = next
  }

  function handleNavClick(item) {
    if (hasChildren(item) && props.activeView === item.key) {
      toggleMenuCollapse(item.key)
      return
    }
    expandMenu(item.key)
    emit('switch-view', item.key)
  }

  function handleRouteTabClick(item) {
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    topSubmenuOpenKey.value = ''
    emit('activate-route-tab', item.key)
    syncDesktopPageTabsAfterRender(true)
  }

  function handleRouteTabClose(item) {
    profileMenuOpen.value = false
    mobileSubmenuOpenKey.value = ''
    topSubmenuOpenKey.value = ''
    emit('close-route-tab', item.key)
    syncDesktopPageTabsAfterRender(false)
  }

  function isPageTabActive(item) {
    if (!item) return false
    return props.activeRouteTab === item.key
  }

  function isTopSubmenuOpen(item) {
    return hasChildren(item) && topSubmenuOpenKey.value === item.key
  }

  function isChildActive(item) {
    return props.activePerformanceTab === performanceTabKey(item.key) || props.activeView === item.key
  }

  function getDesktopPageTabsWrap() {
    const el = desktopPageTabsInnerRef.value
    if (!el) return null
    if (typeof HTMLElement !== 'undefined' && el instanceof HTMLElement) return el
    return el?.$el || null
  }

  function updateDesktopPageTabsScrollState() {
    const wrap = getDesktopPageTabsWrap()
    if (!wrap) {
      desktopTabsOverflowing.value = false
      desktopTabsCanScrollLeft.value = false
      desktopTabsCanScrollRight.value = false
      return
    }
    const maxScrollLeft = Math.max(wrap.scrollWidth - wrap.clientWidth, 0)
    desktopTabsOverflowing.value = maxScrollLeft > 1
    desktopTabsCanScrollLeft.value = desktopTabsOverflowing.value && wrap.scrollLeft > 1
    desktopTabsCanScrollRight.value = desktopTabsOverflowing.value && wrap.scrollLeft < maxScrollLeft - 1
  }

  function scrollDesktopPageTabs(direction) {
    const wrap = getDesktopPageTabsWrap()
    if (!wrap || typeof window === 'undefined') return
    const maxScrollLeft = Math.max(wrap.scrollWidth - wrap.clientWidth, 0)
    const distance = Math.max(180, Math.floor(wrap.clientWidth * 0.72))
    const target = Math.min(
      maxScrollLeft,
      Math.max(0, wrap.scrollLeft + (direction === 'left' ? -distance : distance))
    )
    wrap.scrollTo({ left: target, behavior: 'smooth' })
    window.setTimeout(updateDesktopPageTabsScrollState, 220)
  }

  function scrollActiveDesktopPageTabIntoView() {
    const wrap = getDesktopPageTabsWrap()
    const activeTab = wrap?.querySelector('.desktop-page-tab.active')
    activeTab?.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'smooth' })
    if (typeof window !== 'undefined') {
      window.setTimeout(updateDesktopPageTabsScrollState, 220)
    }
  }

  function syncDesktopPageTabsAfterRender(scrollActive) {
    nextTick(() => {
      updateDesktopPageTabsScrollState()
      if (!scrollActive) return
      nextTick(() => {
        scrollActiveDesktopPageTabIntoView()
        updateDesktopPageTabsScrollState()
      })
    })
  }

  function setupDesktopPageTabsResizeObserver() {
    const wrap = getDesktopPageTabsWrap()
    if (!wrap || typeof ResizeObserver === 'undefined') return
    desktopPageTabsResizeObserver?.disconnect()
    desktopPageTabsResizeObserver = new ResizeObserver(updateDesktopPageTabsScrollState)
    desktopPageTabsResizeObserver.observe(wrap)
  }

  onMounted(() => {
    nextTick(() => {
      setupDesktopPageTabsResizeObserver()
      updateDesktopPageTabsScrollState()
      if (typeof window !== 'undefined') {
        window.addEventListener('resize', updateDesktopPageTabsScrollState)
      }
    })
  })

  onBeforeUnmount(() => {
    desktopPageTabsResizeObserver?.disconnect()
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', updateDesktopPageTabsScrollState)
    }
  })

  watch(
    () => [props.routeTabs.length, props.activeRouteTab, menuLayout.value],
    () => syncDesktopPageTabsAfterRender(true),
    { flush: 'post' }
  )

  return {
    activeNavItem,
    appDisplayTitle,
    appLogoText,
    appLogoUrl,
    closeMobileSubmenu,
    closeProfileMenu,
    closeTopSubmenu,
    desktopPageTabsInnerRef,
    desktopTabsCanScrollLeft,
    desktopTabsCanScrollRight,
    desktopTabsOverflowing,
    handleLogout,
    handleMobileChildClick,
    handleMobileNavClick,
    handleNavClick,
    handleOpenProfile,
    handleRouteTabClick,
    handleRouteTabClose,
    handleTopChildClick,
    handleTopNavClick,
    hasChildren,
    isChildActive,
    isMenuExpanded,
    isPageTabActive,
    isTopSubmenuOpen,
    menuLayout,
    mobileSubmenuOpenItem,
    mobileSubmenuOpenKey,
    navIcon,
    profileMenuOpen,
    scrollDesktopPageTabs,
    sidebarCollapsed,
    toggleMenuLayout,
    toggleProfileMenu,
    toggleSidebar,
    topSubmenuOpenKey,
    updateDesktopPageTabsScrollState,
    userAvatarUrl,
    userInitial
  }
}
