<template>
  <view class="container">
    <view class="search-bar">
      <image src="/static/icons/back.png" mode="aspectFit" class="back-icon" @click="goBack"></image>
      <input 
        type="text" 
        v-model="keyword" 
        placeholder="搜索打卡项目、通知等"
        class="search-input"
        confirm-type="search"
        @confirm="doSearch"
        focus
      />
      <text class="search-btn" @click="doSearch">搜索</text>
    </view>

    <view class="history" v-if="historyList.length > 0 && !keyword">
      <view class="history-header">
        <text class="history-title">搜索历史</text>
        <image src="/static/icons/delete.png" mode="aspectFit" class="delete-icon" @click="clearHistory"></image>
      </view>
      <view class="history-list">
        <text 
          class="history-item" 
          v-for="(item, index) in historyList" 
          :key="index"
          @click="searchByHistory(item)"
        >{{ item }}</text>
      </view>
    </view>

    <view class="result-list" v-if="keyword">
      <view class="result-item" v-for="(item, index) in resultList" :key="index" @click="goDetail(item)">
        <text class="result-title">{{ item.title }}</text>
        <text class="result-type">{{ item.type === 'enroll' ? '打卡任务' : '通知' }}</text>
      </view>

      <view class="empty" v-if="resultList.length === 0 && !loading">
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">未找到相关内容</text>
      </view>
    </view>
  </view>
</template>

<script>
import { homeApi, enrollApi, newsApi } from '../../api/index'

export default {
  data() {
    return {
      keyword: '',
      historyList: [],
      resultList: [],
      loading: false
    }
  },

  onLoad() {
    this.historyList = uni.getStorageSync('searchHistory') || []
  },

  methods: {
    goBack() {
      uni.navigateBack()
    },

    doSearch() {
      if (!this.keyword.trim()) return
      
      this.saveHistory()
      this.search()
    },

    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async search() {
      this.loading = true
      try {
        const uid = this.getUserId()
        const [enrollRes, newsRes] = await Promise.all([
          enrollApi.getList({ user_id: uid }),
          newsApi.getList({ user_id: uid })
        ])

        const enrollRaw = Array.isArray(enrollRes.data) ? enrollRes.data : (enrollRes.data.list || [])
        const newsRaw = Array.isArray(newsRes.data) ? newsRes.data : (newsRes.data.list || [])

        const enrollList = enrollRaw.map(item => ({
          ...item,
          type: 'enroll'
        }))
        
        const newsList = newsRaw.map(item => ({
          ...item,
          type: 'news'
        }))

        this.resultList = [...enrollList, ...newsList]
      } catch (e) {
        console.error('搜索失败', e)
      } finally {
        this.loading = false
      }
    },

    saveHistory() {
      const keyword = this.keyword.trim()
      if (!keyword) return

      let history = [...this.historyList]
      const index = history.indexOf(keyword)
      if (index > -1) {
        history.splice(index, 1)
      }
      history.unshift(keyword)
      
      if (history.length > 10) {
        history = history.slice(0, 10)
      }

      this.historyList = history
      uni.setStorageSync('searchHistory', history)
    },

    searchByHistory(keyword) {
      this.keyword = keyword
      this.doSearch()
    },

    clearHistory() {
      uni.showModal({
        title: '提示',
        content: '确定清空搜索历史吗？',
        success: (res) => {
          if (res.confirm) {
            this.historyList = []
            uni.removeStorageSync('searchHistory')
          }
        }
      })
    },

    goDetail(item) {
      const id = item.id || item._id
      if (item.type === 'enroll') {
        uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${id}` })
      } else {
        uni.navigateTo({ url: `/pages/news/news_detail?id=${id}` })
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.search-bar {
  display: flex;
  align-items: center;
  padding: 20rpx 30rpx;
  background-color: #fff;
}

.back-icon {
  width: 40rpx;
  height: 40rpx;
  margin-right: 20rpx;
}

.search-input {
  flex: 1;
  height: 70rpx;
  background-color: #f5f5f5;
  border-radius: 35rpx;
  padding: 0 30rpx;
  font-size: 28rpx;
}

.search-btn {
  color: #fb454c;
  font-size: 30rpx;
  margin-left: 20rpx;
  white-space: nowrap;
}

.history {
  margin: 20rpx;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.history-title {
  font-size: 30rpx;
  color: #333;
  font-weight: bold;
}

.delete-icon {
  width: 36rpx;
  height: 36rpx;
}

.history-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.history-item {
  padding: 12rpx 28rpx;
  background-color: #f5f5f5;
  border-radius: 30rpx;
  font-size: 26rpx;
  color: #666;
}

.result-list {
  padding: 20rpx;
}

.result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #fff;
  padding: 28rpx 30rpx;
  margin-bottom: 16rpx;
  border-radius: 12rpx;
}

.result-title {
  font-size: 30rpx;
  color: #333;
  flex: 1;
  margin-right: 20rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-type {
  font-size: 24rpx;
  color: #fb454c;
  background-color: rgba(251, 69, 76, 0.1);
  padding: 6rpx 18rpx;
  border-radius: 20rpx;
  white-space: nowrap;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 200rpx;
}

.empty-img {
  width: 240rpx;
  height: 240rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
  margin-top: 30rpx;
}
</style>