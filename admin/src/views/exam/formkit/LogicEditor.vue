<template>
  <div class="logic-editor">
    <div
      v-for="(rule, idx) in modelValue"
      :key="idx"
      class="logic-row"
    >
      <el-input
        v-model="rule.when.questionId"
        placeholder="题目 ID"
        size="small"
        style="width: 90px"
      />
      <el-select v-model="rule.when.operator" size="small" style="width: 100px">
        <el-option label="等于" value="==" />
        <el-option label="不等于" value="!=" />
        <el-option label="大于" value=">" />
        <el-option label="小于" value="<" />
        <el-option label="包含" value="contains" />
        <el-option label="为空" value="empty" />
        <el-option label="非空" value="notEmpty" />
      </el-select>
      <el-input
        v-if="rule.when.operator !== 'empty' && rule.when.operator !== 'notEmpty'"
        v-model="rule.when.value"
        placeholder="值"
        size="small"
        style="flex: 1"
      />
      <el-select v-model="rule.action" size="small" style="width: 90px">
        <el-option label="显示" value="show" />
        <el-option label="隐藏" value="hide" />
        <el-option label="必填" value="require" />
        <el-option label="选填" value="optional" />
      </el-select>
      <el-input
        v-model="rule.target"
        placeholder="目标 ID (留空=自身)"
        size="small"
        style="width: 100px"
      />
      <el-button
        size="small"
        text
        type="danger"
        @click="removeAt(idx)"
      >×</el-button>
    </div>
    <el-button size="small" @click="addRule" plain>+ 添加逻辑</el-button>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ modelValue: any[]; allQuestions: any[] }>()
const emit = defineEmits<{ 'update:modelValue': [v: any[]] }>()

const addRule = () => {
  const list = [...(props.modelValue || [])]
  list.push({
    when: { questionId: '', operator: '==', value: '' },
    action: 'hide',
    target: ''
  })
  emit('update:modelValue', list)
}

const removeAt = (idx: number) => {
  const list = [...props.modelValue]
  list.splice(idx, 1)
  emit('update:modelValue', list)
}
</script>

<style scoped>
.logic-editor {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}
.logic-row {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}
</style>
