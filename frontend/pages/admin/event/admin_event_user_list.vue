<template>
  <view class="container">
    <view class="header">
      <text class="header-title">{{ title }}</text>
      <text class="header-count">共 {{ list.length }} 人参与</text>
    </view>
    <view class="list" v-if="list.length > 0">
      <view class="card" v-for="(item, i) in list" :key="i">
        <view class="card-left">
          <image v-if="item.avatar" :src="item.avatar" mode="aspectFill" class="avatar" />
          <view v-else class="avatar-placeholder">{{ (item.userName || '匿')[0] }}</view>
          <view class="card-info">
            <text class="user-name">{{ item.userName || item.nickname || '匿名' }}</text>
            <text class="user-dept">{{ item.deptName || '' }}</text>
            <text class="user-time" v-if="item.createdAt">报名时间：{{ item.createdAt }}</text>
          </view>
        </view>
        <view class="card-right" v-if="item.forms">
          <view class="form-data" @click="showForm(item)">查看表单</view>
        </view>
      </view>
    </view>
    <view class="empty" v-else-if="!loading">
      <text class="empty-text">暂无参与者</text>
    </view>
    <view class="loading" v-if="loading"><text>加载中...</text></view>

    <view class="overlay" v-if="formVisible" @click="formVisible = false">
      <view class="overlay-content" @click.stop>
        <text class="overlay-title">报名信息</text>
        <view v-for="(val, idx) in formDataList" :key="idx" class="overlay-item">
          <text class="ol-label">{{ val.label || '字段' + (idx+1) }}：</text>
          <text class="ol-val">{{ val.value }}</text>
        </view>
        <view class="overlay-close" @click="formVisible = false">关闭</view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'
export default {
  data() {
    return {
      eventId: '',
      title: '',
      list: [],
      loading: false,
      formVisible: false,
      formDataList: []
    }
  },
  onLoad(opts) {
    if (opts.event_id) { this.eventId = opts.event_id }
    if (opts.title) { this.title = decodeURIComponent(opts.title) }
    this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const res = await adminApi.eventParticipantList({ eventId: this.eventId })
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        this.list = data
      } catch (e) { console.error(e) }
      this.loading = false
    },
    showForm(item) {
      try {
        const forms = typeof item.forms === 'string' ? JSON.parse(item.forms || '[]') : (item.forms || [])
        this.formDataList = forms.map((v, i) => ({ label: '字段' + (i+1), value: typeof v === 'object' && v.label ? v.label : v }))
      } catch (e) {
        this.formDataList = [{ label: '原始数据', value: item.forms }]
      }
      this.formVisible = true
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.header { padding: 20rpx; background-color: #fff; margin-bottom: 20rpx; }
.header-title { font-size: 32rpx; font-weight: bold; color: #333; display: block; }
.header-count { font-size: 24rpx; color: #999; margin-top: 8rpx; display: block; }
.list { padding: 0 20rpx; }
.card { background-color: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 12rpx; display: flex; align-items: center; justify-content: space-between; }
.card-left { display: flex; align-items: center; flex: 1; }
.avatar { width: 64rpx; height: 64rpx; border-radius: 50%; flex-shrink: 0; }
.avatar-placeholder { width: 64rpx; height: 64rpx; border-radius: 50%; background-color: #fb454c; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 28rpx; flex-shrink: 0; }
.card-info { margin-left: 16rpx; }
.user-name { font-size: 28rpx; color: #333; font-weight: bold; display: block; }
.user-dept { font-size: 22rpx; color: #999; }
.user-time { font-size: 22rpx; color: #999; display: block; }
.card-right { flex-shrink: 0; }
.form-data { background-color: #f0f5ff; color: #2b7ef5; padding: 8rpx 20rpx; border-radius: 20rpx; font-size: 24rpx; }
.empty, .loading { display: flex; align-items: center; justify-content: center; padding-top: 200rpx; font-size: 28rpx; color: #999; }
.overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; }
.overlay-content { background-color: #fff; border-radius: 16rpx; padding: 32rpx; width: 80%; max-height: 70vh; overflow-y: auto; }
.overlay-title { font-size: 30rpx; font-weight: bold; color: #333; display: block; margin-bottom: 20rpx; text-align: center; }
.overlay-item { margin-bottom: 12rpx; }
.ol-label { font-size: 26rpx; color: #999; }
.ol-val { font-size: 26rpx; color: #333; }
.overlay-close { margin-top: 20rpx; text-align: center; color: #fb454c; font-size: 28rpx; padding: 16rpx; }
</style>
