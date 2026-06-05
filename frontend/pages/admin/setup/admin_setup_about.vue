<template>
  <view class="main-admin">
    <view class="form-box">
      <view class="form-group" style="width:100%">
        <textarea v-model="formContent" placeholder="请输入内容" style="width:100%;min-height:400rpx" />
      </view>
    </view>
    <view class="btn-bottom-admin">
      <view class="return" @click="goBack">不保存,返回</view>
      <view class="save" @click="submit">保存修改</view>
    </view>
  </view>
</template>

<script>
import { homeApi } from '../../../api/index'
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      key: '',
      title: '',
      formContent: ''
    }
  },

  onLoad(options) {
    if (options.key) {
      this.key = options.key
    }
    if (options.title) {
      this.title = decodeURIComponent(options.title)
    }
    uni.setNavigationBarTitle({
      title: '编辑' + this.title
    })
    this.loadContent()
  },

  methods: {
    async loadContent() {
      try {
        const res = await homeApi.setupGet({ key: this.key })
        if (res && res.data) {
          this.formContent = res.data
        }
      } catch (e) {
        console.error('加载内容失败', e)
      }
    },

    async submit() {
      try {
        await adminApi.setupSetContent({
          key: this.key,
          value: this.formContent
        })
        uni.showToast({ title: '修改成功', icon: 'success' })
        setTimeout(() => {
          uni.navigateBack()
        }, 1500)
      } catch (e) {
        console.error('保存失败', e)
      }
    },

    goBack() {
      uni.navigateBack()
    }
  }
}
</script>

<style scoped>
.main-admin {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20rpx;
  padding-bottom: 120rpx;
}
.form-box {
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}
.form-group textarea {
  font-size: 28rpx;
  color: #333;
  line-height: 1.6;
  box-sizing: border-box;
}
.btn-bottom-admin {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  height: 100rpx;
  background: #fff;
  border-top: 1rpx solid #eee;
}
.btn-bottom-admin .return {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  color: #999;
}
.btn-bottom-admin .save {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  color: #fff;
  background-color: #2499f2;
}
</style>
