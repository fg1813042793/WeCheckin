<template>
  <div class="validate-editor">
    <div
      v-for="(rule, idx) in modelValue"
      :key="idx"
      class="rule-row"
    >
      <el-select v-model="rule.type" placeholder="类型" size="small" style="width: 110px">
        <el-option label="最大长度" value="maxLength" />
        <el-option label="最小长度" value="minLength" />
        <el-option label="正则" value="pattern" />
      </el-select>
      <el-input
        v-model="rule.value"
        placeholder="值"
        size="small"
        style="flex: 1"
      />
      <el-input
        v-model="rule.message"
        placeholder="错误提示"
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
    <el-button size="small" @click="addRule" plain>+ 添加校验</el-button>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ modelValue: any[] }>()
const emit = defineEmits<{ 'update:modelValue': [v: any[]] }>()

const addRule = () => {
  const list = [...(props.modelValue || [])]
  list.push({ type: 'maxLength', value: '', message: '' })
  emit('update:modelValue', list)
}

const removeAt = (idx: number) => {
  const list = [...props.modelValue]
  list.splice(idx, 1)
  emit('update:modelValue', list)
}
</script>

<style scoped>
.validate-editor {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}
.rule-row {
  display: flex;
  align-items: center;
  gap: 6px;
}
</style>
