<template>
  <el-dialog
    v-model="dialog.visible"
    :title="title"
    width="760px"
    append-to-body
    destroy-on-close
    :close-on-click-modal="false"
    class="bank-rich-full-dialog"
    @closed="dialog.content=''"
  >
    <AdvancedRichTextEditor v-model="dialog.content" :min-height="420" placeholder="请输入富文本内容" />
    <template #footer>
      <el-button @click="dialog.visible=false">取消</el-button>
      <el-button type="primary" @click="confirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const AdvancedRichTextEditor = defineAsyncComponent(() => import('@/components/rich-text/AdvancedRichTextEditor.vue'))

type RichEditTarget = 'title' | 'option'

defineProps<{
  dialog: {
    visible: boolean
    target: RichEditTarget
    optionIndex: number
    content: string
  }
  title: string
  confirm: () => void
}>()
</script>

<style scoped>
:global(.bank-rich-full-dialog .el-dialog__body) {
  padding-top: 8px;
}

:global(.bank-rich-full-dialog) {
  max-width: calc(100vw - 32px);
}
</style>
