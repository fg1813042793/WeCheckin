<template>
  <view class="container">
    <view class="page-header">
      <text class="page-title">添加管理员</text>
    </view>

    <view class="form">
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
        <text class="form-label">密码</text>
        <input class="form-input" v-model="form.password" placeholder="请输入密码" type="password" />
      </view>
    </view>

    <view class="form-actions">
      <button class="btn btn-submit" @click="submit">提交</button>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      form: {
        name: '',
        desc: '',
        phone: '',
        password: ''
      }
    }
  },

  methods: {
    async submit() {
      if (!this.form.name) {
        uni.showToast({ title: '请输入姓名', icon: 'none' })
        return
      }
      if (!this.form.password) {
        uni.showToast({ title: '请输入密码', icon: 'none' })
        return
      }

      try {
        await adminApi.mgrInsert(this.form)
        uni.showToast({ title: '添加成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('添加管理员失败', e)
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
