interface PermissionTreeInstance {
  getCheckedKeys?: () => Array<string | number>
  getHalfCheckedKeys?: () => Array<string | number>
}

interface PermissionTreeRef {
  value?: PermissionTreeInstance | null
}

export function checkedKeys(
  treeRef: PermissionTreeRef,
  options: { includeHalfChecked?: boolean; prefix?: string; prefixes?: string[] } = {}
) {
  const checked = treeRef.value?.getCheckedKeys?.() || []
  const halfChecked = options.includeHalfChecked ? (treeRef.value?.getHalfCheckedKeys?.() || []) : []
  const keys = Array.from(new Set([...checked, ...halfChecked])).map(String)
  const prefixes = options.prefixes || (options.prefix ? [options.prefix] : [])
  return prefixes.length === 0 ? keys : keys.filter(key => prefixes.some(prefix => key.startsWith(prefix)))
}
