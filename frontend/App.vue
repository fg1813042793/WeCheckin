<template>
    <view />
</template>

<script>
export default {
  globalData: {
    statusBarHeight: 0,
    customBarHeight: 0,
    userInfo: null,
    token: null
  },
  onLaunch: function() {
    console.log('App Launch')
    
    // #ifdef APP-PLUS
    // APP平台初始化
    this.initApp()
    // #endif
  },
  onShow: function() {
    console.log('App Show')
  },
  onHide: function() {
    console.log('App Hide')
  },
  methods: {
    initApp() {
      // 获取系统信息
      const systemInfo = uni.getSystemInfoSync()
      this.globalData.statusBarHeight = systemInfo.statusBarHeight
      
      // #ifdef APP-PLUS
      if (systemInfo.platform === 'android') {
        console.log('Android 平台')
      } else if (systemInfo.platform === 'ios') {
        console.log('iOS 平台')
      }
      // #endif
      
      // #ifdef HARMONY
      console.log('鸿蒙平台')
      // #endif
    }
  }
}
</script>

<style>
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
</style>