<template>
  <view class="video-fullscreen">
    <view class="video-close-btn" @click="goBack">✕</view>
    <video :src="src" class="video-player" controls autoplay />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

const src = ref('')

onLoad((opt) => {
  if (opt.src) src.value = decodeURIComponent(opt.src)
})

function goBack() {
  uni.navigateBack()
}
</script>

<style scoped>
.video-fullscreen { position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 999; background-color: #000; display: flex; align-items: center; justify-content: center; }
.video-close-btn { position: absolute; top: 88rpx; right: 30rpx; width: 64rpx; height: 64rpx; background-color: rgba(255,255,255,0.5); color: #fff; border-radius: 50%; text-align: center; line-height: 64rpx; font-size: 40rpx; z-index: 1000; }
.video-player { width: 100%; height: 100%; }
</style>
