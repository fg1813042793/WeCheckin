import { h } from 'vue'
import { resolveMenuPageComponent } from '../../router/index'

function renderLoading(sectionTitle) {
  return h('view', { class: 'content-loading-view' }, [
    h('view', { class: 'page-head' }, [
      h('view', {}, [
        h('text', { class: 'page-title' }, sectionTitle),
        h('text', { class: 'page-desc' }, '加载中...')
      ])
    ]),
    h('section', { class: 'panel performance-loading-panel' }, [
      h('view', { class: 'performance-loading-head' }, [
        h('view', { class: 'performance-loading-title' }),
        h('view', { class: 'performance-loading-action' })
      ]),
      h('view', { class: 'performance-loading-table' }, Array.from({ length: 6 }, (_, rowIndex) =>
        h('view', { class: 'performance-loading-row', key: `loading-row-${rowIndex}` }, Array.from({ length: 6 }, (_, cellIndex) =>
          h('view', {
            key: `loading-cell-${rowIndex}-${cellIndex}`,
            class: [
              'performance-loading-cell',
              cellIndex === 0 ? 'wide' : '',
              rowIndex === 0 ? 'head' : ''
            ]
          })
        ))
      ))
    ])
  ])
}

export default {
  name: 'AppContentOutlet',
  props: {
    contentLoading: { type: Boolean, default: false },
    contentView: { type: String, default: 'dashboard' },
    navItemsLength: { type: Number, default: 0 },
    sectionTitle: { type: String, default: '' }
  },
  setup(props) {
    return () => {
      if (props.navItemsLength === 0) {
        return h('view', { class: 'empty no-permission' }, '暂无可用菜单，请联系管理员配置钉钉 H5 权限')
      }
      if (props.contentLoading) {
        return renderLoading(props.sectionTitle)
      }
      const Page = resolveMenuPageComponent(props.contentView)
      return h(Page)
    }
  }
}
