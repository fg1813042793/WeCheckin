<template>
  <view class="main-admin">
    <view class="form-box">
      <view class="checkin" v-if="qrUrl">
        <view class="notice"><text class="icon-scan margin-right-s"></text>放在推广的地方展示</view>
        <image :src="qrUrl" mode="aspectFill" class="loading" @click="preview" />
        <view class="oprt">长按图片保存小程序码</view>
        <view v-if="title" class="oprt title">《{{ title }}》小程序码</view>
      </view>
      <view class="checkin" v-else>
        <view class="oprt">{{ errorMsg || '加载中...' }}</view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      qrUrl: '',
      title: '',
      errorMsg: ''
    }
  },

  onLoad(options) {
    if (options.title) {
      this.title = decodeURIComponent(options.title)
    }
    if (options.qr) {
      this.qrUrl = decodeURIComponent(options.qr)
    }
    this.loadQr()
  },

  methods: {
    async loadQr() {
      if (this.qrUrl) return

      try {
        const res = await adminApi.setupQr()
        if (res && res.data) {
          this.qrUrl = res.data
        } else {
          this.errorMsg = '二维码生成暂不开放'
        }
      } catch (e) {
        this.errorMsg = '加载失败'
        console.error('加载二维码失败', e)
      }
    },

    preview() {
      uni.previewImage({
        urls: [this.qrUrl]
      })
    }
  }
}
</script>

<style scoped>
.main-admin {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
}
.form-box {
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}
.checkin {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.notice {
  font-size: 28rpx;
  color: #333;
  margin-bottom: 30rpx;
}
image {
  width: 400rpx;
  height: 400rpx;
  border-radius: 16rpx;
  background: #f0f0f0;
}
.oprt {
  font-size: 26rpx;
  color: #999;
  margin-top: 20rpx;
}
.oprt.title {
  color: #333;
  font-weight: bold;
}
</style>
