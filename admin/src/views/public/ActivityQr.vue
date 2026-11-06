<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { PublicPortalKind } from '@/types/publicPortal'
import { activityIdentity, errorText, safeQrFilename, safeTargetUrl } from '../public-portal/portalModel'
import { downloadActivityQr, generateActivityQr } from '../public-portal/portalQr'

const props = withDefaults(defineProps<{ url: string; title: string; managementTitle?: string; kind: PublicPortalKind; sourceId: number; category?: string }>(), { managementTitle: '' })
const identifiedTitle = computed(() => `${props.title || '活动二维码'} · ${activityIdentity(props.kind, props.sourceId, props.category)}`)
const root = ref<HTMLElement>()
const dataUrl = ref('')
const loading = ref(false)
const error = ref('')
const expanded = ref(false)
const downloading = ref(false)
const safeUrl = computed(() => safeTargetUrl(props.url))
let active = false
let generation = 0
let observer: IntersectionObserver | undefined
async function generate() {
  if (loading.value || dataUrl.value) return
  const current = ++generation
  loading.value = true
  error.value = ''
  try {
    const image = await generateActivityQr(props.url)
    if (current === generation) dataUrl.value = image
  } catch (cause) { if (current === generation) error.value = errorText(cause, '二维码生成失败，请重试') }
  finally { if (current === generation) loading.value = false }
}
async function download() {
  if (downloading.value || loading.value || !safeUrl.value) return
  downloading.value = true
  try {
    if (!dataUrl.value) await generate()
    if (!dataUrl.value) return
    downloadActivityQr(dataUrl.value, safeQrFilename(props.managementTitle || props.title, props.kind, props.sourceId))
    ElMessage.success('已发起 PNG 下载；手机也可放大后长按保存')
  } catch (cause) { error.value = errorText(cause, '下载失败，请重试或放大后长按保存') }
  finally { downloading.value = false }
}
function enlarge() { if (dataUrl.value) expanded.value = true }
watch(() => props.url, () => {
  ++generation; dataUrl.value = ''; error.value = ''; loading.value = false; expanded.value = false
  if (active) void generate()
})
onMounted(() => {
  if (!root.value || typeof IntersectionObserver === 'undefined') { active = true; void generate(); return }
  observer = new IntersectionObserver(entries => {
    if (entries.some(entry => entry.isIntersecting)) { active = true; observer?.disconnect(); void generate() }
  }, { rootMargin: '200px' })
  observer.observe(root.value)
})
onBeforeUnmount(() => { ++generation; observer?.disconnect() })
</script>

<template>
  <div ref="root" class="activity-qr">
    <button class="qr-thumbnail" :disabled="!dataUrl || loading" :aria-label="`放大 ${identifiedTitle} 的二维码`" @click="enlarge">
      <img v-if="dataUrl" :src="dataUrl" :alt="`${identifiedTitle}的参与二维码`" width="112" height="112" />
      <span v-else role="status">{{ loading ? '生成中…' : error ? '生成失败' : '二维码' }}</span>
    </button>
    <div class="qr-copy">
      <span>扫码在手机上参与</span>
      <p>点击二维码放大，长按可保存</p>
      <el-button :loading="downloading || loading" :disabled="!safeUrl || loading || downloading" @click="download">下载 PNG</el-button>
      <div v-if="error" role="alert" class="qr-error">{{ error }} <el-button link type="primary" :disabled="loading" @click="generate">重试</el-button></div>
    </div>
    <el-dialog v-model="expanded" :title="identifiedTitle" width="min(440px, calc(100vw - 32px))" append-to-body class="public-qr-dialog">
      <p v-if="managementTitle" class="qr-management">管理名称：{{ managementTitle }}</p>
      <img v-if="dataUrl" :src="dataUrl" :alt="`${identifiedTitle}的参与二维码，可长按保存`" class="qr-large" />
      <p class="qr-save-hint">手机长按二维码保存图片，或点击下方下载。</p>
      <template #footer><el-button type="primary" :loading="downloading" @click="download">下载 1024 × 1024 PNG</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.activity-qr { display: flex; align-items: center; gap: var(--admin-ui-space-3); padding-top: var(--admin-ui-space-4); border-top: 1px solid var(--admin-ui-color-border); }
.qr-thumbnail { flex: 0 0 112px; width: 112px; height: 112px; padding: 0; border: 0; background: white; color: var(--admin-ui-color-text-secondary); cursor: zoom-in; }
.qr-thumbnail:disabled { cursor: default; }
.qr-thumbnail:focus-visible { outline: 3px solid var(--admin-ui-color-primary); outline-offset: 2px; }
.qr-thumbnail img { display: block; width: 100%; height: 100%; }
.qr-copy { min-width: 0; font-size: var(--admin-ui-font-size-caption); line-height: 1.6; }
.qr-copy p { color: var(--admin-ui-color-text-secondary); margin: var(--admin-ui-space-1) 0 var(--admin-ui-space-2); }
.qr-copy .el-button { min-height: 44px; }
.qr-error { color: var(--admin-ui-color-danger); overflow-wrap: anywhere; }
.qr-large { display: block; width: 100%; height: auto; background: white; -webkit-touch-callout: default; }
.qr-save-hint { text-align: center; color: var(--admin-ui-color-text-secondary); }
.qr-management { margin: 0 0 var(--admin-ui-space-3); color: var(--admin-ui-color-text-secondary); overflow-wrap: anywhere; }
</style>
