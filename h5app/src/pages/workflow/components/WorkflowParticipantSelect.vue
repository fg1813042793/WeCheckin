<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { resolveWorkflowSelectPlacement } from '../workflow-select-placement'

interface ParticipantOption {
  userId: string
  userName: string
}

interface ParticipantGroup {
  nodeId: string
  nodeName: string
  users: ParticipantOption[]
}

const props = withDefaults(defineProps<{
  modelValue: string[]
  groups: ParticipantGroup[]
  disabled?: boolean
  placeholder?: string
}>(), {
  disabled: false,
  placeholder: '请选择通知对象',
})

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const desktopOpen = ref(false)
const desktopPlacement = ref<'top' | 'bottom'>('bottom')
const desktopRootRef = ref<unknown>(null)
const desktopPanelRef = ref<unknown>(null)
const mobileOpen = ref(false)
const mobileDraft = ref<string[]>([])

const optionMap = computed(() => {
  const values = new Map<string, ParticipantOption>()
  for (const group of props.groups) {
    for (const user of group.users) {
      if (!values.has(user.userId))
        values.set(user.userId, user)
    }
  }
  return values
})

const selectedLabel = computed(() => {
  const names = props.modelValue
    .map(userId => optionMap.value.get(userId)?.userName || '')
    .filter(Boolean)
  if (names.length === 0)
    return ''
  if (names.length <= 2)
    return names.join('、')
  return `${names.slice(0, 2).join('、')} 等 ${names.length} 人`
})

const canClear = computed(() => !props.disabled && props.modelValue.length > 0)

function resolveMobileSelect() {
  try {
    const info = uni.getSystemInfoSync()
    const width = Number(info.windowWidth || info.screenWidth || 0)
    if (width > 0 && width <= 768)
      return true
  }
  catch {
    // 无法读取设备信息时按 PC 处理，避免宽屏出现底部弹层。
  }

  // #ifdef H5
  if (typeof window !== 'undefined' && typeof window.matchMedia === 'function')
    return window.matchMedia('(hover: none) and (pointer: coarse)').matches
  // #endif

  return false
}

function resolveH5Element(value: unknown): HTMLElement | null {
  // #ifdef H5
  if (typeof HTMLElement === 'undefined')
    return null
  if (value instanceof HTMLElement)
    return value
  if (value && typeof value === 'object' && '$el' in value) {
    const element = (value as { $el?: unknown }).$el
    if (element instanceof HTMLElement)
      return element
  }
  // #endif
  return null
}

// #ifdef H5
function resolveVisibleVerticalBounds(element: HTMLElement) {
  let visibleTop = 0
  let visibleBottom = Math.max(window.innerHeight, document.documentElement.clientHeight)
  let ancestor = element.parentElement
  while (ancestor) {
    const style = window.getComputedStyle(ancestor)
    if (['auto', 'scroll', 'hidden', 'clip'].includes(style.overflowY)) {
      const rect = ancestor.getBoundingClientRect()
      visibleTop = Math.max(visibleTop, rect.top)
      visibleBottom = Math.min(visibleBottom, rect.bottom)
    }
    ancestor = ancestor.parentElement
  }
  return { visibleTop, visibleBottom }
}
// #endif

function syncDesktopPlacement() {
  // #ifdef H5
  if (typeof window === 'undefined' || typeof document === 'undefined')
    return
  const control = resolveH5Element(desktopRootRef.value)
  if (!control)
    return
  const panel = resolveH5Element(desktopPanelRef.value)
  const controlRect = control.getBoundingClientRect()
  const { visibleTop, visibleBottom } = resolveVisibleVerticalBounds(control)
  const panelHeight = Math.min(300, panel?.scrollHeight || 300)
  desktopPlacement.value = resolveWorkflowSelectPlacement({
    controlTop: controlRect.top,
    controlBottom: controlRect.bottom,
    visibleTop,
    visibleBottom,
    panelHeight,
  })
  // #endif
}

function openSelect() {
  if (props.disabled || optionMap.value.size === 0)
    return
  if (resolveMobileSelect()) {
    desktopOpen.value = false
    mobileDraft.value = [...props.modelValue]
    mobileOpen.value = true
    return
  }
  if (desktopOpen.value) {
    desktopOpen.value = false
    return
  }
  syncDesktopPlacement()
  desktopOpen.value = true
  void nextTick(syncDesktopPlacement)
}

function selected(values: string[], userId: string) {
  return values.includes(userId)
}

function toggledValues(values: string[], userId: string) {
  return selected(values, userId)
    ? values.filter(value => value !== userId)
    : [...values, userId]
}

function toggleDesktopOption(userId: string) {
  if (!props.disabled)
    emit('update:modelValue', toggledValues(props.modelValue, userId))
}

function toggleMobileOption(userId: string) {
  if (!props.disabled)
    mobileDraft.value = toggledValues(mobileDraft.value, userId)
}

function clearSelection() {
  if (!props.disabled)
    emit('update:modelValue', [])
}

function confirmMobileSelection() {
  emit('update:modelValue', [...mobileDraft.value])
  mobileOpen.value = false
}
</script>

<template>
  <view
    ref="desktopRootRef"
    class="workflow-participant-select"
    :class="{ 'is-open': desktopOpen, 'is-disabled': disabled }"
  >
    <u-input
      type="select"
      :model-value="selectedLabel"
      :clearable="canClear"
      :disabled="disabled"
      :select-open="desktopOpen"
      :placeholder="placeholder"
      @click="openSelect"
      @clear="clearSelection"
    />

    <view v-if="desktopOpen" class="workflow-participant-select__mask" @click="desktopOpen = false" />
    <view
      v-if="desktopOpen"
      ref="desktopPanelRef"
      class="workflow-participant-select__panel"
      :class="{ 'workflow-participant-select__panel--top': desktopPlacement === 'top' }"
      @click.stop
    >
      <scroll-view scroll-y class="workflow-participant-select__list">
        <view v-for="group in groups" :key="group.nodeId" class="workflow-participant-select__group">
          <text class="workflow-participant-select__node">
            {{ group.nodeName }}
          </text>
          <view
            v-for="user in group.users"
            :key="user.userId"
            class="workflow-participant-select__option"
            :class="{ 'is-selected': selected(modelValue, user.userId) }"
            role="option"
            :aria-selected="selected(modelValue, user.userId)"
            tabindex="0"
            @click="toggleDesktopOption(user.userId)"
            @keydown.enter.prevent="toggleDesktopOption(user.userId)"
            @keydown.space.prevent="toggleDesktopOption(user.userId)"
          >
            <view class="workflow-participant-select__checkbox" :class="{ 'is-selected': selected(modelValue, user.userId) }">
              <u-icon v-if="selected(modelValue, user.userId)" name="checkbox-mark" size="12px" color="#fff" />
            </view>
            <text class="workflow-participant-select__option-label">
              {{ user.userName }}
            </text>
          </view>
        </view>
      </scroll-view>
      <view class="workflow-participant-select__footer">
        <text>已选 {{ modelValue.length }} 人</text>
        <u-button size="mini" type="primary" @click="desktopOpen = false">
          完成
        </u-button>
      </view>
    </view>

    <u-popup
      v-model="mobileOpen"
      mode="bottom"
      custom-class="workflow-participant-select__mobile-popup app-pc-control-scope"
      :border-radius="8"
      :safe-area-inset-bottom="true"
    >
      <view class="workflow-participant-select__mobile">
        <view class="workflow-participant-select__mobile-header">
          <u-button plain size="mini" @click="mobileOpen = false">
            取消
          </u-button>
          <text>选择通知对象</text>
          <u-button type="primary" size="mini" @click="confirmMobileSelection">
            确认选择
          </u-button>
        </view>
        <scroll-view scroll-y class="workflow-participant-select__mobile-list">
          <view v-for="group in groups" :key="group.nodeId" class="workflow-participant-select__group">
            <text class="workflow-participant-select__node">
              {{ group.nodeName }}
            </text>
            <view
              v-for="user in group.users"
              :key="user.userId"
              class="workflow-participant-select__option"
              :class="{ 'is-selected': selected(mobileDraft, user.userId) }"
              @click="toggleMobileOption(user.userId)"
            >
              <view class="workflow-participant-select__checkbox" :class="{ 'is-selected': selected(mobileDraft, user.userId) }">
                <u-icon v-if="selected(mobileDraft, user.userId)" name="checkbox-mark" size="12px" color="#fff" />
              </view>
              <text class="workflow-participant-select__option-label">
                {{ user.userName }}
              </text>
            </view>
          </view>
        </scroll-view>
      </view>
    </u-popup>
  </view>
</template>

<style lang="scss" scoped>
.workflow-participant-select {
  position: relative;
  width: 100%;

  &.is-open {
    z-index: 360;
  }
}

.workflow-participant-select__mask {
  position: fixed;
  inset: 0;
  z-index: 1;
  background: transparent;
}

.workflow-participant-select__panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 2;
  overflow: hidden;
  border: 1px solid #dfe5ef;
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 10px 24px rgba(31, 35, 41, 0.12);
}

.workflow-participant-select__panel--top {
  top: auto;
  bottom: calc(100% + 6px);
}

.workflow-participant-select__list {
  max-height: 244px;
}

.workflow-participant-select__group {
  padding: 6px 0;
}

.workflow-participant-select__group + .workflow-participant-select__group {
  border-top: 1px solid #f0f1f2;
}

.workflow-participant-select__node {
  display: block;
  padding: 4px 12px 6px;
  color: #86909c;
  font-size: 12px;
  line-height: 18px;
}

.workflow-participant-select__option {
  display: flex;
  min-height: 36px;
  align-items: center;
  gap: 9px;
  padding: 0 12px;
  color: #1f2329;
  cursor: pointer;

  &:hover,
  &.is-selected {
    background: #f5f8ff;
  }
}

.workflow-participant-select__checkbox {
  width: 16px;
  height: 16px;
  display: flex;
  flex: 0 0 16px;
  align-items: center;
  justify-content: center;
  border: 1px solid #c9cdd4;
  border-radius: 3px;
  background: #fff;

  &.is-selected {
    border-color: #1677ff;
    background: #1677ff;
  }
}

.workflow-participant-select__option-label {
  min-width: 0;
  overflow: hidden;
  font-size: 13px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-participant-select__footer {
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px 0 12px;
  border-top: 1px solid #edf0f3;
  color: #86909c;
  font-size: 12px;
}

.workflow-participant-select__mobile {
  padding-bottom: env(safe-area-inset-bottom);
  background: #fff;
}

.workflow-participant-select__mobile-header {
  min-height: 52px;
  display: grid;
  grid-template-columns: 76px 1fr 76px;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  border-bottom: 1px solid #edf0f3;
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
}

.workflow-participant-select__mobile-list {
  max-height: min(60vh, 480px);
}
</style>
