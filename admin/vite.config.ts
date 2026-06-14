import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
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
      '/home': {
        target: 'http://localhost:8083',
        changeOrigin: true
      }
    }
  }
})
