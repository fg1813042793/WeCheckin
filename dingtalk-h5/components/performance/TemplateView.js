import { h } from 'vue'
import { usePerformanceContext } from './context'

function templatePanel(title, items, renderItem) {
  return h('section', { class: 'panel' }, [
    h('view', { class: 'panel-head' }, [h('text', { class: 'panel-title' }, title)]),
    h('view', { class: 'template-list' }, items.map((item) => h('view', { class: 'template-row' }, renderItem(item))))
  ])
}

export default {
  name: 'TemplateView',
  setup() {
    const ctx = usePerformanceContext()

    return () => h('view', { class: 'template-page' }, [
      h('view', { class: 'page-head' }, [
        h('view', {}, [
          h('text', { class: 'page-title' }, '绩效模版'),
          h('text', { class: 'page-desc' }, '目标模板、价值观标尺和绩效工资系数')
        ]),
        h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新')
      ]),
      h('view', { class: 'template-grid' }, [
        templatePanel('默认目标', ctx.state.template?.objectiveDefaults || [], (item) => `${item.target} · ${item.weight}%`),
        templatePanel('下月目标', ctx.state.template?.nextObjectiveDefaults || [], (item) => `${item.target} · ${item.weight}%`),
        templatePanel('绩效工资系数', ctx.state.template?.gradeLevels || [], (item) => `${item.label} · ${item.grade} · ${item.coefficient}`),
        templatePanel('价值观', ctx.state.template?.values || [], (item) => `${item.name} · ${item.definition}`)
      ])
    ])
  }
}
