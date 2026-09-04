import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import process from 'node:process'
import ts from 'typescript'

const root = process.cwd()

const requiredFiles = [
  'src/api/workflow.ts',
  'src/types/workflow.ts',
  'src/pages/workflow/workflow.menu.ts',
  'src/pages/workflow/workflow-route-keys.ts',
  'src/pages/workflow/workflow.routes.ts',
  'src/pages/workflow/components/WorkflowCenter.vue',
  'src/pages/workflow/components/WorkflowFilterPanel.vue',
  'src/pages/workflow/components/WorkflowRecordTable.vue',
  'src/pages/workflow/components/WorkflowInstancePage.vue',
  'src/pages/workflow/components/WorkflowTaskPage.vue',
  'src/pages/workflow/components/WorkflowStartPage.vue',
  'src/pages/workflow/components/WorkflowSummaryPage.vue',
  'src/pages/workflow/components/WorkflowSummarySection.vue',
  'src/pages/workflow/components/WorkflowHistoryDatePicker.vue',
  'src/pages/workflow/components/WorkflowReadOnlyGraph.vue',
  'src/pages/workflow/workflow-graph-layout.ts',
  'src/pages/workflow/workflow-status.ts',
  'src/pages/workflow/components/WorkflowRuntimeForm.vue',
  'src/pages/workflow/components/WorkflowFieldControl.vue',
  'src/pages/workflow/components/WorkflowAttachmentControl.vue',
  'src/pages/workflow/components/WorkflowImagePicker.vue',
  'src/pages/workflow/components/WorkflowParticipantSelect.vue',
  'src/pages/workflow/components/WorkflowTextarea.vue',
  'src/pages/workflow/components/WorkflowDetailPanel.vue',
  'src/pages/workflow/components/WorkflowNodeProgressList.vue',
  'src/pages/performance/components/PerformanceWorkbench.vue',
  'src/pages/workflow/workflow-form.ts',
  'src/pages/workflow/workflow-select-placement.ts',
  'src/pages/workflow/workflow-history-filter.ts',
  'src/pages/workflow/workflow-task.ts',
]

const requiredContent = [
  {
    file: 'src/common/style.scss',
    patterns: ['--app-pc-content-max-width: 1080px;'],
  },
  {
    file: 'src/config/app-navigation.ts',
    patterns: ['workflowMenuPages', 'workflowRootNavItem'],
  },
  {
    file: 'src/config/app-content-routes.ts',
    patterns: ['workflowContentRoutes', 'resolveWorkflowContentComponent'],
  },
  {
    file: 'src/stores/appContent.ts',
    patterns: ['dynamicTabs', 'openDynamicTab', 'registerTabCloseGuard', 'requestCloseTab', 'focusWorkflowInstance', 'focusedWorkflowTab', 'focusWorkflowTab', 'clearFocusedWorkflowTab', 'WorkflowStartSeed', 'setWorkflowStartSeed', 'takeWorkflowStartSeed'],
  },
  {
    file: 'src/components/app-shell/app-shell.vue',
    patterns: ['dynamicTabs', 'app-shell__content--dynamic', 'workflow-tab-close-prompt', '保存草稿', '不保存', '继续填写', '@wheel="scrollTabsHorizontally"', 'touch-action: pan-x;', '.app-shell__tab {\n  position: relative;\n  flex: 0 0 auto;'],
  },
  {
    file: 'src/pages/index/index.vue',
    patterns: ['dynamicTabs', ':content-key="tab.key"'],
  },
  {
    file: 'src/common/http.interceptor.ts',
    patterns: ['requestUrl.startsWith(\'/api/\')'],
  },
  {
    file: 'src/api/workflow.ts',
    patterns: [
      '/api/v2/dingtalk/h5/workflows',
      'listWorkflowCategories',
      'listWorkflowDefinitions',
      'startWorkflowInstance',
      'getWorkflowStartDraft',
      'saveWorkflowStartDraft',
      'deleteWorkflowStartDraft',
      'deleteWorkflowInstance',
      'commentWorkflowInstance',
      'remindWorkflowInstance',
      'listWorkflowInstances',
      'getWorkflowInstance',
      'withdrawWorkflowInstance',
      'listWorkflowTasks',
      'completeWorkflowTask',
      'uploadWorkflowAttachment',
      'uploadFile',
      'commentWorkflowInstance',
    ],
  },
  {
    file: 'src/types/workflow.ts',
    patterns: ['logoUrl?: string', 'minVisibleRows?: number', 'maxVisibleRows?: number', 'export interface WorkflowAttachment', 'mimeType: string', 'size: number', 'starterName: string', 'starterName?: string', 'assigneeName: string', 'handledByName: string', 'approvalChainKey?: string', 'approvalLayer?: number', 'approvalLayerTotal?: number', 'sourceDepartmentName?: string', 'currentNodeNames: string[]', 'currentAssigneeNames: string[]', 'WorkflowNodeProgressStatus', 'WorkflowNodeProgressSummary', 'nodeProgress?: WorkflowNodeProgressSummary[]', 'WorkflowReminderPolicy', 'WorkflowReminderNode', 'reminderPolicy: WorkflowReminderPolicy', 'reminderNodes: WorkflowReminderNode[]', 'nodes?: WorkflowPublishedNode[]', 'edges?: WorkflowPublishedEdge[]', 'userNames: Record<string, string>', 'WorkflowCommentNotificationRequest', 'WorkflowNotificationChannel', 'notification?: WorkflowCommentNotificationRequest', 'definitionCategory?: string', 'startTimeFrom?: number', 'startTimeTo?: number', 'endTimeFrom?: number', 'endTimeTo?: number'],
  },
  {
    file: 'src/pages/workflow/workflow-form.ts',
    patterns: ['normalizeWorkflowAttachments', 'field.type === \'attachment\''],
  },
  {
    file: 'src/pages/workflow/components/WorkflowFieldControl.vue',
    patterns: ['WorkflowAttachmentControl', 'field.type === \'attachment\''],
  },
  {
    file: 'src/pages/workflow/components/WorkflowAttachmentControl.vue',
    patterns: ['uni.chooseFile', 'uni.previewImage', 'uni.downloadFile', 'uni.openDocument', 'window.open', 'removeAttachment', 'downloadAttachment'],
  },
  {
    file: 'src/types/dingtalk-h5.ts',
    patterns: ['workflowActorId?: string'],
  },
  {
    file: 'src/pages/workflow/workflow.menu.ts',
    patterns: ['workflowRootNavItem', 'key: \'workflow\'', 'title: \'流程审批\''],
  },
  {
    file: 'src/pages/workflow/workflow-status.ts',
    patterns: ['WorkflowStatusMeta', 'workflowInstanceStatusMeta', 'workflowTaskStatusMeta', 'workflowNodeProgressStatusMeta', 'running: { label: \'审批中\', type: \'warning\' }', 'waiting: { label: \'待激活\', type: \'warning\' }', 'processing: { label: \'处理中\', type: \'warning\' }', 'completed: { label: \'已完成\', type: \'success\' }', 'rejected: { label: \'已驳回\', type: \'error\' }', 'cancelled: { label: \'已取消\', type: \'info\' }'],
  },
  {
    file: 'src/pages/workflow/workflow-route-keys.ts',
    patterns: ['workflowStartContentKey', 'workflowDefinitionIdFromContentKey', 'workflowInstanceContentKey', 'workflowInstanceIdFromContentKey', 'workflowTaskContentKey', 'workflowTaskIdFromContentKey', 'workflowTaskInstanceIdFromContentKey'],
  },
  {
    file: 'src/pages/workflow/workflow.routes.ts',
    patterns: ['WorkflowInstancePage', 'WorkflowTaskPage', './workflow-route-keys', 'workflowInstanceContentKey', 'workflowInstanceIdFromContentKey', 'workflowTaskContentKey', 'workflowTaskIdFromContentKey', 'workflowTaskInstanceIdFromContentKey', 'return WorkflowInstancePage', 'return WorkflowTaskPage'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: ['发起审批', '我的待办', '已处理', '我的申请', '抄送我的', '汇总', 'WorkflowSummaryPage', 'activeTab === \'summary\'', 'definitionStartMeta(definition)', '当前周期剩余', '当前周期已达上限', 'openWorkflowStartTab', 'openWorkflowTaskTab', 'openWorkflowInstanceTab', 'workflowInstanceContentKey', 'workflowTaskContentKey', 'resolveMobilePage', 'mobileHidden', 'focusedWorkflowInstanceId', 'focusedWorkflowTab', 'openFocusedWorkflowTab', 'clearFocusedWorkflowTab', 'dingtalk_h5:api:workflow:view', 'dingtalk_h5:api:workflow:start', 'definitionCategory', 'definition.logoUrl', '<image', '@error', 'markDefinitionLogoFailed', 'WorkflowFilterPanel', ':active-count="catalogFilterCount"', 'class="workflow-center__catalog-filters"', ':active-count="applicationFilterCount"', 'WorkflowHistoryDatePicker', 'WorkflowRecordTable', 'recordColumns', 'recordRows', 'openRecord', 'WorkflowRecordFilters', 'recordFilters', 'appliedRecordFilters', 'activeRecordFilters', 'activeAppliedRecordFilters', 'showStarterNameFilter', 'showStatusFilter', 'historyStatusOptions', 'listWorkflowCategories', 'workflowCategories', 'recordCategoryOptions', 'queryRecords', 'resetRecordFilters', 'placeholder="输入发起人用户名"', ':maxlength="50"', 'buildWorkflowHistoryTimeQuery', 'definitionCategory:', 'class="workflow-center__record-filters"', 'class="workflow-center__filter-input"', 'class="workflow-center__filter-select"', ':columns="recordColumns"', ':rows="recordRows"', '#actions="{ row }"', 'activeTab === \'pending\' ? \'办理\' : \'查看\'', 'activeTab === \'pending\'', 'activeTab.value === \'handled\'', 'activeTab.value === \'copied\'', 'name="eye"', 'presentation="history-drawer"', ':application-actions="activeTab === \'started\'"', 'showStarterColumn', '[\'pending\', \'handled\', \'copied\'].includes(activeTab.value)', 'key: \'starterName\'', 'label: \'发起人\'', 'starterDisplayName(instance)', 'key: \'currentNode\'', 'label: \'当前节点\'', 'key: \'currentAssignees\'', 'label: \'节点处理人\'', 'currentNodeDisplay(instance)', 'currentAssigneeDisplay(instance)', '流程分类', '提交时间', '审批状态', '查看', '查询', '重置'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowSummaryPage.vue',
    patterns: ['listWorkflowSummaryDefinitions', 'WorkflowSummarySection', ':definitions="definitions"'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowSummarySection.vue',
    patterns: ['definitionName', '流程名称', 'placeholder="输入流程名称"', 'instance.definitionName', 'instance.definitionId', '请选择同一流程的记录批量导出', ':definitions="definitions"', '.workflow-summary__filter-actions {\n  grid-column: 11 / 13;\n  grid-row: 3;', 'width: fit-content;'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: [
      'taskStarterDisplayName',
      'task.definitionName',
      'task.starterName',
      'size="16px"',
      'size="14px"',
      'size="32px"',
      'custom-class="workflow-center__search-control"',
      'custom-class="workflow-center__loading-icon"',
      'custom-class="workflow-center__pagination-control"',
      ':custom-style="workflowLoadingStyle"',
      'font-size: 12px;',
      'line-height: 34px;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: ['function defaultWorkflowTab()', 'if (canStart.value)', 'return \'start\'', 'if (canSummary.value)', 'return \'summary\'', 'return \'pending\''],
  },
  {
    file: 'src/pages/workflow/components/WorkflowInstancePage.vue',
    patterns: ['presentation="history-page"', 'comment-action'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    patterns: ['dingtalk_h5:api:workflow:remind', 'reminderNodes', 'remindWorkflowInstance', '提醒处理', '今日剩余', 'rejectVisible', 'rejectImages', 'commentImages', 'WorkflowImagePicker', 'WorkflowParticipantSelect', 'commentNotificationGroups', 'commentNotificationUserIds', 'commentNotificationChannels = ref<WorkflowNotificationChannel[]>([\'in_app\'])', ':groups="commentNotificationGroups"', 'value="in_app" label="站内信"', 'value="dingtalk_oa" label="钉钉"', 'notification: commentNotificationUserIds.value.length > 0', 'right.eventTime - left.eventTime'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowParticipantSelect.vue',
    patterns: ['modelValue: string[]', 'selectedLabel', 'toggleDesktopOption', 'workflow-participant-select__panel', 'role="option"', 'u-popup', '确认选择'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowImagePicker.vue',
    patterns: ['uni.chooseImage', 'uploadWorkflowAttachment', 'uni.previewImage', 'maxCount', '超过20MB'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowTaskPage.vue',
    patterns: ['presentation="page"', 'comment-action'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: ['presentation="history-drawer"', 'comment-action'],
  },
  {
    file: 'src/pages/performance/components/PerformanceWorkbench.vue',
    patterns: [
      'listWorkflowTasks',
      'workflowPendingCount',
      'workflowTasks',
      'canViewWorkflowTodos',
      'openWorkflowTodos',
      'workflowTodoTitle',
      'workflowTodoStarter',
      'workflowTodoMeta',
      'visibleWorkflowTasks',
      'v-for="task in visibleWorkflowTasks"',
      'task.definitionName',
      'task.starterName',
      'task.nodeName',
      'appContent.focusWorkflowTab(\'pending\')',
      'appContent.switchContent(\'workflow\')',
      '发起人：',
      '当前节点：',
      '处理',
      'custom-class="workbench__total-tag"',
      'size="18px"',
      'size="20px"',
      'size="14px"',
      'height: 24px;',
      'height: 34px;',
      'font-size: 12px;',
      '@media (max-width: 768px) {',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowFilterPanel.vue',
    patterns: ['const expanded = ref(!resolveMobilePresentation())', 'activeCount?: number', '筛选条件', '已选 {{ activeCount }}', 'workflow-filter-panel__toggle', 'workflow-filter-panel__content', 'name="arrow-up"', 'name="arrow-down"', 'size="14px"', 'v-show="expanded"', 'width <= 768'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: ['workflowInstanceStatusMeta', 'workflowTaskStatusMeta'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: [
      'const showStatusFilter = computed(() => activeTab.value === \'started\' || activeTab.value === \'copied\')',
      'status: showStatusFilter.value ? filters.status || undefined : undefined',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: [
      'definitionName: string',
      'filters.definitionName.trim()',
      'const definitionName = filters.definitionName.trim() || undefined',
      'definitionName,',
      'definitionName: \'\'',
      'v-model="activeRecordFilters.definitionName"',
      'placeholder="输入流程名称"',
      ':maxlength="50"',
    ],
  },
  {
    file: 'src/types/workflow.ts',
    patterns: [
      'export interface WorkflowInstanceQuery {\n  definitionId?: number\n  definitionName?: string',
      'export interface WorkflowTaskQuery {\n  instanceId?: string\n  status?: string\n  definitionName?: string',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: [
      '{ key: \'assigneeName\', label: \'节点处理人\', width: \'minmax(110px, 0.9fr)\', mobileHidden: true }',
      'assigneeName: task.assigneeName.trim() || \'-\'',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowRecordTable.vue',
    patterns: ['WorkflowRecordColumn', 'WorkflowStatusMeta', 'mobileHidden?: boolean', 'WorkflowRecordRow', 'gridTemplateColumns', 'workflow-record-table__header', 'workflow-record-table__header-cell--actions', 'workflow-record-table__row', 'workflow-record-table__cell-label', 'workflow-record-table__cell--mobile-hidden', 'workflow-record-table__actions', 'justify-content: center;', '<u-tag', 'custom-class="workflow-record-table__status-tag"', 'height: 24px;', 'font-size: 12px;', 'name="actions"', '@media screen and (max-width: 900px)'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: [
      'workflow-center__list-head--mobile-hidden',
      'class="workflow-center__filter-field workflow-center__filter-field--date">\n              <text class="workflow-center__filter-label">\n                提交时间',
      '.workflow-center__filter-field--date',
      'grid-column: auto;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowRecordTable.vue',
    patterns: [
      'workflow-record-table__cell--submitted-at',
      '@media screen and (max-width: 768px)',
      'grid-template-columns: minmax(0, 1fr) auto !important;',
      'grid-template-columns: auto minmax(0, 1fr) auto !important;',
      '.workflow-record-table__cell--status {\n    grid-column: 2;',
      '.workflow-record-table__cell--submitted-at {\n    grid-column: 1;',
      'grid-row: 2;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowInstancePage.vue',
    patterns: ['workflowInstanceIdFromContentKey', 'WorkflowDetailPanel', ':display-title="detailTitle"', 'presentation="history-page"', 'appContent.requestCloseTab', 'appContent.removeDynamicTab', 'appContent.requestRefresh', 'appContent.switchContent(\'workflow\')'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowTaskPage.vue',
    patterns: ['workflowTaskIdFromContentKey', 'workflowTaskInstanceIdFromContentKey', 'WorkflowDetailPanel', ':display-title="taskTitle"', 'presentation="page"', 'appContent.requestCloseTab', 'appContent.removeDynamicTab', 'appContent.requestRefresh', 'appContent.switchContent(\'workflow\')'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    patterns: ['workflowTaskTabTitle', 'workflowTaskTabTitle(taskStarterDisplayName(task), taskDefinitionName(task))'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: ['发起审批', '历史记录', '草稿箱', '流程', 'getWorkflowStartDraft', 'saveWorkflowStartDraft', 'listWorkflowInstances', 'WorkflowRuntimeForm', 'WorkflowReadOnlyGraph', 'WorkflowDetailPanel', 'presentation="history-drawer"', 'WorkflowFilterPanel', ':active-count="historyFilterCount"', 'WorkflowHistoryDatePicker', '保存草稿', '提交申请', 'registerTabCloseGuard', 'workflowRequestErrorMessage', 'catch (error)', 'workflowRequestErrorMessage(error, \'草稿保存失败\')', 'historyStatusOptions', 'queryHistory', 'resetHistoryFilters', 'buildWorkflowHistoryTimeQuery', 'workflowInstanceStatusMeta', 'class="workflow-start-page__history-filters"', 'class="workflow-start-page__filter-select"', '审批状态', '发起时间', '完成时间', '查询', '重置', 'workflowInstanceContentKey', 'function resolveMobilePage()', 'if (resolveMobilePage())', 'appContent.openDynamicTab({', 'path: `/pages/index/index?view=', 'encodeURIComponent(key)', 'takeWorkflowStartSeed', 'workflowStartSeedTick', 'appContent.requestRefresh()', '提交成功'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: [
      'class="workflow-start-page__records-scroll"',
      'class="workflow-start-page__records-table"',
      'workflow-start-page__record--header',
      'workflow-start-page__record-cell--name',
      'workflow-start-page__record-cell--key',
      'workflow-start-page__record-cell--version',
      'workflow-start-page__record-cell--start',
      'workflow-start-page__record-cell--end',
      'workflow-start-page__record-cell--status',
      'workflow-start-page__record-cell--action',
      '流程版本',
      'v{{ instance.definitionVersion }}',
      'min-width: 960px;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: [
      '.workflow-start-page__record--header {\n    display: none;',
      'grid-template-columns: auto minmax(0, 1fr) auto;',
      '.workflow-start-page__record-cell--key,\n  .workflow-start-page__record-cell--version,\n  .workflow-start-page__record-cell--end {\n    display: none;',
      '.workflow-start-page__record-cell--name {\n    grid-column: 1 / -1;\n    grid-row: 1;',
      '.workflow-start-page__record-cell--start {\n    grid-column: 2;\n    grid-row: 2;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowHistoryDatePicker.vue',
    patterns: ['resolveMobilePresentation', '<u-calendar', ':is-page="true"', 'class="workflow-history-date-picker__panel"', 'name="calendar"', 'size="16px"'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: [
      'class="workflow-start-page__form-card"',
      '.workflow-start-page__form-card {\n  padding: 20px;\n  border: 1px solid #dfe5ee;\n  border-radius: 6px;\n  background: #ffffff;\n  box-shadow: 0 2px 8px rgba(31, 35, 41, 0.05);',
      '.workflow-start-page__form-card {\n    padding: 12px;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: [
      'const savedDraft = ref<WorkflowStartDraft | null>(null)',
      'const hasSavedDraft = computed(() => Boolean(savedDraft.value',
      'const draftCount = computed(() => hasSavedDraft.value ? 1 : 0)',
      'function continueSavedDraft()',
      'function confirmDeleteSavedDraft()',
      'async function deleteSavedDraft()',
      'savedDraft.value = draft || null',
      '@click="continueSavedDraft"',
      '@click.stop="confirmDeleteSavedDraft"',
      'name="trash"',
      'workflow-start-page__draft-delete',
      'app-icon-button app-icon-button--small',
      'class="workflow-start-page__nav-badge"',
      'item.key === \'drafts\' && draftCount > 0',
      '共 {{ draftCount }} 份',
      'class="workflow-start-page__draft-grid"',
      'class="workflow-start-page__draft-cover-head"',
      'class="workflow-start-page__draft-label"',
      'class="workflow-start-page__draft-footer"',
      'class="workflow-start-page__draft-version"',
      '版本 {{ savedDraft.definitionVersion }}',
      '当前流程版本 {{ definition.version }}',
      '已按当前版本恢复草稿',
      '最后更新：{{ formatTime(draftUpdatedAt) }}',
      'grid-template-columns: repeat(auto-fill, minmax(150px, 160px));',
      'width: 160px;',
      'aspect-ratio: 210 / 297;',
      'border-radius: 2px;',
      'box-shadow: 0 6px 18px rgba(31, 35, 41, 0.12);',
      '.workflow-start-page__draft-grid {\n    grid-template-columns: 1fr;',
      '.workflow-start-page__draft-card {\n    width: 160px;\n    max-width: 100%;',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowReadOnlyGraph.vue',
    patterns: ['assigneeDisplay', '审批人：', '<svg', 'marker-end', 'ResizeObserver', 'viewportWidth.value', 'contentWidth + canvasPadding * 2', 'alignWorkflowTerminalNodes', 'source.node.type === \'start\' || target.node.type === \'end\'', 'if (straight)'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowRuntimeForm.vue',
    patterns: ['WorkflowFormField', 'fieldAccess', 'detail_list', 'WorkflowFieldControl', ':textarea-default-min-rows="2"', ':textarea-default-max-rows="6"', '<u-modal', ':z-index="10120"', 'workflow-help-modal', 'width: 680px', 'helpVisible.value = true', 'validateWorkflowFormData(props.fields, props.modelValue, accessMap.value)', 'readonlyAppearance?: \'disabled\' | \'plain\'', '\'workflow-form--plain-readonly\': readonlyAppearance === \'plain\'', '\'workflow-form--fully-readonly\': readonly && readonlyAppearance === \'plain\'', ':readonly-appearance="readonlyAppearance"', 'class="workflow-form__help-action"', '.workflow-form--fully-readonly {', '.workflow-form--plain-readonly {', ':deep(.u-input--disabled .u-input__input)', ':deep(.u-textarea--disabled .u-textarea__field)', ':deep(.u-switch--disabled)'],
  },
  {
    file: 'src/pages/workflow/workflow-form.ts',
    patterns: ['workflowFieldActionsMap', 'if (!Array.isArray(permission.actions))', 'continue', 'accessMap?: WorkflowFieldAccessMap', 'accessMap[field.key] !== \'write\''],
  },
  {
    file: 'src/pages/workflow/components/WorkflowFieldControl.vue',
    patterns: ['WorkflowFormField', 'WorkflowTextarea', 'textareaDefaultMinRows', 'textareaDefaultMaxRows', 'field.minVisibleRows || textareaDefaultMinRows', 'field.maxVisibleRows || textareaDefaultMaxRows', 'u-input', 'u-select', 'u-checkbox-group', 'readonlyAppearance?: \'disabled\' | \'plain\'', 'const plainReadonlyEmpty = computed', 'value.every((item) =>', 'v-if="plainReadonlyEmpty" class="workflow-control__empty"', '暂无填写', '.workflow-control__empty {'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowFieldControl.vue',
    patterns: ['nextTick', 'resolveWorkflowSelectPlacement', 'desktopSelectRootRef', 'desktopSelectPanelRef', 'syncDesktopSelectPlacement', 'ref="desktopSelectRootRef"', 'ref="desktopSelectPanelRef"', 'workflow-control__select-panel--top', 'bottom: calc(100% + 6px);'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowTextarea.vue',
    patterns: ['<u-textarea', 'custom-class="workflow-textarea__control"', ':height="minHeight"', ':auto-height="true"', ':count="count"', '--workflow-textarea-max-height', 'max-height: var(--workflow-textarea-max-height) !important;', '.uni-textarea-textarea', 'overflow-y: auto !important;'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    patterns: [
      'completeWorkflowTask',
      'withdrawWorkflowInstance',
      'WorkflowRuntimeForm',
      'WorkflowTextarea',
      'WorkflowNodeProgressList',
      'WorkflowReadOnlyGraph',
      'displayTitle?: string',
      'applicationActions?: boolean',
      'presentation?: \'dialog\' | \'history-drawer\' | \'page\'',
      'const historyDrawer = computed(() => props.presentation === \'history-drawer\')',
      'const compactHistoryDialog = ref(resolveCompactHistoryDialog())',
      'const historyDialog = computed(() => historyDrawer.value && compactHistoryDialog.value)',
      'function resolveCompactHistoryDialog()',
      'window.matchMedia(\'(max-width: 1024px)\').matches',
      'window.addEventListener(\'resize\', syncCompactHistoryDialog)',
      'window.removeEventListener(\'resize\', syncCompactHistoryDialog)',
      'const historyPage = computed(() => props.presentation === \'history-page\')',
      'const historyPresentation = computed(() => historyDrawer.value || historyPage.value)',
      'const inlinePresentation = computed(() => pagePresentation.value || historyPage.value)',
      '<view v-if="pagePresentation" class="workflow-detail-panel__page-nav">',
      'v-if="historyPresentation"',
      'const historyEventsExpanded = ref(false)',
      'historyEventsExpanded.value = false',
      'const actionConfirmVisible = ref(false)',
      '<u-modal',
      ':z-index="10160"',
      '@confirm="resolveActionConfirmation(true)"',
      '@cancel="resolveActionConfirmation(false)"',
      'const pagePresentation = computed(() => props.presentation === \'page\')',
      ':mode="popupMode"',
      ':width="popupWidth"',
      ':height="popupHeight"',
      ':custom-class="popupCustomClass"',
      ':border-radius="popupBorderRadius"',
      ':zoom="!historyDialog"',
      'workflow-detail-popup--history-drawer',
      'workflow-detail-popup--history-dialog',
      'workflow-detail-panel--history-dialog',
      'workflow-detail-panel--history-drawer',
      ':global(.workflow-detail-popup--history-drawer .u-drawer-right)',
      ':global(.workflow-detail-popup--history-dialog .u-mode-center-box)',
      'width: clamp(420px, 38vw, 620px) !important;',
      'width: min(720px, calc(100vw - 32px)) !important;',
      'height: min(760px, calc(100vh - 32px)) !important;',
      'max-height: calc(100vh - 32px);',
      'max-width: calc(100vw - 32px);',
      'width: min(560px, 94vw) !important;',
      'class="workflow-detail-panel__title-line"',
      'class="workflow-detail-panel__starter"',
      'custom-class="workflow-detail-panel__status-tag"',
      ':deep(.workflow-detail-panel__status-tag)',
      'height: 24px;',
      'class="workflow-detail-panel__subtitle workflow-detail-panel__subtitle--business"',
      '业务编号：{{ detail.instance.businessKey || \'-\' }}',
      'v-if="!historyPresentation && !pagePresentation" class="workflow-detail-panel__summary"',
      'class="workflow-detail-panel__history-form-section"',
      'class="workflow-detail-panel__history-record-section"',
      'class="workflow-detail-panel__history-section-scroll"',
      'workflow-detail-panel__body--history-page',
      'workflow-detail-panel__history-layout--page',
      ':scroll-y="!historyPage"',
      '-webkit-overflow-scrolling: touch;',
      'touch-action: pan-y;',
      '.workflow-detail-panel--history-page :deep(.u-input--disabled)',
      'pointer-events: none;',
      'readonly-appearance="plain"',
      'class="workflow-detail-panel__subsection-toggle"',
      '@click="historyEventsExpanded = !historyEventsExpanded"',
      ':name="historyEventsExpanded ? \'arrow-up\' : \'arrow-down\'"',
      '<template v-if="historyEventsExpanded">',
      'flex: 1 1 50%;',
      '申请表单',
      '流程流转记录',
      'detail.instance.starterName || \'未知用户\'',
      'taskHandlerName(task)',
      'dingtalk_h5:api:workflow:handle',
      'dingtalk_h5:api:workflow:withdraw',
      'dingtalk_h5:api:workflow:comment',
      'const hasDeleteApplicationPermission = computed',
      'v-if="showApplicationActions && hasDeleteApplicationPermission"',
      'function deleteApplicationUnavailableMessage()',
      '审批中的申请请先撤销后再删除',
      ':disabled="applicationActionBusy"',
      'workflowRequestErrorMessage(error, \'申请删除失败\')',
      'const resubmitApplication = computed',
      '[\'completed\', \'cancelled\'].includes(status)',
      'const modifyApplicationLabel = computed',
      '{{ modifyApplicationLabel }}',
      'resubmitApplication.value ? \'再次提交\' : \'修改申请\'',
      'commentAction?: boolean',
      'commentAction: false',
      'showCommentAction',
      'canCommentInstance',
      'showApplicationActions || showCommentAction',
      'v-if="showApplicationActions"',
      'v-if="showCommentAction"',
      'setWorkflowStartSeed',
      'workflow-detail-panel__application-actions',
      'line-height: 38px;',
      'white-space: nowrap;',
      '撤销',
      '修改',
      '评论',
      '删除',
      '审批处理',
      '流程记录',
      '流程图',
      'workflow-detail-panel__page-card',
      ':nodes="detail.nodes || []"',
      ':edges="detail.edges || []"',
      'workflow-detail-panel__page-cancel',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    patterns: ['workflowInstanceStatusMeta', 'workflowTaskStatusMeta'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowNodeProgressList.vue',
    patterns: [
      'WorkflowNodeProgressSummary',
      'WorkflowTaskSummary',
      'nodeProgress',
      'legacyProgress',
      'workflowNodeProgressStatusMeta',
      'workflowTaskStatusMeta',
      'nodeTypeLabel',
      'instance.starterName || \'未知用户\'',
      'taskHandlerName(task)',
      'taskLayerLabel(task)',
      'task.approvalLayerTotal || task.approvalLayer',
      'custom-class="workflow-node-progress__status-tag"',
      ':deep(.workflow-node-progress__status-tag)',
      'workflow-node-progress__item--warning',
      'workflow-node-progress__item--success',
      'workflow-node-progress__item--info',
      'workflow-node-progress__item--error',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowNodeProgressList.vue',
    patterns: [
      'WorkflowHistorySummary',
      'history: WorkflowHistorySummary[]',
      'commentsByNode',
      'event.eventType !== \'instance_commented\'',
      'commentActorName(comment)',
      'workflow-node-progress__comments',
      '评论人：',
    ],
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    patterns: [
      'auth.user?.workflowActorId || auth.user?.id',
      'isWorkflowTaskAssignedToUser(activeTask.value, currentUserId.value)',
      ':field-access="fieldAccess"',
      ':field-actions="fieldActions"',
      ':readonly="!canHandle"',
      'formRef.value?.validate()',
      'writableWorkflowFormData(current.form || [], formData.value, fieldAccess.value)',
      ':class="{ \'workflow-detail-panel__actions--page\': pagePresentation }"',
      '取消',
      '驳回',
      '同意',
      '提交办理',
      'z-index: 20;',
    ],
  },
  {
    file: 'src/pages/workflow/workflow-history-filter.ts',
    patterns: ['buildWorkflowHistoryTimeQuery', 'startTimeFrom', 'startTimeTo', 'endTimeFrom', 'endTimeTo'],
  },
  {
    file: 'docs/development-guidelines.md',
    patterns: ['PC 端选择与日期控件', '不得在 PC 端使用底部弹层式选择器', 'input[type=date]'],
  },
]

const forbiddenContent = [
  { file: 'src/api/workflow.ts', patterns: ['const WORKFLOW_API = \'/api/v2/workflows\''] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['client:api:workflow:'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: [':loading="definitionsLoading || listLoading || countsLoading"'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['workflow-center__topbar', 'custom-class="workflow-center__refresh"', 'const refreshing = ref(false)', 'async function handleRefresh()'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['workflow-center__list-refresh'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['workflow-start-popup', 'startPopupVisible', 'saveStartDraft()', 'submitStart()'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['definition.description || definition.key'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['grid-template-columns: 180px minmax(320px, 1fr) auto;'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['workflow-center__application-columns', 'workflow-record--application', 'workflow-record__accent', 'workflow-record__main'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['selectedTaskId', ':task-id=', ':presentation="activeTab === \'pending\' ? \'dialog\' : \'history-drawer\'"'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['applicationFilters', 'appliedApplicationFilters', 'applicationCategoryOptions', 'queryApplications', 'resetApplicationFilters', 'workflow-center__application-filters'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['from \'../workflow.routes\''] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['workflow-record__action--delete', '@click.stop="deleteRecord(row.id)"', 'deleteWorkflowInstance'] },
  { file: 'src/pages/workflow/components/WorkflowInstancePage.vue', patterns: ['from \'../workflow.routes\''] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['from \'../workflow.routes\''] },
  { file: 'src/pages/workflow/components/WorkflowTaskPage.vue', patterns: ['from \'../workflow.routes\''] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['margin-top: -68px', 'margin-top: -64px', 'background: rgba(255, 255, 255'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['size="small" type="primary" plain @click="switchSection(\'start\')"'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['max-width: 720px;'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['width: min(1080px, 100%);'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['border-left: 8px solid #0f766e;', '.workflow-start-page__draft-card::before', '.workflow-start-page__draft-card::after'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['restoredDraft', 'if (draft?.definitionVersion === loadedDefinition.version)', '流程已更新，旧草稿未恢复', '@click="switchSection(\'start\')"'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['<u-select', '<picker mode="date"', 'type="date"'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['WorkflowSummarySection', 'getWorkflowSummaryDefinition', '\'summary\'', '流程汇总'] },
  { file: 'src/pages/workflow/components/WorkflowSummaryPage.vue', patterns: ['选择需要汇总的流程', '更换流程', 'selectedDefinition'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['workflow-start-page__record-main', 'workflow-start-page__record-title-row'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['client:api:workflow:'] },
  { file: 'src/pages/workflow/components/WorkflowCenter.vue', patterns: ['function statusMeta('] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['<view v-if="inlinePresentation" class="workflow-detail-panel__page-nav">'] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['function statusMeta('] },
  { file: 'src/pages/workflow/components/WorkflowStartPage.vue', patterns: ['appContent.focusWorkflowInstance(instanceId)'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['function instanceStatusMeta(', 'function taskStatusMeta('] },
  { file: 'src/pages/workflow/components/WorkflowNodeProgressList.vue', patterns: ['function nodeProgressStatusMeta(', 'function taskStatusMeta('] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['#{{ detail.instance.starterId }}'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['处理人 #{{ task.assigneeId }}'] },
  { file: 'src/pages/workflow/components/WorkflowNodeProgressList.vue', patterns: ['用户 #' + '$' + '{instance.starterId}'] },
  { file: 'src/pages/workflow/components/WorkflowNodeProgressList.vue', patterns: ['处理人 #{{ task.handledBy || task.assigneeId }}'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['max-width: 720px;', '<scroll-view v-if="historyDrawer" class="workflow-detail-panel__body workflow-detail-panel__body--history-drawer" scroll-y>'] },
  { file: 'src/pages/workflow/components/WorkflowRuntimeForm.vue', patterns: ['uni.showModal({'] },
  { file: 'src/pages/workflow/components/WorkflowFieldControl.vue', patterns: ['<u-textarea', 'workflow-control__textarea'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['<u-textarea', 'workflow-detail-panel__comment-input'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['uni.showModal({'] },
  { file: 'src/pages/workflow/components/WorkflowTextarea.vue', patterns: ['resize:', 'workflow-resizable-textarea', 'workflow-textarea__count'] },
  { file: 'src/pages/workflow/components/WorkflowFieldControl.vue', patterns: ['WorkflowResizableTextarea'] },
  { file: 'src/pages/workflow/components/WorkflowDetailPanel.vue', patterns: ['WorkflowResizableTextarea'] },
  { file: 'src/pages/workflow/components/WorkflowRuntimeForm.vue', patterns: ['\'workflow-form--plain-readonly\': readonly && readonlyAppearance === \'plain\''] },
  { file: 'src/pages/performance/components/PerformanceWorkbench.vue', patterns: ['max-width: 640px;', 'align-self: flex-start;'] },
]

const orderedContent = [
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    patterns: ['@click="cancelStart"', '@click="submitStart"', '@click="saveDraft"'],
  },
  {
    file: 'src/pages/workflow/components/WorkflowRuntimeForm.vue',
    patterns: ['class="workflow-detail__row"', 'class="workflow-detail__add"'],
  },
]

const componentUsageCounts = [
  {
    file: 'src/pages/workflow/components/WorkflowCenter.vue',
    pattern: /<WorkflowHistoryDatePicker\b/g,
    expected: 2,
    label: 'workflow application submit-time calendar picker',
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    pattern: /<WorkflowHistoryDatePicker\b/g,
    expected: 4,
    label: 'workflow history calendar picker',
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    pattern: /<select\b/g,
    expected: 1,
    label: 'workflow history desktop dropdown',
  },
  {
    file: 'src/pages/workflow/components/WorkflowStartPage.vue',
    pattern: /width: min\(var\(--app-pc-content-max-width, 1080px\), 100%\);/g,
    expected: 2,
    label: 'workflow PC content width global variable',
  },
  {
    file: 'src/pages/workflow/components/WorkflowFieldControl.vue',
    pattern: /<WorkflowTextarea\b/g,
    expected: 2,
    label: 'workflow field textarea',
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    pattern: /<WorkflowTextarea\b/g,
    expected: 3,
    label: 'workflow interaction textarea',
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    pattern: /<WorkflowNodeProgressList\b/g,
    expected: 2,
    label: 'workflow node progress list',
  },
  {
    file: 'src/pages/workflow/components/WorkflowDetailPanel.vue',
    pattern: /readonly-appearance="plain"/g,
    expected: 2,
    label: 'workflow readable readonly form appearance',
  },
]

const failures = []

async function loadTypeScriptModule(relativePath) {
  const filename = resolve(root, relativePath)
  const source = readFileSync(filename, 'utf8')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.ES2020,
      target: ts.ScriptTarget.ES2020,
    },
    fileName: filename,
  }).outputText
  return import(`data:text/javascript;charset=utf-8,${encodeURIComponent(output)}`)
}

for (const file of requiredFiles) {
  if (!existsSync(resolve(root, file))) {
    failures.push(`missing file: ${file}`)
  }
}

for (const expectation of requiredContent) {
  const path = resolve(root, expectation.file)
  if (!existsSync(path)) {
    failures.push(`missing file: ${expectation.file}`)
    continue
  }
  const text = readFileSync(path, 'utf8')
  for (const pattern of expectation.patterns) {
    if (!text.includes(pattern)) {
      failures.push(`${expectation.file} missing pattern: ${pattern}`)
    }
  }
}

for (const expectation of forbiddenContent) {
  const path = resolve(root, expectation.file)
  if (!existsSync(path))
    continue
  const text = readFileSync(path, 'utf8')
  for (const pattern of expectation.patterns) {
    if (text.includes(pattern)) {
      failures.push(`${expectation.file} contains forbidden pattern: ${pattern}`)
    }
  }
}

for (const expectation of orderedContent) {
  const path = resolve(root, expectation.file)
  if (!existsSync(path))
    continue
  const text = readFileSync(path, 'utf8')
  let previousIndex = -1
  for (const pattern of expectation.patterns) {
    const index = text.indexOf(pattern)
    if (index < 0 || index <= previousIndex) {
      failures.push(`${expectation.file} must contain ordered pattern: ${pattern}`)
      break
    }
    previousIndex = index
  }
}

for (const expectation of componentUsageCounts) {
  const path = resolve(root, expectation.file)
  if (!existsSync(path))
    continue
  const text = readFileSync(path, 'utf8')
  const usageCount = text.match(expectation.pattern)?.length || 0
  if (usageCount !== expectation.expected)
    failures.push(`${expectation.file} must render every ${expectation.label} with WorkflowTextarea: ${usageCount}/${expectation.expected}`)
}

const graphLayoutPath = 'src/pages/workflow/workflow-graph-layout.ts'
if (existsSync(resolve(root, graphLayoutPath))) {
  try {
    const { alignWorkflowTerminalNodes } = await loadTypeScriptModule(graphLayoutPath)
    assert.equal(typeof alignWorkflowTerminalNodes, 'function')
    const alignedNodes = alignWorkflowTerminalNodes([
      { node: { id: 'start', type: 'start', name: '开始' }, x: 160, y: 0, width: 132, height: 72 },
      { node: { id: 'approval', type: 'approval', name: '审批' }, x: 100, y: 150, width: 236, height: 96 },
      { node: { id: 'end', type: 'end', name: '结束' }, x: 146, y: 300, width: 132, height: 72 },
    ], [
      { id: 'start-approval', source: 'start', target: 'approval' },
      { id: 'approval-end', source: 'approval', target: 'end' },
    ])
    const alignedNodeMap = new Map(alignedNodes.map(item => [item.node.id, item]))
    const centerX = item => item.x + item.width / 2
    assert.equal(centerX(alignedNodeMap.get('start')), centerX(alignedNodeMap.get('approval')))
    assert.equal(centerX(alignedNodeMap.get('end')), centerX(alignedNodeMap.get('approval')))
    assert.equal(alignedNodeMap.get('approval').x, 100)
  }
  catch (error) {
    failures.push(`${graphLayoutPath} terminal alignment check failed: ${error instanceof Error ? error.message : String(error)}`)
  }
}

const historyFilterPath = 'src/pages/workflow/workflow-history-filter.ts'
if (existsSync(resolve(root, historyFilterPath))) {
  try {
    const { buildWorkflowHistoryTimeQuery } = await loadTypeScriptModule(historyFilterPath)
    const filters = {
      startDateFrom: '2026-09-01',
      startDateTo: '2026-09-02',
      endDateFrom: '2026-09-03',
      endDateTo: '2026-09-04',
    }
    const query = buildWorkflowHistoryTimeQuery(filters)
    assert.deepEqual(query, {
      startTimeFrom: new Date(2026, 8, 1).getTime(),
      startTimeTo: new Date(2026, 8, 3).getTime() - 1,
      endTimeFrom: new Date(2026, 8, 3).getTime(),
      endTimeTo: new Date(2026, 8, 5).getTime() - 1,
    })
    assert.equal(buildWorkflowHistoryTimeQuery({ ...filters, startDateFrom: '2026-09-05' }), null)
    assert.deepEqual(buildWorkflowHistoryTimeQuery({}), {})
  }
  catch (error) {
    failures.push(`${historyFilterPath} date range check failed: ${error instanceof Error ? error.message : String(error)}`)
  }
}

const workflowTaskPath = 'src/pages/workflow/workflow-task.ts'
if (existsSync(resolve(root, workflowTaskPath))) {
  try {
    const { isWorkflowTaskAssignedToUser, workflowTaskTabTitle } = await loadTypeScriptModule(workflowTaskPath)
    assert.equal(workflowTaskTabTitle('张三', '绩效单'), '张三-绩效单')
    assert.equal(workflowTaskTabTitle('  ', '请假申请'), '未知用户-请假申请')
    assert.equal(isWorkflowTaskAssignedToUser({ assigneeId: ' 66 ' }, 66), true)
    assert.equal(isWorkflowTaskAssignedToUser({ assigneeId: '67' }, 66), false)
  }
  catch (error) {
    failures.push(`${workflowTaskPath} task presentation check failed: ${error instanceof Error ? error.message : String(error)}`)
  }
}

const workflowSelectPlacementPath = 'src/pages/workflow/workflow-select-placement.ts'
if (existsSync(resolve(root, workflowSelectPlacementPath))) {
  try {
    const { resolveWorkflowSelectPlacement } = await loadTypeScriptModule(workflowSelectPlacementPath)
    assert.equal(resolveWorkflowSelectPlacement({ controlTop: 500, controlBottom: 536, visibleTop: 0, visibleBottom: 600, panelHeight: 180 }), 'top')
    assert.equal(resolveWorkflowSelectPlacement({ controlTop: 120, controlBottom: 156, visibleTop: 0, visibleBottom: 600, panelHeight: 180 }), 'bottom')
    assert.equal(resolveWorkflowSelectPlacement({ controlTop: 120, controlBottom: 156, visibleTop: 100, visibleBottom: 220, panelHeight: 180 }), 'bottom')
  }
  catch (error) {
    failures.push(`${workflowSelectPlacementPath} placement check failed: ${error instanceof Error ? error.message : String(error)}`)
  }
}

const workflowDetailPath = 'src/pages/workflow/components/WorkflowDetailPanel.vue'
if (existsSync(resolve(root, workflowDetailPath))) {
  const source = readFileSync(resolve(root, workflowDetailPath), 'utf8')
  if (source.includes('workflow-detail-panel__comment'))
    failures.push(`${workflowDetailPath} must not render a persistent handling-comment field`)
  if (source.includes('操作人 #') || source.includes('用户 #'))
    failures.push(`${workflowDetailPath} must display workflow history actor names instead of user IDs`)
}

const workflowStatusPath = 'src/pages/workflow/workflow-status.ts'
if (existsSync(resolve(root, workflowStatusPath))) {
  try {
    const {
      workflowInstanceStatusMeta,
      workflowNodeProgressStatusMeta,
      workflowTaskStatusMeta,
    } = await loadTypeScriptModule(workflowStatusPath)
    assert.deepEqual(workflowInstanceStatusMeta('running'), { label: '审批中', type: 'warning' })
    assert.deepEqual(workflowInstanceStatusMeta('completed'), { label: '已完成', type: 'success' })
    assert.deepEqual(workflowInstanceStatusMeta('rejected'), { label: '已驳回', type: 'error' })
    assert.deepEqual(workflowInstanceStatusMeta('cancelled'), { label: '已取消', type: 'info' })
    assert.equal(workflowTaskStatusMeta('waiting').type, 'warning')
    assert.equal(workflowTaskStatusMeta('approved').type, 'success')
    assert.equal(workflowTaskStatusMeta('cancelled').type, 'info')
    assert.equal(workflowNodeProgressStatusMeta('processing').type, 'warning')
    assert.equal(workflowNodeProgressStatusMeta('completed').type, 'success')
    assert.equal(workflowNodeProgressStatusMeta('terminated', 'rejected').type, 'error')
    assert.equal(workflowNodeProgressStatusMeta('terminated', 'withdrawn').type, 'info')
  }
  catch (error) {
    failures.push(`${workflowStatusPath} status mapping check failed: ${error instanceof Error ? error.message : String(error)}`)
  }
}

if (failures.length > 0) {
  console.error(failures.join('\n'))
  process.exit(1)
}

console.log('workflow module structure checks passed')
