<script setup lang="ts">
import type { WorkflowStatusMeta } from '../workflow-status'
import { computed } from 'vue'

interface WorkflowRecordColumn {
  key: string
  label: string
  width?: string
  mobileHidden?: boolean
}

interface WorkflowRecordRow {
  id: string
  cells: Record<string, string>
  status?: WorkflowStatusMeta
  deletable?: boolean
  deleting?: boolean
}

const props = defineProps<{
  columns: WorkflowRecordColumn[]
  rows: WorkflowRecordRow[]
}>()

defineSlots<{
  actions: (props: { row: WorkflowRecordRow }) => unknown
}>()

const gridStyle = computed(() => ({
  gridTemplateColumns: props.columns
    .map(column => column.width || 'minmax(100px, 1fr)')
    .join(' '),
}))
</script>

<template>
  <view class="workflow-record-table">
    <view class="workflow-record-table__header" :style="gridStyle">
      <text
        v-for="column in columns"
        :key="column.key"
        :class="{ 'workflow-record-table__header-cell--actions': column.key === 'actions' }"
      >
        {{ column.label }}
      </text>
    </view>

    <view
      v-for="row in rows"
      :key="row.id"
      class="workflow-record-table__row"
      :style="gridStyle"
    >
      <view
        v-for="column in columns"
        :key="column.key"
        class="workflow-record-table__cell"
        :class="{
          'workflow-record-table__cell--name': column.key === 'name',
          'workflow-record-table__cell--status': column.key === 'status',
          'workflow-record-table__cell--submitted-at': column.key === 'submittedAt',
          'workflow-record-table__cell--mobile-hidden': column.mobileHidden,
          'workflow-record-table__actions': column.key === 'actions',
        }"
      >
        <text class="workflow-record-table__cell-label">
          {{ column.label }}
        </text>
        <slot v-if="column.key === 'actions'" name="actions" :row="row" />
        <u-tag
          v-else-if="column.key === 'status' && row.status"
          custom-class="workflow-record-table__status-tag"
          :text="row.status.label"
          :type="row.status.type"
          size="mini"
        />
        <text v-else class="workflow-record-table__cell-value">
          {{ row.cells[column.key] || '-' }}
        </text>
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
.workflow-record-table {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}

.workflow-record-table__header,
.workflow-record-table__row {
  display: grid;
  align-items: center;
  gap: 14px;
  box-sizing: border-box;
}

.workflow-record-table__header {
  min-height: 44px;
  padding: 0 16px;
  border-bottom: 1px solid #e5eaf3;
  background: #f7f8fa;
  color: #6b7785;
  font-size: 12px;
  font-weight: 600;
}

.workflow-record-table__row {
  min-height: 68px;
  padding: 12px 16px;
  background: #fff;
}

.workflow-record-table__row + .workflow-record-table__row {
  border-top: 1px solid #edf0f5;
}

.workflow-record-table__row:hover {
  background: #f8fafc;
}

.workflow-record-table__cell {
  min-width: 0;
}

.workflow-record-table__cell-label {
  display: none;
}

.workflow-record-table__cell-value {
  display: block;
  overflow: hidden;
  color: #4e5969;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-record-table__cell--name .workflow-record-table__cell-value {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
}

.workflow-record-table__header-cell--actions {
  text-align: center;
}

.workflow-record-table__actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.workflow-record-table__status-tag,
:deep(.workflow-record-table__status-tag) {
  width: auto;
  height: 24px;
  min-height: 24px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 22px;
  white-space: nowrap;
  box-sizing: border-box;
}

@media screen and (max-width: 900px) {
  .workflow-record-table__header {
    display: none;
  }

  .workflow-record-table__row {
    min-height: 0;
    padding: 14px;
    grid-template-columns: minmax(0, 1fr) auto !important;
    align-items: start;
    gap: 14px 18px;
  }

  .workflow-record-table__cell {
    display: grid;
    gap: 5px;
  }

  .workflow-record-table__cell--mobile-hidden {
    display: none;
  }

  .workflow-record-table__cell-label {
    display: block;
    color: #9ca3af;
    font-size: 11px;
  }

  .workflow-record-table__cell--name .workflow-record-table__cell-label,
  .workflow-record-table__actions .workflow-record-table__cell-label {
    display: none;
  }

  .workflow-record-table__cell--name {
    grid-column: 1;
    grid-row: 1;
  }

  .workflow-record-table__cell--status {
    grid-column: 2;
    grid-row: 1;
    justify-self: end;
  }

  .workflow-record-table__cell--submitted-at {
    grid-column: 1;
    grid-row: 2;
  }

  .workflow-record-table__actions {
    grid-column: 2;
    grid-row: 2;
    display: flex;
    align-self: end;
    justify-self: end;
    justify-content: flex-end;
  }
}

@media screen and (max-width: 768px) {
  .workflow-record-table__row {
    padding: 12px;
    grid-template-columns: auto minmax(0, 1fr) auto !important;
    align-items: center;
    gap: 10px 12px;
  }

  .workflow-record-table__cell {
    display: flex;
    align-items: center;
  }

  .workflow-record-table__cell--mobile-hidden {
    display: none;
  }

  .workflow-record-table__cell-label {
    display: none;
  }

  .workflow-record-table__cell--name {
    grid-column: 1 / -1;
    grid-row: 1;
  }

  .workflow-record-table__cell--status {
    grid-column: 1;
    grid-row: 2;
    justify-self: start;
  }

  .workflow-record-table__cell--submitted-at {
    grid-column: 2;
    grid-row: 2;
  }

  .workflow-record-table__actions {
    grid-column: 3;
    grid-row: 2;
  }
}
</style>
