import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Proxy /api ke backend Go saat dev dan saat `vite preview` (dipakai
// scripts/run.sh). Di production, nginx di deploy/ yang mem-proxy /api.
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:8080',
      '/health': 'http://127.0.0.1:8080',
    },
  },
  preview: {
    port: 4173,
    host: '127.0.0.1',
    proxy: {
      '/api': 'http://127.0.0.1:8080',
      '/health': 'http://127.0.0.1:8080',
    },
  },
})
