<template>
  <view class="container">
    <view class="page-header">
      <text class="page-title">数据导出</text>
    </view>

    <view class="info-card">
      <text class="card-title">当前数据文件</text>
      <view class="file-row" v-if="fileUrl">
        <text class="file-url-text">{{ fileUrl }}</text>
        <view class="file-actions">
          <button class="btn btn-sm" @click="copyUrl">复制</button>
          <button class="btn btn-sm btn-download" @click="downloadFile">下载</button>
          <button class="btn btn-sm btn-danger" @click="deleteFile">删除</button>
        </view>
      </view>
      <view class="no-file" v-else>
        <text class="no-file-text">暂无导出文件</text>
      </view>
    </view>

    <view class="action-card">
      <button class="btn btn-generate" @click="generateFile" :loading="generating">
        {{ generating ? '生成中...' : '生成导出文件' }}
      </button>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      fileUrl: '',
      generating: false
    }
  },

  onShow() {
    this.loadFile()
  },

  methods: {
    async loadFile() {
      try {
        const res = await adminApi.userDataGet()
        this.fileUrl = res.data || ''
      } catch (e) {
        console.error('获取导出文件信息失败', e)
      }
    },

    async generateFile() {
      this.generating = true
      try {
        const res = await adminApi.userDataExport()
        this.fileUrl = res.data || ''
        uni.showToast({ title: '生成成功', icon: 'success' })
      } catch (e) {
        console.error('生成导出文件失败', e)
      } finally {
        this.generating = false
      }
    },

    copyUrl() {
      uni.setClipboardData({
        data: this.fileUrl,
        success: () => {
          uni.showToast({ title: '已复制链接', icon: 'success' })
        }
      })
    },

    downloadFile() {
      if (!this.fileUrl) return
      uni.downloadFile({
        url: this.fileUrl,
        success: (res) => {
          if (res.statusCode === 200) {
            uni.saveFile({
              tempFilePath: res.tempFilePath,
              success: () => {
                uni.showToast({ title: '下载成功', icon: 'success' })
              }
            })
          }
        },
        fail: (e) => {
          console.error('下载失败', e)
          uni.showToast({ title: '下载失败', icon: 'none' })
        }
      })
    },

    deleteFile() {
      uni.showModal({
        title: '提示',
        content: '确定要删除导出的数据文件吗？',
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.userDataDel()
              this.fileUrl = ''
              uni.showToast({ title: '已删除', icon: 'success' })
            } catch (e) {
              console.error('删除文件失败', e)
            }
          }
        }
      })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
}

.page-header {
  margin-bottom: 20rpx;
}

.page-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
}

.info-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 16rpx;
}

.file-row {
  display: flex;
  flex-direction: column;
}
.file-row .file-url-text {
  margin-bottom: 16rpx;
}

.file-actions {
  display: flex;
}
.file-actions .btn {
  margin-right: 16rpx;
}
.file-actions .btn:last-child {
  margin-right: 0;
}
.file-url-text {
  font-size: 24rpx;
  color: #1976d2;
  word-break: break-all;
}

.btn {
  border: none;
  border-radius: 8rpx;
  padding: 12rpx 24rpx;
  font-size: 26rpx;
  color: #fff;
  background-color: #fb454c;
  text-align: center;
}

.btn-sm {
  font-size: 24rpx;
  padding: 8rpx 20rpx;
  background-color: #1976d2;
}

.btn-download {
  background-color: #2e7d32;
}

.btn-danger {
  background-color: #c62828;
}

.no-file {
  padding: 40rpx 0;
  text-align: center;
}

.no-file-text {
  font-size: 26rpx;
  color: #999;
}

.action-card {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.btn-generate {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background-color: #fb454c;
  color: #fff;
  font-size: 32rpx;
  border-radius: 16rpx;
  text-align: center;
}
</style>
