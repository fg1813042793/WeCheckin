import { inject, onScopeDispose } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ADMIN_ROUTE_TAB_REPLACE_KEY, ADMIN_ROUTE_TAB_TITLE_KEY, formDesignerTabTitle } from './routeTabTitles'

export function useFormDesignerTab(kind: 'survey' | 'exam') {
  const route = useRoute()
  const router = useRouter()
  const openedPath = route.path || `/${kind}/designer`
  const openedFullPath = route.fullPath || openedPath
  const openedQuery = { ...route.query }
  const updateTitle = inject(ADMIN_ROUTE_TAB_TITLE_KEY, undefined)
  const replaceTab = inject(ADMIN_ROUTE_TAB_REPLACE_KEY, undefined)
  let disposed = false
  onScopeDispose(() => { disposed = true })

  return (id: number, managementTitle: string, created = false) => {
    if (disposed) return
    const title = formDesignerTabTitle(kind, id, managementTitle)
    updateTitle?.(openedFullPath, title)
    if (!created || !id) return
    const target = router.resolve({ path: openedPath, query: { ...openedQuery, id: String(id) } })
    // Creation can finish while another cached page is active. Replace the
    // originating tab without redirecting the user away from their current work.
    if (router.currentRoute.value.fullPath !== openedFullPath) {
      replaceTab?.(openedFullPath, target, title)
      return
    }
    void router.replace(target.fullPath).then((failure) => {
      if (!failure) replaceTab?.(openedFullPath, target, title)
    }).catch(() => {
      ElMessage.warning('内容已保存，页面跳转失败，请从列表重新打开')
    })
  }
}
