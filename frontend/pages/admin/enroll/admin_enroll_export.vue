<template>
  <view class="container">
    <view class="card">
      <text class="card-title">导出打卡数据</text>
      <text class="desc">选择时间范围后生成 CSV 文件并自动下载，可用 Excel 打开</text>

      <view class="form-item">
        <text class="label">开始日期</text>
        <picker mode="date" :value="form.startTime" @change="onStartChange">
          <view class="picker">{{ form.startTime || '全部' }}</view>
        </picker>
      </view>

      <view class="form-item">
        <text class="label">结束日期</text>
        <picker mode="date" :value="form.endTime" @change="onEndChange">
          <view class="picker">{{ form.endTime || '全部' }}</view>
        </picker>
      </view>

      <button class="btn" :loading="generating" @click="handleExport">生成并下载</button>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      enrollId: '',
      form: {
        startTime: '',
        endTime: ''
      },
      generating: false
    }
  },

  onLoad(options) {
    this.enrollId = options.enrollId
  },

  methods: {
    onStartChange(e) {
      this.form.startTime = e.detail.value
    },
    onEndChange(e) {
      this.form.endTime = e.detail.value
    },

    async handleExport() {
      this.generating = true
      try {
        const res = await adminApi.enrollJoinDataExport({
          enrollId: this.enrollId,
          startTime: this.form.startTime,
          endTime: this.form.endTime
        })
        const url = typeof res.data === 'string' ? res.data : (res.data && res.data.url ? res.data.url : '')
        if (!url) {
          uni.showToast({ title: '获取下载链接失败', icon: 'none' })
          return
        }

        // #ifdef H5
        window.open(url, '_blank')
        // #endif

        // #ifndef H5
        uni.showModal({
          title: '下载链接',
          content: url,
          confirmText: '复制链接',
          success: (r) => {
            if (r.confirm) {
              uni.setClipboardData({ data: url })
            }
          }
        })
        // #endif

        uni.showToast({ title: '生成成功', icon: 'success' })
      } catch (e) {
        console.error('导出失败', e)
      } finally {
        this.generating = false
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 30rpx;
}

.card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 40rpx 30rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 12rpx;
}

.desc {
  font-size: 26rpx;
  color: #999;
  line-height: 1.6;
  display: block;
  margin-bottom: 30rpx;
}

.form-item {
  display: flex;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.label {
  font-size: 28rpx;
  color: #333;
  width: 140rpx;
  flex-shrink: 0;
}

.picker {
  flex: 1;
  font-size: 28rpx;
  color: #333;
  height: 60rpx;
  line-height: 60rpx;
}

.btn {
  margin-top: 40rpx;
  background-color: #fb454c;
  color: #fff;
  font-size: 30rpx;
  border-radius: 48rpx;
  height: 88rpx;
  line-height: 88rpx;
}

.btn::after {
  border: none;
}
</style>
