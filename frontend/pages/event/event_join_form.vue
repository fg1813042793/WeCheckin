<template>
  <view class="container">
    <view class="form-wrap" v-if="event.id">
      <view class="event-info">
        <text class="event-title">{{ event.title }}</text>
        <text class="event-desc">{{ event.desc }}</text>
      </view>
      <view class="form-list">
        <form-render
          :schema="event.forms"
          v-model="formData"
          @change="onFormChange"
        />
      </view>
    </view>
    <view class="submit-bar">
      <view class="submit-btn" @click="handleSubmit">提交报名</view>
    </view>
  </view>
</template>

<script>
import { eventApi, formkitApi } from '../../api/index'
import FormRender from '../../components/formkit/FormRender.vue'
import { isOldSchema, normalizeSchema, initAnswers, serializeAnswers } from '../../utils/formkit.js'

export default {
  components: { FormRender },
  data() {
    return {
      id: '',
      event: {},
      formData: null,
      isOld: true
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
        const isOld = isOldSchema(this.event.forms)
        this.isOld = isOld
        const questions = normalizeSchema(this.event.forms)
        this.formData = initAnswers(questions, isOld)
      } catch (e) { console.error(e) }
    },
    onFormChange() {
      // 表单变化，可在此处触发后端实时校验（v1 暂不调）
    },
    async handleSubmit() {
      const uid = this.getUserId()
      if (!uid) { uni.showToast({ title: '请先登录', icon: 'none' }); return }
      // 提交前应用 (CalcValue + Logic) 并校验
      try {
        let answersToSubmit = this.formData
        if (!this.isOld) {
          // 新格式：跑 apply 让 CalcValue 生效
          const applyRes = await formkitApi.apply({ schema: this.event.forms, answers: this.formData || {} })
          answersToSubmit = applyRes.data.answers
          // 跑 validate
          const valRes = await formkitApi.validate({ schema: this.event.forms, answers: answersToSubmit })
          if (!valRes.data.ok) {
            const first = (valRes.data.errors || [])[0]
            uni.showToast({ title: first ? first.message : '校验失败', icon: 'none' })
            return
          }
        }
        // 老格式 / 新格式都直接序列化
        const forms = serializeAnswers(answersToSubmit)
        await eventApi.participate({ event_id: this.id, user_id: uid, forms })
        uni.showToast({ title: '报名成功', icon: 'success' })
        setTimeout(() => { uni.navigateBack() }, 1500)
      } catch (e) {
        uni.showToast({ title: e.msg || '报名失败', icon: 'none' })
      }
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
.submit-bar { padding: 20rpx; padding-bottom: calc(20rpx + env(safe-area-inset-bottom)); }
.submit-btn { height: 88rpx; background-color: #fb454c; border-radius: 44rpx; display: flex; align-items: center; justify-content: center; color: #fff; font-size: 32rpx; font-weight: bold; }
</style>
