export interface AppNavItem {
  key: string
  label: string
  icon: string
  path?: string
  children: AppNavItem[]
}
