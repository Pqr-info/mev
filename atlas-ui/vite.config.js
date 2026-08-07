import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 9080,
    host: '0.0.0.0',
    proxy: {
      '/sos': 'http://127.0.0.1:4052',
      '/lpv': 'http://127.0.0.1:4050',
      '/api': 'http://127.0.0.1:4050',
      '/antigravity': 'http://127.0.0.1:4050',
      '/v1': 'http://127.0.0.1:8200'
    }
  },
  build: {
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        portal: resolve(__dirname, 'portal.html')
      }
    }
  }
})
