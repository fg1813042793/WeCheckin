<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { PublicActivityItem } from '@/types/publicPortal'
import { activityCategory, activityIdentity, canOpenActivity, copyPortalLink, errorText, formatActivityTime, publicActivityUrl, stateLabels } from '../public-portal/portalModel'
import ActivityQr from './ActivityQr.vue'

const props = defineProps<{ item: PublicActivityItem }>()
const identity = computed(() => activityIdentity(props.item.kind, props.item.sourceId, props.item.category))
const safeUrl = computed(() => publicActivityUrl(props.item.url))
const canOpen = computed(() => canOpenActivity(props.item))
const statusType = computed(() => props.item.state === 'active' ? 'success' : props.item.state === 'upcoming' ? 'warning' : 'info')
const copying = ref(false)
const copyError = ref('')
async function copy() {
  if (!safeUrl.value || copying.value) return
  copying.value = true
  copyError.value = ''
  try { await copyPortalLink(safeUrl.value); ElMessage.success('活动链接已复制') }
  catch (cause) { copyError.value = errorText(cause) }
  finally { copying.value = false }
}
function guardOpen(event: MouseEvent) { if (!canOpenActivity(props.item)) event.preventDefault() }
</script>

<template>
  <article class="activity-card" :aria-labelledby="`activity-${item.id} activity-identity-${item.id}`">
    <div class="card-labels"><span class="kind-label">{{ activityCategory(item.kind, item.category) }}</span><el-tag :type="statusType" effect="light">{{ stateLabels[item.state] }}</el-tag></div>
    <h2 :id="`activity-${item.id}`">{{ item.title }}</h2>
    <p v-if="item.managementTitle" class="management-title"><span>管理名称</span>{{ item.managementTitle }}</p>
    <p :id="`activity-identity-${item.id}`" class="activity-identity">{{ identity }}</p>
    <p class="description">{{ item.description || '请查看活动安排，并按页面提示参与。' }}</p>
    <dl class="activity-times"><div><dt>开始时间</dt><dd>{{ formatActivityTime(item.startTime) }}</dd></div><div><dt>结束时间</dt><dd>{{ formatActivityTime(item.endTime) }}</dd></div></dl>
    <p class="requirements">{{ [item.loginRequired ? '需登录后参与' : '', item.passwordRequired ? '需输入访问密码' : ''].filter(Boolean).join(' · ') || '打开后按页面提示参与' }}</p>
    <div class="card-actions">
      <a v-if="canOpen" :href="safeUrl" target="_blank" rel="noopener noreferrer" class="participate" @click="guardOpen">{{ item.kind === 'survey' ? '立即填写' : '立即参加' }}</a>
      <el-button v-else type="primary" disabled>{{ item.state === 'upcoming' ? '尚未开始' : item.state === 'ended' ? '活动已结束' : '链接不可用' }}</el-button>
      <el-button :disabled="!safeUrl || copying" :loading="copying" @click="copy">复制链接</el-button>
    </div>
    <div v-if="copyError" class="copy-error" role="alert">{{ copyError }}<span class="manual-link">{{ safeUrl }}</span></div>
    <ActivityQr :url="safeUrl" :title="item.title" :management-title="item.managementTitle" :kind="item.kind" :source-id="item.sourceId" :category="item.category" />
  </article>
</template>

<style scoped>
.activity-card { display: flex; flex-direction: column; gap: var(--admin-ui-space-3); min-width: 0; padding: var(--admin-ui-space-6); border: 1px solid var(--admin-ui-color-border); border-radius: var(--admin-ui-radius-lg); background: var(--admin-ui-color-surface); }
.card-labels { display: flex; justify-content: space-between; align-items: center; gap: var(--admin-ui-space-2); }
.kind-label { min-width: 0; overflow-wrap: anywhere; color: var(--admin-ui-color-primary); background: var(--admin-ui-color-primary-soft); padding: var(--admin-ui-space-1) var(--admin-ui-space-2); border-radius: var(--admin-ui-radius-sm); font-size: var(--admin-ui-font-size-caption); font-weight: 600; }
h2 { font-size: 20px; line-height: 1.5; margin: 0; overflow-wrap: anywhere; }
.management-title { display: flex; flex-wrap: wrap; gap: var(--admin-ui-space-2); margin: 0; color: var(--admin-ui-color-text); line-height: 1.6; font-size: var(--admin-ui-font-size-body); overflow-wrap: anywhere; }
.management-title span, .activity-identity { color: var(--admin-ui-color-text-secondary); font-size: var(--admin-ui-font-size-caption); }
.activity-identity { margin: 0; line-height: 1.5; }
.description { margin: 0; white-space: pre-wrap; overflow-wrap: anywhere; font-size: var(--admin-ui-font-size-body); line-height: 1.75; color: var(--admin-ui-color-text-secondary); max-height: 7em; overflow: auto; }
.activity-times { display: grid; gap: var(--admin-ui-space-1); margin: auto 0 0; padding-top: var(--admin-ui-space-3); font-size: var(--admin-ui-font-size-caption); }
.activity-times div { display: flex; flex-wrap: wrap; gap: var(--admin-ui-space-2); }
dt { color: var(--admin-ui-color-text-secondary); }
dd { margin: 0; }
.requirements { margin: 0; font-size: var(--admin-ui-font-size-caption); color: var(--admin-ui-color-text-secondary); }
.card-actions { display: flex; gap: var(--admin-ui-space-2); flex-wrap: wrap; }
.card-actions .el-button { min-height: 44px; margin-left: 0; }
.participate { box-sizing: border-box; display: inline-flex; justify-content: center; align-items: center; min-height: 44px; padding: 0 var(--admin-ui-space-5); color: white; background: var(--admin-ui-color-primary); text-decoration: none; border-radius: var(--admin-ui-radius-sm); font-size: var(--admin-ui-font-size-body); }
.participate:hover { background: var(--admin-ui-color-primary-hover); }
.participate:focus-visible { outline: 3px solid var(--admin-ui-color-primary); outline-offset: 3px; }
.copy-error { color: var(--admin-ui-color-danger); font-size: var(--admin-ui-font-size-caption); overflow-wrap: anywhere; }
.manual-link { display: block; user-select: all; }
</style>
