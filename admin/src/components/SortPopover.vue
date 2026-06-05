<template>
  <el-popover trigger="click" placement="bottom-end" :width="320" @show="initRules">
    <template #reference>
      <el-button circle icon="Sort" :title="'排序' + (activeCount ? ' (' + activeCount + ')' : '')" :type="activeCount ? 'primary' : ''" />
    </template>
    <div style="padding:4px 0;">
      <div v-for="(col, i) in localCols" :key="col.field" style="display:flex;align-items:center;gap:6px;margin-bottom:6px">
        <el-checkbox v-model="col._checked" @change="onCheckChange" />
        <span style="flex:1;font-size:13px;white-space:nowrap">{{ col.label }}</span>
        <el-select v-model="col._order" size="small" style="width:90px" :disabled="!col._checked" @change="emitChange">
          <el-option label="升序" value="asc" />
          <el-option label="降序" value="desc" />
        </el-select>
        <el-button size="small" circle :icon="Close" @click="clearSort(col)" v-if="col._checked" />
      </div>
      <div v-if="activeCount === 0" style="color:#999;font-size:13px;text-align:center;padding:12px 0">勾选列名并选择排序方式</div>
      <div style="display:flex;gap:8px;justify-content:flex-end;margin-top:8px;border-top:1px solid #eee;padding-top:8px">
        <el-button size="small" @click="clearAll">清除排序</el-button>
        <el-button size="small" type="primary" @click="applySort">排序</el-button>
      </div>
    </div>
  </el-popover>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { Close } from '@element-plus/icons-vue'

const props = defineProps<{
  columns: { label: string; field: string }[]
  modelValue: { field: string; order: string }[]
}>()

type LocalCol = { label: string; field: string; _checked: boolean; _order: string }
const localCols = ref<LocalCol[]>([])

const emit = defineEmits<{
  (e: 'update:modelValue', v: { field: string; order: string }[]): void
  (e: 'change', v: { field: string; order: string }[]): void
}>()

function initRules() {
  localCols.value = props.columns.map(c => {
    const existing = props.modelValue.find(s => s.field === c.field)
    return { ...c, _checked: !!existing, _order: existing?.order || 'asc' }
  })
}

const activeCount = computed(() => localCols.value.filter(c => c._checked).length)

function onCheckChange() {
  const checked = localCols.value.filter(c => c._checked)
  if (checked.length === 0) {
    emitChange()
  }
}

function emitChange() {
  const result = localCols.value.filter(c => c._checked).map(c => ({ field: c.field, order: c._order }))
  emit('update:modelValue', result)
}

function clearSort(col: any) {
  col._checked = false
  emitChange()
}

function clearAll() {
  localCols.value.forEach(c => { c._checked = false; c._order = 'asc' })
  emitChange()
}

function applySort() {
  const result = localCols.value.filter(c => c._checked).map(c => ({ field: c.field, order: c._order }))
  emit('update:modelValue', result)
  emit('change', result)
}
</script>
