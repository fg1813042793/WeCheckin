<template>
  <view class="container">
    <view class="search-bar" @click="goSearch">
      <image src="/static/icons/search.png" mode="aspectFit" class="search-icon"></image>
      <text class="search-text">搜索打卡项目</text>
    </view>

    <scroll-view scroll-y class="main-content" @scrolltolower="loadMore">
      <swiper class="banner" :indicator-dots="true" :autoplay="true" :interval="3000" :duration="500">
        <swiper-item v-for="(item, index) in banners" :key="index">
          <image :src="item.url" mode="aspectFill" class="banner-img"></image>
        </swiper-item>
      </swiper>

      <view class="section">
        <view class="section-header">
          <text class="section-title">热门打卡</text>
          <text class="section-more" @click="goEnroll">更多 ></text>
        </view>
        <view class="enroll-list">
          <view class="enroll-item" v-for="(item, index) in hotList" :key="index" @click="goDetail(item.id)">
            <image :src="item.img || '/static/default.png'" mode="aspectFill" class="enroll-img"></image>
            <view class="enroll-info">
              <text class="enroll-title">{{ item.title }}</text>
              <view class="enroll-meta">
                <text class="join-count">{{ item.joinCount || 0 }}人参与</text>
                <text class="enroll-time">{{ formatTime(item.timeStart) }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="section" v-if="newList.length > 0">
        <view class="section-header">
          <text class="section-title">最新通知</text>
          <text class="section-more" @click="goNews">更多 ></text>
        </view>
        <view class="news-list">
          <view class="news-item" v-for="(item, index) in newList" :key="index" @click="goNewsDetail(item.id)">
            <text class="news-title">{{ item.title }}</text>
            <text class="news-time">{{ item._createTime }}</text>
          </view>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script>
import { homeApi } from '../../api/index'
import { pageHelper } from '../../utils/page'

export default {
  data() {
    return {
      cur: 'hot',
      banners: [],
      hotList: [],
      newList: [],
      page: 1,
      pageSize: 10
    }
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.page = 1
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    async loadData() {
      try {
        const res = await homeApi.getList()
        if (res.data) {
          this.banners = res.data.banners || []
          this.hotList = res.data.hotList || []
          this.newList = res.data.newList || []
        }
      } catch (e) {
        console.error('加载数据失败', e)
      }
    },

    loadMore() {
      this.page++
      // 加载更多数据
    },

    goSearch() {
      uni.navigateTo({ url: '/pages/search/search' })
    },

    goEnroll() {
      uni.switchTab({ url: '/pages/enroll/enroll_index' })
    },

    goNews() {
      uni.switchTab({ url: '/pages/news/news_index' })
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${id}` })
    },

    goNewsDetail(id) {
      uni.navigateTo({ url: `/pages/news/news_detail?id=${id}` })
    },

    formatTime(time) {
      return pageHelper.formatDate(time)
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
  margin: 20rpx;
  padding: 20rpx 30rpx;
  background-color: #fff;
  border-radius: 40rpx;
  display: flex;
  align-items: center;
}

.search-icon {
  width: 36rpx;
  height: 36rpx;
  margin-right: 16rpx;
}

.search-text {
  color: #999;
  font-size: 28rpx;
}

.banner {
  height: 300rpx;
  margin: 20rpx;
  border-radius: 16rpx;
  overflow: hidden;
}

.banner-img {
  width: 100%;
  height: 100%;
}

.section {
  margin: 20rpx;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.section-more {
  font-size: 26rpx;
  color: #999;
}

.enroll-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.enroll-item {
  display: flex;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.enroll-item:last-child {
  border-bottom: none;
}

.enroll-img {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  margin-right: 24rpx;
}

.enroll-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.enroll-title {
  font-size: 30rpx;
  color: #333;
  font-weight: 500;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.enroll-meta {
  display: flex;
  justify-content: space-between;
  font-size: 24rpx;
  color: #999;
}

.news-list {
  display: flex;
  flex-direction: column;
}

.news-item {
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.news-item:last-child {
  border-bottom: none;
}

.news-title {
  font-size: 28rpx;
  color: #333;
  flex: 1;
  margin-right: 20rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.news-time {
  font-size: 24rpx;
  color: #999;
}
</style>