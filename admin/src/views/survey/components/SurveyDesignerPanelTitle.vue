<template>
  <header class="props-panel-title">
    <strong>{{ title }}</strong>
    <span v-if="context" class="props-panel-title-context" :title="context">{{ context }}</span>
  </header>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  context?: string
}>(), {
  context: '',
})
</script>

<style scoped>
.props-panel-title {
  position:sticky; top:0; z-index:2; min-height:48px; box-sizing:border-box;
  margin:-14px -14px 12px; padding:10px 14px 9px 25px;
  border-bottom:1px solid var(--designer-border); background:rgba(255,255,255,0.96);
  backdrop-filter:saturate(120%) blur(8px); display:flex; flex-direction:column; justify-content:center; gap:2px;
}
.props-panel-title::before {
  content:''; position:absolute; left:14px; top:14px; width:3px; height:14px;
  border-radius:999px; background:var(--designer-accent);
}
.props-panel-title strong {
  color:#1f2937; font-size:14px; font-weight:650; line-height:20px;
  overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
}
.props-panel-title-context {
  color:#98a2b3; font-size:11px; line-height:16px;
  overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
}
</style>
