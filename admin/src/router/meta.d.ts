import 'vue-router'

export type AdminUiPattern = 'list' | 'filter-list' | 'form' | 'detail' | 'workspace'

export interface AdminUiContract {
  version: 1
  pattern: AdminUiPattern
}

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    menuPath?: string
    allowWithoutMenu?: boolean
    public?: boolean
    adminUi?: AdminUiContract
  }
}

export {}
