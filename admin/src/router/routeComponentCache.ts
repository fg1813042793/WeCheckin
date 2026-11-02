import { createVNode, defineComponent, markRaw } from 'vue'
import type { Component, PropType, VNode } from 'vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'

type CacheRoute = Pick<RouteLocationNormalizedLoaded, 'fullPath' | 'name' | 'path'>

export function getRouteCacheName(fullPath: string, name: string) {
  let hash = 0
  for (let i = 0; i < fullPath.length; i++) hash = ((hash << 5) - hash + fullPath.charCodeAt(i)) | 0
  const safeName = name.replace(/[^a-zA-Z0-9_]/g, '_') || 'Route'
  return `AdminRouteCache_${safeName}_${Math.abs(hash).toString(36)}`
}

export function createAdminRouteComponentCache() {
  const entries = new Map<string, { name: string; source: VNode['type']; component: Component }>()

  function resolve(vnode: VNode | undefined, route: CacheRoute): Component | undefined {
    if (!vnode) return undefined
    const name = getRouteCacheName(route.fullPath, String(route.name || route.path))
    const entry = entries.get(route.fullPath)
    // RouterView supplies a new VNode on every render; its component type is stable.
    if (entry?.source === vnode.type && entry.name === name) return entry.component

    const component = markRaw(defineComponent({
      name,
      inheritAttrs: false,
      props: { routeNode: { type: Object as PropType<VNode>, required: true } },
      setup(props) {
        // Rebuild from current route inputs. Never reuse a mounted VNode's
        // component instance, parent or KeepAlive flags inside this wrapper.
        return () => createVNode(props.routeNode.type, props.routeNode.props, props.routeNode.children)
      },
    }))
    entries.set(route.fullPath, { name, source: vnode.type, component })
    return component
  }

  return {
    resolve,
    remove: (fullPath: string) => { entries.delete(fullPath) },
    prune: (activePaths: ReadonlySet<string>) => {
      for (const path of entries.keys()) if (!activePaths.has(path)) entries.delete(path)
    },
  }
}
