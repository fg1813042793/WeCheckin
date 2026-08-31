<template>
  <div class="flow-insert">
    <span class="flow-insert__line" />
    <el-popover ref="popoverRef" placement="right-start" :width="196" trigger="click" popper-class="workflow-insert-popper">
      <template #reference>
        <button class="flow-insert__trigger" type="button" title="添加流程节点" :disabled="disabled" @click.stop>
          <el-icon><Plus /></el-icon>
        </button>
      </template>
      <div class="insert-menu">
        <button type="button" @click="choose('approval')">
          <span class="insert-menu__icon approval"><el-icon><UserFilled /></el-icon></span>
          <span><b>审批人</b><small>指定人员或角色处理</small></span>
        </button>
        <button type="button" @click="choose('exclusive')">
          <span class="insert-menu__icon condition"><el-icon><Switch /></el-icon></span>
          <span><b>条件分支</b><small>按条件选择一条路径</small></span>
        </button>
        <button type="button" @click="choose('parallel')">
          <span class="insert-menu__icon parallel"><el-icon><Grid /></el-icon></span>
          <span><b>并行分支</b><small>多条路径同时执行</small></span>
        </button>
      </div>
    </el-popover>
    <span class="flow-insert__line" />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { Grid, Plus, Switch, UserFilled } from '@element-plus/icons-vue'
import type { WorkflowNodeType } from '../../types'

const props = defineProps<{ edgeId: string; disabled?: boolean }>()
const emit = defineEmits<{
  insert: [payload: { edgeId: string; type: Extract<WorkflowNodeType, 'approval' | 'exclusive' | 'parallel'> }]
}>()
const popoverRef = ref<{ hide?: () => void }>()

function choose(type: Extract<WorkflowNodeType, 'approval' | 'exclusive' | 'parallel'>) {
  emit('insert', { edgeId: props.edgeId, type })
  popoverRef.value?.hide?.()
}
</script>

<style scoped>
.flow-insert { position: relative; display: flex; min-height: 64px; flex-direction: column; align-items: center; justify-content: center; }
.flow-insert__line { width: 1px; min-height: 15px; flex: 1; background: #c8d0da; }
.flow-insert__trigger { display: grid; width: 28px; height: 28px; flex: 0 0 28px; place-items: center; padding: 0; border: 0; border-radius: 50%; color: #fff; background: #4b8ffb; box-shadow: 0 3px 9px rgb(75 143 251 / 30%); cursor: pointer; transition: transform .18s, background .18s; }
.flow-insert__trigger:hover { background: #337cf1; transform: scale(1.08); }
.flow-insert__trigger:disabled { cursor: not-allowed; opacity: .5; transform: none; }
.insert-menu { display: grid; gap: 4px; }
.insert-menu button { display: flex; align-items: center; gap: 10px; width: 100%; padding: 8px; border: 0; border-radius: 6px; color: #344054; background: transparent; text-align: left; cursor: pointer; }
.insert-menu button:hover { background: #f4f7fb; }
.insert-menu__icon { display: grid; width: 34px; height: 34px; flex: 0 0 34px; place-items: center; border-radius: 6px; font-size: 17px; }
.insert-menu__icon.approval { color: #d97706; background: #fff4e5; }
.insert-menu__icon.condition { color: #16a36a; background: #ebf8f1; }
.insert-menu__icon.parallel { color: #2878d0; background: #eaf3ff; }
.insert-menu b, .insert-menu small { display: block; }
.insert-menu b { font-size: 13px; font-weight: 600; }
.insert-menu small { margin-top: 3px; color: #98a2b3; font-size: 11px; }
</style>
