<template>
  <div class="options-editor">
    <div
      v-for="(opt, idx) in modelValue"
      :key="idx"
      class="opt-row"
    >
      <el-input
        v-model="opt.label"
        placeholder="显示文本"
        size="small"
        style="flex: 1"
      />
      <el-input
        v-model="opt.value"
        placeholder="值"
        size="small"
        style="flex: 1"
      />
      <el-button
        size="small"
        text
        type="danger"
        @click="removeAt(idx)"
      >×</el-button>
    </div>
    <el-button size="small" @click="addOption" plain>+ 添加选项</el-button>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ modelValue: any[] }>()
const emit = defineEmits<{ 'update:modelValue': [v: any[]] }>()

const updateList = (list: any[]) => {
  emit('update:modelValue', list)
}

const addOption = () => {
  const list = [...(props.modelValue || [])]
  const n = list.length + 1
  list.push({ label: `选项 ${n}`, value: `opt${n}` })
  updateList(list)
}

const removeAt = (idx: number) => {
  const list = [...props.modelValue]
  list.splice(idx, 1)
  updateList(list)
}
</script>

<style scoped>
.options-editor {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}
.opt-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
</style>
