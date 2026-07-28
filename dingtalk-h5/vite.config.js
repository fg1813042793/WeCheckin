import { defineConfig, loadEnv } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_DEV_PROXY_TARGET || 'http://localhost:8083'

  return {
    plugins: [uni()],
    server: {
      proxy: {
        '/api/v2': {
          target: proxyTarget,
          changeOrigin: true
        }
      }
    }
  }
})
