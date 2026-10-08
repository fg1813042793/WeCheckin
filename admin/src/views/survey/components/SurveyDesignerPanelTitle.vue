<template>
  <header class="props-panel-title">
    <div class="props-panel-title-copy">
      <strong>{{ title }}</strong>
      <span v-if="context" class="props-panel-title-context" :title="context">{{ context }}</span>
    </div>
    <button v-if="closable" type="button" class="props-panel-title-close" aria-label="关闭属性面板" @click="emit('close')">
      <el-icon><Close /></el-icon>
    </button>
  </header>
</template>

<script setup lang="ts">
import { Close } from '@element-plus/icons-vue'

withDefaults(defineProps<{
  title: string
  context?: string
  closable?: boolean
}>(), {
  context: '',
  closable: false,
})
const emit = defineEmits<{ close: [] }>()
</script>

<style scoped>
.props-panel-title {
  position:sticky; top:0; z-index:2; min-height:48px; box-sizing:border-box;
  margin:-14px -14px 12px; padding:10px 46px 9px 25px;
  border-bottom:1px solid var(--designer-border); background:rgba(255,255,255,0.96);
  backdrop-filter:saturate(120%) blur(8px); display:flex; align-items:center;
}
.props-panel-title::before {
  content:''; position:absolute; left:14px; top:14px; width:3px; height:14px;
  border-radius:999px; background:var(--designer-accent);
}
.props-panel-title-copy { min-width:0; display:flex; flex:1; flex-direction:column; justify-content:center; gap:2px; }
.props-panel-title strong {
  color:#1f2937; font-size:14px; font-weight:650; line-height:20px;
  overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
}
.props-panel-title-context {
  color:#98a2b3; font-size:11px; line-height:16px;
  overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
}
.props-panel-title-close {
  position:absolute; top:9px; right:10px; width:30px; height:30px; padding:0;
  border:0; border-radius:7px; background:transparent; color:#667085; cursor:pointer;
  display:flex; align-items:center; justify-content:center;
}
.props-panel-title-close:hover { background:#f2f4f7; color:#1f2937; }
.props-panel-title-close:focus-visible { outline:2px solid var(--designer-accent); outline-offset:2px; }
</style>
