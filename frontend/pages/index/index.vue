<template>
  <view class="main">
    <view class="up">
      <image mode="widthFix" src="/static/images/home.jpg"></image>
    </view>
    <view class="down">
      <view class="menu card-project">
        <view class="item" @click="goEnrollCate(1)">
          <view class="item-inner">
            <view class="img img1">
              <image src="/static/images/menu/1.png" mode="aspectFit" />
            </view>
            <view class="title">问卷</view>
          </view>
        </view>
        <view class="item" @click="goEnrollCate(2)">
          <view class="item-inner">
            <view class="img img2">
              <image src="/static/images/menu/2.png" mode="aspectFit" />
            </view>
            <view class="title">打卡</view>
          </view>
        </view>
        <view class="item" @click="goEnrollCate(3)">
          <view class="item-inner">
            <view class="img img3">
              <image src="/static/images/menu/3.png" mode="aspectFit" />
            </view>
            <view class="title">活动</view>
          </view>
        </view>
        <view class="item" @click="goEnrollCate(4)">
          <view class="item-inner">
            <view class="img img4">
              <image src="/static/images/menu/4.png" mode="aspectFit" />
            </view>
            <view class="title">比赛</view>
          </view>
        </view>
        <view class="item" @click="goEnrollCate(5)">
          <view class="item-inner">
            <view class="img img5">
              <image src="/static/images/menu/5.png" mode="aspectFit" />
            </view>
            <view class="title">工作</view>
          </view>
        </view>
      </view>

      <view class="tab-line">
        <view class="item">推荐</view>
        <!--<view class="item1" @click="goEnrollAll">全部</view>-->
      </view>

      <view v-if="!vouchList" class="loading-tip">加载中...</view>

      <view class="scroll-x" v-if="vouchList && vouchList.length > 0">
        <scroll-view scroll-x class="scroll-list">
          <view
            class="scroll-item"
            v-for="(item, idx) in vouchList"
            :key="idx"
            @click="goDetail(item)"
          >
            <view class="cover">
              <image v-if="item.img" lazy-load mode="aspectFill" :src="item.img"></image>
              <view v-else class="cover-placeholder" :style="{ background: getPlaceholderBg(idx) }">
                <text class="cover-placeholder-text">{{ item.title }}</text>
              </view>
              <view class="cover-title text-cut">{{ item.title }}</view>
            </view>
            <view class="users">
              <view class="pic-group">
                <template v-for="(u, ui) in (item.userList ? item.userList.slice(0, 3) : [])" :key="ui">
                  <image v-if="u.pic" class="pic round" mode="aspectFill" lazy-load :src="u.pic" />
                  <text v-else class="pic round text-avatar">{{ (u.name || u.userName || '?').charAt(0) }}</text>
                </template>
              </view>
              <text class="type-tag">{{ kindText(item.kind) }}</text>
              <text class="num">{{ '+' + item.userCnt }}人参与</text>
            </view>
          </view>
        </scroll-view>
      </view>

      <view class="tab-line1">
        <view
          class="tab-item"
          :class="{ cur: cur === 'hot' }"
          @click="switchTab('hot')"
        >热门</view>
        <view
          class="tab-item"
          :class="{ cur: cur === 'new' }"
          @click="switchTab('new')"
        >最新</view>
      </view>

      <view v-if="!newList || !hotList" class="loading-tip">加载中...</view>

      <view class="list" v-if="cur === 'new'">
        <view
          class="list-item"
          v-for="(item, idx) in newList"
          :key="idx"
          @click="goDetail(item)"
        >
          <view v-if="item.img" class="list-img">
            <image mode="aspectFill" lazy-load :src="item.img" class="list-img-inner" />
          </view>
          <view v-else class="list-img-placeholder" :style="{ background: getPlaceholderBg(idx) }">
            <text class="list-placeholder-text">{{ item.title }}</text>
          </view>
          <view class="list-right">
            <view class="list-title">{{ item.title }}</view>
            <view class="list-desc">
              <text class="kind-tag">{{ kindText(item.kind) }}</text>
              <text class="cate-tag">{{ item.cateName }}</text>
              <text>{{ item.userCnt }}人参与</text>
              <text class="list-time">{{ formatTime(item._createTime) }}</text>
            </view>
          </view>
        </view>
      </view>

      <view class="list" v-if="cur === 'hot'">
        <view
          class="list-item"
          v-for="(item, idx) in hotList"
          :key="idx"
          @click="goDetail(item)"
        >
          <view v-if="item.img" class="list-img">
            <image mode="aspectFill" lazy-load :src="item.img" class="list-img-inner" />
          </view>
          <view v-else class="list-img-placeholder" :style="{ background: getPlaceholderBg(idx) }">
            <text class="list-placeholder-text">{{ item.title }}</text>
          </view>
          <view class="list-right">
            <view class="list-title">{{ item.title }}</view>
            <view class="list-desc">
              <text class="kind-tag">{{ kindText(item.kind) }}</text>
              <text class="cate-tag">{{ item.cateName }}</text>
              <text>{{ item.userCnt }}人参与</text>
              <text class="list-time">{{ formatTime(item._createTime) }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { homeApi } from '../../api/index'

export default {
  data() {
    return {
      cur: 'hot',
      vouchList: null,
      hotList: [],
      newList: []
    }
  },

  onLoad() {
    this.loadData()
  },

  onShow() {
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
        const userInfo = uni.getStorageSync('userInfo')
        const userId = (userInfo && (userInfo.miniOpenID || userInfo.id)) || ''
        const res = await homeApi.getList({ user_id: userId })
        if (res.data) {
          this.vouchList = res.data.vouchList || []
          this.hotList = res.data.hotList || []
          this.newList = res.data.newList || []
        }
      } catch (e) {
        console.error('加载数据失败', e)
      }
    },

    switchTab(tab) {
      this.cur = tab
    },

    goEnrollCate(id) {
      uni.switchTab({ url: '/pages/enroll/enroll_index' })
    },

    goEnrollAll() {
      uni.setStorageSync('enrollTab', 'all')
      uni.switchTab({ url: '/pages/enroll/enroll_index' })
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

    formatTime(ts) {
      if (!ts) return ''
      const d = new Date(ts)
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      return y + '-' + m + '-' + day
    },

    goDetail(item) {
      if (item.kind === 'enroll') {
        uni.navigateTo({ url: `/pages/enroll/enroll_detail?id=${item.id}` })
      } else {
        uni.navigateTo({ url: `/pages/event/event_detail?id=${item.id}` })
      }
    },
    kindText(kind) {
      const map = { enroll: '打卡', activity: '活动', competition: '赛事' }
      return map[kind] || kind || ''
    },
    getCateName(id) {
      const map = { '1': '问卷', '2': '打卡', '3': '活动', '4': '赛事', '5': '工作' }
      return map[String(id)] || ''
    }
  }
}
</script>

<style scoped>
page {
  background-color: #fff;
  overflow-x: hidden;
}

.main {
  padding: 0 0 100rpx;
  overflow-x: hidden;
}

.up {
  width: 100%;
}

.up image {
  width: 100%;
  border-radius: 0% 0% 72% 39% / 10% 11% 16% 0%;
  box-shadow: 0rpx 15rpx 8px -10rpx rgba(236, 28, 28, 0.4);
}

.down {
  width: 100%;
  box-sizing: border-box;
  padding: 10rpx 0;
  margin-top: 0;
}

.menu {
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  background-color: #fff;
  padding: 20rpx 10rpx;
  border-radius: 20rpx;
  box-sizing: border-box;
}

.menu .item {
  width: 20%;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 180rpx;
}

.menu .item .item-inner {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.menu .item-inner .img {
  width: 90rpx;
  height: 90rpx;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 40%;
  box-shadow: 1rpx 8rpx 8rpx rgba(230, 225, 226, 0.8);
}

.menu .item-inner .img.img1 {
  background-color: #feb702;
}

.menu .item-inner .img.img2 {
  background-color: #e12b7f;
}

.menu .item-inner .img.img3 {
  background-color: #9e5ffe;
}

.menu .item-inner .img.img4 {
  background-color: #3ebaf7;
}

.menu .item-inner .img.img5 {
  background-color: #fb454c;
}

.menu .item-inner .img image {
  width: 50rpx;
  height: 50rpx;
}

.menu .item-inner .title {
  margin-top: 20rpx;
  font-size: 28rpx;
  color: #333;
}

.tab-line {
  width: 100%;
  display: flex;
  justify-content: space-between;
  padding: 10rpx 30rpx;
  align-items: center;
  margin-top: 20rpx;
  box-sizing: border-box;
}

.tab-line .item {
  font-size: 34rpx;
  font-weight: bold;
  color: #000;
}

.tab-line .item1 {
  font-size: 24rpx;
  color: #333;
  padding: 6rpx 24rpx;
  border: 2rpx solid #bbb;
  border-radius: 30rpx;
}

.loading-tip {
  text-align: center;
  padding: 40rpx;
  color: #999;
  font-size: 26rpx;
}

.scroll-x {
  width: 100%;
  padding: 0 28rpx;
  box-sizing: border-box;
}

.scroll-list {
  width: 100%;
  margin-top: 10rpx;
  white-space: nowrap;
}
.scroll-list .scroll-item {
  display: inline-flex;
  flex-direction: column;
  padding: 0;
  margin-right: 20rpx;
  border-radius: 20rpx;
  overflow: hidden;
  width: 370rpx;
  flex-shrink: 0;
  box-shadow: 1rpx 8rpx 8rpx rgba(230, 225, 226, 0.6);
  margin-bottom: 30rpx;
  background-color: #fff;
  vertical-align: top;
}

.scroll-list .scroll-item .cover {
  height: 180rpx;
  width: 100%;
  border-radius: 20rpx;
  position: relative;
  overflow: hidden;
}

.scroll-list .scroll-item .cover image,
.scroll-list .scroll-item .cover .cover-placeholder {
  height: 100%;
  width: 100%;
  border-radius: inherit;
}
.scroll-list .scroll-item .cover .cover-placeholder {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}
.scroll-list .scroll-item .cover .cover-placeholder .cover-placeholder-text {
  font-size: 24rpx;
  color: #fff;
  font-weight: bold;
  text-align: center;
  word-break: break-all;
  line-height: 1.3;
  padding: 0 10rpx;
}

.scroll-list .scroll-item .cover .cover-title {
  position: absolute;
  bottom: 0;
  z-index: 99;
  width: 100%;
  background-color: rgba(241, 21, 21, 0.7);
  color: #fff;
  font-size: 24rpx;
  display: flex;
  align-items: center;
  padding: 6rpx 10rpx;
}

.scroll-list .users {
  width: 100%;
  font-size: 24rpx;
  padding: 10rpx;
  color: #999;
  display: flex;
  align-items: center;
  height: 68rpx;
}

.scroll-list .users .pic-group {
  display: flex;
  padding: 0 10rpx 0 0;
  align-items: center;
}

.scroll-list .users .pic-group .pic {
  width: 48rpx !important;
  height: 48rpx !important;
  border: 4rpx solid #fff;
  margin-left: -20rpx;
  border-radius: 50%;
}

.scroll-list .users .pic-group .pic:first-child {
  margin-left: 0;
}
.scroll-list .users .pic-group .text-avatar {
  font-size: 20rpx;
  background-color: #fb454c;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}

.scroll-list .users .num {
  font-size: 24rpx;
  color: #999;
}
.scroll-list .users .type-tag {
  font-size: 22rpx;
  color: #fb454c;
  margin-right: 8rpx;
  flex-shrink: 0;
}

.tab-line1 {
  width: 100%;
  display: flex;
  justify-content: center;
  padding: 10rpx 30rpx;
  align-items: flex-end;
  margin-top: 20rpx;
  margin-bottom: 10rpx;
  box-sizing: border-box;
}

.tab-line1 .tab-item {
  font-size: 30rpx;
  color: #666;
  padding: 0 30rpx;
  position: relative;
}

.tab-line1 .tab-item.cur {
  font-size: 34rpx;
  color: #000;
  font-weight: bold;
}

.tab-line1 .tab-item.cur::after {
  width: 20%;
  position: absolute;
  bottom: -16rpx;
  left: 50%;
  height: 8rpx;
  content: '';
  border-radius: 30%;
  background-color: #fb454c;
  transform: translateX(-50%);
}

.list {
  width: 100%;
  padding: 20rpx 0;
  display: flex;
  flex-direction: column;
}

.list-item {
  width: 100%;
  display: flex;
  align-items: center;
  padding: 10rpx 30rpx;
  margin-bottom: 15rpx;
  overflow: hidden;
  border-bottom: 2rpx dashed #eee;
  box-sizing: border-box;
}

.list-item:last-child {
  border-bottom: 0;
}

.list-img {
  width: 100rpx;
  height: 100rpx;
  border-radius: 10rpx;
  margin-right: 25rpx;
  overflow: hidden;
}
.list-img-inner {
  width: 100%;
  height: 100%;
}
.list-img-placeholder {
  width: 100rpx;
  height: 100rpx;
  border-radius: 10rpx;
  margin-right: 25rpx;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.list-placeholder-text {
  font-size: 20rpx;
  color: #fff;
  font-weight: bold;
  text-align: center;
  word-break: break-all;
  line-height: 1.3;
  padding: 0 6rpx;
}

.list-right {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.list-title {
  width: 100%;
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 20rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-desc {
  width: 100%;
  font-size: 26rpx;
  color: #999;
  margin-bottom: 10rpx;
  display: flex;
  align-items: baseline;
}

.list-desc .kind-tag {
  font-size: 22rpx;
  color: #fb454c;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  background-color: #fff1f0;
}
.list-desc .cate-tag {
  font-size: 22rpx;
  color: #2b7ef5;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  background-color: #f0f5ff;
}
.list-desc .tag {
  background-color: rgba(251, 69, 76, 0.1);
  color: #fb454c;
  padding: 2rpx 12rpx;
  border-radius: 8rpx;
  font-size: 22rpx;
  margin-right: 10rpx;
}
.list-desc .list-time {
  margin-left: auto;
  font-size: 22rpx;
  color: #bbb;
}


</style>
