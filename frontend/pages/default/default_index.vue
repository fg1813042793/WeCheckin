<template>
  <view class="container">
    <view class="banner-section">
      <image src="/static/images/default-banner.png" mode="aspectFill" class="banner-img"></image>
    </view>

    <view class="menu-section">
      <view class="menu-item" v-for="(menu, index) in categoryMenus" :key="index" @click="goEnroll(menu.id)">
        <image :src="menu.icon" mode="aspectFit" class="menu-icon"></image>
        <text class="menu-text">{{ menu.name }}</text>
      </view>
    </view>

    <view class="section" v-if="vouchList.length > 0">
      <view class="section-header">
        <text class="section-title">推荐</text>
      </view>
      <scroll-view scroll-x class="vouch-scroll">
        <view class="vouch-list">
          <view class="vouch-item" v-for="(item, index) in vouchList" :key="index" @click="goDetail(item.id)">
            <image :src="item.img || '/static/default.png'" mode="aspectFill" class="vouch-img"></image>
            <text class="vouch-title">{{ item.title }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="tabs">
      <view class="tab-item" :class="{ active: curTab === 'hot' }" @click="switchTab('hot')">热门</view>
      <view class="tab-item" :class="{ active: curTab === 'new' }" @click="switchTab('new')">最新</view>
    </view>

    <view class="list-section">
      <view class="enroll-list" v-if="list.length > 0">
        <view class="enroll-card" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
          <image :src="item.img || '/static/default.png'" mode="aspectFill" class="card-img"></image>
          <view class="card-body">
            <text class="card-title">{{ item.title }}</text>
            <text class="card-desc">{{ item.desc || '' }}</text>
            <view class="card-footer">
              <text class="join-count">{{ item.joinCount || 0 }}人参与</text>
            </view>
          </view>
        </view>
      </view>

      <view class="empty" v-else>
        <text class="empty-text">暂无数据</text>
      </view>
    </view>
  </view>
</template>

<script>
import { homeApi } from '../../api/index'

export default {
  data() {
    return {
      categoryMenus: [
        { id: 1, name: '问卷', icon: '/static/images/menu/1.png' },
        { id: 2, name: '打卡', icon: '/static/images/menu/2.png' },
        { id: 3, name: '活动', icon: '/static/images/menu/3.png' },
        { id: 4, name: '比赛', icon: '/static/images/menu/4.png' },
        { id: 5, name: '工作', icon: '/static/images/menu/5.png' }
      ],
      vouchList: [],
      curTab: 'hot',
      list: [],
      hotList: [],
      newList: []
    }
  },

  onLoad() {
    this.loadData()
  },

  onPullDownRefresh() {
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    async loadData() {
      try {
        const res = await homeApi.getList()
        if (res.data) {
          this.vouchList = res.data.vouchList || []
          this.hotList = res.data.hotList || []
          this.newList = res.data.newList || []
          this.updateList()
        }
      } catch (e) {
        console.error('加载数据失败', e)
      }
    },

    switchTab(tab) {
      this.curTab = tab
      this.updateList()
    },

    updateList() {
      this.list = this.curTab === 'hot' ? this.hotList : this.newList
    },

    goEnroll(categoryId) {
      uni.navigateTo({ url: `/pages/enroll/enroll_index?categoryId=${categoryId}` })
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${id}` })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.banner-section {
  height: 320rpx;
}

.banner-img {
  width: 100%;
  height: 100%;
}

.menu-section {
  display: flex;
  background-color: #fff;
  padding: 30rpx 20rpx;
  margin-bottom: 20rpx;
}

.menu-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.menu-icon {
  width: 80rpx;
  height: 80rpx;
  margin-bottom: 12rpx;
}

.menu-text {
  font-size: 26rpx;
  color: #333;
}

.section {
  margin: 20rpx;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
}

.section-header {
  margin-bottom: 24rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.vouch-scroll {
  white-space: nowrap;
}

.vouch-list {
  display: flex;
}
.vouch-list .vouch-item {
  margin-right: 20rpx;
}
.vouch-list .vouch-item:last-child {
  margin-right: 0;
}

.vouch-item {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  width: 160rpx;
}

.vouch-img {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  margin-bottom: 12rpx;
}

.vouch-title {
  font-size: 26rpx;
  color: #333;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  width: 100%;
}

.tabs {
  display: flex;
  background-color: #fff;
  margin: 20rpx;
  border-radius: 16rpx;
  padding: 20rpx;
}

.tab-item {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #666;
  position: relative;
}

.tab-item.active {
  color: #fb454c;
  font-weight: bold;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 60rpx;
  height: 4rpx;
  background-color: #fb454c;
  border-radius: 2rpx;
}

.list-section {
  margin: 0 20rpx 20rpx;
}

.enroll-list {
  display: flex;
  flex-direction: column;
}
.enroll-list .enroll-card {
  margin-bottom: 20rpx;
}
.enroll-list .enroll-card:last-child {
  margin-bottom: 0;
}

.enroll-card {
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
  display: flex;
  padding: 24rpx;
}

.card-img {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  margin-right: 24rpx;
  flex-shrink: 0;
}

.card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-desc {
  font-size: 26rpx;
  color: #666;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  font-size: 24rpx;
  color: #999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 100rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #999;
}
</style>
