import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    port: 25099,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:25000',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://127.0.0.1:25000',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
