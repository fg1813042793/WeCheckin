<template>
  <view class="container">
    <view class="header" v-if="id">
      <text class="header-title">打卡记录</text>
    </view>
    <scroll-view scroll-y class="scroll-area" @scrolltolower="loadMore" v-if="list.length > 0">
      <view class="timeline">
        <view class="timeline-item" v-for="(item, index) in list" :key="index">
          <view class="timeline-dot" :class="{ first: index === 0 }"></view>
          <view class="timeline-line" v-if="index < list.length - 1"></view>
          <view class="timeline-content" @click="showDetail(item)">
            <view class="tl-header">
              <text class="tl-count">第{{ list.length - index }}次</text>
              <text class="tl-status" :class="'s' + item.status">{{ statusText(item.status) }}</text>
            </view>
            <text class="tl-time">{{ formatTime(item._createTime || item.addTime) }}</text>
            <text class="tl-day" v-if="item.day">日期：{{ item.day }}</text>
          </view>
        </view>
      </view>
      <view class="load-more" v-if="hasMore">
        <text>加载更多...</text>
      </view>
    </scroll-view>
    <view class="empty" v-else-if="!loading">
      <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
      <text class="empty-text">暂无打卡记录</text>
    </view>
    <view class="loading-full" v-if="loading && list.length === 0">
      <text>加载中...</text>
    </view>

    <view class="modal-mask" v-if="showModal" @click="showModal = false">
      <view class="modal-content" @click.stop>
        <text class="modal-title">打卡详情</text>
        <text class="modal-time">{{ formatTime(modalItem._createTime || modalItem.addTime) }}</text>
        <text class="modal-day" v-if="modalItem.day">打卡日期：{{ modalItem.day }}</text>
        <text class="modal-status">状态：{{ statusText(modalItem.status) }}</text>
        <view class="modal-forms" v-if="modalItem.forms">
          <text class="modal-forms-title">表单内容：</text>
          <text class="modal-forms-text">{{ modalItem.forms }}</text>
        </view>
        <view class="modal-btn" @click="showModal = false">关闭</view>
      </view>
    </view>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'

export default {
  data() {
    return {
      id: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      loading: false,
      showModal: false,
      modalItem: {}
    }
  },

  onLoad(options) {
    if (options && options.id) {
      this.id = options.id
      this.loadData()
    }
  },

  methods: {
    async loadData() {
      if (this.loading || (!this.hasMore && this.page > 1)) return
      this.loading = true
      try {
        const uid = uni.getStorageSync('userInfo')
        const token = uni.getStorageSync('token')
        const userID = (uid && (uid.miniOpenID || uid.id)) || token || ''
        const res = await enrollApi.myJoinList({ enrollId: this.id, page: this.page, pageSize: this.pageSize, user_id: userID })
        const data = res.data ? (Array.isArray(res.data) ? res.data : (res.data.list || [])) : []
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
        this.hasMore = data.length >= this.pageSize
      } catch (e) {
        console.error('加载打卡记录失败', e)
      }
      this.loading = false
    },

    loadMore() {
      if (this.hasMore && !this.loading) {
        this.page++
        this.loadData()
      }
    },

    formatTime(ts) {
      if (!ts) return '-'
      const d = new Date(Number(ts))
      const pad = n => String(n).padStart(2, '0')
      return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
    },

    statusText(status) {
      const map = { 1: '已通过', 0: '待审核', 2: '未通过' }
      return map[status] || '未知'
    },

    showDetail(item) {
      this.modalItem = item
      this.showModal = true
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; display: flex; flex-direction: column; }
.header { background-color: #fff; padding: 24rpx 30rpx; border-bottom: 1rpx solid #f0f0f0; }
.header-title { font-size: 32rpx; font-weight: bold; color: #333; }
.scroll-area { flex: 1; overflow-y: auto; }

.timeline { padding: 30rpx; position: relative; }
.timeline-item { position: relative; padding-left: 50rpx; margin-bottom: 30rpx; }
.timeline-dot {
  position: absolute; left: 8rpx; top: 12rpx;
  width: 20rpx; height: 20rpx; border-radius: 50%;
  background-color: #fb454c; z-index: 1;
}
.timeline-dot.first { width: 24rpx; height: 24rpx; left: 6rpx; top: 10rpx; background-color: #fb454c; }
.timeline-line {
  position: absolute; left: 17rpx; top: 32rpx; bottom: -30rpx;
  width: 2rpx; background-color: #e0e0e0;
}
.timeline-content {
  background-color: #fff; border-radius: 16rpx; padding: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0,0,0,0.06);
}
.tl-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8rpx; }
.tl-count { font-size: 28rpx; font-weight: bold; color: #333; }
.tl-status { font-size: 22rpx; padding: 4rpx 16rpx; border-radius: 20rpx; }
.tl-status.s1 { background-color: #f0fff0; color: #52c41a; }
.tl-status.s0 { background-color: #fff7e6; color: #fa8c16; }
.tl-status.s2 { background-color: #fff1f0; color: #f5222d; }
.tl-time { font-size: 24rpx; color: #999; display: block; }
.tl-day { font-size: 24rpx; color: #999; display: block; margin-top: 4rpx; }

.load-more { text-align: center; padding: 30rpx; font-size: 24rpx; color: #999; }

.empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding-top: 200rpx; }
.empty-img { width: 240rpx; height: 240rpx; }
.empty-text { font-size: 28rpx; color: #999; margin-top: 30rpx; }
.loading-full { display: flex; align-items: center; justify-content: center; height: 60vh; font-size: 28rpx; color: #999; }

.modal-mask { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background-color: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal-content { background-color: #fff; border-radius: 20rpx; padding: 40rpx; width: 600rpx; max-width: 80vw; }
.modal-title { font-size: 32rpx; font-weight: bold; color: #333; display: block; margin-bottom: 20rpx; text-align: center; }
.modal-time, .modal-day, .modal-status { font-size: 28rpx; color: #666; display: block; margin-bottom: 12rpx; }
.modal-forms { margin-top: 16rpx; }
.modal-forms-title { font-size: 26rpx; font-weight: bold; color: #333; display: block; margin-bottom: 8rpx; }
.modal-forms-text { font-size: 26rpx; color: #666; line-height: 1.6; white-space: pre-wrap; }
.modal-btn { margin-top: 30rpx; background-color: #fb454c; color: #fff; text-align: center; padding: 20rpx; border-radius: 12rpx; font-size: 28rpx; }
</style>