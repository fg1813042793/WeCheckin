<script setup lang="ts">
defineProps<{ form: { title?: string; displayTitle?: string } }>()
</script>

<template>
  <el-form class="exam-name-fields" label-position="top" size="small">
    <el-form-item label="管理名称" required>
      <el-input v-model="form.title" aria-label="管理名称" placeholder="用于后台识别和管理考试" maxlength="100" />
      <div class="exam-name-fields__help">用于后台列表、考试记录和管理导出。</div>
    </el-form-item>
    <el-form-item label="显示名称">
      <el-input v-model="form.displayTitle" aria-label="显示名称" placeholder="留空时使用管理名称" maxlength="100" clearable />
      <div class="exam-name-fields__help">考生答题和查看结果时显示，留空则使用管理名称。</div>
    </el-form-item>
  </el-form>
</template>

<style scoped>
.exam-name-fields__help { margin-top: 4px; color: var(--admin-ui-color-text-muted); font-size: 12px; line-height: 1.5; }
</style>
