<template>
  <el-menu-item v-if="isMenu" :index="item.path">
    <el-icon v-if="resolveAdminIcon(item.icon)">
      <component :is="resolveAdminIcon(item.icon)" />
    </el-icon>
    <span>{{ item.name }}</span>
  </el-menu-item>

  <el-sub-menu v-else-if="isDirectory" :index="menuIndex">
    <template #title>
      <el-icon v-if="resolveAdminIcon(item.icon)">
        <component :is="resolveAdminIcon(item.icon)" />
      </el-icon>
      <span>{{ item.name }}</span>
    </template>
    <AdminMenuNode
      v-for="child in renderableChildren"
      :key="child.path || child.id"
      :item="child"
    />
  </el-sub-menu>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import type { AdminMenuItem } from '../../api/types'
import { resolveAdminIcon } from '../../icons'

const props = defineProps<{
  item: AdminMenuItem
}>()

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
</script>
