<template>
  <div class="calc-editor">
    <el-input
      :model-value="modelValue?.expr || ''"
      @update:model-value="onExprChange"
      placeholder="如 q1 + q2 * 0.1，留空则不计算"
      size="small"
    >
      <template #append>
        <el-button size="small" @click="evalNow" :loading="evaluating">试算</el-button>
      </template>
    </el-input>
    <div v-if="modelValue?.target" class="hint">
      计算结果将写入题目: <code>{{ modelValue.target }}</code>
    </div>
    <el-input
      v-if="modelValue"
      v-model="modelValue.target"
      placeholder="目标题目 ID (留空则写入此题)"
      size="small"
      style="margin-top: 4px"
    />
    <div v-if="evalResult !== null" class="result" :class="{ error: evalError }">
      {{ evalError || '= ' + JSON.stringify(evalResult) }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { adminApi } from '../../../api'

const props = defineProps<{ modelValue: any; env?: Record<string, any> }>()
const emit = defineEmits<{ 'update:modelValue': [v: any] }>()

const evaluating = ref(false)
const evalResult = ref<any>(null)
const evalError = ref('')

const onExprChange = (val: string) => {
  if (val === '') {
    emit('update:modelValue', null)
  } else {
    emit('update:modelValue', { ...(props.modelValue || {}), expr: val })
  }
}

const evalNow = async () => {
  const expr = props.modelValue?.expr
  if (!expr) return
  evaluating.value = true
  evalError.value = ''
  try {
    const res = await adminApi.formkitEval({ expr, env: props.env || {} })
    evalResult.value = res.data?.value
  } catch (e: any) {
    evalError.value = e?.msg || '试算失败'
    evalResult.value = null
  } finally {
    evaluating.value = false
  }
}
</script>

<style scoped>
.calc-editor {
  display: flex;
  flex-direction: column;
  width: 100%;
}
.hint {
  font-size: 11px;
  color: #909399;
  margin-top: 4px;
}
.hint code {
  background: #f0f9ff;
  color: #409eff;
  padding: 1px 4px;
  border-radius: 3px;
}
.result {
  font-size: 12px;
  margin-top: 4px;
  color: #67c23a;
  font-family: monospace;
}
.result.error {
  color: #f56c6c;
}
</style>
