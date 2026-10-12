import { defineComponent, h, type App, type Component } from 'vue'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const elementIconMap = Object.fromEntries(
  Object.entries(ElementPlusIconsVue).filter(([name]) => name !== 'default')
) as Record<string, Component>

type SvgNode = {
  tag: 'path' | 'circle' | 'rect' | 'ellipse' | 'line' | 'polyline'
  attrs: Record<string, string | number>
}

function createAdminSvgIcon(nodes: SvgNode[]): Component {
  return defineComponent({
    inheritAttrs: false,
    setup(_, { attrs }) {
      return () => h('svg', {
        ...attrs,
        viewBox: '0 0 24 24',
        fill: 'none',
        stroke: 'currentColor',
        strokeWidth: 1.8,
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        xmlns: 'http://www.w3.org/2000/svg',
      }, nodes.map(node => h(node.tag, node.attrs)))
    },
  })
}

const customAdminIconMap: Record<string, Component> = {
  AdminDashboard: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 3, y: 3, width: 7, height: 7, rx: 1 } },
    { tag: 'rect', attrs: { x: 14, y: 3, width: 7, height: 7, rx: 1 } },
    { tag: 'rect', attrs: { x: 3, y: 14, width: 7, height: 7, rx: 1 } },
    { tag: 'rect', attrs: { x: 14, y: 14, width: 7, height: 7, rx: 1 } },
  ]),
  AdminAnalytics: createAdminSvgIcon([
    { tag: 'line', attrs: { x1: 4, y1: 20, x2: 4, y2: 11 } },
    { tag: 'line', attrs: { x1: 10, y1: 20, x2: 10, y2: 5 } },
    { tag: 'line', attrs: { x1: 16, y1: 20, x2: 16, y2: 14 } },
    { tag: 'line', attrs: { x1: 22, y1: 20, x2: 22, y2: 8 } },
  ]),
  AdminUsers: createAdminSvgIcon([
    { tag: 'circle', attrs: { cx: 9, cy: 8, r: 3 } },
    { tag: 'path', attrs: { d: 'M3 20c0-3.3 2.7-6 6-6s6 2.7 6 6' } },
    { tag: 'path', attrs: { d: 'M16 5.5a3 3 0 0 1 0 5.8M18 14.5a5.5 5.5 0 0 1 3 5.5' } },
  ]),
  AdminTeam: createAdminSvgIcon([
    { tag: 'circle', attrs: { cx: 8, cy: 8, r: 3 } },
    { tag: 'circle', attrs: { cx: 17, cy: 9, r: 2.5 } },
    { tag: 'path', attrs: { d: 'M2.5 20c.3-3.5 2.4-5.5 5.5-5.5s5.2 2 5.5 5.5M14 15c2.8-.5 5.2 1.2 6 5' } },
  ]),
  AdminShield: createAdminSvgIcon([
    { tag: 'path', attrs: { d: 'M12 3l8 3v5c0 4.9-3.4 8.3-8 10-4.6-1.7-8-5.1-8-10V6l8-3z' } },
    { tag: 'path', attrs: { d: 'm8.5 12 2.2 2.2 4.8-5' } },
  ]),
  AdminKey: createAdminSvgIcon([
    { tag: 'circle', attrs: { cx: 8, cy: 15, r: 4 } },
    { tag: 'path', attrs: { d: 'm11 12 9-9M16 6l2 2M13.5 9.5l2 2' } },
  ]),
  AdminCalendar: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 3, y: 5, width: 18, height: 16, rx: 2 } },
    { tag: 'line', attrs: { x1: 3, y1: 10, x2: 21, y2: 10 } },
    { tag: 'line', attrs: { x1: 8, y1: 3, x2: 8, y2: 7 } },
    { tag: 'line', attrs: { x1: 16, y1: 3, x2: 16, y2: 7 } },
  ]),
  AdminWorkflow: createAdminSvgIcon([
    { tag: 'circle', attrs: { cx: 6, cy: 6, r: 2.5 } },
    { tag: 'circle', attrs: { cx: 18, cy: 6, r: 2.5 } },
    { tag: 'circle', attrs: { cx: 12, cy: 18, r: 2.5 } },
    { tag: 'path', attrs: { d: 'M8.5 6h7M7.5 8l3 7.5M16.5 8l-3 7.5' } },
  ]),
  AdminDatabase: createAdminSvgIcon([
    { tag: 'ellipse', attrs: { cx: 12, cy: 5, rx: 8, ry: 3 } },
    { tag: 'path', attrs: { d: 'M4 5v7c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12v7c0 1.7 3.6 3 8 3s8-1.3 8-3v-7' } },
  ]),
  AdminCloud: createAdminSvgIcon([
    { tag: 'path', attrs: { d: 'M7 19h10a4 4 0 0 0 .6-8A6 6 0 0 0 6 9a5 5 0 0 0 1 10z' } },
  ]),
  AdminApi: createAdminSvgIcon([
    { tag: 'polyline', attrs: { points: '8 4 3 12 8 20' } },
    { tag: 'polyline', attrs: { points: '16 4 21 12 16 20' } },
    { tag: 'line', attrs: { x1: 14, y1: 3, x2: 10, y2: 21 } },
  ]),
  AdminSurvey: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 5, y: 3, width: 14, height: 18, rx: 2 } },
    { tag: 'line', attrs: { x1: 8, y1: 8, x2: 16, y2: 8 } },
    { tag: 'line', attrs: { x1: 8, y1: 12, x2: 16, y2: 12 } },
    { tag: 'line', attrs: { x1: 8, y1: 16, x2: 13, y2: 16 } },
  ]),
  AdminExam: createAdminSvgIcon([
    { tag: 'path', attrs: { d: 'M6 3h12a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z' } },
    { tag: 'path', attrs: { d: 'm8 11 2 2 5-5M8 17h8' } },
  ]),
  AdminWebhook: createAdminSvgIcon([
    { tag: 'path', attrs: { d: 'M8 8a4 4 0 1 1 4 4H7a4 4 0 1 0 4 4' } },
    { tag: 'path', attrs: { d: 'M16 16a4 4 0 1 1-4-4h5a4 4 0 1 0-4-4' } },
  ]),
  AdminMail: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 3, y: 5, width: 18, height: 14, rx: 2 } },
    { tag: 'polyline', attrs: { points: '3 7 12 13 21 7' } },
  ]),
  AdminChat: createAdminSvgIcon([
    { tag: 'path', attrs: { d: 'M4 5h16a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H9l-5 3v-3a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2z' } },
    { tag: 'line', attrs: { x1: 7, y1: 10, x2: 17, y2: 10 } },
    { tag: 'line', attrs: { x1: 7, y1: 14, x2: 14, y2: 14 } },
  ]),
  AdminImage: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 3, y: 4, width: 18, height: 16, rx: 2 } },
    { tag: 'circle', attrs: { cx: 8, cy: 9, r: 1.5 } },
    { tag: 'path', attrs: { d: 'm4 17 5-5 3 3 2-2 6 6' } },
  ]),
  AdminReport: createAdminSvgIcon([
    { tag: 'path', attrs: { d: 'M5 3h10l4 4v14H5z' } },
    { tag: 'path', attrs: { d: 'M15 3v5h4M8 16v-3M12 16V9M16 16v-5' } },
  ]),
  AdminServer: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 4, y: 3, width: 16, height: 7, rx: 2 } },
    { tag: 'rect', attrs: { x: 4, y: 14, width: 16, height: 7, rx: 2 } },
    { tag: 'line', attrs: { x1: 8, y1: 6.5, x2: 8, y2: 6.5 } },
    { tag: 'line', attrs: { x1: 8, y1: 17.5, x2: 8, y2: 17.5 } },
  ]),
  AdminTask: createAdminSvgIcon([
    { tag: 'rect', attrs: { x: 4, y: 4, width: 16, height: 16, rx: 2 } },
    { tag: 'path', attrs: { d: 'm8 12 2.5 2.5L16 9' } },
  ]),
  AdminCode: createAdminSvgIcon([
    { tag: 'polyline', attrs: { points: '8 6 2 12 8 18' } },
    { tag: 'polyline', attrs: { points: '16 6 22 12 16 18' } },
    { tag: 'line', attrs: { x1: 14, y1: 4, x2: 10, y2: 20 } },
  ]),
}

export const adminIconMap = { ...elementIconMap, ...customAdminIconMap }

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
