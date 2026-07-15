import type { App, Component } from 'vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

export const adminIconMap = Object.fromEntries(
  Object.entries(ElementPlusIconsVue).filter(([name]) => name !== 'default')
) as Record<string, Component>

export type AdminIconName = string

export const ADMIN_ICON_NAMES = Object.keys(adminIconMap) as AdminIconName[]

export const DEFAULT_ADMIN_ICON_NAME: AdminIconName = 'Menu'

export function resolveAdminIcon(name?: string) {
  if (!name) return undefined
  return adminIconMap[name as AdminIconName] || adminIconMap[DEFAULT_ADMIN_ICON_NAME]
}

export function registerAdminIcons(app: App) {
  for (const [name, component] of Object.entries(adminIconMap)) {
    app.component(name, component)
  }
}
