<template>
  <view class="container">
    <view v-if="loading" class="loading">
      <text>加载中...</text>
    </view>

    <block v-else>
      <view class="enroll-header">
        <text class="enroll-title">{{ enrollTitle || '打卡' }}</text>
        <text v-if="dailyLimit > 1 && mode === 'join'" class="daily-limit-hint">每日可打卡 {{ dailyLimit }} 次</text>
      </view>

      <view class="form-section" v-if="formFields.length > 0">
        <view class="section-title">{{ mode === 'enroll' ? '报名信息' : '打卡信息' }}</view>
        <view class="form-item" v-for="(field, fi) in formFields" :key="fi">

          <text class="form-label">{{ field.label }}<text v-if="field.required" class="required">*</text></text>

          <input
            v-if="field.type === 'text' || !field.type"
            class="form-input"
            :class="{ 'input-error': field.error }"
            :placeholder="'请输入' + field.label"
            v-model="field.value"
            @input="field.error = false"
          />

          <input
            v-else-if="field.type === 'number'"
            class="form-input"
            :class="{ 'input-error': field.error }"
            type="number"
            :placeholder="'请输入' + field.label"
            v-model="field.value"
            @input="field.error = false"
          />

          <picker
            v-else-if="field.type === 'select'"
            :range="field.options || []"
            @change="(e) => { field.value = (field.options || [])[e.detail.value]; field.error = false }"
          >
            <view class="form-input" :class="{ 'input-error': field.error, 'input-placeholder': !field.value }">
              {{ field.value || '请选择' + field.label }}
            </view>
          </picker>

          <textarea
            v-else-if="field.type === 'textarea'"
            class="form-textarea"
            :class="{ 'input-error': field.error }"
            :placeholder="'请输入' + field.label"
            v-model="field.value"
            @input="field.error = false"
          />

          <view v-else-if="field.type === 'image'" class="img-list">
            <view class="img-item" v-for="(img, ii) in (field.value || [])" :key="ii">
              <image :src="img" mode="aspectFill" class="preview-img"></image>
              <view class="img-del" @click="removeFieldImg(fi, ii)">×</view>
            </view>
            <view class="img-item add-btn" @click="chooseFieldImage(fi)" v-if="(field.value || []).length < 9">
              <text class="add-icon">+</text>
              <text class="add-text">上传</text>
            </view>
          </view>

          <view v-else-if="field.type === 'location'" class="location-wrapper">
            <view class="location-row" @click="toggleFieldLocation(fi)">
              <text class="location-label">
                {{ field.value && field.value.enabled ? (field.value.addr || '获取中...') : '添加位置' }}
              </text>
              <switch :checked="field.value && field.value.enabled" color="#fb454c" style="transform: scale(0.8)" />
            </view>
          </view>

          <text v-if="field.error" class="error-text">
            {{ field.type === 'image' ? '请上传' + field.label : '请填写' + field.label }}
          </text>
        </view>
      </view>

      <view class="form-section" v-if="mode === 'join'">
        <view class="section-title">备注</view>
        <textarea class="form-textarea" placeholder="其他想说的..." v-model="remark" />
      </view>

      <view class="submit-area">
        <button class="submit-btn" :loading="submitting" @click="handleSubmit">{{ mode === 'enroll' ? '提交报名' : '提交打卡' }}</button>
      </view>
    </block>
  </view>
</template>

<script>
import { enrollApi } from '../../api/index'
import CONFIG from '../../config'

export default {
  data() {
    return {
      mode: 'join',
      enrollId: '',
      enrollTitle: '',
      dailyLimit: 1,
      loading: true,
      formFields: [],
      remark: '',
      submitting: false
    }
  },

  onLoad(options) {
    if (options.id) {
      this.enrollId = options.id
      this.mode = options.mode || 'join'
      this.loadEnroll()
    }
  },

  methods: {
    async loadEnroll() {
      this.loading = true
      try {
        const res = await enrollApi.detail({ id: this.enrollId })
        if (res.data) {
          this.enrollTitle = res.data.title || ''
          let src = []
          try {
            if (this.mode === 'enroll') {
              src = JSON.parse(res.data.forms || '[]')
            } else {
              src = JSON.parse(res.data.joinForms || res.data.forms || '[]')
            }
          } catch (e) {}
          this.formFields = src.map(f => {
            let value = ''
            if (f.type === 'image') value = []
            else if (f.type === 'location') value = { enabled: false, addr: '', lat: 0, lng: 0 }
            return { ...f, value, error: false }
          })
          this.dailyLimit = res.data.dailyLimit || 1
        }
      } catch (e) {
        console.error('加载打卡项目失败', e)
      } finally {
        this.loading = false
      }
    },

    chooseFieldImage(fi) {
      const field = this.formFields[fi]
      const current = field.value || []
      uni.chooseImage({
        count: 9 - current.length,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: (res) => {
          uni.showLoading({ title: '上传中...' })
          let uploaded = 0
          const total = res.tempFilePaths.length
          for (const tempFile of res.tempFilePaths) {
            uni.uploadFile({
              url: CONFIG.BASE_URL + '/upload',
              filePath: tempFile,
              name: 'file',
              success: (uploadRes) => {
                if (uploadRes.statusCode !== 200) {
                  const msg = uploadRes.statusCode === 413 ? '上传文件过大' : ('上传失败(状态' + uploadRes.statusCode + ')')
                  uni.showToast({ title: msg, icon: 'none' })
                } else {
                  try {
                    const data = JSON.parse(uploadRes.data)
                    if (data.code === 0 && data.data.url) {
                      field.value.push(data.data.url)
                      field.error = false
                    } else {
                      uni.showToast({ title: data.msg || '上传失败', icon: 'none' })
                    }
                  } catch (e) {
                    uni.showToast({ title: '上传失败，文件可能过大或不支持', icon: 'none' })
                  }
                }
              },
              fail: (err) => {
                uni.showToast({ title: err.errMsg || '网络异常，上传失败', icon: 'none' })
              },
              complete: () => {
                uploaded++
                if (uploaded >= total) {
                  uni.hideLoading()
                }
              }
            })
          }
        }
      })
    },

    removeFieldImg(fi, ii) {
      const field = this.formFields[fi]
      if (field.value) {
        field.value.splice(ii, 1)
      }
    },

    toggleFieldLocation(fi) {
      const field = this.formFields[fi]
      const val = field.value || {}
      val.enabled = !val.enabled
      if (val.enabled) {
        this.getLocation(fi)
      }
      field.value = val
    },

    getLocation(fi) {
      const field = this.formFields[fi]
      uni.getLocation({
        type: 'gcj02',
        isHighAccuracy: true,
        success: (res) => {
          field.value = {
            enabled: true,
            addr: res.latitude.toFixed(6) + ', ' + res.longitude.toFixed(6),
            lat: res.latitude,
            lng: res.longitude
          }
        },
        fail: () => {
          navigator.geolocation.getCurrentPosition(
            (pos) => {
              field.value = {
                enabled: true,
                addr: pos.coords.latitude.toFixed(6) + ', ' + pos.coords.longitude.toFixed(6),
                lat: pos.coords.latitude,
                lng: pos.coords.longitude
              }
            },
            () => {
              uni.showToast({ title: '获取位置失败，请检查定位权限', icon: 'none' })
              field.value = { enabled: false, addr: '', lat: 0, lng: 0 }
            }
          )
        }
      })
    },

    async handleSubmit() {
      let hasError = false
      for (const field of this.formFields) {
        if (!field.required) continue
        if (field.type === 'image' && (!field.value || field.value.length === 0)) {
          field.error = true
          hasError = true
        } else if (field.type === 'location' && (!field.value || !field.value.enabled || !field.value.addr)) {
          field.error = true
          hasError = true
        } else if (!['image', 'location'].includes(field.type) && !field.value) {
          field.error = true
          hasError = true
        }
      }
      if (hasError) {
        uni.showToast({ title: '请填写所有必填项', icon: 'none' })
        return
      }

      const forms = []
      for (const field of this.formFields) {
        if (field.type === 'image' && field.value) {
          for (const url of field.value) {
            forms.push({ label: field.label, value: url, type: 'image' })
          }
        } else if (field.type === 'location' && field.value && field.value.enabled && field.value.addr) {
          forms.push({ label: field.label + '-地址', value: field.value.addr, type: 'location', locField: '地址' })
          forms.push({ label: field.label + '-纬度', value: String(field.value.lat), type: 'location', locField: '纬度' })
          forms.push({ label: field.label + '-经度', value: String(field.value.lng), type: 'location', locField: '经度' })
        } else {
          forms.push({ label: field.label, value: field.value, type: field.type || 'text' })
        }
      }

      if (this.mode === 'join' && this.remark) {
        forms.push({ label: '备注', value: this.remark })
      }

      const today = new Date()
      const dayStr = today.getFullYear() + '-' + String(today.getMonth() + 1).padStart(2, '0') + '-' + String(today.getDate()).padStart(2, '0')

      const userInfo = uni.getStorageSync('userInfo')
      const token = uni.getStorageSync('token')
      const uid = (userInfo && (userInfo.miniOpenID || userInfo.id)) || token || ''

      this.submitting = true
      try {
        if (this.mode === 'enroll') {
          await enrollApi.enrollSubmit({
            enroll_id: this.enrollId,
            user_id: uid,
            forms: JSON.stringify(forms)
          })
          uni.showToast({ title: '报名成功', icon: 'success' })
        } else {
          await enrollApi.join({
            enroll_id: this.enrollId,
            user_id: uid,
            day: dayStr,
            forms: JSON.stringify(forms)
          })
          uni.showToast({ title: '打卡成功', icon: 'success' })
        }
        setTimeout(() => { uni.navigateBack() }, 1500)
      } catch (e) {
        console.error('操作失败', e)
      } finally {
        this.submitting = false
      }
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 120rpx;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60vh;
  color: #999;
  font-size: 28rpx;
}

.enroll-header {
  background: linear-gradient(135deg, #fb454c, #d32f2f);
  margin: 20rpx;
  border-radius: 16rpx;
  padding: 36rpx 30rpx;
}

.enroll-title {
  font-size: 36rpx;
  font-weight: bold;
  color: #fff;
}

.daily-limit-hint {
  display: inline-block;
  font-size: 22rpx;
  color: rgba(255,255,255,0.85);
  background: rgba(255,255,255,0.2);
  padding: 4rpx 16rpx;
  border-radius: 20rpx;
  margin-top: 12rpx;
}

.form-section {
  background-color: #fff;
  margin: 20rpx;
  border-radius: 16rpx;
  padding: 30rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 24rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.form-label {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 10rpx;
}

.required {
  color: #fb454c;
}

.form-input {
  height: 80rpx;
  background-color: #f5f5f5;
  border-radius: 12rpx;
  padding: 0 24rpx;
  font-size: 28rpx;
  color: #333;
}

.form-textarea {
  width: 100%;
  min-height: 120rpx;
  background-color: #f5f5f5;
  border-radius: 12rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  color: #333;
  box-sizing: border-box;
}

.img-list {
  display: flex;
  flex-wrap: wrap;
}

.img-item {
  width: 180rpx;
  height: 180rpx;
  border-radius: 12rpx;
  overflow: hidden;
  position: relative;
  margin-right: 16rpx;
  margin-bottom: 16rpx;
}

.preview-img {
  width: 100%;
  height: 100%;
}

.img-del {
  position: absolute;
  top: 6rpx;
  right: 6rpx;
  width: 36rpx;
  height: 36rpx;
  background-color: rgba(0, 0, 0, 0.5);
  color: #fff;
  border-radius: 50%;
  text-align: center;
  line-height: 36rpx;
  font-size: 28rpx;
}

.add-btn {
  border: 2rpx dashed #ccc;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #fafafa;
}

.add-icon {
  font-size: 48rpx;
  color: #999;
  line-height: 1;
}

.add-text {
  font-size: 22rpx;
  color: #999;
  margin-top: 8rpx;
}

.location-wrapper {
  width: 100%;
}

.location-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 80rpx;
  background-color: #f5f5f5;
  border-radius: 12rpx;
  padding: 0 24rpx;
}

.location-label {
  font-size: 28rpx;
  color: #333;
}

.input-error {
  border: 2rpx solid #fb454c !important;
  background-color: #fff5f5 !important;
}

.error-text {
  font-size: 22rpx;
  color: #fb454c;
  margin-top: 6rpx;
  display: block;
}

.input-placeholder {
  color: #bbb;
}

.submit-area {
  padding: 40rpx 30rpx;
}

.submit-btn {
  background-color: #fb454c;
  color: #fff;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 44rpx;
  font-size: 32rpx;
}
</style>
