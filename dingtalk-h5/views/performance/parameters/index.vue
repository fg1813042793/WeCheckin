<script>
import { h, ref } from 'vue'
import { usePerformanceContext } from '../common/context'

function readInputValue(event) {
  return event.detail?.value ?? event.target.value
}

function emptyTemplate() {
  return {
    objectiveDefaults: [],
    nextObjectiveDefaults: [],
    gradeLevels: [],
    values: []
  }
}

function cloneTemplate(template) {
  const source = template || emptyTemplate()
  return JSON.parse(JSON.stringify({
    objectiveDefaults: source.objectiveDefaults || [],
    nextObjectiveDefaults: source.nextObjectiveDefaults || [],
    gradeLevels: source.gradeLevels || [],
    values: source.values || []
  }))
}

function templatePanel(title, items, renderItem, actions = []) {
  return h('section', { class: 'panel template-panel' }, [
    h('view', { class: 'panel-head' }, [
      h('text', { class: 'panel-title' }, title),
      actions.length > 0 ? h('view', { class: 'template-panel-actions' }, actions) : null
    ]),
    h('view', { class: 'template-list' }, [
      items.length > 0
        ? items.map((item, index) => h('view', { class: 'template-row', key: `${title}-${index}` }, renderItem(item, index)))
        : h('view', { class: 'template-empty' }, '暂无模板项')
    ])
  ])
}

function addObjective(template, field, prefix) {
  template[field].push({ id: `${prefix}-${Date.now()}`, target: '', weight: 0 })
}

function addGrade(template) {
  template.gradeLevels.push({ label: '', grade: '', coefficient: 1 })
}

function addValue(template) {
  template.values.push({
    id: `value-${Date.now()}`,
    name: '',
    definition: '',
    rubric: [
      { label: '卓越', score: 50, description: '持续超出要求，对团队或业务产生明显正向影响' },
      { label: '优秀', score: 40, description: '高质量完成要求，表现稳定且有主动贡献' },
      { label: '良好', score: 30, description: '符合岗位要求，能够稳定完成相关表现' },
      { label: '及格', score: 20, description: '基本达到要求，但仍有明显提升空间' },
      { label: '较差', score: 10, description: '未达到要求，需要重点改进' }
    ]
  })
}

function removeItem(items, index) {
  items.splice(index, 1)
}

function renderTemplateDeleteButton(onClick) {
  return h('button', { class: 'dt-btn dt-btn-danger-light small template-delete-btn', onClick }, '删除')
}

function renderObjectiveItem(item, index, editing, onRemove) {
  if (!editing) {
    return h('view', { class: 'template-read-row' }, [
      h('text', { class: 'template-row-main' }, item.target),
      h('text', { class: 'template-weight-badge' }, `${item.weight || 0}%`)
    ])
  }
  return h('view', { class: 'template-edit-row' }, [
    h('textarea', {
      class: 'template-editor-textarea',
      value: item.target,
      placeholder: '目标描述',
      onInput: (event) => { item.target = readInputValue(event) }
    }),
    h('view', { class: 'template-inline-fields' }, [
      h('input', {
        class: 'template-editor-input',
        type: 'number',
        value: item.weight,
        placeholder: '权重%',
        onInput: (event) => { item.weight = Number(readInputValue(event)) }
      }),
      renderTemplateDeleteButton(onRemove)
    ])
  ])
}

function renderGradeItem(item, index, editing, onRemove) {
  if (!editing) {
    return h('view', { class: 'template-read-row' }, [
      h('text', { class: 'template-row-main' }, `${item.label} · ${item.grade}`),
      h('text', { class: 'template-weight-badge' }, String(item.coefficient))
    ])
  }
  return h('view', { class: 'template-grade-edit-row' }, [
    h('input', {
      class: 'template-editor-input',
      value: item.label,
      placeholder: '等级标签',
      onInput: (event) => { item.label = readInputValue(event) }
    }),
    h('input', {
      class: 'template-editor-input',
      value: item.grade,
      placeholder: '档位',
      onInput: (event) => { item.grade = readInputValue(event) }
    }),
    h('input', {
      class: 'template-editor-input',
      type: 'number',
      value: item.coefficient,
      placeholder: '系数',
      onInput: (event) => { item.coefficient = Number(readInputValue(event)) }
    }),
    renderTemplateDeleteButton(onRemove)
  ])
}

function renderValueItem(item, index, editing, onRemove) {
  if (!editing) {
    const rubrics = Array.isArray(item.rubric) ? item.rubric.filter((rubric) => rubric?.label) : []
    return h('view', { class: 'template-read-row value' }, [
      h('text', { class: 'template-row-main' }, item.name),
      h('text', { class: 'template-row-desc' }, item.definition),
      rubrics.length ? h('view', { class: 'template-rubric-preview' }, rubrics.map((rubric) =>
        h('text', { class: 'template-rubric-preview-item' }, `${rubric.score || 0}分 · ${rubric.label}${rubric.description ? `：${rubric.description}` : ''}`)
      )) : null
    ])
  }
  return h('view', { class: 'template-edit-row value' }, [
    h('view', { class: 'template-inline-fields' }, [
      h('input', {
        class: 'template-editor-input',
        value: item.name,
        placeholder: '价值观名称',
        onInput: (event) => { item.name = readInputValue(event) }
      }),
      renderTemplateDeleteButton(onRemove)
    ]),
    h('textarea', {
      class: 'template-editor-textarea',
      value: item.definition,
      placeholder: '价值观定义',
      onInput: (event) => { item.definition = readInputValue(event) }
    }),
    h('view', { class: 'template-rubric-list' }, [
      ...(item.rubric || []).map((rubric, rubricIndex) => h('view', { class: 'template-rubric-row', key: `rubric-${rubricIndex}` }, [
        h('input', {
          class: 'template-editor-input',
          value: rubric.label,
          placeholder: '评分名称',
          onInput: (event) => { rubric.label = readInputValue(event) }
        }),
        h('input', {
          class: 'template-editor-input',
          type: 'number',
          value: rubric.score,
          placeholder: '分值',
          onInput: (event) => { rubric.score = Number(readInputValue(event)) }
        }),
        h('textarea', {
          class: 'template-editor-textarea template-rubric-description',
          value: rubric.description || '',
          placeholder: '评分说明',
          onInput: (event) => { rubric.description = readInputValue(event) }
        }),
        renderTemplateDeleteButton(() => removeItem(item.rubric, rubricIndex))
      ])),
      h('button', {
        class: 'dt-btn dt-btn-light small',
        onClick: () => {
          if (!item.rubric) item.rubric = []
          item.rubric.push({ label: '', score: 0, description: '' })
        }
      }, '添加评分')
    ])
  ])
}

export default {
  name: 'TemplatePage',
  setup() {
    const ctx = usePerformanceContext()
    const editing = ref(false)
    const saving = ref(false)
    const draft = ref(cloneTemplate(ctx.state.template))
    const canEdit = () => ctx.canEditTemplate()

    function startEdit() {
      draft.value = cloneTemplate(ctx.state.template)
      editing.value = true
    }

    function cancelEdit() {
      draft.value = cloneTemplate(ctx.state.template)
      editing.value = false
    }

    async function submitTemplate() {
      if (saving.value) return
      saving.value = true
      try {
        await ctx.saveTemplate(draft.value)
        editing.value = false
      } finally {
        saving.value = false
      }
    }

    return () => {
      const template = editing.value ? draft.value : cloneTemplate(ctx.state.template)
      return h('view', { class: 'template-page' }, [
        h('view', { class: 'page-head' }, [
          h('view', { class: 'template-page-head-copy' }, [
            h('text', { class: 'page-title' }, ctx.sectionTitle.value),
            h('text', { class: 'page-desc' }, '目标模板、价值观标尺和绩效工资系数')
          ]),
          h('view', { class: 'template-toolbar' }, [
            canEdit() && !editing.value ? h('button', { class: 'dt-btn dt-btn-primary', onClick: startEdit }, '编辑') : null,
            !editing.value ? h('button', { class: 'dt-btn dt-btn-light', onClick: ctx.refreshData }, '刷新') : null,
            editing.value ? h('button', { class: 'dt-btn dt-btn-light', onClick: cancelEdit }, '取消') : null,
            editing.value ? h('button', { class: 'dt-btn dt-btn-primary', loading: saving.value, onClick: submitTemplate }, '保存') : null
          ])
        ]),
        h('view', { class: ['template-grid', editing.value ? 'editing' : ''] }, [
          templatePanel('默认目标', template.objectiveDefaults, (item, index) => renderObjectiveItem(item, index, editing.value, () => removeItem(template.objectiveDefaults, index)), editing.value ? [
            h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addObjective(template, 'objectiveDefaults', 'objective') }, '添加')
          ] : []),
          templatePanel('下月目标', template.nextObjectiveDefaults, (item, index) => renderObjectiveItem(item, index, editing.value, () => removeItem(template.nextObjectiveDefaults, index)), editing.value ? [
            h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addObjective(template, 'nextObjectiveDefaults', 'next') }, '添加')
          ] : []),
          templatePanel('绩效工资系数', template.gradeLevels, (item, index) => renderGradeItem(item, index, editing.value, () => removeItem(template.gradeLevels, index)), editing.value ? [
            h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addGrade(template) }, '添加')
          ] : []),
          templatePanel('价值观', template.values, (item, index) => renderValueItem(item, index, editing.value, () => removeItem(template.values, index)), editing.value ? [
            h('button', { class: 'dt-btn dt-btn-light small', onClick: () => addValue(template) }, '添加')
          ] : [])
        ])
      ])
    }
  }
}
</script>

<style>
/* 页面专属样式：从 styles/performance.css 拆分 */
.template-list {
  display: grid;
  gap: 10px;
  padding: 14px;
}

.template-row {
  padding: 14px;
  border: 1px solid #edf0f5;
  border-radius: 4px;
  background: #fff;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.template-grid.editing {
  grid-template-columns: 1fr;
}

.template-grid.editing .template-panel {
  min-width: 0;
}

.template-toolbar,
.template-panel-actions,
.template-read-row {
  display: flex;
  align-items: center;
}

.template-toolbar,
.template-panel-actions {
  gap: 8px;
}

.template-panel .panel-head {
  min-height: 46px;
}

.template-read-row {
  min-width: 0;
  justify-content: space-between;
  gap: 12px;
}

.template-read-row.value {
  align-items: flex-start;
  display: grid;
  gap: 6px;
}

.template-rubric-preview {
  display: grid;
  gap: 4px;
}

.template-rubric-preview-item {
  color: #4e5969;
  font-size: 12px;
  line-height: 1.5;
}

.template-row-main {
  min-width: 0;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.45;
}

.template-row-desc {
  color: #4e5969;
  font-size: 13px;
  line-height: 1.55;
}

.template-weight-badge {
  flex: 0 0 auto;
  min-width: 42px;
  height: 24px;
  padding: 0 8px;
  border-radius: 999px;
  background: #f2f7ff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 800;
  line-height: 24px;
  text-align: center;
}

.template-edit-row {
  display: grid;
  gap: 10px;
}

.template-edit-row.value {
  gap: 12px;
}

.template-inline-fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  min-width: 0;
  gap: 8px;
}

.template-grade-edit-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(70px, 90px) minmax(70px, 90px) auto;
  gap: 8px;
  align-items: center;
}

.template-editor-textarea,
.template-editor-input {
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
  border: 1px solid #dfe6f2;
  border-radius: 4px;
  background: #fff;
  color: #1f2329;
  font-size: 13px;
}

.template-editor-textarea {
  min-height: 74px;
  padding: 10px;
  line-height: 1.55;
}

.template-editor-input {
  height: 34px;
  min-height: 34px;
  padding: 0 10px;
  line-height: 34px;
}

.template-rubric-list {
  display: grid;
  gap: 8px;
  padding-top: 2px;
}

.template-rubric-row {
  display: grid;
  grid-template-columns: minmax(120px, 160px) minmax(70px, 88px) minmax(260px, 1fr) auto;
  gap: 8px;
  align-items: start;
}

.template-rubric-description {
  min-height: 56px;
  line-height: 1.45;
  resize: vertical;
}

.template-delete-btn {
  min-width: 52px;
  height: 32px;
  min-height: 32px;
  padding: 0 12px;
  border-radius: 6px;
  line-height: 30px;
  white-space: nowrap;
  flex-shrink: 0;
}

.template-empty {
  padding: 20px 12px;
  color: #86909c;
  font-size: 13px;
  text-align: center;
}

.template-page {
  width: 100%;
  max-width: 1480px;
  margin: 0 auto;
  display: grid;
  gap: 16px;
}

@media (max-width: 960px) {
  .template-page {
    max-width: none;
    padding: 16px;
  }

  .template-page-head-copy {
    display: none;
  }

  .template-page > .page-head {
    justify-content: flex-start;
    margin-bottom: 12px;
  }

  .template-row {
    padding: 12px;
  }

  .template-toolbar {
    width: 100%;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .template-grid,
  .template-grid.editing {
    grid-template-columns: 1fr;
  }

  .template-inline-fields,
  .template-rubric-row,
  .template-grade-edit-row {
    align-items: center;
  }

  .template-inline-fields {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .template-grade-edit-row {
    grid-template-columns: minmax(0, 1fr) minmax(70px, 86px) auto;
  }

  .template-grade-edit-row .template-editor-input:first-child {
    grid-column: 1 / -1;
  }

  .template-rubric-row {
    grid-template-columns: 1fr;
  }

  .template-rubric-row .template-delete-btn {
    width: 100%;
  }
}
</style>
