import type { TabbarItem } from 'uview-pro/types/global'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useTabbarStore = defineStore('tabbar', () => {
  const activeIndex = ref(0)
  const tabbarList = ref<TabbarItem[]>([])

  const setActiveIndex = (index: number) => {
    activeIndex.value = index
  }

  const updateBadge = (index: number, count: number) => {
    if (!tabbarList.value[index]) {
      return
    }
    tabbarList.value[index].count = count
  }

  const updateIsDot = (index: number, isDot: boolean) => {
    if (!tabbarList.value[index]) {
      return
    }
    tabbarList.value[index].isDot = isDot
  }

  return {
    activeIndex,
    tabbarList,
    setActiveIndex,
    updateBadge,
    updateIsDot,
  }
}, {
  persist: true,
})
