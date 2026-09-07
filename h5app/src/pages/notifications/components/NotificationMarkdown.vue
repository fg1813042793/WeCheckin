<script setup lang="ts">
import DOMPurify from 'dompurify'
import { computed } from 'vue'
import {
  renderNotificationMarkdownSource,
  renderNotificationMarkdownWithoutHtml,
} from '@/pages/notifications/notification-markdown'

const props = defineProps<{
  content?: string
}>()

const purifier = typeof window === 'undefined' ? null : DOMPurify(window)
const safeHtml = computed(() => {
  if (!purifier?.sanitize)
    return renderNotificationMarkdownWithoutHtml(props.content || '')
  return purifier.sanitize(renderNotificationMarkdownSource(props.content || ''), {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'textarea', 'select', 'option', 'iframe', 'object', 'embed'],
    FORBID_ATTR: ['style'],
    ALLOW_DATA_ATTR: false,
  })
})
</script>

<template>
  <rich-text class="notification-markdown" :nodes="safeHtml" />
</template>

<style scoped lang="scss">
.notification-markdown {
  display: block;
  min-width: 0;
  color: #4e5969;
  font-size: 14px;
  line-height: 24px;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.notification-markdown :deep(h1),
.notification-markdown :deep(h2),
.notification-markdown :deep(h3),
.notification-markdown :deep(h4),
.notification-markdown :deep(h5),
.notification-markdown :deep(h6) {
  margin: 18px 0 10px;
  color: #1f2329;
  font-weight: 700;
  line-height: 1.45;
}

.notification-markdown :deep(h1) {
  font-size: 22px;
}

.notification-markdown :deep(h2) {
  font-size: 19px;
}

.notification-markdown :deep(h3),
.notification-markdown :deep(h4),
.notification-markdown :deep(h5),
.notification-markdown :deep(h6) {
  font-size: 16px;
}

.notification-markdown :deep(h1:first-child),
.notification-markdown :deep(h2:first-child),
.notification-markdown :deep(h3:first-child),
.notification-markdown :deep(p:first-child),
.notification-markdown :deep(ul:first-child),
.notification-markdown :deep(ol:first-child),
.notification-markdown :deep(blockquote:first-child),
.notification-markdown :deep(pre:first-child) {
  margin-top: 0;
}

.notification-markdown :deep(p) {
  margin: 0 0 12px;
}

.notification-markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.notification-markdown :deep(ul),
.notification-markdown :deep(ol) {
  margin: 10px 0 12px;
  padding-left: 24px;
}

.notification-markdown :deep(li) {
  margin: 4px 0;
}

.notification-markdown :deep(blockquote) {
  margin: 12px 0;
  padding: 8px 12px;
  border-left: 3px solid #00a67e;
  background: #f2fbf8;
  color: #646a73;
}

.notification-markdown :deep(a) {
  color: #2563eb;
  text-decoration: underline;
}

.notification-markdown :deep(code) {
  padding: 2px 5px;
  border-radius: 3px;
  background: #f2f3f5;
  color: #c93756;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.notification-markdown :deep(pre) {
  margin: 12px 0;
  padding: 12px;
  overflow-x: auto;
  border-radius: 4px;
  background: #1f2329;
  color: #f7f8fa;
  white-space: pre;
}

.notification-markdown :deep(pre code) {
  padding: 0;
  background: transparent;
  color: inherit;
}

.notification-markdown :deep(table) {
  width: 100%;
  margin: 12px 0;
  border-collapse: collapse;
  table-layout: fixed;
}

.notification-markdown :deep(th),
.notification-markdown :deep(td) {
  padding: 8px 10px;
  border: 1px solid #dfe3e8;
  overflow-wrap: anywhere;
  text-align: left;
  vertical-align: top;
}

.notification-markdown :deep(th) {
  background: #f7f8fa;
  color: #1f2329;
  font-weight: 600;
}

.notification-markdown :deep(img) {
  max-width: 100%;
  height: auto;
}

.notification-markdown :deep(hr) {
  margin: 16px 0;
  border: 0;
  border-top: 1px solid #dfe3e8;
}
</style>
