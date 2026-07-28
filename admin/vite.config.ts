import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return
          if (id.includes('/element-plus/') || id.includes('/@element-plus/')) return 'vendor-element-plus'
          if (id.includes('/vue/') || id.includes('/vue-router/')) return 'vendor-vue'
          if (id.includes('/echarts/') || id.includes('/vue-echarts/')) return 'vendor-echarts'
          if (id.includes('/@vueup/vue-quill/') || id.includes('/quill/')) return 'vendor-editor'
          if (id.includes('/html5-qrcode/') || id.includes('/leaflet/')) return 'vendor-qrcode-map'
          return 'vendor'
        }
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      '/admin': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      '/passport': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      '/upload': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      '/user_form_fields': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      '/survey': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        bypass(req) {
          const path = req.url || ''
          if (/^\/(survey(\/?$|\?)|survey\/(designer|formkit|responses|statistic|preview)(\/|\?|$))/.test(path)) {
            return req.url
          }
        }
      },
      '/exam': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        bypass(req) {
          const path = req.url || ''
          if (/^\/(exam(\/?$|\?)|exam\/(list|designer|responses|statistic|formkit)(\/|\?|$))/.test(path)) {
            return req.url
          }
        }
      },
      '/home': {
        target: 'http://localhost:8083',
        changeOrigin: true
      }
    }
  }
})
