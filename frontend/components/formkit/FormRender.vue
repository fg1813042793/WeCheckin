<template>
  <view class="form-render">
    <view v-for="q in questions" :key="q.id" class="fr-item" :class="{ 'fr-hidden': isHidden(q), 'fr-layout': isLayout(q.type) }">
      <!-- 布局题：分割线 -->
      <view v-if="q.type === 'divider'" class="fr-divider">
        <text class="fr-divider-text" v-if="q.options && q.options.text">{{ q.options.text }}</text>
        <view class="fr-divider-line" />
      </view>

      <!-- 布局题：说明文字 -->
      <view v-else-if="q.type === 'description'" class="fr-desc">
        <text class="fr-desc-text">{{ (q.options && q.options.text) || q.description || '' }}</text>
      </view>

      <!-- 输入型 -->
      <block v-else>
        <view class="fr-label">
          <text class="fr-title">{{ q.title }}</text>
          <text v-if="q.required" class="fr-required">*</text>
        </view>
        <view v-if="q.description" class="fr-helper">{{ q.description }}</view>

        <!-- input / text / phone / email / idCard / password -->
        <input
          v-if="['input','text','phone','email','idCard','password'].includes(q.type)"
          :type="inputType(q.type)"
          :value="getVal(q)"
          @input="(e) => setVal(q, e.detail.value)"
          :placeholder="q.placeholder || placeholderFor(q.type)"
          :password="q.type === 'password'"
          class="fr-input"
        />

        <!-- number -->
        <input
          v-else-if="q.type === 'number'"
          type="number"
          :value="getVal(q)"
          @input="(e) => setVal(q, parseFloat(e.detail.value))"
          :placeholder="q.placeholder || '请输入数字'"
          class="fr-input"
        />

        <!-- textarea -->
        <textarea
          v-else-if="q.type === 'textarea'"
          :value="getVal(q)"
          @input="(e) => setVal(q, e.detail.value)"
          :placeholder="q.placeholder || '请输入'"
          class="fr-textarea"
        />

        <!-- select / picker -->
        <picker
          v-else-if="q.type === 'select' || q.type === 'picker'"
          :range="optionLabels(q)"
          @change="(e) => onPickerChange(q, e)"
        >
          <view class="fr-picker">
            <text :class="getVal(q) ? 'fr-picker-val' : 'fr-picker-ph'">
              {{ getVal(q) ? optionLabel(q, getVal(q)) : (q.placeholder || '请选择') }}
            </text>
          </view>
        </picker>

        <!-- radio -->
        <view v-else-if="q.type === 'radio'" class="fr-radio-group">
          <label
            v-for="opt in q.options"
            :key="opt.value"
            class="fr-radio"
            @click="setVal(q, opt.value)"
          >
            <view class="fr-radio-dot" :class="{ active: getVal(q) === opt.value }" />
            <text class="fr-radio-text">{{ opt.label }}</text>
          </label>
        </view>

        <!-- checkbox -->
        <view v-else-if="q.type === 'checkbox'" class="fr-checkbox-group">
          <label
            v-for="opt in q.options"
            :key="opt.value"
            class="fr-checkbox"
            @click="onCheckboxToggle(q, opt.value)"
          >
            <view class="fr-checkbox-box" :class="{ active: isChecked(q, opt.value) }">
              <text v-if="isChecked(q, opt.value)" class="fr-checkbox-tick">✓</text>
            </view>
            <text class="fr-checkbox-text">{{ opt.label }}</text>
          </label>
        </view>

        <!-- switch -->
        <switch
          v-else-if="q.type === 'switch'"
          :checked="getVal(q) === '1' || getVal(q) === true"
          @change="(e) => setVal(q, e.detail.value ? '1' : '0')"
          color="#3b82f6"
        />

        <!-- rating -->
        <view v-else-if="q.type === 'rating'" class="fr-rating">
          <text
            v-for="i in (q.options && q.options.max) || 5"
            :key="i"
            class="fr-star"
            :class="{ active: i <= parseInt(getVal(q) || 0) }"
            @click="setVal(q, i)"
          >★</text>
        </view>

        <!-- date -->
        <picker
          v-else-if="q.type === 'date'"
          mode="date"
          :value="getVal(q)"
          @change="(e) => setVal(q, e.detail.value)"
        >
          <view class="fr-picker">
            <text :class="getVal(q) ? 'fr-picker-val' : 'fr-picker-ph'">
              {{ getVal(q) || (q.placeholder || '请选择日期') }}
            </text>
          </view>
        </picker>

        <!-- time -->
        <picker
          v-else-if="q.type === 'time'"
          mode="time"
          :value="getVal(q)"
          @change="(e) => setVal(q, e.detail.value)"
        >
          <view class="fr-picker">
            <text :class="getVal(q) ? 'fr-picker-val' : 'fr-picker-ph'">
              {{ getVal(q) || (q.placeholder || '请选择时间') }}
            </text>
          </view>
        </picker>

        <!-- file -->
        <view v-else-if="q.type === 'file'" class="fr-file">
          <text class="fr-file-path" v-if="getVal(q)">{{ getVal(q) }}</text>
          <button size="mini" @click="onFilePick(q)">选择文件</button>
        </view>

        <!-- location -->
        <view v-else-if="q.type === 'location'" class="fr-location">
          <text class="fr-location-addr" v-if="getLoc(q)">{{ getLoc(q).address }}</text>
          <button size="mini" @click="onLocationPick(q)">获取位置</button>
        </view>

        <!-- signature -->
        <view v-else-if="q.type === 'signature'" class="fr-signature">
          <text v-if="getVal(q)" class="fr-sig-val">已签名</text>
          <button size="mini" @click="onSignature(q)">{{ getVal(q) ? '重新签名' : '点击签名' }}</button>
        </view>

        <!-- dateRange -->
        <view v-else-if="q.type === 'dateRange'" class="fr-daterange">
          <picker mode="date" :value="getRangeVal(q, 0)" @change="(e) => onRangeChange(q, 0, e.detail.value)">
            <view class="fr-picker half">
              <text :class="getRangeVal(q, 0) ? 'fr-picker-val' : 'fr-picker-ph'">
                {{ getRangeVal(q, 0) || '开始日期' }}
              </text>
            </view>
          </picker>
          <text class="fr-daterange-sep">~</text>
          <picker mode="date" :value="getRangeVal(q, 1)" @change="(e) => onRangeChange(q, 1, e.detail.value)">
            <view class="fr-picker half">
              <text :class="getRangeVal(q, 1) ? 'fr-picker-val' : 'fr-picker-ph'">
                {{ getRangeVal(q, 1) || '结束日期' }}
              </text>
            </view>
          </picker>
        </view>

        <!-- matrixRadio -->
        <view v-else-if="q.type === 'matrixRadio'" class="fr-matrix">
          <view v-for="row in (q.options && q.options.rows) || []" :key="row.value" class="fr-matrix-row">
            <text class="fr-matrix-label">{{ row.label }}</text>
            <view class="fr-matrix-cols">
              <label
                v-for="col in (q.options && q.options.columns) || []"
                :key="col.value"
                class="fr-radio"
                @click="setMatrixVal(q, row.value, col.value)"
              >
                <view class="fr-radio-dot" :class="{ active: getMatrixVal(q, row.value) === col.value }" />
                <text class="fr-radio-text">{{ col.label }}</text>
              </label>
            </view>
          </view>
        </view>

        <!-- autopop -->
        <view v-else-if="q.type === 'autopop'" class="fr-autopop">
          <text class="fr-autopop-val">{{ getVal(q) || '(自动填充)' }}</text>
        </view>

        <!-- 未知/未实现类型 -->
        <view v-else class="fr-unknown">
          <text>暂不支持的题型: {{ q.type }}</text>
        </view>
      </block>
    </view>
  </view>
</template>

<script>
import { isOldSchema, normalizeSchema, initAnswers, getAnswerValue, setAnswerValue } from '../../utils/formkit.js'

export default {
  name: 'FormRender',
  props: {
    schema: { type: [String, Object], default: '' },
    value: { type: [Array, Object], default: null }
  },
  emits: ['update:value', 'change'],
  data() {
    return {
      questions: [],
      answers: null,
      isOld: true
    }
  },
  watch: {
    schema: {
      immediate: true,
      handler(v) {
        this.isOld = isOldSchema(v)
        this.questions = normalizeSchema(v)
        if (this.value !== null && this.value !== undefined) {
          this.answers = Array.isArray(this.value) || typeof this.value === 'object'
            ? JSON.parse(JSON.stringify(this.value))
            : initAnswers(this.questions, this.isOld)
        } else {
          this.answers = initAnswers(this.questions, this.isOld)
        }
      }
    }
  },
  methods: {
    isHidden(q) {
      // 简化：不做 logic 求值，后续 P3+ 集成
      return false
    },
    isLayout(t) {
      return t === 'divider' || t === 'description'
    },
    getVal(q) {
      return getAnswerValue(this.answers, q, this.isOld)
    },
    setVal(q, v) {
      setAnswerValue(this.answers, q, v, this.isOld)
      this.$emit('update:value', this.answers)
      this.$emit('change', { id: q.id, value: v })
    },
    inputType(t) {
      if (t === 'phone' || t === 'idCard' || t === 'password' || t === 'email' || t === 'number') return t
      return 'text'
    },
    placeholderFor(t) {
      const map = { phone: '请输入手机号', email: '请输入邮箱', idCard: '请输入身份证号', password: '请输入密码' }
      return map[t] || '请输入'
    },
    optionLabels(q) {
      return (q.options || []).map((o) => o.label)
    },
    optionLabel(q, value) {
      const opt = (q.options || []).find((o) => o.value === value)
      return opt ? opt.label : value
    },
    onPickerChange(q, e) {
      const idx = parseInt(e.detail.value)
      const opt = (q.options || [])[idx]
      if (opt) this.setVal(q, opt.value)
    },
    isChecked(q, val) {
      const arr = this.getVal(q) || []
      return arr.indexOf(val) >= 0
    },
    onCheckboxToggle(q, val) {
      let arr = [...(this.getVal(q) || [])]
      const idx = arr.indexOf(val)
      if (idx >= 0) arr.splice(idx, 1)
      else arr.push(val)
      this.setVal(q, arr)
    },
    getRangeVal(q, idx) {
      const arr = this.getVal(q)
      if (Array.isArray(arr) && arr.length === 2) return arr[idx] || ''
      return ''
    },
    onRangeChange(q, idx, v) {
      let arr = Array.isArray(this.getVal(q)) && this.getVal(q).length === 2 ? [...this.getVal(q)] : ['', '']
      arr[idx] = v
      this.setVal(q, arr)
    },
    getMatrixVal(q, rowValue) {
      const m = this.getVal(q) || {}
      return m[rowValue] || ''
    },
    setMatrixVal(q, rowValue, colValue) {
      const m = { ...(this.getVal(q) || {}) }
      m[rowValue] = colValue
      this.setVal(q, m)
    },
    getLoc(q) {
      const v = this.getVal(q)
      if (typeof v === 'object' && v && v.address) return v
      return null
    },
    onLocationPick(q) {
      uni.chooseLocation({
        success: (res) => {
          this.setVal(q, { address: res.address, lat: String(res.latitude), lng: String(res.longitude) })
        },
        fail: () => {
          uni.showToast({ title: '获取位置失败', icon: 'none' })
        }
      })
    },
    onFilePick(q) {
      uni.chooseImage({
        count: 1,
        success: (res) => {
          const path = res.tempFilePaths[0]
          uni.uploadFile({
            url: '/upload',
            filePath: path,
            name: 'file',
            success: (up) => {
              try {
                const data = JSON.parse(up.data)
                this.setVal(q, data.data || path)
              } catch (e) {
                this.setVal(q, path)
              }
            }
          })
        }
      })
    },
    onSignature(q) {
      uni.showToast({ title: '请使用专业签名版组件', icon: 'none' })
    }
  }
}
</script>

<style>
.form-render { width: 100%; }
.fr-item { margin-bottom: 24rpx; }
.fr-hidden { display: none; }
.fr-layout { padding: 0; }
.fr-label { display: flex; align-items: center; margin-bottom: 8rpx; }
.fr-title { font-size: 28rpx; color: #333; }
.fr-required { color: #fb454c; margin-left: 6rpx; }
.fr-helper { font-size: 22rpx; color: #999; margin-bottom: 8rpx; }
.fr-input {
  height: 80rpx;
  background: #f7f8fa;
  border-radius: 12rpx;
  padding: 0 20rpx;
  font-size: 28rpx;
  width: 100%;
  box-sizing: border-box;
}
.fr-textarea {
  width: 100%;
  min-height: 160rpx;
  background: #f7f8fa;
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  font-size: 28rpx;
  box-sizing: border-box;
}
.fr-picker {
  height: 80rpx;
  background: #f7f8fa;
  border-radius: 12rpx;
  padding: 0 20rpx;
  display: flex;
  align-items: center;
  font-size: 28rpx;
}
.fr-picker.half { flex: 1; margin: 0 8rpx; }
.fr-picker-val { color: #333; }
.fr-picker-ph { color: #999; }
.fr-radio-group, .fr-checkbox-group { display: flex; flex-direction: column; gap: 16rpx; }
.fr-radio, .fr-checkbox { display: flex; align-items: center; gap: 12rpx; }
.fr-radio-dot {
  width: 32rpx; height: 32rpx; border-radius: 50%; border: 2rpx solid #cbd5e1;
  display: flex; align-items: center; justify-content: center;
}
.fr-radio-dot.active { border-color: #3b82f6; background: #3b82f6; }
.fr-radio-dot.active::after {
  content: ''; width: 12rpx; height: 12rpx; background: #fff; border-radius: 50%;
}
.fr-checkbox-box {
  width: 32rpx; height: 32rpx; border: 2rpx solid #cbd5e1; border-radius: 6rpx;
  display: flex; align-items: center; justify-content: center;
}
.fr-checkbox-box.active { border-color: #3b82f6; background: #3b82f6; }
.fr-checkbox-tick { color: #fff; font-size: 22rpx; }
.fr-radio-text, .fr-checkbox-text { font-size: 28rpx; color: #333; }
.fr-rating { display: flex; gap: 12rpx; }
.fr-star { font-size: 48rpx; color: #cbd5e1; }
.fr-star.active { color: #fbbf24; }
.fr-file, .fr-location, .fr-signature, .fr-autopop { padding: 16rpx 0; }
.fr-file-path, .fr-location-addr, .fr-sig-val, .fr-autopop-val { font-size: 26rpx; color: #333; margin-right: 20rpx; }
.fr-daterange { display: flex; align-items: center; }
.fr-daterange-sep { padding: 0 12rpx; color: #999; }
.fr-matrix-row { margin-bottom: 16rpx; }
.fr-matrix-label { font-size: 26rpx; color: #333; margin-bottom: 8rpx; display: block; }
.fr-matrix-cols { display: flex; gap: 16rpx; flex-wrap: wrap; }
.fr-unknown { padding: 16rpx; background: #fef2f2; border-radius: 8rpx; color: #b91c1c; font-size: 24rpx; }
.fr-divider { display: flex; align-items: center; padding: 16rpx 0; }
.fr-divider-text { font-size: 24rpx; color: #999; margin-right: 16rpx; }
.fr-divider-line { flex: 1; height: 1rpx; background: #e4e7ed; }
.fr-desc { padding: 16rpx 0; }
.fr-desc-text { font-size: 26rpx; color: #666; line-height: 1.6; }
</style>
