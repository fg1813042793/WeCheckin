<template>
  <view />
</template>

<script>
import CONFIG from './config'
import { initDingTalkBridge } from './utils/dingtalk'

export default {
  globalData: {
    appName: CONFIG.APP_NAME,
    corpId: CONFIG.DINGTALK_CORP_ID,
    dingtalkReady: false,
    userInfo: null,
    token: ''
  },
  onLaunch() {
    initDingTalkBridge().then((result) => {
      this.globalData.dingtalkReady = result.ready
    })
  }
}
</script>

<style>
@import "./styles/performance.css";

page {
  min-height: 100%;
  background: #f5f7fb;
  color: #1f2937;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

view,
text,
button,
input,
select,
textarea {
  box-sizing: border-box;
}

button {
  margin: 0;
}

.page-shell {
  min-height: 100vh;
  padding: 32rpx 32rpx 56rpx;
}
</style>
