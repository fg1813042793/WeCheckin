<template>
  <view class="container">
    <view class="page-header">
      <text class="page-title">编辑管理员</text>
    </view>

    <view class="form">
      <view class="avatar-section">
        <text class="form-label">头像</text>
        <view class="avatar-wrap" @click="chooseAvatar">
          <image v-if="form.pic" :src="form.pic" mode="aspectFill" class="avatar-img"></image>
          <view v-else class="avatar-placeholder">
            <text class="avatar-placeholder-text">+</text>
          </view>
        </view>
      </view>

      <view class="form-item">
        <text class="form-label">姓名</text>
        <input class="form-input" v-model="form.name" placeholder="请输入姓名" />
      </view>
      <view class="form-item">
        <text class="form-label">描述</text>
        <input class="form-input" v-model="form.desc" placeholder="请输入描述" />
      </view>
      <view class="form-item">
        <text class="form-label">手机号</text>
        <input class="form-input" v-model="form.phone" placeholder="请输入手机号" type="number" />
      </view>
      <view class="form-item">
        <text class="form-label">新密码</text>
        <input class="form-input" v-model="form.password" placeholder="留空则不修改密码" type="password" />
      </view>
    </view>

    <view class="form-actions">
      <button class="btn btn-submit" @click="submit">保存</button>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      id: null,
      form: {
        name: '',
        desc: '',
        pic: '',
        phone: '',
        password: ''
      }
    }
  },

  onLoad(options) {
    this.id = options.id
    if (this.id) {
      this.loadDetail()
    }
  },

  methods: {
    async loadDetail() {
      try {
        const res = await adminApi.mgrDetail(this.id)
        const data = res.data || {}
        this.form.name = data.name || ''
        this.form.desc = data.desc || ''
        this.form.pic = data.pic || ''
        this.form.phone = data.phone || ''
      } catch (e) {
        console.error('加载管理员信息失败', e)
      }
    },

    chooseAvatar() {
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
                this.form.pic = data.data.url
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
    },

    async submit() {
      if (!this.form.name) {
        uni.showToast({ title: '请输入姓名', icon: 'none' })
        return
      }

      try {
        const data = { id: this.id, name: this.form.name, desc: this.form.desc, pic: this.form.pic, phone: this.form.phone }
        if (this.form.password) {
          data.password = this.form.password
        }
        await adminApi.mgrEdit(data)
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('编辑管理员失败', e)
      }
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

.form {
  background-color: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 12rpx;
}

.form-input {
  height: 80rpx;
  border: 1rpx solid #eee;
  border-radius: 8rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #333;
  background-color: #fafafa;
}

.avatar-section {
  margin-bottom: 24rpx;
}

.avatar-wrap {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  overflow: hidden;
  border: 2rpx solid #eee;
}

.avatar-img {
  width: 100%;
  height: 100%;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background-color: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-placeholder-text {
  font-size: 48rpx;
  color: #ccc;
}

.form-actions {
  margin-top: 40rpx;
  padding: 0 20rpx;
}

.btn-submit {
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
