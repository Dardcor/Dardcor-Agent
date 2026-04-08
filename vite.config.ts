import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    port: 25000,
    strictPort: true,
    hmr: {
      port: 25000,
      clientPort: 25000,
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
