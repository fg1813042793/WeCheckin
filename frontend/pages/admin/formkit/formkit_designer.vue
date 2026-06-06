<template>
  <view class="formkit-h5">
    <view class="h5-header">
      <text class="h5-title">表单设计器</text>
      <view class="h5-actions">
        <button class="h5-btn" @click="loadFromJson">导入</button>
        <button class="h5-btn primary" @click="exportToJson">导出</button>
        <button class="h5-btn" @click="validateSchema" :disabled="validating">{{ validating ? '校验中' : '校验' }}</button>
      </view>
    </view>

    <view class="h5-categories">
      <scroll-view scroll-x class="cat-scroll">
        <view
          v-for="cat in categories"
          :key="cat.name"
          class="cat-tab"
          :class="{ active: activeCategory === cat.name }"
          @click="activeCategory = cat.name"
        >{{ cat.label }}</view>
      </scroll-view>
    </view>

    <view class="h5-types">
      <view
        v-for="t in (typesByCategory[activeCategory] || [])"
        :key="t.type"
        class="type-chip"
        @click="addQuestion(t)"
      >
        <text class="chip-name">{{ t.displayName }}</text>
      </view>
      <view v-if="(typesByCategory[activeCategory] || []).length === 0" class="empty-cat">
        该分类暂无题型
      </view>
    </view>

    <view class="h5-questions">
      <view class="q-header">
        <text>题目列表（{{ questions.length }}）</text>
        <text v-if="questions.length > 0" class="clear" @click="clearAll">清空</text>
      </view>
      <view v-if="questions.length === 0" class="empty-msg">点击上方题型添加</view>
      <view
        v-for="(q, idx) in questions"
        :key="q.id"
        class="q-card"
        :class="{ selected: selectedId === q.id }"
        @click="selectQuestion(q.id)"
      >
        <view class="q-main">
          <view class="q-line">
            <text class="q-type">{{ typeName(q.type) }}</text>
            <text class="q-title">{{ q.title || q.id }}</text>
            <text v-if="q.required" class="q-req">*</text>
          </view>
          <text class="q-id">ID: {{ q.id }}</text>
        </view>
        <view class="q-actions">
          <text class="q-act" @click.stop="moveUp(idx)" v-if="idx > 0">↑</text>
          <text class="q-act" @click.stop="moveDown(idx)" v-if="idx < questions.length - 1">↓</text>
          <text class="q-act del" @click.stop="removeAt(idx)">删</text>
        </view>
      </view>
    </view>

    <!-- 题目属性编辑弹窗 -->
    <view v-if="selected" class="q-modal-mask" @click="closeModal">
      <view class="q-modal" @click.stop>
        <view class="modal-header">
          <text class="modal-title">编辑题目</text>
          <text class="modal-close" @click="closeModal">×</text>
        </view>
        <scroll-view scroll-y class="modal-body">
          <view class="form-row">
            <text class="form-label">题目 ID</text>
            <input v-model="selected.id" :disabled="!!selected._existing" class="form-input" />
          </view>
          <view class="form-row">
            <text class="form-label">标题</text>
            <input v-model="selected.title" class="form-input" />
          </view>
          <view class="form-row">
            <text class="form-label">说明</text>
            <textarea v-model="selected.description" class="form-textarea" />
          </view>
          <view class="form-row">
            <text class="form-label">必填</text>
            <switch :checked="!!selected.required" @change="onRequiredChange" color="#3b82f6" />
          </view>
          <view class="form-row">
            <text class="form-label">占位提示</text>
            <input v-model="selected.placeholder" class="form-input" />
          </view>
          <view v-if="hasOptions(selected)" class="form-row">
            <text class="form-label">选项</text>
            <view class="opts-list">
              <view v-for="(opt, oi) in selected.props.options" :key="oi" class="opt-row">
                <input v-model="opt.label" placeholder="显示文本" class="form-input small" />
                <input v-model="opt.value" placeholder="值" class="form-input small" />
                <text class="opt-del" @click="removeOpt(oi)">删</text>
              </view>
              <view class="opt-add" @click="addOpt">+ 添加选项</view>
            </view>
          </view>
          <view class="form-row">
            <text class="form-label">计算表达式</text>
            <input v-model="calcExpr" @input="onCalcExprChange" placeholder="如 q1 + q2" class="form-input" />
            <text class="form-hint">结果将写入本题目（除非下方指定目标）</text>
            <input v-model="calcTarget" @input="onCalcTargetChange" placeholder="目标题目 ID (留空=自身)" class="form-input" style="margin-top: 8rpx" />
          </view>
        </scroll-view>
        <view class="modal-footer">
          <button class="modal-btn" @click="closeModal">完成</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { adminApi } from '../../api/admin.js'

export default {
  data() {
    return {
      types: [],
      activeCategory: 'base',
      questions: [],
      selectedId: null,
      validating: false,
      calcExpr: '',
      calcTarget: '',
      categories: [
        { name: 'base', label: '基础' },
        { name: 'select', label: '选择' },
        { name: 'media', label: '媒体' },
        { name: 'layout', label: '布局' },
        { name: 'advanced', label: '高级' }
      ]
    }
  },
  computed: {
    typesByCategory() {
      const m = {}
      for (const t of this.types) {
        if (!m[t.category]) m[t.category] = []
        m[t.category].push(t)
      }
      return m
    },
    selected() {
      return this.questions.find((q) => q.id === this.selectedId) || null
    }
  },
  async mounted() {
    try {
      const res = await adminApi.formkitTypes()
      this.types = res.data || []
    } catch (e) {
      uni.showToast({ title: '加载题型失败', icon: 'none' })
    }
  },
  watch: {
    selected: {
      immediate: false,
      handler(v) {
        if (v) {
          this.calcExpr = v.calcValue?.expr || ''
          this.calcTarget = v.calcValue?.target || ''
        } else {
          this.calcExpr = ''
          this.calcTarget = ''
        }
      }
    }
  },
  methods: {
    typeName(t) {
      const map = {
        input: '单行文本', text: '文本', textarea: '多行', number: '数字',
        select: '下拉', radio: '单选', checkbox: '多选', picker: '选择器',
        rating: '评分', date: '日期', time: '时间', dateRange: '日期范围',
        file: '文件', signature: '签名', location: '位置',
        phone: '手机', email: '邮箱', idCard: '身份证', password: '密码',
        switch: '开关', matrixRadio: '矩阵单选', autopop: '自动填充',
        divider: '分割线', description: '说明'
      }
      return map[t] || t
    },
    hasOptions(q) {
      return ['select', 'radio', 'checkbox', 'picker'].includes(q.type)
    },
    nextId() {
      this._idCounter = (this._idCounter || 0) + 1
      return `q${this._idCounter}`
    },
    addQuestion(t) {
      const id = this.nextId()
      const q = {
        id,
        type: t.type,
        title: t.displayName,
        description: '',
        required: false,
        placeholder: (t.defaultProps && t.defaultProps.placeholder) || '',
        props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {},
        validate: [],
        logic: []
      }
      this.questions.push(q)
      this.selectedId = id
    },
    selectQuestion(id) {
      this.selectedId = id
    },
    removeAt(idx) {
      this.questions.splice(idx, 1)
      if (this.questions.length === 0) this.selectedId = null
    },
    moveUp(idx) {
      if (idx <= 0) return
      const arr = [...this.questions]
      ;[arr[idx - 1], arr[idx]] = [arr[idx], arr[idx - 1]]
      this.questions = arr
    },
    moveDown(idx) {
      if (idx >= this.questions.length - 1) return
      const arr = [...this.questions]
      ;[arr[idx], arr[idx + 1]] = [arr[idx + 1], arr[idx]]
      this.questions = arr
    },
    clearAll() {
      uni.showModal({
        title: '确认',
        content: '清空所有题目？',
        success: (r) => {
          if (r.confirm) {
            this.questions = []
            this.selectedId = null
            this._idCounter = 0
          }
        }
      })
    },
    onRequiredChange(e) {
      if (this.selected) this.selected.required = e.detail.value
    },
    addOpt() {
      if (!this.selected) return
      if (!this.selected.props) this.selected.props = {}
      if (!this.selected.props.options) this.selected.props.options = []
      const n = this.selected.props.options.length + 1
      this.selected.props.options.push({ label: `选项 ${n}`, value: `opt${n}` })
    },
    removeOpt(oi) {
      if (!this.selected) return
      this.selected.props.options.splice(oi, 1)
    },
    onCalcExprChange(e) {
      if (!this.selected) return
      const v = e.detail.value
      if (v === '') this.selected.calcValue = null
      else this.selected.calcValue = { ...(this.selected.calcValue || {}), expr: v }
    },
    onCalcTargetChange(e) {
      if (!this.selected) return
      const v = e.detail.value
      if (!this.selected.calcValue) this.selected.calcValue = { expr: '' }
      this.selected.calcValue.target = v
    },
    closeModal() {
      this.selectedId = null
    },
    buildSchemaJson() {
      return JSON.stringify({
        version: '2.0',
        questions: this.questions.map((q) => {
          const { _existing, ...rest } = q
          return rest
        })
      })
    },
    exportToJson() {
      const json = this.buildSchemaJson()
      uni.setClipboardData({
        data: json,
        success: () => uni.showToast({ title: '已复制到剪贴板' })
      })
    },
    loadFromJson() {
      uni.getClipboardData({
        success: (res) => {
          uni.showModal({
            title: '从剪贴板导入？',
            content: res.data.substring(0, 100) + (res.data.length > 100 ? '...' : ''),
            success: (r) => {
              if (r.confirm) this.parseAndImport(res.data)
            }
          })
        }
      })
    },
    async parseAndImport(json) {
      this.validating = true
      try {
        const res = await adminApi.formkitParseSchema(json)
        this.questions = (res.data.questions || []).map((q) => ({ ...q, _existing: true }))
        this._idCounter = this.questions.length
        uni.showToast({ title: '导入成功' })
      } catch (e) {
        uni.showToast({ title: e.msg || '解析失败', icon: 'none' })
      } finally {
        this.validating = false
      }
    },
    async validateSchema() {
      this.validating = true
      try {
        await adminApi.formkitParseSchema(this.buildSchemaJson())
        uni.showToast({ title: '校验通过' })
      } catch (e) {
        uni.showToast({ title: e.msg || '校验失败', icon: 'none' })
      } finally {
        this.validating = false
      }
    }
  }
}
</script>

<style>
.formkit-h5 {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f7fa;
}
.h5-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 32rpx;
  background: #fff;
  border-bottom: 1rpx solid #e4e7ed;
}
.h5-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #303133;
}
.h5-actions {
  display: flex;
  gap: 12rpx;
}
.h5-btn {
  font-size: 24rpx;
  padding: 8rpx 20rpx;
  background: #fff;
  border: 1rpx solid #dcdfe6;
  border-radius: 8rpx;
}
.h5-btn.primary {
  background: #3b82f6;
  color: #fff;
  border-color: #3b82f6;
}
.h5-categories {
  background: #fff;
  border-bottom: 1rpx solid #e4e7ed;
}
.cat-scroll {
  white-space: nowrap;
}
.cat-tab {
  display: inline-block;
  padding: 20rpx 32rpx;
  font-size: 26rpx;
  color: #606266;
  border-bottom: 4rpx solid transparent;
}
.cat-tab.active {
  color: #3b82f6;
  border-bottom-color: #3b82f6;
}
.h5-types {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  padding: 20rpx;
  background: #fff;
}
.type-chip {
  background: #f5f7fa;
  border: 1rpx dashed #c0c4cc;
  border-radius: 12rpx;
  padding: 16rpx 24rpx;
}
.chip-name {
  font-size: 24rpx;
  color: #303133;
}
.empty-cat {
  font-size: 24rpx;
  color: #909399;
  padding: 20rpx;
}
.h5-questions {
  flex: 1;
  overflow-y: auto;
  padding: 20rpx;
}
.q-header {
  display: flex;
  justify-content: space-between;
  font-size: 26rpx;
  color: #606266;
  margin-bottom: 16rpx;
}
.clear {
  color: #f56c6c;
}
.empty-msg {
  text-align: center;
  color: #909399;
  padding: 60rpx 0;
  font-size: 24rpx;
}
.q-card {
  display: flex;
  align-items: center;
  background: #fff;
  border: 1rpx solid #e4e7ed;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  margin-bottom: 16rpx;
}
.q-card.selected {
  border-color: #3b82f6;
  background: #ecf5ff;
}
.q-main {
  flex: 1;
  min-width: 0;
}
.q-line {
  display: flex;
  align-items: center;
  gap: 12rpx;
}
.q-type {
  font-size: 20rpx;
  padding: 4rpx 12rpx;
  background: #f0f9ff;
  color: #3b82f6;
  border-radius: 6rpx;
}
.q-title {
  font-size: 28rpx;
  color: #303133;
  font-weight: 500;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.q-req {
  color: #f56c6c;
  font-size: 28rpx;
}
.q-id {
  font-size: 20rpx;
  color: #909399;
  font-family: monospace;
  margin-top: 8rpx;
  display: block;
}
.q-actions {
  display: flex;
  gap: 8rpx;
}
.q-act {
  font-size: 24rpx;
  color: #3b82f6;
  padding: 8rpx 12rpx;
}
.q-act.del {
  color: #f56c6c;
}
.q-modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  display: flex;
  align-items: flex-end;
}
.q-modal {
  width: 100%;
  max-height: 85vh;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  display: flex;
  flex-direction: column;
}
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 32rpx;
  border-bottom: 1rpx solid #e4e7ed;
}
.modal-title {
  font-size: 30rpx;
  font-weight: 600;
}
.modal-close {
  font-size: 40rpx;
  color: #909399;
  padding: 0 8rpx;
}
.modal-body {
  flex: 1;
  padding: 16rpx 32rpx;
  max-height: 60vh;
}
.form-row {
  margin-bottom: 24rpx;
}
.form-label {
  display: block;
  font-size: 24rpx;
  color: #606266;
  margin-bottom: 8rpx;
}
.form-input {
  width: 100%;
  padding: 16rpx 20rpx;
  font-size: 26rpx;
  border: 1rpx solid #dcdfe6;
  border-radius: 8rpx;
  background: #fff;
  box-sizing: border-box;
}
.form-input.small {
  flex: 1;
}
.form-textarea {
  width: 100%;
  padding: 16rpx 20rpx;
  font-size: 26rpx;
  border: 1rpx solid #dcdfe6;
  border-radius: 8rpx;
  background: #fff;
  height: 120rpx;
  box-sizing: border-box;
}
.form-hint {
  font-size: 20rpx;
  color: #909399;
  display: block;
  margin-top: 8rpx;
}
.opts-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}
.opt-row {
  display: flex;
  align-items: center;
  gap: 8rpx;
}
.opt-del {
  color: #f56c6c;
  font-size: 24rpx;
  padding: 8rpx 12rpx;
}
.opt-add {
  text-align: center;
  padding: 16rpx;
  color: #3b82f6;
  border: 1rpx dashed #3b82f6;
  border-radius: 8rpx;
  font-size: 24rpx;
}
.modal-footer {
  padding: 20rpx 32rpx 40rpx;
  border-top: 1rpx solid #e4e7ed;
}
.modal-btn {
  width: 100%;
  background: #3b82f6;
  color: #fff;
  font-size: 28rpx;
  padding: 20rpx;
  border-radius: 12rpx;
  border: none;
}
</style>
