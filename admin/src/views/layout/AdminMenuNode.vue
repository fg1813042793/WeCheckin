<template>
  <el-menu-item v-if="isMenu" :index="item.path">
    <el-icon
      v-if="showIconSlot"
      :class="{ 'admin-menu-node__icon--placeholder': !iconComponent }"
      :aria-hidden="!iconComponent"
    >
      <component v-if="iconComponent" :is="iconComponent" />
    </el-icon>
    <span>{{ item.name }}</span>
  </el-menu-item>

  <el-sub-menu v-else-if="isDirectory" :index="menuIndex">
    <template #title>
      <el-icon
        v-if="showIconSlot"
        :class="{ 'admin-menu-node__icon--placeholder': !iconComponent }"
        :aria-hidden="!iconComponent"
      >
        <component v-if="iconComponent" :is="iconComponent" />
      </el-icon>
      <span>{{ item.name }}</span>
    </template>
    <AdminMenuNode
      v-for="child in renderableChildren"
      :key="child.path || child.id"
      :item="child"
      :depth="depth + 1"
    />
  </el-sub-menu>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { AdminMenuItem } from '../../api/types'
import { resolveAdminIcon } from '../../icons'

const props = withDefaults(defineProps<{
  item: AdminMenuItem
  depth?: number
}>(), {
  depth: 0,
})

function isRenderableMenuNode(item: AdminMenuItem): boolean {
  if (item.status !== 1) return false
  if (item.type === 1) return !!item.path
  if (item.type !== 0) return false
  return (item.children || []).some(isRenderableMenuNode)
}

const renderableChildren = computed(() => (props.item.children || []).filter(isRenderableMenuNode))
const isMenu = computed(() => props.item.status === 1 && props.item.type === 1 && !!props.item.path)
const isDirectory = computed(() => props.item.status === 1 && props.item.type === 0 && renderableChildren.value.length > 0)
const menuIndex = computed(() => props.item.path || `directory:${props.item.id}`)
const iconComponent = computed(() => resolveAdminIcon(props.item.icon))
const showIconSlot = computed(() => !!iconComponent.value || props.depth > 0)
</script>

<style scoped>
.admin-menu-node__icon--placeholder {
  visibility: hidden;
}
</style>
