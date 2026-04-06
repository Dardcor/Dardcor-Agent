import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    port: 25000,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:25001',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://127.0.0.1:25001',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
