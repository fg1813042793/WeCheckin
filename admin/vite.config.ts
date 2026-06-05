import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/admin': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/passport': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/upload': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/user_form_fields': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/home': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
