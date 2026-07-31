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
      '/api': 'http://localhost:4050',
      '/antigravity': 'http://localhost:4050'
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
