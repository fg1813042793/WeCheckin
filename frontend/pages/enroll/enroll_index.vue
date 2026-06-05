<template>
  <view class="container" :style="{ paddingTop: containerPad }">
    <view class="header-sticky" :style="{ top: fixedTop }">
      <view class="search-bar">
        <input v-model="keyword" placeholder="搜索打卡项目" class="search-input" @confirm="handleSearch" />
        <text class="search-btn" @click="handleSearch">搜索</text>
      </view>

      <view class="tabs">
        <view 
          class="tab-item" 
          :class="{ active: cur === 'join' }" 
          @click="switchTab('join')"
        >我参与的</view>
        <view 
          class="tab-item" 
          :class="{ active: cur === 'all' }" 
          @click="switchTab('all')"
        >全部</view>
      </view>
    </view>

    <view class="content">
      <view class="enroll-list" v-if="list.length > 0">
        <view class="enroll-card" v-for="(item, index) in list" :key="index" @click="goDetail(item.id)">
          <view v-if="item.img" class="card-img">
            <image :src="item.img" mode="aspectFill" class="card-img-inner" />
          </view>
          <view v-else class="card-img-placeholder" :style="{ background: getPlaceholderBg(index) }">
            <text class="placeholder-text">{{ item.title }}</text>
          </view>
          <view class="card-body">
            <view class="card-title-row">
              <text class="card-title">{{ item.title }}</text>
              <text class="fav-icon" :class="{ active: favSet.has(String(item.id)) }" @click.stop="toggleFav($event, item)">{{ favSet.has(String(item.id)) ? '♥' : '♡' }}</text>
            </view>
            <text class="card-desc">{{ item.desc || '快来参与打卡吧！' }}</text>
            <view class="card-footer">
              <view class="card-info">
                <text class="info-item">{{ item.userCnt || 0 }}人参与</text>
                <text class="info-item">{{ item.checkinCount || 0 }}次打卡</text>
              </view>
              <view v-if="cur === 'join'" class="card-btn" @click.stop="goDetail(item.id)">去打卡</view>
              <view v-else-if="!item.isJoin" class="card-btn" @click.stop="handleJoin($event, item)">立即参与</view>
              <view v-else class="card-btn joined">已参与</view>
            </view>
          </view>
        </view>
      </view>

      <view class="empty" v-else>
        <image src="/static/empty.png" mode="aspectFit" class="empty-img"></image>
        <text class="empty-text">暂无打卡任务</text>
      </view>
    </view>

    <view class="fab" @click="goAdmin" v-if="isAdmin">
      <text class="fab-text">+</text>
    </view>
  </view>
</template>

<script>
import { enrollApi, favApi } from '../../api/index'

export default {
  data() {
    return {
      cur: 'join',
      list: [],
      page: 1,
      pageSize: 10,
      keyword: '',
      isAdmin: false,
      hasMore: true,
      favSet: new Set(),
      fixedTop: '0px',
      containerPad: '0px'
    }
  },

  onLoad() {
    this.checkAdmin()
    this.loadData()
    const sys = uni.getSystemInfoSync()
    const pxScale = 750 / sys.windowWidth
    if (sys.platform === 'android') {
      this.fixedTop = '12rpx'
      this.containerPad = '192rpx'
    } else {
      const navOffset = (sys.statusBarHeight || 0) + 44
      this.fixedTop = (navOffset + 6) + 'px'
      this.containerPad = (navOffset + 6 + Math.round(180 / pxScale)) + 'px'
    }
  },

  onShow() {
    const tab = uni.getStorageSync('enrollTab')
    if (tab === 'all') {
      this.cur = 'all'
      uni.removeStorageSync('enrollTab')
    }
    this.page = 1
    this.hasMore = true
    this.loadData()
  },

  onReachBottom() {
    this.loadMore()
  },

  onPullDownRefresh() {
    this.page = 1
    this.hasMore = true
    this.loadData().then(() => {
      uni.stopPullDownRefresh()
    })
  },

  methods: {
    checkAdmin() {
      const userInfo = uni.getStorageSync('userInfo')
      this.isAdmin = userInfo && userInfo.role === 'admin'
    },

    handleSearch() {
      this.page = 1
      this.list = []
      this.hasMore = true
      this.loadData()
    },

    switchTab(tab) {
      this.cur = tab
      this.keyword = ''
      this.page = 1
      this.list = []
      this.hasMore = true
      this.loadData()
    },

    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },

    async loadFav() {
      try {
        const uid = this.getUserId()
        if (!uid) return
        const res = await favApi.list({ user_id: uid })
        const list = Array.isArray(res.data) ? res.data : (res.data.list || [])
        this.favSet = new Set(list.map(f => String(f.id)))
      } catch (e) {
        console.error('加载收藏数据失败', e)
      }
    },

    async toggleFav(e, item) {
      e.stopPropagation()
      const uid = this.getUserId()
      if (!uid) {
        uni.showToast({ title: '请先登录', icon: 'none' })
        return
      }
      const id = String(item.id)
      try {
        if (this.favSet.has(id)) {
          await favApi.del({ oid: id, user_id: uid })
          this.favSet.delete(id)
        } else {
          await favApi.insert({ oid: id, title: item.title, typ: 'enroll', user_id: uid, path: '/pages/enroll/enroll_detail?id=' + id })
          this.favSet.add(id)
        }
      } catch (e) {
        console.error('操作收藏失败', e)
      }
    },

    async loadData() {
      try {
        const uid = this.getUserId()
        const params = { page: this.page, pageSize: this.pageSize, user_id: uid }
        let res
        if (this.cur === 'join') {
          res = await enrollApi.myJoinList({ user_id: uid })
        } else {
          res = await enrollApi.getList({ ...params, keyword: this.keyword })
        }
        const data = Array.isArray(res.data) ? res.data : (res.data.list || [])
        if (this.cur === 'join') {
          this.list = data
          this.hasMore = false
        } else if (this.page === 1) {
          this.list = data
        } else {
          this.list = [...this.list, ...data]
        }
        if (this.cur !== 'join' && data.length < this.pageSize) {
          this.hasMore = false
        }
        this.loadFav()
      } catch (e) {
        console.error('加载打卡任务失败', e)
      }
    },

    loadMore() {
      if (!this.hasMore) return
      this.page++
      this.loadData()
    },

    getPlaceholderBg(index) {
      const colors = [
        'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
        'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
        'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
        'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
        'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
        'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
        'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)',
        'linear-gradient(135deg, #f5576c 0%, #ff758c 100%)',
        'linear-gradient(135deg, #3b82f6 0%, #2dd4bf 100%)',
      ]
      return colors[index % colors.length]
    },

    goDetail(id) {
      uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${id}` })
    },

    async handleJoin(e, item) {
      e.stopPropagation()
      const uid = this.getUserId()
      if (!uid) {
        uni.showToast({ title: '请先登录', icon: 'none' })
        return
      }
      // Check if enrollment form fields exist
      let enrollForms = []
      try {
        enrollForms = typeof item.forms === 'string' ? JSON.parse(item.forms || '[]') : (item.forms || [])
      } catch (e) {}
      if (enrollForms.length > 0) {
        uni.navigateTo({ url: '/pages/enroll/enroll_join_form?id=' + item.id + '&mode=enroll' })
        return
      }
      // No enrollment form, join directly
      try {
        await enrollApi.enrollSubmit({ enroll_id: item.id, user_id: uid, forms: '[]' })
        item.isJoin = true
        uni.showToast({ title: '参与成功', icon: 'success' })
      } catch (e) {
        console.error('参与失败', e)
      }
    },

    goAdmin() {
      uni.navigateTo({ url: '/pages/admin/admin_home' })
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header-sticky {
  position: fixed;
  left: 0;
  right: 0;
  z-index: 10;
  background-color: #fff;
}

.header-sticky::before {
  content: '';
  position: absolute;
  top: -12rpx;
  left: 0;
  right: 0;
  height: 12rpx;
  background-color: #f5f5f5;
}

.search-bar {
  display: flex;
  align-items: center;
  padding:20rpx;
}
.search-bar .search-input {
  margin-right: 16rpx;
}

.search-input {
  flex: 1;
  height: 64rpx;
  background-color: #f5f5f5;
  border-radius: 32rpx;
  padding: 0 24rpx;
  font-size: 26rpx;
  color: #333;
}

.search-btn {
  font-size: 26rpx;
  color: #fb454c;
  flex-shrink: 0;
}

.tabs {
  display: flex;
  background-color: #fff;
  padding: 0 20rpx 20rpx;
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



.enroll-list {
  padding: 20rpx 0;
}
.enroll-list .enroll-card {
  margin: 0 4rpx 20rpx;
}
.enroll-list .enroll-card:last-child {
  margin-bottom: 0;
}

.enroll-card {
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.card-img {
  width: 100%;
  height: 300rpx;
  overflow: hidden;
}
.card-img-inner {
  width: 100%;
  height: 100%;
}
.card-img-placeholder {
  width: 100%;
  height: 300rpx;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}
.placeholder-text {
  font-size: 40rpx;
  color: #fff;
  font-weight: bold;
  text-align: center;
  word-break: break-all;
  line-height: 1.4;
  padding: 0 20rpx;
}

.enroll-card {
  background-color: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.fav-icon {
  font-size: 36rpx;
  color: #ccc;
  flex-shrink: 0;
  padding: 4rpx;
}

.fav-icon.active {
  color: #fb454c;
}

.card-body {
  padding: 24rpx;
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.card-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 12rpx;
}

.card-desc {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 20rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-info {
  display: flex;
}
.card-info .info-item {
  margin-right: 20rpx;
}
.card-info .info-item:last-child {
  margin-right: 0;
}

.info-item {
  font-size: 24rpx;
  color: #999;
}

.card-btn {
  background-color: #fb454c;
  color: #fff;
  padding: 12rpx 32rpx;
  border-radius: 30rpx;
  font-size: 26rpx;
}

.card-btn.joined {
  background-color: #eee;
  color: #999;
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
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

.fab {
  position: fixed;
  right: 40rpx;
  bottom: 200rpx;
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background-color: #fb454c;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4rpx 20rpx rgba(251, 69, 76, 0.4);
}

.fab-text {
  color: #fff;
  font-size: 48rpx;
  line-height: 1;
}
</style>