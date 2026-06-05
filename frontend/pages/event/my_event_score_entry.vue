<template>
  <view class="container">
    <view class="header">
      <text class="header-title">{{ title }}</text>
      <view class="header-row">
        <text class="header-info">为参与者录入成绩</text>
        <view class="header-actions">
          <view class="action-btn" @click="exportCSV">导出</view>
          <view class="action-btn" @click="triggerImport">导入</view>
        </view>
      </view>
    </view>
    <view class="search-bar">
      <input v-model="keyword" placeholder="搜索参与者姓名" class="search-input" confirm-type="search" @confirm="keyword = keyword" />
    </view>
    <view class="loading" v-if="loading"><text>加载中...</text></view>
    <view class="list" v-else-if="filteredParticipants.length > 0">
      <view class="item" v-for="(p, i) in filteredParticipants" :key="i">
        <view class="item-left">
          <text class="item-name">{{ p.userName || p.nickname || '匿名' }}</text>
          <text class="item-info">{{ [p.deptName, p.topDeptName].filter(Boolean).join(' / ') }}</text>
        </view>
        <view class="item-right">
          <view v-if="scoreFields.length > 0" class="score-fields">
            <view class="score-field-row" v-for="(sf, j) in scoreFields" :key="j">
              <text class="score-field-label">{{ sf.name }}</text>
              <input v-if="sf.type === 'text'" v-model="p._scores[j]" placeholder="输入" class="score-input-sm" />
              <view v-else-if="sf.type === 'select'" class="score-input-sm score-select" @click="showSelectPicker(p, j, sf)">{{ p._scores[j] || '选择' }}</view>
              <input v-else v-model="p._scores[j]" type="digit" placeholder="0" class="score-input-sm" />
            </view>
          </view>
          <input v-else v-model="p._score" type="digit" placeholder="输入成绩" class="score-input" />
          <view class="save-btn" @click="saveScore(p)">保存</view>
      </view>
      </view>
    </view>
    <view class="empty" v-else>
      <text class="empty-text">{{ keyword ? '未找到匹配的参与者' : '暂无参与者' }}</text>
    </view>
  </view>
</template>

<script>
import { eventApi } from '../../api/index'
export default {
  data() {
    return {
      eventId: '',
      title: '',
      participants: [],
      scoreFields: [],
      loading: false,
      keyword: ''
    }
  },
  computed: {
    filteredParticipants() {
      const kw = (this.keyword || '').trim().toLowerCase()
      if (!kw) return this.participants
      return this.participants.filter(p => {
        const name = (p.userName || p.nickname || '').toLowerCase()
        return name.includes(kw)
      })
    }
  },
  onLoad(opts) {
    if (opts.event_id) { this.eventId = opts.event_id }
    if (opts.title) { this.title = decodeURIComponent(opts.title) }
    this.loadData()
  },
  methods: {
    getUserId() {
      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      return (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''
    },
    async loadData() {
      this.loading = true
      try {
        const [partRes, detailRes] = await Promise.all([
          eventApi.participantList({ event_id: this.eventId, page: 1, pageSize: 500 }),
          eventApi.getDetail({ id: this.eventId })
        ])
        const data = Array.isArray(partRes.data) ? partRes.data : (partRes.data.list || [])
        let scoreFields = []
        if (detailRes.data && detailRes.data.scoreFields) {
          try { scoreFields = typeof detailRes.data.scoreFields === 'string' ? JSON.parse(detailRes.data.scoreFields) : detailRes.data.scoreFields } catch {}
        }
        this.scoreFields = scoreFields || []
        // Load existing scores
        let scoreMap = {}
        try {
          const scoreRes = await eventApi.scores({ event_id: this.eventId })
          const scoreList = Array.isArray(scoreRes.data) ? scoreRes.data : (scoreRes.data.list || [])
          scoreList.forEach(s => { scoreMap[s.participantId || s.miniOpenId] = s.score })
        } catch {}
        this.participants = data.map(p => {
          const existing = scoreMap[p.miniOpenId] || ''
          if (scoreFields.length > 0) {
            let scores = []
            try { scores = typeof existing === 'string' ? JSON.parse(existing) : (existing || []) } catch { scores = [] }
            scores = scores.map(s => typeof s === 'object' ? String(s.score || '') : String(s || ''))
            scoreFields.forEach((sf, j) => { if (scores[j] === undefined) scores[j] = '' })
            return { ...p, _scores: scores, _score: existing }
          }
          return { ...p, _score: existing }
        })
      } catch (e) { console.error(e) }
      this.loading = false
    },
    showSelectPicker(p, j, sf) {
      const opts = (sf.options || '').split(',').map(s => s.trim()).filter(Boolean)
      if (opts.length === 0) return
      uni.showActionSheet({ itemList: opts, success: (res) => { p._scores[j] = opts[res.tapIndex] } })
    },
    async saveScore(p) {
      let scoreVal
      if (this.scoreFields.length > 0) {
        const allFilled = p._scores.every(s => s !== '' && s !== undefined && s !== null)
        if (!allFilled) { uni.showToast({ title: '请填写所有评分项', icon: 'none' }); return }
        scoreVal = JSON.stringify(this.scoreFields.map((sf, j) => ({ name: sf.name, score: p._scores[j] })))
      } else {
        if (!p._score && p._score !== 0) { uni.showToast({ title: '请输入成绩', icon: 'none' }); return }
        scoreVal = p._score
      }
      try {
        const uid = this.getUserId()
        await eventApi.scoreSave({ event_id: this.eventId, judge_id: uid, participant_id: p.miniOpenId, score: scoreVal })
        p._score = scoreVal
        uni.showToast({ title: '保存成功', icon: 'success' })
      } catch (e) { uni.showToast({ title: '保存失败', icon: 'none' }) }
    },
    buildCSV() {
      const headers = ['姓名', '部门', ...this.scoreFields.map(sf => sf.name)]
      const rows = this.participants.map(p => {
        const name = p.userName || p.nickname || '匿名'
        const dept = p.deptName || ''
        if (this.scoreFields.length > 0) {
          const vals = this.scoreFields.map((sf, j) => p._scores[j] || '')
          return [name, dept, ...vals]
        }
        return [name, dept, p._score || '']
      })
      const esc = v => '"' + String(v).replace(/"/g, '""') + '"'
      return [headers, ...rows].map(r => r.map(esc).join(',')).join('\n')
    },
    exportCSV() {
      const csv = '\uFEFF' + this.buildCSV()
      const title = this.title || '成绩'
      try {
        const a = document.createElement('a')
        a.href = 'data:text/csv;charset=utf-8,' + encodeURIComponent(csv)
        a.download = title + '.csv'
        a.click()
      } catch (e) {
        uni.setClipboardData({ data: csv, fail: () => uni.showToast({ title: '导出失败', icon: 'none' }) })
      }
    },
    triggerImport() {
      try {
        const input = document.createElement('input')
        input.type = 'file'
        input.accept = '.csv'
        input.addEventListener('change', (e) => {
          const file = e.target?.files?.[0]
          if (!file) return
          const reader = new FileReader()
          reader.onload = (ev) => this.parseCSV(ev.target?.result || '')
          reader.readAsText(file, 'UTF-8')
          e.target.value = ''
        })
        input.click()
      } catch (e) {
        uni.showToast({ title: '当前环境不支持导入', icon: 'none' })
      }
    },
    parseCSV(text) {
      const lines = text.split('\n').filter(l => l.trim())
      if (lines.length < 2) { uni.showToast({ title: 'CSV 格式错误', icon: 'none' }); return }
      const parseLine = l => {
        const result = []; let cur = ''; let inQ = false
        for (const ch of l) {
          if (ch === '"') { inQ = !inQ; continue }
          if (ch === ',' && !inQ) { result.push(cur.trim()); cur = ''; continue }
          cur += ch
        }
        result.push(cur.trim())
        return result
      }
      const headers = parseLine(lines[0])
      const sfNames = this.scoreFields.map(sf => sf.name)
      const colIdx = h => headers.findIndex(x => x === h)
      const nameIdx = colIdx('姓名'); const deptIdx = colIdx('部门')
      if (nameIdx < 0) { uni.showToast({ title: 'CSV 缺少"姓名"列', icon: 'none' }); return }
      let updated = 0; let notFound = 0
      for (let i = 1; i < lines.length; i++) {
        const cols = parseLine(lines[i])
        const name = cols[nameIdx] || ''
        const p = this.participants.find(x => (x.userName || x.nickname) === name)
        if (!p) { notFound++; continue }
        if (this.scoreFields.length > 0) {
          sfNames.forEach((sn, j) => {
            const ci = colIdx(sn)
            if (ci >= 0 && cols[ci] !== undefined) p._scores[j] = cols[ci]
          })
        } else {
          if (cols[deptIdx + 1] !== undefined) p._score = cols[deptIdx + 1]
        }
        updated++
      }
      uni.showToast({ title: `导入完成：更新 ${updated} 人${notFound ? '，未匹配 ' + notFound + ' 人' : ''}`, icon: 'none' })
    }
  }
}
</script>

<style scoped>
.container { min-height: 100vh; background-color: #f5f5f5; }
.header { padding: 20rpx; background-color: #fff; margin-bottom: 20rpx; }
.header-title { font-size: 32rpx; font-weight: bold; color: #333; display: block; }
.header-info { font-size: 26rpx; color: #999; display: block; }
.header-row { display: flex; align-items: center; justify-content: space-between; margin-top: 8rpx; }
.header-actions { display: flex; }
.action-btn { background-color: #f5f5f5; padding: 8rpx 24rpx; border-radius: 20rpx; font-size: 24rpx; color: #333; margin-left: 16rpx; }
.action-btn:first-child { margin-left: 0; }
.search-bar { padding: 0 20rpx 16rpx; }
.search-input { height: 60rpx; background-color: #fff; border-radius: 30rpx; padding: 0 24rpx; font-size: 26rpx; color: #333; border: 2rpx solid #eee; }
.list { padding: 0 20rpx; }
.item { background-color: #fff; border-radius: 12rpx; padding: 20rpx; margin-bottom: 12rpx; display: flex; align-items: center; justify-content: space-between; }
.item-left { flex: 1; margin-right: 12rpx; }
.item-name { font-size: 28rpx; color: #333; font-weight: bold; display: block; }
.item-info { font-size: 24rpx; color: #999; margin-top: 4rpx; display: block; }
.item-right { display: flex; align-items: center; flex-shrink: 0; }
.score-input { width: 140rpx; height: 60rpx; background-color: #f5f5f5; border-radius: 8rpx; padding: 0 12rpx; font-size: 26rpx; text-align: center; margin-left: 12rpx; }
.save-btn { background-color: #fb454c; color: #fff; padding: 8rpx 24rpx; border-radius: 20rpx; font-size: 24rpx; margin-left: 12rpx; }
.score-fields { display: flex; flex-direction: column; }
.score-field-row { display: flex; align-items: center; margin-top: 6rpx; }
.score-field-row:first-child { margin-top: 0; }
.score-field-label { font-size: 22rpx; color: #666; white-space: nowrap; min-width: 60rpx; margin-right: 6rpx; }
.score-input-sm { width: 100rpx; height: 52rpx; background-color: #f5f5f5; border-radius: 6rpx; padding: 0 8rpx; font-size: 24rpx; text-align: center; margin-left: 6rpx; }
.score-select { display: flex; align-items: center; justify-content: center; }
.empty, .loading { display: flex; align-items: center; justify-content: center; padding-top: 200rpx; font-size: 28rpx; color: #999; }
</style>
