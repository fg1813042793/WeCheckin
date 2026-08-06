import { computed, h, ref, unref, watch } from 'vue'

const DEFAULT_TABLE_PAGE_SIZE = 10

function normalizeRows(rowsRef) {
  const value = typeof rowsRef === 'function' ? rowsRef() : unref(rowsRef)
  return Array.isArray(value) ? value : []
}

export function useTablePagination(rowsRef, pageSize = DEFAULT_TABLE_PAGE_SIZE) {
  const page = ref(1)
  const tableRows = computed(() => normalizeRows(rowsRef))
  const total = computed(() => tableRows.value.length)
  const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))
  const rows = computed(() => {
    const current = Math.min(Math.max(1, page.value), pageCount.value)
    const start = (current - 1) * pageSize
    return tableRows.value.slice(start, start + pageSize)
  })

  function reset() {
    page.value = 1
  }

  function prev() {
    if (page.value > 1) page.value -= 1
  }

  function next() {
    if (page.value < pageCount.value) page.value += 1
  }

  watch(tableRows, () => {
    reset()
  })

  watch(pageCount, (count) => {
    if (page.value > count) page.value = count
  })

  return {
    page,
    pageSize,
    pageCount,
    rows,
    total,
    reset,
    prev,
    next
  }
}

export function renderTablePagination(pagination) {
  if (!pagination || pagination.total.value <= 0) return null
  const start = (pagination.page.value - 1) * pagination.pageSize + 1
  const end = Math.min(pagination.page.value * pagination.pageSize, pagination.total.value)
  return h('view', { class: 'table-pagination table-pagination-outside' }, [
    h('text', { class: 'table-pagination-info' }, `共 ${pagination.total.value} 条，${start}-${end}`),
    h('view', { class: 'table-pagination-actions' }, [
      h('button', {
        class: ['table-pagination-btn', pagination.page.value <= 1 ? 'disabled' : ''],
        disabled: pagination.page.value <= 1,
        onClick: pagination.prev
      }, '上一页'),
      h('text', { class: 'table-pagination-page' }, `${pagination.page.value} / ${pagination.pageCount.value}`),
      h('button', {
        class: ['table-pagination-btn', pagination.page.value >= pagination.pageCount.value ? 'disabled' : ''],
        disabled: pagination.page.value >= pagination.pageCount.value,
        onClick: pagination.next
      }, '下一页')
    ])
  ])
}
