<template>
  <view class="main-admin">
    <view class="form-box">
      <view class="form-group">
        <view class="title must">旧密码</view>
        <input maxlength="30" type="password" placeholder="请填写旧密码" v-model="formOldPassword" />
      </view>
      <view class="form-group">
        <view class="title must">新密码</view>
        <input maxlength="30" type="password" placeholder="请填写新密码" v-model="formPassword" />
      </view>
      <view class="form-group">
        <view class="title must">新密码再次填写</view>
        <input maxlength="30" type="password" placeholder="请再次填写新密码" v-model="formPassword2" />
      </view>
    </view>
    <button class="btn-admin margin-top" @click="submit">提交修改</button>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      formOldPassword: '',
      formPassword: '',
      formPassword2: ''
    }
  },

  methods: {
    async submit() {
      if (!this.formOldPassword) {
        uni.showToast({ title: '请填写旧密码', icon: 'none' })
        return
      }
      if (!this.formPassword) {
        uni.showToast({ title: '请填写新密码', icon: 'none' })
        return
      }
      if (this.formPassword !== this.formPassword2) {
        uni.showToast({ title: '两次输入的新密码不一致', icon: 'none' })
        return
      }
      if (this.formPassword.length < 6) {
        uni.showToast({ title: '新密码至少6位', icon: 'none' })
        return
      }

      try {
        await adminApi.mgrPwd({
          oldPassword: this.formOldPassword,
          password: this.formPassword,
          password2: this.formPassword2
        })
        uni.showToast({ title: '修改成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('修改密码失败', e)
      }
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
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}
.form-group {
  margin-bottom: 30rpx;
}
.form-group .title {
  font-size: 28rpx;
  color: #333;
  margin-bottom: 16rpx;
}
.form-group .title.must::before {
  content: '*';
  color: #f44336;
  margin-right: 4rpx;
}
input {
  height: 80rpx;
  border: 1rpx solid #eee;
  border-radius: 8rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  color: #333;
  background: #fafafa;
  box-sizing: border-box;
  width: 100%;
}
.btn-admin {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  background-color: #2499f2;
  color: #fff;
  font-size: 32rpx;
  border-radius: 16rpx;
  text-align: center;
  border: none;
  margin-top: 40rpx;
}
</style>
