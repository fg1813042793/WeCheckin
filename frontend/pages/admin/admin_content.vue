<template>
  <view class="container">
    <view class="content-area">
      <textarea
        class="editor"
        v-model="content"
        placeholder="请输入内容..."
        placeholder-class="placeholder"
        :maxlength="10000"
        auto-height
      />
      <text class="word-count">{{ content.length }} / 10000</text>
    </view>

    <view class="actions">
      <button class="btn btn-save" @click="handleSave">保存</button>
      <button class="btn btn-back" @click="handleBack">不保存,返回</button>
    </view>
  </view>
</template>

<script>
export default {
  data() {
    return {
      content: ''
    }
  },

  onLoad(options) {
    if (options.content) {
      this.content = options.content
    }
  },

  methods: {
    handleSave() {
      const pages = getCurrentPages()
      const prevPage = pages[pages.length - 2]
      if (prevPage) {
        prevPage.$vm.content = this.content
      }
      uni.navigateBack()
    },

    handleBack() {
      uni.navigateBack()
    }
  }
}
</script>

<style scoped>
.container {
  min-height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
  padding: 20rpx;
}

.content-area {
  flex: 1;
  background-color: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
  position: relative;
}

.editor {
  width: 100%;
  min-height: 400rpx;
  font-size: 30rpx;
  color: #333;
  line-height: 1.6;
}

.placeholder {
  color: #ccc;
}

.word-count {
  display: block;
  text-align: right;
  font-size: 24rpx;
  color: #999;
  margin-top: 16rpx;
}

.actions {
  display: flex;
}
.actions .btn {
  margin-right: 20rpx;
}
.actions .btn:last-child {
  margin-right: 0;
}

.btn {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 30rpx;
  border-radius: 12rpx;
  text-align: center;
  border: none;
  margin: 0;
}

.btn::after {
  border: none;
}

.btn-save {
  background: linear-gradient(135deg, #c0392b 0%, #e74c3c 100%);
  color: #fff;
}

.btn-back {
  background-color: #fff;
  color: #666;
  border: 2rpx solid #ddd;
}
</style>
