<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchPublicActivities, fetchPublicPortalPreview } from '@/api/publicPortal'
import type { PublicActivities } from '@/types/publicPortal'
import { activityFilters, errorText, filterActivities, isPreviewQuery, safeLogoUrl } from '../public-portal/portalModel'
import ActivityCard from './ActivityCard.vue'

const route = useRoute()
const preview = computed(() => isPreviewQuery(route.query.preview))
const data = ref<PublicActivities | null>(null)
const loading = ref(false)
const error = ref('')
const category = ref('all')
const keyword = ref('')
const logoFailed = ref(false)
const logo = computed(() => safeLogoUrl(data.value?.logoUrl || ''))
const title = computed(() => data.value?.title || '公共活动')
const filters = computed(() => activityFilters(data.value?.items || []))
const categorySummary = computed(() => filters.value.slice(1).map(option => option.label).join(' · '))
const filtered = computed(() => filterActivities(data.value?.items || [], category.value, keyword.value))
const originalTitle = document.title
let requestId = 0
let controller: AbortController | undefined
async function load() {
  const current = ++requestId
  controller?.abort()
  controller = new AbortController()
  loading.value = true
  error.value = ''
  data.value = null
  logoFailed.value = false
  try {
    // Read admin storage only when the explicit preview query has been selected.
    const result = preview.value
      ? await fetchPublicPortalPreview(localStorage.getItem('admin_token') || '', controller.signal)
      : await fetchPublicActivities(controller.signal)
    if (current === requestId) {
      data.value = result
      if (!filters.value.some(option => option.value === category.value)) category.value = 'all'
    }
  } catch (cause) {
    if (current === requestId) error.value = errorText(cause, '活动加载失败，请检查网络后重试')
  } finally { if (current === requestId) loading.value = false }
}
function reset() { category.value = 'all'; keyword.value = '' }
watch(preview, () => { reset(); void load() }, { immediate: true })
watch(title, value => { document.title = value }, { immediate: true })
onBeforeUnmount(() => { ++requestId; controller?.abort(); document.title = originalTitle })
</script>

<template>
  <main class="public-activities" :aria-busy="loading">
    <div class="public-content">
      <el-alert v-if="preview" title="管理员预览 · 仅展示已保存内容" description="预览绕过页面总开关，仍检查来源公开状态、展示开关和活动时间设置。" type="warning" show-icon :closable="false" />
      <header class="portal-heading">
        <img v-if="logo && !logoFailed" :src="logo" :alt="`${title} Logo`" class="portal-logo" width="64" height="64" referrerpolicy="no-referrer" @error="logoFailed = true" />
        <span v-else class="portal-monogram" aria-hidden="true">W</span>
        <div class="heading-copy"><p v-if="categorySummary" class="eyebrow">{{ categorySummary }}</p><h1>{{ title }}</h1><p v-if="data?.description" class="portal-description">{{ data.description }}</p></div>
      </header>
      <section v-if="loading" class="page-state" role="status" aria-label="正在加载活动"><el-skeleton :rows="6" animated /><p>正在加载活动…</p></section>
      <section v-else-if="error" class="page-state" role="alert"><el-alert :title="error" type="error" show-icon :closable="false" /><el-button type="primary" @click="load">重新加载</el-button></section>
      <section v-else-if="data && !data.enabled && !preview" class="page-state"><el-empty description="公共入口暂未开放，请稍后再来。" /><el-button @click="load">刷新</el-button></section>
      <template v-else-if="data">
        <section class="portal-filters" aria-label="活动筛选">
          <el-radio-group v-model="category" aria-label="内容分类" class="category-filters"><el-radio-button v-for="option in filters" :key="option.value" :value="option.value">{{ option.label }}</el-radio-button></el-radio-group>
          <el-input v-model="keyword" clearable placeholder="搜索名称、说明或编号" aria-label="搜索名称、说明或编号" class="search-input" />
          <span class="result-count" role="status">{{ filtered.length }} 项活动</span>
          <el-button @click="load">刷新</el-button>
        </section>
        <div v-if="filtered.length" class="activity-grid"><ActivityCard v-for="item in filtered" :key="item.id" :item="item" /></div>
        <section v-else class="page-state"><el-empty :description="data.items.length ? '没有匹配的活动，换个关键词试试。' : '暂无可展示的活动，欢迎稍后再来。'" /><el-button v-if="data.items.length" @click="reset">清除筛选</el-button></section>
      </template>
      <footer class="public-footer">WeCheckin · 让参与更简单</footer>
    </div>
  </main>
</template>

<style scoped>
.public-activities { min-height: 100dvh; box-sizing: border-box; background: var(--admin-ui-color-surface-muted); color: var(--admin-ui-color-text); padding: var(--admin-ui-space-8) var(--admin-ui-page-gutter) max(var(--admin-ui-space-6), env(safe-area-inset-bottom)); }
.public-content { max-width: 1320px; margin: 0 auto; display: grid; gap: var(--admin-ui-space-6); }
.portal-heading { display: flex; gap: var(--admin-ui-space-5); align-items: flex-start; padding: var(--admin-ui-space-6) 0; }
.portal-logo, .portal-monogram { flex: 0 0 64px; width: 64px; height: 64px; object-fit: contain; border-radius: var(--admin-ui-radius-lg); background: var(--admin-ui-color-surface); }
.portal-monogram { display: grid; place-items: center; color: var(--admin-ui-color-primary); background: var(--admin-ui-color-primary-soft); font-size: 32px; font-weight: 750; }
.heading-copy { min-width: 0; }
.eyebrow { margin: 0 0 var(--admin-ui-space-2); color: var(--admin-ui-color-primary); font-size: var(--admin-ui-font-size-caption); letter-spacing: .12em; font-weight: 600; line-height: 1.6; overflow-wrap: anywhere; }
h1 { margin: 0; font-size: clamp(24px, 3vw, 36px); line-height: 1.35; overflow-wrap: anywhere; }
.portal-description { white-space: pre-wrap; overflow-wrap: anywhere; color: var(--admin-ui-color-text-secondary); margin: var(--admin-ui-space-3) 0 0; line-height: 1.75; }
.portal-filters { display: flex; align-items: center; flex-wrap: wrap; gap: var(--admin-ui-space-3); }
.category-filters { max-width: 100%; row-gap: var(--admin-ui-space-2); }
.category-filters :deep(.el-radio-button) { max-width: 100%; }
.category-filters :deep(.el-radio-button__inner) { max-width: 100%; white-space: normal; overflow-wrap: anywhere; }
.search-input { width: min(100%, 320px); }
.result-count { margin-left: auto; color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-body); }
.activity-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--admin-ui-space-5); }
.page-state { display: grid; justify-items: center; gap: var(--admin-ui-space-4); padding: var(--admin-ui-space-8); background: var(--admin-ui-color-surface); border-radius: var(--admin-ui-radius-lg); }
.page-state > .el-skeleton, .page-state > .el-alert { width: 100%; }
.portal-filters :deep(.el-input__wrapper), .portal-filters :deep(.el-radio-button__inner), .portal-filters > .el-button, .page-state > .el-button { min-height: 44px; box-sizing: border-box; }
.portal-filters :deep(.el-radio-button__inner) { display: inline-flex; align-items: center; }
.public-footer { padding: var(--admin-ui-space-6) 0; text-align: center; color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
@media (max-width: 1199px) { .activity-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 767px) {
  .public-activities { padding: var(--admin-ui-space-4) var(--admin-ui-page-gutter-mobile); }
  .activity-grid { grid-template-columns: minmax(0, 1fr); }
  .portal-heading { gap: var(--admin-ui-space-3); padding: var(--admin-ui-space-4) 0; }
  .portal-logo, .portal-monogram { flex-basis: 48px; width: 48px; height: 48px; }
  .search-input { width: 100%; order: 1; }
  .page-state { padding: var(--admin-ui-space-4); }
}
</style>
