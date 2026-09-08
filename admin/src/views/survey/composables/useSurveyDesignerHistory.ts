import { computed, ref } from 'vue'

interface SurveyDesignerHistoryOptions {
  capture: () => string
  restore: (snapshot: string) => void
  delay?: number
  maxEntries?: number
}

export function useSurveyDesignerHistory(options: SurveyDesignerHistoryOptions) {
  const entries = ref<string[]>([])
  const cursor = ref(-1)
  const isRestoring = ref(false)
  const canUndo = computed(() => cursor.value > 0)
  const canRedo = computed(() => cursor.value >= 0 && cursor.value < entries.value.length - 1)
  let timer: ReturnType<typeof setTimeout> | undefined
  let pendingSnapshot = ''

  function clearTimer() {
    if (timer) clearTimeout(timer)
    timer = undefined
  }

  function commit(snapshot: string) {
    if (!snapshot || entries.value[cursor.value] === snapshot) return
    entries.value.splice(cursor.value + 1)
    entries.value.push(snapshot)
    if (entries.value.length > (options.maxEntries || 100)) entries.value.shift()
    cursor.value = entries.value.length - 1
  }

  function reset(snapshot = options.capture()) {
    clearTimer()
    pendingSnapshot = ''
    entries.value = snapshot ? [snapshot] : []
    cursor.value = entries.value.length - 1
  }

  function record(snapshot = options.capture()) {
    if (isRestoring.value) return
    pendingSnapshot = snapshot
    clearTimer()
    timer = setTimeout(() => {
      timer = undefined
      commit(pendingSnapshot)
      pendingSnapshot = ''
    }, options.delay ?? 300)
  }

  function flush() {
    if (!timer) return
    clearTimer()
    commit(pendingSnapshot)
    pendingSnapshot = ''
  }

  function restoreAt(nextCursor: number) {
    clearTimer()
    pendingSnapshot = ''
    cursor.value = nextCursor
    isRestoring.value = true
    options.restore(entries.value[nextCursor])
    queueMicrotask(() => { isRestoring.value = false })
  }

  function undo() {
    flush()
    if (canUndo.value) restoreAt(cursor.value - 1)
  }

  function redo() {
    if (canRedo.value) restoreAt(cursor.value + 1)
  }

  function dispose() {
    clearTimer()
    pendingSnapshot = ''
  }

  return { canUndo, canRedo, isRestoring, reset, record, undo, redo, dispose }
}
