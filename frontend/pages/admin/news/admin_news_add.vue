<template>
  <view class="container">
    <scroll-view scroll-y class="form-scroll">
      <view class="form-group">
        <text class="form-label">标题 <text class="required">*</text></text>
        <input v-model="form.title" placeholder="请输入通知标题" class="form-input" />
      </view>

      <view class="form-group">
        <text class="form-label">分类</text>
        <picker :range="categories" range-key="name" @change="onCateChange">
          <view class="picker-value" :class="{ placeholder: !form.cateId }">{{ form.cateName || '请选择分类' }}</view>
        </picker>
      </view>

      <view class="form-group">
        <text class="form-label">排序</text>
        <input v-model="form.order" type="digit" placeholder="输入排序值（越小越靠前）" class="form-input" />
      </view>

      <view class="form-group">
        <text class="form-label">描述</text>
        <textarea v-model="form.desc" placeholder="简短描述（可选，自动生成则留空）" class="form-textarea" />
      </view>

      <view class="form-group">
        <text class="form-label">内容 <text class="required">*</text></text>
        <textarea v-model="form.content" placeholder="请输入通知内容" class="form-textarea content-area" />
      </view>

      <view class="form-group">
        <text class="form-label">封面图片</text>
        <view class="cover-upload" @click="chooseCover">
          <image v-if="form.img" :src="form.img" mode="aspectFill" class="cover-preview"></image>
          <view v-else class="cover-placeholder">
            <text class="cover-plus">+</text>
            <text class="cover-hint">点击上传封面</text>
          </view>
        </view>
      </view>

      <view class="form-actions">
        <view class="btn-submit" @click="handleSubmit">提交</view>
        <view class="btn-cancel" @click="goBack">取消</view>
      </view>
    </scroll-view>


  </view>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { adminApi, dictApi } from '../../../api/admin'

const form = reactive({
  title: '',
  cateId: 0,
  cateName: '',
  order: 0,
  desc: '',
  content: '',
  img: ''
})
const categories = ref([])

async function loadCategories() {
  try {
    const res = await dictApi.items('content_type')
    const list = Array.isArray(res.data) ? res.data : (res.data.list || [])
    categories.value = list.map(d => ({ id: d.value, name: d.label }))
  } catch (e) {
    console.error('加载分类失败', e)
  }
}

function onCateChange(e) {
  const index = e.detail.value
  const c = categories.value[index]
  if (c) {
    form.cateId = c.id
    form.cateName = c.name
  }
}

async function handleSubmit() {
  if (!form.title.trim()) {
    uni.showToast({ title: '请输入标题', icon: 'none' })
    return
  }
  if (!form.content.trim()) {
    uni.showToast({ title: '请输入内容', icon: 'none' })
    return
  }
  try {
    const data = { ...form }
    if (!data.cateId) {
      data.cateId = 0
      delete data.cateName
    }
    await adminApi.newsInsert(data)
    uni.showToast({ title: '添加成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 500)
  } catch (e) {
    console.error('添加通知失败', e)
  }
}

function chooseCover() {
  uni.chooseImage({
    count: 1,
    success: (res) => {
      const tempPath = res.tempFilePaths[0]
      uni.uploadFile({
        url: 'http://localhost:8080/upload',
        filePath: tempPath,
        name: 'file',
        header: {
          Authorization: uni.getStorageSync('admin_token') || ''
        },
        success: (uploadRes) => {
          const data = JSON.parse(uploadRes.data)
          if (data && data.data && data.data.url) {
            form.img = data.data.url
          } else {
            uni.showToast({ title: '上传失败', icon: 'none' })
          }
        },
        fail: () => {
          uni.showToast({ title: '上传失败', icon: 'none' })
        }
      })
    }
  })
}

function goBack() {
  uni.navigateBack()
}

onLoad(() => {
  loadCategories()
})
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.form-scroll {
  padding: 30rpx;
  max-width: 750rpx;
  margin: 0 auto;
  box-sizing: border-box;
}

.form-group {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
}

.form-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  font-weight: 500;
  margin-bottom: 16rpx;
}

.required {
  color: #ff4d4f;
}

.form-input {
  height: 72rpx;
  border: 1rpx solid #e8e8e8;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 26rpx;
  color: #333;
}

.form-textarea {
  width: 100%;
  min-height: 120rpx;
  border: 1rpx solid #e8e8e8;
  border-radius: 12rpx;
  padding: 20rpx;
  font-size: 26rpx;
  color: #333;
  box-sizing: border-box;
}

.content-area {
  min-height: 300rpx;
}

.picker-value {
  font-size: 28rpx;
  color: #333;
  height: 60rpx;
  line-height: 60rpx;
}

.picker-value.placeholder {
  color: #999;
}

.cover-upload {
  width: 100%;
  height: 300rpx;
  border-radius: 12rpx;
  overflow: hidden;
}

.cover-preview {
  width: 100%;
  height: 100%;
  background-color: #f0f0f0;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  border: 2rpx dashed #ddd;
  border-radius: 12rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
}

.cover-plus {
  font-size: 60rpx;
  color: #ccc;
  line-height: 1;
}

.cover-hint {
  font-size: 26rpx;
  color: #999;
  margin-top: 12rpx;
}

.form-actions {
  display: flex;
  margin-top: 40rpx;
  padding-bottom: 60rpx;
}
.form-actions .btn-submit {
  margin-right: 20rpx;
}
.form-actions .btn-submit:last-child {
  margin-right: 0;
}

.btn-submit {
  flex: 1;
  background-color: #fb454c;
  color: #fff;
  text-align: center;
  padding: 24rpx 0;
  border-radius: 40rpx;
  font-size: 30rpx;
}

.btn-cancel {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  border-radius: 40rpx;
  font-size: 30rpx;
  color: #666;
  background-color: #fff;
  border: 1rpx solid #ddd;
}

</style>
