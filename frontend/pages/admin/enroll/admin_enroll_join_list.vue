<template>
  <view class="container">
    <view class="header">
      <text class="header-title">{{ enrollTitle || '打卡记录管理' }}</text>
    </view>

    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索用户名称" class="search-input" @confirm="handleSearch" />
      <text class="search-btn" @click="handleSearch">搜索</text>
    </view>

    <scroll-view scroll-y class="content" @scrolltolower="loadMore">
        <view class="table" v-if="list.length > 0">
          <view class="table-header">
            <text class="col col-user">用户</text>
            <text class="col col-day">打卡日期</text>
            <text class="col col-content">打卡内容</text>
            <text class="col col-action">操作</text>
          </view>
          <view class="table-row" v-for="(item, index) in list" :key="index" @click="showDetail(item)">
            <text class="col col-user">{{ item.enrollTitle || item.userId || '-' }}</text>
            <text class="col col-day">{{ item.day || '-' }}</text>
            <text class="col col-content">{{ formatForms(item.forms) }}</text>
            <text class="col col-action action-del" @click.stop="handleDelete(item)">删除</text>
          </view>
        </view>

      <view class="empty" v-else-if="!loading">
        <text class="empty-text">暂无打卡记录</text>
      </view>
      <view class="loading-more" v-if="loading">
        <text class="loading-text">加载中...</text>
      </view>
    </scroll-view>

    <view class="overlay" v-if="detailItem" @click="closeDetail">
      <view class="detail-card" @click.stop>
        <text class="detail-title">打卡详情</text>
        <view class="detail-body">
          <view class="detail-row">
            <text class="detail-label">用户</text>
            <text class="detail-value">{{ detailItem.enrollTitle || detailItem.userId || '-' }}</text>
          </view>
          <view class="detail-row">
            <text class="detail-label">用户ID</text>
            <text class="detail-value">{{ detailItem.userId || '-' }}</text>
          </view>
          <view class="detail-row">
            <text class="detail-label">打卡日期</text>
            <text class="detail-value">{{ detailItem.day || '-' }}</text>
          </view>
          <view class="detail-row" v-if="detailItem._createTime">
            <text class="detail-label">打卡时间</text>
            <text class="detail-value">{{ formatTime(detailItem._createTime) }}</text>
          </view>
          <view class="detail-section" v-if="detailItem.forms">
            <text class="detail-section-title">表单内容</text>
            <view class="form-row" v-for="(f, fi) in formDataArr(detailItem.forms)" :key="fi">
              <text class="form-label">{{ f.label }}：</text>
              <text class="form-value">{{ f.value || '-' }}</text>
            </view>
          </view>
        </view>
        <view class="detail-actions">
          <view class="detail-btn btn-close" @click="closeDetail">关闭</view>
          <view class="detail-btn btn-del" @click="handleDelete(detailItem)">删除</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../../api/admin'

export default {
  data() {
    return {
      enrollId: '',
      enrollTitle: '',
      list: [],
      page: 1,
      pageSize: 20,
      hasMore: true,
      keyword: '',
      loading: false,
      detailItem: null
    }
  },

  onLoad(options) {
    this.enrollId = options.enrollId
    this.loadEnrollTitle()
    this.loadData()
  },

  methods: {
    async loadEnrollTitle() {
      try {
        const res = await adminApi.enrollDetail(this.enrollId)
        const data = res.data || {}
        this.enrollTitle = data.title || ''
      } catch (e) {
        console.error('加载项目名称失败', e)
      }
    },


    handleSearch() {
      this.page = 1
      this.list = []
      this.hasMore = true
      this.loadData()
    },

    async loadData() {
      if ((!this.hasMore && this.page > 1) || this.loading) return
      this.loading = true
      try {
        const res = await adminApi.enrollJoinList({ enrollId: this.enrollId, keyword: this.keyword, page: this.page, pageSize: this.pageSize })
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        const total = res.data.total || 0
        if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
        this.hasMore = this.list.length < total
      } catch (e) {
        console.error('加载打卡记录失败', e)
      } finally {
        this.loading = false
      }
    },

    loadMore() {
      if (this.hasMore && !this.loading) {
        this.page++
        this.loadData()
      }
    },

    showDetail(item) {
      this.detailItem = item
    },

    closeDetail() {
      this.detailItem = null
    },

    formDataArr(forms) {
      if (!forms) return []
      try {
        const arr = JSON.parse(forms)
        return Array.isArray(arr) ? arr : []
      } catch {
        return []
      }
    },

    formatTime(ts) {
      if (!ts || ts === 0) return '-'
      const d = new Date(ts)
      const pad = n => String(n).padStart(2, '0')
      return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
    },

    formatForms(forms) {
      if (!forms) return '-'
      try {
        const arr = JSON.parse(forms)
        return arr.map(v => v.value || v.label).join(' / ')
      } catch {
        return forms
      }
    },

    handleDelete(item) {
      uni.showModal({
        title: '提示',
        content: '确定要删除该打卡记录吗？',
        success: async (res) => {
          if (res.confirm) {
            try {
              await adminApi.enrollJoinDel({ enrollJoinId: item.id })
              uni.showToast({ title: '删除成功', icon: 'success' })
              this.list = this.list.filter(v => v.id !== item.id)
            } catch (e) {
              console.error('删除失败', e)
            }
          }
        }
      })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  background-color: #fff;
  padding: 30rpx;
  border-bottom: 2rpx solid #eee;
}

.header-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #333;
}

.search-bar {
  display: flex;
  align-items: center;
  padding: 16rpx 20rpx;
  background-color: #fff;
  border-bottom: 2rpx solid #f0f0f0;
}

.search-input {
  flex: 1;
  height: 60rpx;
  background-color: #f5f5f5;
  border-radius: 30rpx;
  padding: 0 24rpx;
  font-size: 26rpx;
  color: #333;
}

.search-btn {
  font-size: 26rpx;
  color: #fb454c;
  margin-left: 16rpx;
  flex-shrink: 0;
}

.content {
  height: calc(100vh - 170rpx);
}

.loading-more {
  text-align: center;
  padding: 20rpx;
}

.loading-text {
  font-size: 24rpx;
  color: #999;
}
.content::-webkit-scrollbar {
  display: none;
}

.table {
  background-color: #fff;
  margin: 20rpx;
  border-radius: 16rpx;
  overflow: hidden;
}

.table-header {
  display: flex;
  background-color: #fafafa;
  padding: 20rpx 16rpx;
  border-bottom: 2rpx solid #f0f0f0;
}

.table-row {
  display: flex;
  padding: 20rpx 16rpx;
  border-bottom: 2rpx solid #f5f5f5;
}

.table-row:last-child {
  border-bottom: none;
}

.col {
  font-size: 26rpx;
  color: #333;
}

.col-user {
  flex: 2;
}

.col-day {
  flex: 2;
}

.col-content {
  flex: 3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-action {
  flex: 1;
  text-align: center;
}

.action-del {
  color: #fb454c;
}

.overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-card {
  width: 85%;
  max-height: 80vh;
  background-color: #fff;
  border-radius: 20rpx;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.detail-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  text-align: center;
  padding: 30rpx;
  border-bottom: 2rpx solid #f0f0f0;
}

.detail-body {
  flex: 1;
  overflow-y: auto;
  padding: 20rpx 30rpx;
}

.detail-row {
  display: flex;
  padding: 12rpx 0;
  border-bottom: 2rpx solid #f5f5f5;
}

.detail-label {
  font-size: 26rpx;
  color: #999;
  width: 140rpx;
  flex-shrink: 0;
}

.detail-value {
  font-size: 26rpx;
  color: #333;
  flex: 1;
}

.detail-section {
  margin-top: 20rpx;
}

.detail-section-title {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 12rpx;
  display: block;
}

.form-row {
  padding: 8rpx 0;
}

.form-label {
  font-size: 26rpx;
  color: #666;
}

.form-value {
  font-size: 26rpx;
  color: #333;
}

.detail-actions {
  display: flex;
  border-top: 2rpx solid #f0f0f0;
}

.detail-btn {
  flex: 1;
  text-align: center;
  padding: 24rpx 0;
  font-size: 28rpx;
}

.btn-close {
  color: #666;
}

.btn-del {
  color: #fb454c;
  border-left: 2rpx solid #f0f0f0;
}

.empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 200rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}
</style>
