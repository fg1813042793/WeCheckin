<template>
    <view />
</template>

<script>
import { hasAnyAuth } from './utils/auth'

export default {
  globalData: {
    statusBarHeight: 0,
    customBarHeight: 0,
    userInfo: null,
    token: null
  },
  onLaunch: function() {
    if (!hasAnyAuth()) {
      uni.reLaunch({ url: '/pages/login/login' })
    }
  },
  methods: {
    initApp() {
      // 获取系统信息
      const systemInfo = uni.getSystemInfoSync()
      this.globalData.statusBarHeight = systemInfo.statusBarHeight
    }
  }
}
</script>

<style>
@import '/static/css/icon.css';

page {
  background-color: #f1f1f1;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

/* 全局样式 */
.container {
  padding: 20rpx;
}

.text-center {
  text-align: center;
}

.flex {
  display: flex;
}

.flex-column {
  flex-direction: column;
}

.flex-center {
  justify-content: center;
  align-items: center;
}

.flex-between {
  justify-content: space-between;
}

/* 修复 H5 下拉刷新 passive event listener 警告 */
uni-page-wrapper, uni-page-body {
  touch-action: pan-y !important;
}
</style>
