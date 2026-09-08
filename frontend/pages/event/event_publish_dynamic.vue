<template>
  <view class="container">
    <view class="form-wrap">
      <view class="form-item">
        <text class="form-label">动态标题</text>
        <input v-model="title" placeholder="请输入动态标题" class="input" maxlength="100" />
      </view>
      <view class="form-item">
        <text class="form-label">动态内容</text>
        <textarea v-model="content" placeholder="请输入动态内容..." class="textarea" maxlength="2000" />
      </view>
      <view class="form-item">
        <text class="form-label">图片上传</text>
        <view class="upload-area">
          <view class="upload-list">
            <view class="upload-item" v-for="(img, i) in images" :key="i">
              <image :src="img" mode="aspectFill" class="upload-preview" />
              <text class="upload-del" @click="delImage(i)">×</text>
            </view>
            <view class="upload-add" @click="chooseImage" v-if="images.length < 9">
              <text class="upload-add-icon">+</text>
            </view>
          </view>
        </view>
      </view>
      <view class="form-item">
        <text class="form-label">视频上传</text>
        <view class="upload-area">
          <view class="upload-list">
            <view class="video-item" v-for="(v, i) in videos" :key="i">
              <video :src="v" class="video-preview" controls />
              <text class="upload-del" @click="delVideo(i)">×</text>
            </view>
            <view class="upload-add" @click="chooseVideo" v-if="videos.length < 3">
              <text class="upload-add-icon">+</text>
            </view>
          </view>
        </view>
      </view>
    </view>
    <view class="submit-bar">
      <view class="submit-btn" :class="{ disabled: submitting }" @click="handleSubmit">
        <text v-if="!submitting">发布动态</text>
        <text v-else>发布中...</text>
      </view>
    </view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'
import { getClientUserId } from '../../utils/auth'
import { uploadFile } from '../../utils/upload'

export default {
  data() {
    return {
      eventId: '',
      title: '',
      content: '',
      images: [],
      videos: [],
      submitting: false
    }
  },
  onLoad(options) {
    if (options.event_id) {
      this.eventId = options.event_id
    }
  },
  methods: {
    getUserId() {
      return getClientUserId()
    },
    async chooseImage() {
      const remain = 9 - this.images.length
      const res = await new Promise((resolve) => {
        uni.chooseImage({
          count: remain,
          sizeType: ['compressed'],
          sourceType: ['album', 'camera'],
          success: resolve,
          fail: () => resolve(null)
        })
      })
      if (!res) return
      const tempFiles = res.tempFilePaths || []
      for (const path of tempFiles) {
        try {
          const uploaded = await uploadFile(path, { scope: 'client' })
          this.images.push(uploaded.url)
        } catch (e) {
          uni.showToast({ title: e.message || '上传失败', icon: 'none' })
        }
      }
    },
    async chooseVideo() {
      const res = await new Promise((resolve) => {
        uni.chooseVideo({
          sourceType: ['album', 'camera'],
          compressed: true,
          maxDuration: 60,
          success: resolve,
          fail: () => resolve(null)
        })
      })
      if (!res) return
      try {
        const uploaded = await uploadFile(res.tempFilePath, { scope: 'client' })
        this.videos.push(uploaded.url)
      } catch (e) {
        uni.showToast({ title: e.message || '上传失败', icon: 'none' })
      }
    },
    delImage(i) {
      this.images.splice(i, 1)
    },
    delVideo(i) {
      this.videos.splice(i, 1)
    },
    async handleSubmit() {
      if (this.submitting) return
      if (!this.title && !this.content && this.images.length === 0 && this.videos.length === 0) {
        uni.showToast({ title: '请填写标题或内容', icon: 'none' })
        return
      }
      this.submitting = true
      try {
        const uid = this.getUserId()
        await eventApi.dynamicInsert({
          event_id: this.eventId,
          user_id: uid,
          title: this.title,
          content: this.content,
          images: JSON.stringify(this.images),
          videos: JSON.stringify(this.videos)
        })
        uni.showToast({ title: '发布成功', icon: 'success' })
        setTimeout(() => { uni.navigateBack() }, 1500)
      } catch (e) {
        console.error('发布失败', e)
      } finally {
        this.submitting = false
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}
.form-wrap {
  padding: 20rpx;
}
.form-item {
  background-color: #fff;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.form-label {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 16rpx;
}
.input {
  height: 72rpx;
  font-size: 28rpx;
  color: #333;
  border-bottom: 1rpx solid #eee;
}
.textarea {
  width: 100%;
  min-height: 240rpx;
  font-size: 26rpx;
  color: #333;
  line-height: 1.6;
}
.upload-area {
  width: 100%;
}
.upload-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.upload-item {
  position: relative;
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  overflow: hidden;
}
.upload-preview {
  width: 100%;
  height: 100%;
}
.video-item {
  position: relative;
  width: 320rpx;
  height: 240rpx;
  border-radius: 12rpx;
  overflow: hidden;
}
.video-preview {
  width: 100%;
  height: 100%;
}
.upload-del {
  position: absolute;
  top: 4rpx;
  right: 4rpx;
  width: 36rpx;
  height: 36rpx;
  background-color: rgba(0,0,0,0.5);
  color: #fff;
  border-radius: 50%;
  text-align: center;
  line-height: 36rpx;
  font-size: 28rpx;
  z-index: 1;
}
.upload-add {
  width: 180rpx;
  height: 180rpx;
  border: 2rpx dashed #ddd;
  border-radius: 12rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
}
.upload-add-icon {
  font-size: 60rpx;
  color: #ccc;
}
.submit-bar {
  padding: 20rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
}
.submit-btn {
  height: 88rpx;
  background-color: #fb454c;
  border-radius: 44rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  color: #fff;
  font-weight: bold;
}
.submit-btn.disabled {
  opacity: 0.6;
}
</style>
