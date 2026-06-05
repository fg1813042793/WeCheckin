<template>
  <view class="container">
    <view class="form-wrap" v-if="event.id">
      <view class="event-info">
        <text class="event-title">{{ event.title }}</text>
        <text class="event-desc">{{ event.desc }}</text>
      </view>
      <view class="form-list">
        <view class="form-item" v-for="(field, i) in formFields" :key="i">
          <text class="form-label">{{ field.label }}<text v-if="field.required" class="required">*</text></text>
          <input v-if="field.type === 'input' || field.type === 'text'" v-model="formData[i]" :placeholder="field.placeholder || '请输入'" class="form-input" />
          <textarea v-else-if="field.type === 'textarea'" v-model="formData[i]" :placeholder="field.placeholder || '请输入'" class="form-textarea" />
          <picker v-else-if="field.type === 'select' || field.type === 'picker'" :range="field.options || []" @change="(e) => { formData[i] = (field.options || [])[e.detail.value] }">
            <view class="form-picker">{{ formData[i] || (field.placeholder || '请选择') }}</view>
          </picker>
          <input v-else-if="field.type === 'number'" v-model="formData[i]" type="number" :placeholder="field.placeholder || '请输入'" class="form-input" />
          <input v-else v-model="formData[i]" :placeholder="field.placeholder || '请输入'" class="form-input" />
        </view>
      </view>
    </view>
    <view class="submit-bar">
      <view class="submit-btn" @click="handleSubmit">提交报名</view>
    </view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'
export default {
  data() {
    return {
      id: '',
      event: {},
      formFields: [],
      formData: []
    }
  },
  onLoad(opts) {
    if (opts.id) { this.id = opts.id; this.loadEvent() }
  },
  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },
    async loadEvent() {
      try {
        const uid = this.getUserId()
        const res = await eventApi.getDetail({ id: this.id, user_id: uid })
        this.event = res.data || {}
        let forms = []
        try { forms = typeof this.event.forms === 'string' ? JSON.parse(this.event.forms || '[]') : (this.event.forms || []) } catch (e) {}
        this.formFields = forms
        this.formData = forms.map(f => '')
      } catch (e) { console.error(e) }
    },
    async handleSubmit() {
      const uid = this.getUserId()
      if (!uid) { uni.showToast({ title: '请先登录', icon: 'none' }); return }
      for (let i = 0; i < this.formFields.length; i++) {
        if (this.formFields[i].required && !this.formData[i]) {
          uni.showToast({ title: '请填写' + this.formFields[i].label, icon: 'none' })
          return
        }
      }
      try {
        await eventApi.participate({ event_id: this.id, user_id: uid, forms: JSON.stringify(this.formData) })
        uni.showToast({ title: '报名成功', icon: 'success' })
        setTimeout(() => { uni.navigateBack() }, 1500)
      } catch (e) { uni.showToast({ title: '报名失败', icon: 'none' }) }
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.form-wrap { padding: 20rpx; }
.event-info { background-color: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 20rpx; }
.event-title { font-size: 32rpx; font-weight: bold; color: #333; display: block; margin-bottom: 8rpx; }
.event-desc { font-size: 26rpx; color: #666; }
.form-list { background-color: #fff; border-radius: 16rpx; padding: 20rpx; }
.form-item { margin-bottom: 24rpx; }
.form-label { font-size: 28rpx; color: #333; display: block; margin-bottom: 12rpx; }
.required { color: #fb454c; }
.form-input { height: 72rpx; background-color: #f5f5f5; border-radius: 12rpx; padding: 0 20rpx; font-size: 26rpx; }
.form-textarea { width: 100%; height: 160rpx; background-color: #f5f5f5; border-radius: 12rpx; padding: 16rpx 20rpx; font-size: 26rpx; box-sizing: border-box; }
.form-picker { height: 72rpx; background-color: #f5f5f5; border-radius: 12rpx; padding: 0 20rpx; font-size: 26rpx; line-height: 72rpx; color: #999; }
.submit-bar { padding: 20rpx; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); }
.submit-btn { height: 88rpx; background-color: #fb454c; border-radius: 44rpx; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 32rpx; font-weight: bold; }
</style>
