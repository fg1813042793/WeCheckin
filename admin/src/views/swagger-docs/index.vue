<template>
  <section class="swagger-docs">
    <header class="swagger-toolbar">
      <div class="swagger-toolbar__title">
        <el-icon><Document /></el-icon>
        <span>接口文档</span>
      </div>
      <div class="swagger-toolbar__actions">
        <el-button
          circle
          :icon="RefreshRight"
          :loading="loading"
          title="刷新接口文档"
          aria-label="刷新接口文档"
          @click="loadSwagger"
        />
        <el-button
          circle
          :icon="TopRight"
          title="在新窗口打开接口文档"
          aria-label="在新窗口打开接口文档"
          @click="openSwagger"
        />
      </div>
    </header>

    <div v-loading="loading" class="swagger-viewer">
      <el-result
        v-if="errorMessage"
        icon="error"
        title="接口文档加载失败"
        :sub-title="errorMessage"
      >
        <template #extra>
          <el-button type="primary" :icon="RefreshRight" @click="loadSwagger">重试</el-button>
        </template>
      </el-result>
      <iframe
        v-else-if="frameVisible"
        :key="frameKey"
        class="swagger-frame"
        :src="swaggerIndexUrl"
        title="Swagger 接口文档"
        @load="handleFrameLoad"
        @error="handleFrameError"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { Document, RefreshRight, TopRight } from '@element-plus/icons-vue'

const swaggerDocUrl = '/swagger/doc.json'
const swaggerIndexUrl = '/swagger/index.html'
const loading = ref(true)
const errorMessage = ref('')
const frameVisible = ref(false)
const frameKey = ref(0)
let loadController: AbortController | null = null

async function loadSwagger() {
  loadController?.abort()
  const controller = new AbortController()
  loadController = controller
  loading.value = true
  errorMessage.value = ''
  frameVisible.value = false

  try {
    const response = await fetch(swaggerDocUrl, {
      cache: 'no-store',
      headers: { Accept: 'application/json' },
      signal: controller.signal,
    })
    const contentType = response.headers.get('content-type') || ''
    if (!response.ok || !contentType.includes('json')) {
      throw new Error(`Swagger 服务响应异常 (${response.status})`)
    }
    if (controller.signal.aborted) return
    frameKey.value += 1
    frameVisible.value = true
  } catch (error) {
    if (controller.signal.aborted) return
    errorMessage.value = error instanceof Error
      ? error.message
      : '请确认后端 Swagger 服务与代理配置'
    loading.value = false
  }
}

function handleFrameLoad() {
  loading.value = false
}

function handleFrameError() {
  loading.value = false
  frameVisible.value = false
  errorMessage.value = 'Swagger 页面加载失败，请稍后重试'
}

function openSwagger() {
  window.open(swaggerIndexUrl, '_blank', 'noopener,noreferrer')
}

onMounted(loadSwagger)
onBeforeUnmount(() => loadController?.abort())
</script>

<style scoped>
.swagger-docs {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: calc(100vh - var(--admin-header-height) - 80px);
  min-height: 480px;
  background: #fff;
  border: 1px solid var(--admin-border);
  border-radius: 6px;
  overflow: hidden;
  box-sizing: border-box;
}

.swagger-toolbar {
  flex: 0 0 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  border-bottom: 1px solid var(--admin-border);
  box-sizing: border-box;
}

.swagger-toolbar__title,
.swagger-toolbar__actions {
  display: flex;
  align-items: center;
}

.swagger-toolbar__title {
  min-width: 0;
  gap: 8px;
  color: var(--admin-text);
  font-size: 15px;
  font-weight: 600;
}

.swagger-toolbar__actions {
  flex-shrink: 0;
  gap: 8px;
}

.swagger-viewer {
  flex: 1;
  min-width: 0;
  min-height: 0;
  background: #fff;
}

.swagger-frame {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
}

.swagger-viewer :deep(.el-result) {
  height: 100%;
  box-sizing: border-box;
}

@media (max-width: 768px) {
  .swagger-docs {
    height: calc(100vh - var(--admin-header-height) - 64px);
    min-height: 420px;
  }
}
</style>
