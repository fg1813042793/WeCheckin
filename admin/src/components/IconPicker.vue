<template>
  <el-popover placement="bottom" :width="420" trigger="click" @show="loadIcons">
    <template #reference>
        <el-input :model-value="modelValue" placeholder="选择图标" clearable @clear="$emit('update:modelValue', '')" @input="$emit('update:modelValue', $event)">
        <template #prefix>
          <el-icon v-if="modelValue" :size="18"><component :is="iconComponent" /></el-icon>
        </template>
      </el-input>
    </template>
    <el-input v-model="keyword" size="small" class="icon-search" placeholder="搜索图标名称" clearable />
    <div v-if="filteredIcons.length" class="icon-grid">
      <div
        v-for="icon in filteredIcons" :key="icon"
        class="icon-item"
        :class="{ active: modelValue === icon }"
        :title="icon"
        @click="selectIcon(icon)"
      >
        <el-icon :size="22"><component :is="icon" /></el-icon>
        <span class="icon-name">{{ icon }}</span>
      </div>
    </div>
    <el-empty v-else description="没有匹配的图标" :image-size="48" />
  </el-popover>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue'
import { ADMIN_ICON_NAMES, resolveAdminIcon } from '../icons'

const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: string): void }>()

const iconComponent = computed(() => {
  if (!props.modelValue) return null
  return resolveAdminIcon(props.modelValue) || null
})

const icons = ref<string[]>([])
const keyword = ref('')

const filteredIcons = computed(() => filterIcons(icons.value))

function loadIcons() {
  if (icons.value.length > 0) return
  icons.value = ADMIN_ICON_NAMES
}

function filterIcons(source: string[]) {
  const value = keyword.value.trim().toLowerCase()
  return value ? source.filter(icon => icon.toLowerCase().includes(value)) : source
}

function selectIcon(name: string) {
  emit('update:modelValue', name)
}
</script>

<style scoped>
.icon-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  max-height: 360px;
  overflow-y: auto;
}
.icon-search { margin-bottom:10px; }
.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 64px;
  cursor: pointer;
  border-radius: 6px;
  border: 1px solid transparent;
  transition: all 0.2s;
}
.icon-item:hover {
  background: #ecf5ff;
  border-color: #409eff;
}
.icon-item.active {
  background: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}
.icon-name {
  font-size: 11px;
  margin-top: 2px;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 68px;
}
</style>
