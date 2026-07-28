<template>
  <el-dialog
    v-model="dialog.visible"
    :title="title"
    width="760px"
    append-to-body
    :close-on-click-modal="false"
    class="bank-rich-full-dialog"
    @closed="dialog.content=''"
  >
    <div class="bank-rich-full-editor">
      <QuillEditor
        v-model:content="dialog.content"
        content-type="html"
        :options="options"
      />
    </div>
    <template #footer>
      <el-button @click="dialog.visible=false">取消</el-button>
      <el-button type="primary" @click="confirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'

type RichEditTarget = 'title' | 'option'

defineProps<{
  dialog: {
    visible: boolean
    target: RichEditTarget
    optionIndex: number
    content: string
  }
  title: string
  options: Record<string, any>
  confirm: () => void
}>()
</script>

<style scoped>
:global(.bank-rich-full-dialog .el-dialog__body) {
  padding-top: 8px;
}

:global(.bank-rich-full-editor) {
  display: flex;
  flex-direction: column;
  height: min(58vh, 560px);
  min-height: 420px;
}

:global(.bank-rich-full-editor .ql-toolbar) {
  flex-shrink: 0;
  border-color: #dcdfe6;
  border-radius: 8px 8px 0 0;
  background: #fff;
}

:global(.bank-rich-full-editor .ql-container) {
  flex: 1;
  min-height: 0;
  border-color: #dcdfe6;
  border-radius: 0 0 8px 8px;
  font-size: 14px;
}

:global(.bank-rich-full-editor .ql-editor) {
  min-height: 100%;
  padding: 14px 16px;
}
</style>
