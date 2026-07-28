<template>
  <view class="page">
    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索考试" confirm-type="search" @confirm="load" />
    </view>

    <view v-if="list.length > 0" class="exam-list">
      <view v-for="s in list" :key="s.id" class="exam-item card">
        <view class="s-head" :class="{ 's-head--disabled': !canFill(s) }" @click="goFill(s)">
          <view class="s-title">{{ s.title }}</view>
          <view class="s-desc" v-if="s.description">{{ s.description }}</view>
          <view class="s-meta">
            <text class="meta-tag" v-if="s.category">{{ s.category }}</text>
            <text class="meta-tag" v-if="s.showScore===1">显示分数</text>
            <text class="meta-tag" v-if="s.visibility===1">需登录</text>
            <text class="meta-tag" v-if="s.visibility===2">部门限定</text>
            <text class="meta-tag meta-tag--count" v-if="s.maxAttempts>0">{{ s.maxAttempts }}次机会</text>
            <text class="meta-tag" v-if="s.duration>0">{{ s.duration }}分钟</text>
          </view>
          <view class="s-time" v-if="s.startTime || s.endTime">
            <text v-if="s.startTime">开始: {{ formatDT(s.startTime) }}</text>
            <text v-if="s.endTime">结束: {{ formatDT(s.endTime) }}</text>
          </view>
          <view class="s-limit" v-if="limitReason(s)">
            <text class="meta-tag meta-tag--limit">{{ limitReason(s) }}</text>
          </view>
        </view>
        <view class="s-foot">
          <view v-if="myRecordMap[s.id]" class="s-my-rec">
            <text class="s-my-badge">已参考</text>
            <text class="s-my-time">{{ formatDT(myRecordMap[s.id].submitTime) }}</text>
            <text class="s-my-duration" v-if="myRecordMap[s.id].timeSpent>0">耗时 {{ myRecordMap[s.id].timeSpent }}秒</text>
          </view>
          <button class="btn-fill" :class="{ 'btn-filled': !canFill(s) }" :disabled="!canFill(s)" @click.stop="goFill(s)">
            {{ btnLabel(s) }}
          </button>
        </view>
      </view>
    </view>

    <view class="empty" v-else>
      <text>{{ loading ? '加载中...' : '暂无可参加的考试' }}</text>
    </view>
  </view>
</template>

<script>
import { examApi } from '../../api/index'
import { getClientUserInfo, hasClientAuth } from '../../utils/auth'

export default {
  data() {
    return {
      keyword: '',
      list: [],
      myRecordMap: {},
      limitsMap: {},
      loading: false
    }
  },
  computed: {
    isLogged() {
      return hasClientAuth()
    },
    myDeptId() {
      try {
        const info = getClientUserInfo()
        return (info && info.deptId) || 0
      } catch (e) { return 0 }
    }
  },
  onShow() {
    this.load()
  },
  onPullDownRefresh() {
    this.load().finally(() => { uni.stopPullDownRefresh() })
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const params = { page: 1, pageSize: 50, keyword: this.keyword }
        params['deviceId'] = this.getDeviceId()
        const res = await examApi.getList(params)
        this.list = (res.data && res.data.list) || []
        this.limitsMap = (res.data && res.data.limits) || {}
        if (this.isLogged) {
          const myRes = await examApi.myRecords()
          const myList = (myRes.data && myRes.data.list) || []
          const map = {}
          myList.forEach(r => { map[r.examId] = r })
          this.myRecordMap = map
        } else {
          this.myRecordMap = {}
        }
      } catch (e) {
        console.error(e)
      } finally { this.loading = false }
    },
    inDept(s) {
      if (s.visibility !== 2 || !s.deptIds) return true
      const allowed = String(s.deptIds).split(',').map(v => Number(v)).filter(v => v > 0)
      return allowed.length === 0 || allowed.includes(this.myDeptId)
    },
    isLimitFull(s) {
      const lim = this.limitsMap[s.id]
      return lim && (lim.deviceFull || lim.ipFull)
    },
    limitReason(s) {
      const lim = this.limitsMap[s.id]
      if (!lim) return ''
      if (lim.deviceFull && lim.ipFull) return '设备+IP已达上限'
      if (lim.deviceFull) return '设备已达上限'
      if (lim.ipFull) return 'IP已达上限'
      return ''
    },
    canFill(s) {
      if (this.isLimitFull(s)) return false
      if (this.myRecordMap[s.id] && s.allowMulti !== 1) return false
      if ((s.visibility === 1 || s.visibility === 2) && !this.isLogged) return false
      if (s.visibility === 2 && !this.inDept(s)) return false
      if (s.startTime > 0 && Date.now() < s.startTime) return false
      if (s.endTime > 0 && Date.now() > s.endTime) return false
      return true
    },
    btnLabel(s) {
      const reason = this.limitReason(s)
      if (reason) return reason
      if (this.myRecordMap[s.id] && s.allowMulti !== 1) return '已参考'
      if ((s.visibility === 1 || s.visibility === 2) && !this.isLogged) return '需登录'
      if (s.visibility === 2 && !this.inDept(s)) return '无权参加'
      if (s.startTime > 0 && Date.now() < s.startTime) return '未开始'
      if (s.endTime > 0 && Date.now() > s.endTime) return '已结束'
      return '开始考试'
    },
    goFill(s) {
      if (!this.canFill(s)) return
      uni.navigateTo({ url: `/pages/exam/fill?id=${s.id}` })
    },
    formatDT(ms) {
      if (!ms) return '-'
      const d = new Date(ms)
      return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
    },
    getDeviceId() {
      const key = '_device_id'
      let id = uni.getStorageSync(key)
      if (!id) {
        id = 'd_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 10)
        uni.setStorageSync(key, id)
      }
      return id
    },
    goBack() { uni.navigateBack() }
  }
}
</script>

<style scoped>
.page { padding-bottom: 200rpx; }
.search-bar { padding: 20rpx 30rpx; background: #fff;}
.search-bar input { background: #f5f5f5; height: 70rpx; border-radius: 35rpx; padding: 0 24rpx; font-size: 28rpx; }
.exam-list { padding: 20rpx 30rpx; }
.exam-item { background: #fff; border-radius: 16rpx; margin-bottom: 20rpx; overflow: hidden; }
.card { box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06); }
.s-head { padding: 30rpx 30rpx 0; min-height: 160rpx; }
.s-head--disabled { opacity: 0.6; }
.s-title { font-size: 32rpx; font-weight: 500; color: #333; margin-bottom: 12rpx; }
.s-desc { font-size: 24rpx; color: #888; margin-bottom: 14rpx; line-height: 1.5; }
.s-meta { display: flex; gap: 14rpx; margin-bottom: 12rpx; flex-wrap: wrap; }
.meta-tag { background: #f0f5ff; color: #4a7af0; font-size: 22rpx; padding: 4rpx 14rpx; border-radius: 6rpx; }
.meta-tag--count { background: #fff7e6; color: #d48806; }
.meta-tag--limit { background: #fff1f0; color: #f5222d; }
.s-time { color: #888; font-size: 24rpx; margin-bottom: 18rpx; }
.s-limit { margin-bottom: 18rpx; }
.s-foot { display: flex; align-items: center; justify-content: flex-end; gap: 16rpx; padding: 20rpx 30rpx; border-top: 1rpx solid #f0f0f0; margin-top: 16rpx; background: #fff; }
.s-my-rec { flex: 1; min-width: 0; display: flex; align-items: center; gap: 12rpx; flex-wrap: wrap; }
.s-my-badge { background: #e8f5e8; color: #4caf50; font-size: 22rpx; padding: 2rpx 12rpx; border-radius: 4rpx; }
.s-my-time { font-size: 22rpx; color: #999; }
.s-my-duration { font-size: 20rpx; color: #aaa; }
.btn-fill { margin: 0 0 0 auto; background: linear-gradient(90deg, #fb454c, #ff6b6b); color: #fff; border-radius: 50rpx; font-size: 26rpx; height: 64rpx; line-height: 64rpx; padding: 0 32rpx; border: none; flex-shrink: 0; }
.btn-filled { background: linear-gradient(90deg, #999, #bbb); }
.btn-fill:disabled { opacity: 0.6; }
.empty { text-align: center; padding: 100rpx 0; color: #aaa; font-size: 28rpx; }
</style>
